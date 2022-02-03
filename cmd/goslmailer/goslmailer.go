package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pja237/goslmailer/connectors/msteams"
	"github.com/pja237/goslmailer/internal/lookup"
	"github.com/pja237/goslmailer/internal/slurmjob"
)

type CmdParams struct {
	Subject string
	Other   []string
}

type Receivers []struct {
	scheme string
	target string
}

// Holder of:
// 1. command line parameters (via log package)
// 2. receivers: from parsed command line (1.) comming from: --mail-user userx,mailto:usery@domain,skype:userid via (*invocationContext) generateReceivers() method
type invocationContext struct {
	CmdParams
	Receivers
}

type connector interface {
	SendMessage(*slurmjob.JobContext, string, *log.Logger) error
}

type connectors map[string]connector

type configContainer struct {
	Logfile          string                       `json:"logfile"`
	DefaultConnector string                       `json:"defaultconnector"`
	Connectors       map[string]map[string]string `json:"connectors"`
}

// Populate the map 'connectors' with connectors specified in config file and their instance from package.
func (c *connectors) populateConnectors(i *invocationContext, conf *configContainer, l *log.Logger) error {
	// Iterate through map of connectors from config file.
	for k, v := range conf.Connectors {
		switch k {
		case "msteams":
			// For each recognized, call the connectorpkg.NewConnector() and...
			con, err := msteams.NewConnector(v)
			if err != nil {
				l.Printf("Problem with %s connector configuration. Ignoring.\n", k)
				break
			}
			l.Printf("%s connector configured.\n", k)
			// ...asign its return object value to the connectors map.
			(*c)[k] = con
		default:
			l.Printf("Unsupported connector found. Ignoring %#v : %#v\n", k, v)
		}
	}
	return nil
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

// populate ic.Receivers (scheme:target) from ic.CmdParams.Other using defCon (defaultconnector) config parameter for undefined schemes
func (m *invocationContext) generateReceivers(defCon string, l *log.Logger) {
	for _, v := range m.CmdParams.Other {
		targets := strings.Split(v, ",")
		for i, t := range targets {
			targetsSplit := strings.Split(t, ":")
			l.Printf("generateReceivers: target %d = %#v\n", i, targetsSplit)
			switch len(targetsSplit) {
			case 1:
				m.Receivers = append(m.Receivers, struct {
					scheme string
					target string
				}{
					// receivers with unspecified connector scheme get global config key "DefaultConnector" set here:
					scheme: defCon,
					// in case an external database has to be used to translate between user_ids for a certain scheme, use the lookup package
					target: lookup.ExtLookupUser(targetsSplit[0], defCon),
				})
			case 2:
				m.Receivers = append(m.Receivers, struct {
					scheme string
					target string
				}{
					scheme: targetsSplit[0],
					// in case an external database has to be used to translate between user_ids for a certain scheme, use the lookup package
					target: lookup.ExtLookupUser(targetsSplit[1], targetsSplit[0]),
				})
			}
		}
	}
}

func (ic *invocationContext) dumpReceivers(l *log.Logger) {
	l.Println("DUMP RECEIVERS:")
	l.Printf("Receivers: %#v\n", ic.Receivers)
	l.Printf("invocationContext: %#v\n", ic)
	l.Println("--------------------------------------------------------------------------------")
}

func (p *CmdParams) getCMDLine() {
	flag.StringVar(&p.Subject, "s", "Default Blank Subject", "e-mail subject")
	flag.Parse()
	p.Other = flag.Args()
}

func (p *CmdParams) dumpCMDLine(l *log.Logger) {
	l.Println("Parsing CMDLine:")
	l.Printf("CMD subject: %#v\n", p.Subject)
	l.Printf("CMD others: %#v\n", p.Other)
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
	ic.CmdParams.getCMDLine()
	ic.CmdParams.dumpCMDLine(log)

	// get SLURM_* environment variables
	job.GetSlurmEnvVars()

	// get job statistics based on the SLURM_JOB_ID from slurmEnv struct
	// only if job is END or FAIL(?)
	job.GetJobStats(job.SlurmEnvironment)

	// generate hints based on SlurmEnv and JobStats (e.g. "too much memory requested" or "walltime << requested queue")
	// only if job is END or fail(?)
	job.GenerateHints()

	// parse CmdParams and generate a list of {scheme, target} receivers (e.g. [ {skype, skypeid}, {msteams, msteamsid}, ...])
	ic.generateReceivers(config.DefaultConnector, log)
	ic.dumpReceivers(log)

	// populate map with configured referenced connectors
	conns.populateConnectors(&ic, &config, log)

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
