package matrix

import (
	"log"
        "bytes"

	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

func NewConnector(conf map[string]string) (*Connector, error) {
	c := Connector{
		username:     conf["username"],
		password:     conf["password"],
		homeserver:   conf["homeserver"],
		roomid:       conf["roomid"],
		template:     conf["template"],
		format:       conf["format"],
	}
	return &c, nil
}

func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {
	var (
		err     error = nil
		buffer  bytes.Buffer
                enduser string = "testuser"
	)

        buffer = bytes.Buffer{}
        err = renderer.RenderTemplate(c.template, c.format, mp.JobContext, enduser, &buffer)
        if err != nil {
                return err
        }

	l.Printf("Logging into", c.homeserver, "as", c.username, "\n")
	client, err := mautrix.NewClient(c.homeserver, "", "")
	if err != nil {
		panic(err)
	}
	_, err = client.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: c.username},
		Password:         c.password,
		StoreCredentials: true,
	})
	if err != nil {
		panic(err)
	}
	l.Printf("Login successful\n")

        _, _ = client.SendText(id.RoomID(c.roomid), buffer.String());

	return err;
}
