/* Main package */
package main

import (
	"flag"
	log "logging"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

var (
	sig         chan os.Signal
	port        int
	verbose     int
	ncpu        int
	httpprofile bool
	cpuprofile  string
)

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
	flagBool(&httpprofile, []string{"httpprofile"}, false, "Run net/http/pprof server")
	flagString(&cpuprofile, []string{"cpuprofile"}, "", "Write cpuprofile info to file")
	sig = make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
}

func main() {
	flag.Parse()
	if httpprofile {
		go func() {
			log.Info("Run profile on localhost:6060")
			log.Err(http.ListenAndServe("localhost:6060", nil).Error())
		}()
	}
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Err(err.Error())
		}
		log.Info("Writing cpuprofile to %v", cpuprofile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	log.SetVerbosity(verbose)
	log.Info("Running gocache on %v cores", ncpu)
	runtime.GOMAXPROCS(ncpu)
	go runServer(port)
	s := <-sig
	log.Info("Got signal: %v", s)
}
