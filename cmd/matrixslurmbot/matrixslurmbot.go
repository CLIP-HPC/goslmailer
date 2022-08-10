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

const app = "matrixslurmbot"

func leaveAndForgetRoom(c *mautrix.Client, rid id.RoomID, l *log.Logger) error {
	l.Printf("Room - leaving: %s\n", rid)
	_, err := c.LeaveRoom(rid)
	if err != nil {
		l.Printf("Error leaving room: %s", err)
		return err
	}
	l.Printf("Room - forgetting: %s\n", rid)
	_, err = c.ForgetRoom(rid)
	if err != nil {
		l.Printf("Error forgetting room: %s", err)
		return err
	}
	return nil
}

func testRoomEmpty(c *mautrix.Client, rid id.RoomID, l *log.Logger) (bool, error) {
	l.Printf("Test room if room empty: %s\n", rid)
	jmResp, err := c.JoinedMembers(rid)
	if err != nil {
		l.Printf("Error looking up joined members: %v", err)
		return false, err
	}
	if len(jmResp.Joined) == 1 {
		//The bot is alone here
		return true, nil
	}
	return false, nil
}

func leaveIfRoomEmpty(c *mautrix.Client, rid id.RoomID, l *log.Logger) error {
	var rooms []id.RoomID
	//START TEST FOR LEAVING EMPTY ROOMS
	if rid != "" {
		// test one specific room, as OnEvent() response
		rooms = append(rooms, rid)
	} else {
		// find ALL rooms bot is joined, on startup
		jrResp, err := c.JoinedRooms()
		if err != nil {
			l.Printf("Error looking up joined rooms: %v", err)
			return err
		}
		rooms = append(rooms, jrResp.JoinedRooms...)
	}

	// traverse target rooms, test if empty, then leave & forget
	for _, r := range rooms {
		alone, err := testRoomEmpty(c, r, l)
		if err != nil {
			return err
		}
		if alone {
			err = leaveAndForgetRoom(c, r, l)
			if err != nil {
				return err
			}
		}
	}
	//END TEST FOR LEAVING EMPTY ROOMS

	return nil
}

func main() {
	var (
		l              *log.Logger
		err            error
		matrixBotUser  string
		matrixBotToken string
		matrixServer   string
	)

	// parse command line params
	cmd, err := cmdline.NewCmdArgs(app)
	if err != nil {
		log.Fatalf("ERROR: parse command line failed with: %q\n", err)
	}

	if *(cmd.Version) {
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

	// let's cleanup leaving/forgetting empty rooms
	l.Printf("Test and leave empty joined rooms...\n")
	err = leaveIfRoomEmpty(client, "", l)
	if err != nil {
		l.Printf("ERROR: leaveIfRoomEmpty: %s\n", err)
	}

	// here comes the work:
	syncer := client.Syncer.(*mautrix.DefaultSyncer)

	// don't pull old events, they've been dealt with
	oldEventIgnorer := mautrix.OldEventIgnorer{
		UserID: id.UserID(matrixBotUser),
	}
	oldEventIgnorer.Register(syncer)

	// do we need to keep any state at all?
	syncer.OnEvent(client.Store.(*mautrix.InMemoryStore).UpdateState)

	/*
		        //START code for responding to user messages
		        //disabled for now since it only works on unencrypted channels

			syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, event *event.Event) {
		            //TODO: implement this for encrypted channels, only works for
		            //unencrypted right now
			    body := event.Content.Raw["body"].(string)
		            if strings.HasPrefix(body, "!bot"){
		                client.SendText(event.RoomID, fmt.Sprintf("Sorry, I'm still a bit dumb. Use this switch in your job submission script and i'll get back to you:\n--mail-user=matrix:%s\n", string(event.RoomID)))
		            }
			})
		        //END code for responding to user messages
	*/

	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, event *event.Event) {
		l.Printf("--------------------------------------------------------------------------------\n")
		// skip non-member events and member events that don't pertain
		// to us

		content := event.Content.AsMember()
		l.Printf("Membership: %#v\n", content.Membership.IsInviteOrJoin())
		l.Printf("Membership: %#v\n", content.Membership)
		l.Printf("Rooms: %#v", client.Store.(*mautrix.InMemoryStore).Rooms)
		l.Printf("StateKey = %s\n", *event.StateKey)
		l.Printf("event.Type.Type = %s\n", event.Type.Type)

		// do we need to test event.Type? we're already triggering on this OnEventType(event.StateMember) ?
		//if *event.StateKey != matrixBotUser || event.Type.Type != "m.room.member" {
		if *event.StateKey != matrixBotUser {
			// someone else joined or left the room
			l.Printf("Event not triggered by us, test and leave empty rooms.\n")
			// if someone else "leave"s, we test and leave if room is left empty, otherwise we ignore this event
			if content.Membership == "leave" {
				l.Printf("Someone left the room, test if we're alone\n")
				err = leaveIfRoomEmpty(client, event.RoomID, l)
				if err != nil {
					l.Printf("ERROR: leaveIfRoomEmpty: %s\n", err)
				}
			} else {
				l.Printf("Someone joined the room, ignore the event.\n")
			}
			return
		}

		// skip non-invite messages
		// test with content.Membership.IsInviteOrJoin() ?
		if content.Membership != "invite" {
			l.Printf("Event not invite.\n")
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
		client.SendText(event.RoomID, fmt.Sprintf("Hello, use this switch in your job submission script and i'll get back to you:\n--mail-user='matrix:%s'\n", string(event.RoomID)))
		l.Printf("--------------------------------------------------------------------------------\n")
	})

	l.Printf("Ready and waiting for events...\n")
	err = client.Sync()
	if err != nil {
		l.Printf("ERROR: Sync(): %s\n", err)
	}
	l.Printf("Done.\n")
	// eowork
}
