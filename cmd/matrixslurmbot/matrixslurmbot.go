package main

import (
        "fmt"
        "log"
        "flag"
        "strings"

        "maunium.net/go/mautrix"
        "maunium.net/go/mautrix/id"
        "maunium.net/go/mautrix/event"
)

func main() {
        var (
            matrixBotUser string
            matrixBotToken string
            matrixServer string
        )

        flag.StringVar(&matrixServer, "s", "matrix.org", "Matrix server URL")
        flag.StringVar(&matrixBotUser, "u", "", "Bot username")
        flag.StringVar(&matrixBotToken, "t", "", "Bot user token")

        flag.Parse()

	if matrixBotUser == "" || matrixBotToken == "" {
		log.Fatalf("ERROR: please provide the matrixBotUser (-u) and matrixBotToken (-t) arguments")
	}

        client, err := mautrix.NewClient(matrixServer, id.UserID(matrixBotUser), matrixBotToken)
        if err != nil {
                log.Fatalf("ERROR: %q\n", err)
        }

        log.Printf("SUCCESS: %#v\n", client)

        syncer := client.Syncer.(*mautrix.DefaultSyncer)
        syncer.OnEvent(client.Store.(*mautrix.InMemoryStore).UpdateState)
        syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, event *event.Event) {
                fmt.Printf("GOT EVENT %#v\n SOURCE: %#v\n", event, source)
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
                fmt.Printf("Joining room %s...\n", event.RoomID)
                _, err := client.JoinRoomByID(event.RoomID)
                if err != nil {
                        panic(err)
                }
                fmt.Printf("Joined room %s\n", event.RoomID)
                //TODO: replacing the room ID's ":" with "@". See related TODO
                //in connectors/matrix/matrix.go
                client.SendText(event.RoomID, fmt.Sprintf("Hello, use this switch in your job submission script and i'll get back to you: --mail-user=matrix:%s\n", strings.Replace(string(event.RoomID),":","@",1)))
        })

        fmt.Println("Looking for rooms to join...")
        err = client.Sync()
        if err != nil {
            panic(err)
        }
}
