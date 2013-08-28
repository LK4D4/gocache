/* Main package */
package main

import (
	"flag"
	log "logging"
)

var port int
var verbose int
var ncpu int

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
}

func main() {
	flag.Parse()
	log.SetVerbosity(verbose)
	log.Info("Running gocache on %v cores", ncpu)
	runtime.GOMAXPROCS(ncpu)
	runServer(port)
}
