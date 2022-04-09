package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pja237/goslmailer/internal/config"
	"github.com/pja237/goslmailer/internal/version"
	tele "gopkg.in/telebot.v3"
)

func main() {

	// todo: separate config, logging etc...
	log := log.New(os.Stderr, "tgslurmbot:", log.Lshortfile|log.Ldate|log.Lmicroseconds)

	cfg := config.NewConfigContainer()
	err := cfg.GetConfig("/etc/slurm/goslmailer.conf")
	if err != nil {
		log.Fatalf("MAIN: getConfig(gobconfig) failed: %s\n", err)
	}

	log.Println("======================= tgslurmbot start =======================================")

	version.DumpVersion(log)

	if _, ok := cfg.Connectors["telegram"]["token"]; !ok {
		log.Fatalf("MAIN: fetching config[connectors][telegram][token] failed: %s\n", err)
	}

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
		str := fmt.Sprintf("Welcome to CLIP SlurmBot,\nplease use this switch in your job submission script in addition to '--mail-type' and i'll get back to you:\n '--mail-user=telegram:%d'", c.Chat().ID)
		return c.Send(str)
	})

	b.Start()

}
