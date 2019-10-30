package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/macroblock/cpbftpchk/xftp"
	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
	"github.com/macroblock/web/cmd/ftpchecksrv/testdata"
)

var (
	log = zlog.Instance("main")

	flagPort   string
	flagTest   bool
	flagSource string

	urlStruct *url.URL
	globList  []xftp.TEntry

	cwd string

	refreshTime     time.Time
	refreshInterval time.Duration = 30 * time.Second
	retryInterval   time.Duration = 5 * time.Second

	mtx sync.Mutex
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
}

func listHandler(w http.ResponseWriter, r *http.Request) {

	mtx.Lock()
	data := globList
	mtx.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	w.Header().Set("content-type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		panic(err)
	}

	return
}

func reloadList(urlStruct *url.URL) error {

	if flagTest {
		mtx.Lock()
		globList = testdata.Data
		mtx.Unlock()
		return nil
	}

	ftp, err := xftp.New2(urlStruct)
	if err != nil {
		return err
	}
	defer ftp.Quit()

	list, err := ftp.List(urlStruct.Path)
	if err != nil {
		return err
	}

	mtx.Lock()
	globList = list
	mtx.Unlock()

	return nil
}

func ftpProcess(urlStruct *url.URL) {
	go func() {
		for {
			fmt.Printf("--- loop ---\n")

			diff := time.Since(refreshTime)
			if diff < refreshInterval {
				time.Sleep(refreshInterval - diff)
				continue
			}

			err := reloadList(urlStruct)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				time.Sleep(retryInterval)
				continue
			}
			// for i, v := range globList {
			// 	fmt.Printf("%3d - %v\n", i, v.Name)
			// }

			refreshTime = time.Now()
		}
	}()
}

func conformFlags() error {
	port, err := strconv.Atoi(flagPort)
	if err != nil || port < 0 {
		return fmt.Errorf("Invalid port number: %q", flagPort)
	}
	flagPort = ":" + strconv.Itoa(port)

	if flagTest {
		return nil
	}

	urlStruct, err = url.Parse(flagSource)
	if err != nil {
		return fmt.Errorf("Parse url error: %v", err)
	}
	if urlStruct.Host == "" {
		return fmt.Errorf("host name is empty")
	}
	return nil
}

func mainFunc() error {
	err := conformFlags()
	if err != nil {
		return err
	}

	// fmt.Printf("url struct:\n%v\n", urlStruct)
	ftpProcess(urlStruct)

	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	mux.HandleFunc("/list", listHandler)

	err = http.ListenAndServe(flagPort, mux)
	return err
}

func main() {
	// setup log
	newLogger := misc.NewSimpleLogger
	if misc.IsTerminal() {
		newLogger = misc.NewAnsiLogger
	}
	log.Add(
		newLogger(loglevel.Warning.OrLower(), ""),
		newLogger(loglevel.Info.Only().Include(loglevel.Notice.Only()), "~x\n"),
	)

	// defer func() {
	// 	if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
	// 		misc.PauseTerminal()
	// 	}
	// }()

	// command line interface
	cmdLine := cli.New("!PROG! is a test server.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information."),
		cli.Flag("-h --help   : help", cmdLine.PrintHelp).Terminator(), // Why does this work ?
		cli.Flag("-p --port   : port to listen.", &flagPort),
		cli.Flag("-t --test   : test mode on.", &flagTest),
		cli.Flag("-s --source : [proto://][username[:password]@]host[/path][:port]", &flagSource),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
