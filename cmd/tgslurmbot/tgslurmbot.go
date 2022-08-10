package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/cmdline"
	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/logger"
	"github.com/CLIP-HPC/goslmailer/internal/version"
	tele "gopkg.in/telebot.v3"
)

const app = "tgslurmbot"

func main() {

	var (
		l   *log.Logger
		err error
	)

	// parse command line params
	cmd, err := cmdline.NewCmdArgs(app)
	if err != nil {
		log.Fatalf("ERROR: parse command line failed with: %q\n", err)
	}

	if *(cmd.Version) == true {
		l = log.New(os.Stderr, app+":", log.Lshortfile|log.Ldate|log.Lmicroseconds)
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
	l, err = logger.SetupLogger(cfg.Logfile, "tgslurmbot")
	if err != nil {
		log.Fatalf("setuplogger(%s) failed with: %q\n", cfg.Logfile, err)
	}

	l.Println("======================= tgslurmbot start =======================================")

	version.DumpVersion(l)

	if _, ok := cfg.Connectors["telegram"]["token"]; !ok {
		l.Fatalf("MAIN: fetching config[connectors][telegram][token] failed: %s\n", err)
	}

	l.Printf("Starting: %q\n", cfg.Connectors["telegram"]["name"])

	pref := tele.Settings{
		Token:  cfg.Connectors["telegram"]["token"],
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnText, func(c tele.Context) error {
		return c.Send("Sorry, i'm not programmed to reply, to get the slurm command line switch to receive messages type: /start")
	})

	b.Handle("/start", func(c tele.Context) error {
		// todo: logging of msg exchanges?
		str := fmt.Sprintf("Welcome to %s,\nplease use this switch in your job submission script in addition to '--mail-type' and i'll get back to you:\n '--mail-user=telegram:%d'", cfg.Connectors["telegram"]["name"], c.Chat().ID)
		return c.Send(str)
	})

	b.Start()

}
