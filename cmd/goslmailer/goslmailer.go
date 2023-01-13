package main

import (
	"log"
	"os"

	_ "github.com/CLIP-HPC/goslmailer/connectors/discord"
	_ "github.com/CLIP-HPC/goslmailer/connectors/mailto"
	_ "github.com/CLIP-HPC/goslmailer/connectors/matrix"
	_ "github.com/CLIP-HPC/goslmailer/connectors/msteams"
	_ "github.com/CLIP-HPC/goslmailer/connectors/telegram"
	_ "github.com/CLIP-HPC/goslmailer/connectors/slack"
	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/logger"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/slurmjob"
	"github.com/CLIP-HPC/goslmailer/internal/version"
)

const goslmailer_config_file = "/etc/slurm/goslmailer.conf"

func main() {

	var (
		ic  invocationContext
		job slurmjob.JobContext
	)

	// get ENV var GOSLMAILERCONF if it's set, if not, use default /etc...
	cf, pres := os.LookupEnv("GOSLMAILER_CONF")
	if !pres || cf == "" {
		cf = goslmailer_config_file
	}

	// read config file
	cfg := config.NewConfigContainer()
	err := cfg.GetConfig(cf)
	if err != nil {
		log.Fatalf("ERROR: getConfig() failed: %s\n", err)
	}

	// setup logger
	l, err := logger.SetupLogger(cfg.Logfile, "goslmailer")
	if err != nil {
		l.Fatalf("setuplogger(%s) failed with: %q\n", cfg.Logfile, err)
	}

	l.Println("======================== START OF RUN ==========================================")

	version.DumpVersion(l)

	// cfg.DumpConfig(l)

	// get '-s "subject" userid' command line parameters with which we're called
	ic.getCMDLine()
	ic.dumpCMDLine(l)

	// parse CmdParams and generate a list of {scheme, target} receivers (e.g. [ {skype, skypeid}, {msteams, msteamsid}, ...])
	ic.generateReceivers(cfg.DefaultConnector, l)
	ic.dumpReceivers(l)

	// get SLURM_* environment variables
	job.GetSlurmEnvVars()

	// get job statistics based on the SLURM_JOB_ID from slurmEnv struct
	// only if job is END or FAIL(?)
	err = job.GetJobStats(ic.CmdParams.Subject, cfg.Binpaths, l)
	if err != nil {
		l.Fatalf("Unable to retrieve job stats. Error: %v", err)
	}

	// generate hints based on SlurmEnv and JobStats (e.g. "too much memory requested" or "walltime << requested queue")
	// only if job is END or fail(?)
	job.GenerateHints(cfg.QosMap)

	// populate map with configured referenced connectors
	connectors.ConMap.PopulateConnectors(cfg, l)

	// Iterate over 'Receivers' map and for each call the connector.SendMessage() (if the receiver scheme is configured in conf file AND has an object in connectors map)
	if ic.Receivers == nil {
		l.Fatalln("No receivers defined. Aborting!")
	}
	// here we loop through requested receivers and invoke SendMessage()
	for _, v := range ic.Receivers {
		mp, err := message.NewMsgPack(v.scheme, v.target, &job)
		if err != nil {
			l.Printf("ERROR in message.NewMsgPack(%s): %q\n", v.scheme, err)
		}
		con, ok := connectors.ConMap[v.scheme]
		if !ok {
			l.Printf("%s connector is not initialized for target %s. Ignoring.\n", v.scheme, v.target)
		} else {
			// useSpool == true when called from here, for connectors that use this capability
			err := con.SendMessage(mp, true, l)
			if err != nil {
				l.Printf("ERROR in %s.SendMessage(): %q\n", v.scheme, err)
			}
		}
	}

	l.Println("========================== END OF RUN ==========================================")
}
