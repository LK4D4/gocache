/* Main package */
package main

import (
	"flag"
	log "logging"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

var port int
var verbose int
var ncpu int
var profile bool

func flagBool(f *bool, aliases []string, value bool, usage string) {
	for _, alias := range aliases {
		flag.BoolVar(f, alias, value, usage)
	}
}

func flagInt(f *int, aliases []string, value int, usage string) {
	for _, alias := range aliases {
		flag.IntVar(f, alias, value, usage)
	}
}

func flagString(f *string, aliases []string, value string, usage string) {
	for _, alias := range aliases {
		flag.StringVar(f, alias, value, usage)
	}
}

func init() {
	flagInt(&port, []string{"port", "p"}, 6090, "Port for incomming connections")
	flagInt(&verbose, []string{"verbose", "v"}, 4, "Logging verbosity")
	flagInt(&ncpu, []string{"ncpu", "n"}, 1, "Number of max used cores")
	flagBool(&profile, []string{"profile"}, false, "Run net/http/pprof server")
}

func main() {
	flag.Parse()
	log.SetVerbosity(verbose)
	log.Info("Running gocache on %v cores", ncpu)
	if profile {
		go func() {
			log.Info("Run profile on localhost:6060")
			log.Err(http.ListenAndServe("localhost:6060", nil).Error())
		}()
	}
	runtime.GOMAXPROCS(ncpu)
	runServer(port)
}
