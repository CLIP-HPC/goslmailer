package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pja237/goslmailer/internal/config"
	"github.com/pja237/goslmailer/internal/connectors"
	"github.com/pja237/goslmailer/internal/message"
	"github.com/pja237/goslmailer/internal/slurmjob"
)

func main() {

	var (
		ic      invocationContext
		job     slurmjob.JobContext
		conns   = make(connectors.Connectors)
		logFile io.Writer
	)

	// read configuration
	// how to handle hardcoding config file?
	cfg := config.NewConfigContainer()
	err := cfg.GetConfig("/etc/slurm/goslmailer.conf")
	if err != nil {
		fmt.Printf("getConfig failed: %s", err)
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
	log := log.New(logFile, "goslmailer:", log.Lshortfile|log.Ldate|log.Lmicroseconds)

	log.Println("======================== START OF RUN ==========================================")
	cfg.DumpConfig(log)

	// get '-s "subject" userid' command line parameters with which we're called
	ic.getCMDLine()
	ic.dumpCMDLine(log)

	// parse CmdParams and generate a list of {scheme, target} receivers (e.g. [ {skype, skypeid}, {msteams, msteamsid}, ...])
	ic.generateReceivers(cfg.DefaultConnector, log)
	ic.dumpReceivers(log)

	// get SLURM_* environment variables
	job.GetSlurmEnvVars()

	// get job statistics based on the SLURM_JOB_ID from slurmEnv struct
	// only if job is END or FAIL(?)
	job.GetJobStats(log)

	// generate hints based on SlurmEnv and JobStats (e.g. "too much memory requested" or "walltime << requested queue")
	// only if job is END or fail(?)
	job.GenerateHints(cfg.QosMap)

	// populate map with configured referenced connectors
	conns.PopulateConnectors(cfg, log)

	// Iterate over 'Receivers' map and for each call the connector.SendMessage() (if the receiver scheme is configured in conf file AND has an object in connectors map)
	if ic.Receivers == nil {
		log.Fatalln("No receivers defined. Aborting!")
	}
	// here we loop through requested receivers and invoke SendMessage()
	for _, v := range ic.Receivers {
		mp, err := message.NewMsgPack(v.scheme, v.target, &job)
		if err != nil {
			log.Printf("ERROR in message.NewMsgPack(%s): %q\n", v.scheme, err)
		}
		con, ok := conns[v.scheme]
		if !ok {
			log.Printf("%s connector is not initialized for target %s. Ignoring.\n", v.scheme, v.target)
		} else {
			// useSpool == true when called from here, for connectors that use this capability
			err := con.SendMessage(mp, true, log)
			if err != nil {
				log.Printf("ERROR in %s.SendMessage(): %q\n", v.scheme, err)
			}
		}
	}

	log.Println("========================== END OF RUN ==========================================")
}
