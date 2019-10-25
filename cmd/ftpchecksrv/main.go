package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

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

	cwd string
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	if flagTest {
		jsonData, err := json.Marshal(testdata.Data)
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
	return
}

func checkURL(s string) (*url.URL, error) {
	u, err := url.Parse(flagSource)
	if err != nil {
		return nil, fmt.Errorf("Parse url error: %v", err)
	}

	// if u.Path != "" {
	// 	u.Path = "/" + u.Path
	// }

	// if u.Scheme == "" {
	// 	switch u.Port() {
	// 	case "21":
	// 		u.Scheme = "ftp"
	// 	case "22":
	// 		u.Scheme = "sftp"
	// 	}
	// }
	// if u.Port() == "" {
	// 	switch u.Scheme {
	// 	case "":
	// 		u.Scheme = "ftp"
	// 		u.Host = u.Hostname() + ":21"
	// 	case "ftp":
	// 		u.Host = u.Hostname() + ":21"
	// 	case "sftp":
	// 		u.Host = u.Hostname() + ":22"
	// 	}
	// }

	if u.Host == "" {
		return nil, fmt.Errorf("host name is empty")
	}

	// if cs.Password == "" && cs.Username != "" {
	// 	fmt.Println("Enter Password:")
	// 	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	// fmt.Println("\nPassword typed: " + string(bytePassword))
	// 	cs.Password = string(bytePassword)
	// }
	return u, nil
}

func mainFunc() error {
	u, err := checkURL(flagSource)
	_ = u

	port, err := strconv.Atoi(flagPort)
	if err != nil || port < 0 {
		return fmt.Errorf("Invalid port number: %q", flagPort)
	}
	flagPort = ":" + strconv.Itoa(port)

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
	cmdLine := cli.New("!PROG! test server.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information."),
		cli.Flag("-h --help   : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-p --port   : port to listen.", &flagPort),
		cli.Flag("-t --test   : test mode on.", &flagTest),
		cli.Flag("-s --sourse : [proto://][username[:password]@]host[/path][:port]", &flagSource),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
