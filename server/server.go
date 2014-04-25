package main

import (
	"flag"
	"github.com/liyinhgqw/oracle"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt)
		go cactchKill(interrupt)
	}

	log.Println("Timestamp Oracle Started")
	orc := oracle.NewOracle()
	orc.Recover()
	orc.WaitForClientConnections()
}

func cactchKill(interrupt chan os.Signal) {
	<-interrupt
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	log.Fatalln("Caught Signal")
}
