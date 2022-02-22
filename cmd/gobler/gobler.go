package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/pja237/goslmailer/internal/config"
	"github.com/pja237/goslmailer/internal/connectors"
	"github.com/pja237/goslmailer/internal/message"
)

var lock sync.Mutex

type MsgList []message.MessagePack

func main() {

	var (
		conns   = make(connectors.Connectors)
		logFile io.Writer
		wg      sync.WaitGroup
	)

	// read gobler configuration
	cfg := config.NewConfigContainer()
	err := cfg.GetConfig("/etc/slurm/gobler.conf")
	if err != nil {
		fmt.Printf("MAIN: getConfig(gobconfig) failed: %s", err)
		os.Exit(1)
	}

	// setup logger
	if cfg.Logfile != "" {
		logFile, err = os.OpenFile(cfg.Logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("ERROR: can not open configured log file. Exiting.")
			os.Exit(1)
		}
	} else {
		logFile = os.Stderr
	}
	log := log.New(logFile, "gobler:", log.Lshortfile|log.Ldate|log.Lmicroseconds)

	log.Println("======================= Gobler start ===========================================")
	cfg.DumpConfig(log)

	// populate map with configured referenced connectors
	err = conns.PopulateConnectors(cfg, log)
	if err != nil {
		log.Printf("MAIN: PopulateConnectors() failed with: %s\n", err)
	}

	// iterate and spin up monitor,picker and sender routines
	for con := range cfg.Connectors {
		spd, ok := cfg.Connectors[con]["spoolDir"]
		if ok {
			log.Printf("MAIN: spoolDir exists: %s - %s\n", cfg.Connectors[con]["spoolDir"], spd)

			cm, err := NewConMon(con, cfg.Connectors[con])
			if err != nil {
				log.Printf("MAIN: NewConMon(%s) failed with: %s\n", con, err)
			}
			// func (cm *conMon) SpinUp(conns connectors.Connectors, wg sync.WaitGroup, log *log.Logger) error {
			err = cm.SpinUp(conns, &wg, log)
			if err != nil {
				log.Printf("MAIN: SpinUp(%s) failed with: %s\n", con, err)
			}
		} else {
			log.Printf("MAIN: connector %s doesn't have spoolDir defined\n", con)
		}
	}

	log.Printf("MAIN: Waiting for routines to finish...\n")
	wg.Wait()
	log.Printf("MAIN: All routines finished, exiting main\n")

	log.Println("======================= Gobler end =============================================")
}
