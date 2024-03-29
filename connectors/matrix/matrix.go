package matrix

import (
	"bytes"
	"log"

	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

func init() {
	connectors.Register(connectorName, connMatrix)
}

func (c *Connector) ConfigConnector(conf map[string]string) error {

	c.username = conf["username"]
	c.token = conf["token"]
	c.homeserver = conf["homeserver"]
	c.template = conf["template"]

	// here we need some test if the connectors "minimal" configuration is satisfied, e.g. must have url at minimum
	//
	// if ok, return nil error
	return nil
}

func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {
	var (
		err    error = nil
		buffer bytes.Buffer
		roomid string = mp.TargetUser
	)

	buffer = bytes.Buffer{}
	err = renderer.RenderTemplate(c.template, "text", mp.JobContext, roomid, &buffer)
	if err != nil {
		return err
	}

	l.Println("Logging into", c.homeserver, "as", c.username)
	client, err := mautrix.NewClient(c.homeserver, id.UserID(c.username), c.token)
	if err != nil {
		return err
	}

	content := format.RenderMarkdown(buffer.String(), true, true)
	content.MsgType = event.MsgNotice
	_, err = client.SendMessageEvent(id.RoomID(roomid), event.EventMessage, content)

	return err
}
