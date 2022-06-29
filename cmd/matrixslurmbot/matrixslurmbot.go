package main

import (
	"fmt"
	"log"
	"os"

	"github.com/CLIP-HPC/goslmailer/internal/cmdline"
	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/logger"
	"github.com/CLIP-HPC/goslmailer/internal/version"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func main() {
	var (
		l              *log.Logger
		err            error
		matrixBotUser  string
		matrixBotToken string
		matrixServer   string
	)

	// parse command line params
	cmd, err := cmdline.NewCmdArgs("matrixslurmbot")
	if err != nil {
		log.Fatalf("ERROR: parse command line failed with: %q\n", err)
	}

	if *(cmd.Version) {
		l = log.New(os.Stderr, "matrixslurmbot:", log.Lshortfile|log.Ldate|log.Lmicroseconds)
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
	l, err = logger.SetupLogger(cfg.Logfile, "matrixslurmbot")
	if err != nil {
		log.Fatalf("setuplogger(%s) failed with: %q\n", cfg.Logfile, err)
	}

	l.Println("======================= matrixslurmbot start ===================================")

	switch {
	case cfg.Connectors["matrix"]["username"] == "":
		l.Fatalf("ERROR: please provide the matrixBotUser in config file")
	case cfg.Connectors["matrix"]["token"] == "":
		l.Fatalf("ERROR: please provide the matrixBotToken in config file")
	case cfg.Connectors["matrix"]["homeserver"] == "":
		l.Fatalf("ERROR: please provide the matrixHomeServer in config file")
	}

	matrixBotUser = cfg.Connectors["matrix"]["username"]
	matrixBotToken = cfg.Connectors["matrix"]["token"]
	matrixServer = cfg.Connectors["matrix"]["homeserver"]

	client, err := mautrix.NewClient(matrixServer, id.UserID(matrixBotUser), matrixBotToken)
	if err != nil {
		l.Fatalf("ERROR NewClient(): %q\n", err)
	}
	//l.Printf("SUCCESS: %#v\n", client)

	// here comes the work:
	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEvent(client.Store.(*mautrix.InMemoryStore).UpdateState)
	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, event *event.Event) {
		l.Printf("--------------------------------------------------------------------------------\n")
		l.Printf("GOT EVENT %#v\n SOURCE: %#v\n", event, source.String())
		// skip non-member events and member events that don't pertain
		// to us
		if *event.StateKey != matrixBotUser || event.Type.Type != "m.room.member" {
			return
		}
		content := event.Content.AsMember()
		// skip non-invite messages
		if content.Membership != "invite" {
			return
		}
		l.Printf("Joining room %s\n", event.RoomID)
		_, err := client.JoinRoomByID(event.RoomID)
		if err != nil {
			l.Printf("ERROR: joinroom(): %s\n", err)
		}
		l.Printf("Joined room %s\n", event.RoomID)
		client.SendText(event.RoomID, fmt.Sprintf("Hello, use this switch in your job submission script and i'll get back to you: --mail-user=matrix:%s\n", string(event.RoomID)))
		l.Printf("--------------------------------------------------------------------------------\n")
	})

	l.Println("Looking for rooms to join")
	err = client.Sync()
	if err != nil {
		l.Printf("ERROR: Sync(): %s\n", err)
	}
	// eowork
}
