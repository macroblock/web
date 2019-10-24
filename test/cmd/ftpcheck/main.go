package main

import (
	"net/http"
	"os"

	"github.com/macroblock/imed/pkg/cli"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log = zlog.Instance("main")

	flagPort string
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
}

func lsHandler(w http.ResponseWriter, r *http.Request) {
}

func mainFunc() error {
	// port, err := strconv.Atoi(flagPort)
	// if err != nil || port < 0 {
	// 	return fmt.Errorf("Wrong port number: %q", flagPort)
	// }
	// flagPort = ":" + strconv.Itoa(port)

	mux := http.NewServeMux()

	mux.HandleFunc("/", mainHandler)
	mux.HandleFunc("/ls", lsHandler)

	err := http.ListenAndServe(flagPort, mux)
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

	defer func() {
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			misc.PauseTerminal()
		}
	}()

	// command line interface
	cmdLine := cli.New("!PROG! test server.", mainFunc)
	cmdLine.Elements(
		cli.Usage("!PROG! {flags|<...>}"),
		// cli.Hint("Use '!PROG! help <flag>' for more information."),
		cli.Flag("-h --help   : help", cmdLine.PrintHelp).Terminator(), // Why is this works ?
		cli.Flag("-p --port   : port to listen.", &flagPort),
		cli.OnError("Run '!PROG! -h' for usage.\n"),
	)

	err := cmdLine.Parse(os.Args)

	log.Error(err)
	log.Info(cmdLine.GetHint())
}
