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

	//START TEST FOR LEAVING EMPTY ROOMS
	jrResp, err := client.JoinedRooms()
	if err != nil {
		log.Printf("Error looking up joined rooms: %v", err)
		return
	}
	for _, roomID := range jrResp.JoinedRooms {
		l.Println("Found bot joined to room: ", roomID)
		jmResp, err := client.JoinedMembers(roomID)
		if err != nil {
			log.Printf("Error looking up joined members: %v", err)
			return
		}
		if len(jmResp.Joined) == 1 { //The bot is the only one there, leave and forget
			l.Printf("Room: %s empty, leaving.\n", roomID)
			_, err := client.LeaveRoom(roomID)
			if err != nil {
				log.Printf("Error leaving room: %v", err)
				return
			}
			l.Printf("Room: %s forgetting.\n", roomID)
			_, err2 := client.ForgetRoom(roomID)
			if err2 != nil {
				log.Printf("Error forgetting room: %v", err2)
				return
			}
		}
	}
	//END TEST FOR LEAVING EMPTY ROOMS

	// here comes the work:
	syncer := client.Syncer.(*mautrix.DefaultSyncer)

	// don't pull old events, they've been dealt with
	oldEventIgnorer := mautrix.OldEventIgnorer{
		UserID: id.UserID(matrixBotUser),
	}
	oldEventIgnorer.Register(syncer)

	// do we need to keep any state at all?
	syncer.OnEvent(client.Store.(*mautrix.InMemoryStore).UpdateState)
	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, event *event.Event) {
		l.Printf("--------------------------------------------------------------------------------\n")
		// skip non-member events and member events that don't pertain
		// to us

		l.Printf("Rooms: %#v", client.Store.(*mautrix.InMemoryStore).Rooms)
		//l.Printf("StateKey = %s\n", *event.StateKey)
		//l.Printf("event.Type.Type = %s\n", event.Type.Type)
		if *event.StateKey != matrixBotUser || event.Type.Type != "m.room.member" {
			l.Printf("Event not intended for us or not m.room.member\n")
			return
		}
		content := event.Content.AsMember()
		// skip non-invite messages
		if content.Membership != "invite" {
			l.Printf("Event not invite\n")
			return
		}
		l.Printf("GOT EVENT %#v\n SOURCE: %#v\n", event, source.String())
		l.Printf("Joining room %s\n", event.RoomID)
		_, err := client.JoinRoomByID(event.RoomID)
		if err != nil {
			l.Printf("ERROR: joinroom(): %s\n", err)
		}
		l.Printf("Joined room %s\n", event.RoomID)

		// send slurm line upon joining
		client.SendText(event.RoomID, fmt.Sprintf("Hello, use this switch in your job submission script and i'll get back to you:\n--mail-user=matrix:%s\n", string(event.RoomID)))
		l.Printf("--------------------------------------------------------------------------------\n")
	})

	err = client.Sync()
	if err != nil {
		l.Printf("ERROR: Sync(): %s\n", err)
	}
	// eowork
}
