package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pja237/goslmailer/internal/slurmjob"
)

type configContainer struct {
	Logfile          string                       `json:"logfile"`
	DefaultConnector string                       `json:"defaultconnector"`
	Connectors       map[string]map[string]string `json:"connectors"`
}

// Read & unmarshall configuration from 'name' file into configContainer structure
func (cc *configContainer) getConfig(name string) error {
	f, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	err = json.Unmarshal(f, cc)
	if err != nil {
		return err
	}
	return nil
}

func (cc *configContainer) dumpConfig(l *log.Logger) {
	l.Println("DUMP CONFIG:")
	l.Printf("CONFIGURATION: %#v\n", cc)
	l.Printf("CONFIGURATION logfile: %s\n", cc.Logfile)
	l.Printf("CONFIGURATION msteams.name: %s\n", cc.Connectors["msteams"]["name"])
	l.Println("--------------------------------------------------------------------------------")
}

func main() {

	var (
		ic      invocationContext
		config  configContainer
		job     slurmjob.JobContext
		conns   connectors = make(connectors)
		logFile io.Writer
	)

	// read configuration
	// how to handle hardcoding config file?
	err := config.getConfig("/etc/slurm/goslmailer.conf")
	if err != nil {
		fmt.Printf("getConfig failed: %s", err)
		os.Exit(1)
	}

	// setup logger
	if config.Logfile != "" {
		logFile, err = os.OpenFile(config.Logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("ERROR: can not open configured log file. Exiting.")
			os.Exit(1)
		}
	} else {
		logFile = os.Stderr
	}
	log := log.New(logFile, "goslmailer:", log.Lshortfile|log.Ldate|log.Lmicroseconds)

	log.Println("======================== START OF RUN ==========================================")
	config.dumpConfig(log)

	// get '-s "subject" userid' command line parameters with which we're called
	ic.getCMDLine()
	ic.dumpCMDLine(log)

	// parse CmdParams and generate a list of {scheme, target} receivers (e.g. [ {skype, skypeid}, {msteams, msteamsid}, ...])
	ic.generateReceivers(config.DefaultConnector, log)
	ic.dumpReceivers(log)

	// get SLURM_* environment variables
	job.GetSlurmEnvVars()

	// get job statistics based on the SLURM_JOB_ID from slurmEnv struct
	// only if job is END or FAIL(?)
	job.GetJobStats()

	// generate hints based on SlurmEnv and JobStats (e.g. "too much memory requested" or "walltime << requested queue")
	// only if job is END or fail(?)
	job.GenerateHints()

	// populate map with configured referenced connectors
	conns.populateConnectors(&config, log)

	// Iterate over 'Receivers' map and for each call the connector.SendMessage() (if the receiver scheme is configured in conf file AND has an object in connectors map)
	if ic.Receivers == nil {
		log.Fatalln("No receivers defined. Aborting!")
	}
	// here we loop through requested receivers and invoke SendMessage()
	for _, v := range ic.Receivers {
		con, ok := conns[v.scheme]
		if !ok {
			log.Printf("%s connector is not initialized for target %s. Ignoring.\n", v.scheme, v.target)
		} else {
			err := con.SendMessage(&job, v.target, log)
			if err != nil {
				log.Printf("ERROR in %s.SendMessage(): %s\n", v.scheme, err)
			}
		}
	}

	log.Println("========================== END OF RUN ==========================================")
}
