package matrix

import (
	"log"
        "bytes"
        "strings"

	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/event"
)

func NewConnector(conf map[string]string) (*Connector, error) {
	c := Connector{
		username:     conf["username"],
		token:     conf["token"],
		homeserver:   conf["homeserver"],
		template:     conf["template"],
	}
	return &c, nil
}

func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {
	var (
		err     error = nil
		buffer  bytes.Buffer
                roomid string = mp.TargetUser
	)

        //TODO: someone doesn't seem to like a ":" in the value, so use "@"
        //instead and replace it here. See if there's a better way.
        roomid = strings.Replace(roomid, "@",":",-1)

        buffer = bytes.Buffer{}
        err = renderer.RenderTemplate(c.template, "text", mp.JobContext, roomid, &buffer)
        if err != nil {
                return err
        }

	l.Printf("Logging into", c.homeserver, "as", c.username, "\n")
        client, err := mautrix.NewClient(c.homeserver, id.UserID(c.username), c.token)
	if err != nil {
                return err
	}

        content := format.RenderMarkdown(buffer.String(), true, true)
        content.MsgType = event.MsgNotice
        _, err = client.SendMessageEvent(id.RoomID(roomid), event.EventMessage, content)

	return err;
}
