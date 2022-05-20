package main

import (
	"log"
	"os"
	"sync"

	"github.com/CLIP-HPC/goslmailer/internal/cmdline"
	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/logger"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/version"
)

var lock sync.Mutex

type MsgList []message.MessagePack

func main() {

	var (
		conns = make(connectors.Connectors)
		wg    sync.WaitGroup
	)

	// parse command line params
	cmd, err := cmdline.NewCmdArgs("gobler")
	if err != nil {
		log.Fatalf("ERROR: parse command line failed with: %q\n", err)
	}

	if *(cmd.Version) == true {
		l := log.New(os.Stderr, "gobler:", log.Lshortfile|log.Ldate|log.Lmicroseconds)
		version.DumpVersion(l)
		os.Exit(0)
	}

	// read config file
	cfg := config.NewConfigContainer()
	err = cfg.GetConfig(*(cmd.CfgFile))
	if err != nil {
		log.Fatalf("ERROR: getConfig() failed: %s\n", err)
	}

	// setup logger
	l, err := logger.SetupLogger(cfg.Paths["logfile"], "gobler")
	if err != nil {
		log.Fatalf("setuplogger(%s) failed with: %q\n", cfg.Paths["logfile"], err)
	}

	l.Println("======================= Gobler start ===========================================")

	version.DumpVersion(l)

	cfg.DumpConfig(l)

	// populate map with configured referenced connectors
	err = conns.PopulateConnectors(cfg, l)
	if err != nil {
		l.Printf("MAIN: PopulateConnectors() failed with: %s\n", err)
	}

	// iterate and spin up monitor,picker and sender routines
	for con := range cfg.Connectors {
		spd, ok := cfg.Connectors[con]["spoolDir"]
		if ok {
			l.Printf("MAIN: %s spoolDir exists: %s - %s\n", con, cfg.Connectors[con]["spoolDir"], spd)

			cm, err := NewConMon(con, cfg.Connectors[con], l)
			if err != nil {
				l.Printf("MAIN: NewConMon(%s) failed with: %s\n", con, err)
				l.Printf("MAIN: skipping %s...\n", con)
				continue
			}
			// func (cm *conMon) SpinUp(conns connectors.Connectors, wg sync.WaitGroup, l *log.Logger) error {
			err = cm.SpinUp(conns, &wg, l)
			if err != nil {
				l.Printf("MAIN: SpinUp(%s) failed with: %s\n", con, err)
			}
		} else {
			l.Printf("MAIN: connector %s doesn't have spoolDir defined\n", con)
		}
	}

	l.Printf("MAIN: Waiting for routines to finish...\n")
	wg.Wait()
	l.Printf("MAIN: All routines finished, exiting main\n")

	l.Println("======================= Gobler end =============================================")
}
