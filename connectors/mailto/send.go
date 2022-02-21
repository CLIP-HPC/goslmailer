package mailto

import (
	"bytes"
	"log"
	"os/exec"
	"text/template"

	"github.com/pja237/goslmailer/internal/message"
)

func NewConnector(conf map[string]string) (*Connector, error) {
	// here we need some test if the connectors "minimal" configuration is satisfied, e.g. must have url at minimum
	c := Connector{
		name:          conf["name"],
		mailCmd:       conf["mailCmd"],
		mailCmdParams: conf["mailCmdParams"],
		mailTemplate:  conf["mailTemplate"],
		allowList:     conf["allowList"],
		blockList:     conf["blockList"],
	}
	return &c, nil
}

func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {
	var (
		e         error
		cmdparams = bytes.Buffer{}
	)

	tmpl := template.Must(template.New("cmdparams").Parse(c.mailCmdParams))
	e = tmpl.Execute(&cmdparams, mp.JobContext)
	if e != nil {
		return e
	}

	l.Printf("mailto params %#v\n", c)
	// todo:
	// - call lookup on targetUserId
	// - test if in allowList/blockList
	cmd := exec.Command(c.mailCmd, cmdparams.String(), mp.TargetUser)
	//cmd.Stdin = bytes.NewBuffer([]byte{0x04})
	out, e := cmd.Output()
	if e != nil {
		return e
	}

	l.Println(string(out))

	return e
}
