package mailto

import (
	"bytes"
	"errors"
	"log"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"
)

func init() {
	connectors.Register(connectorName, connMailto)
}

func (c *Connector) ConfigConnector(conf map[string]string) error {
	c.name = conf["name"]
	c.mailCmd = conf["mailCmd"]
	c.mailCmdParams = conf["mailCmdParams"]
	c.mailTemplate = conf["mailTemplate"]
	c.mailFormat = conf["mailFormat"]
	c.allowList = conf["allowList"]
	c.blockList = conf["blockList"]

	// here we need some test if the connectors "minimal" configuration is satisfied, e.g. must have url at minimum
	//
	// if ok, return nil error
	return nil
}

func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {
	var (
		e         error
		cmdparams = bytes.Buffer{}
		body      = bytes.Buffer{}
	)

	// render mail command line params (-s "mail subject" et.al.)
	tmpl := template.Must(template.New("cmdparams").Parse(c.mailCmdParams))
	e = tmpl.Execute(&cmdparams, mp.JobContext)
	if e != nil {
		return e
	}

	// render mail body
	err := renderer.RenderTemplate(c.mailTemplate, c.mailFormat, mp.JobContext, mp.TargetUser, &body)
	if err != nil {
		return err
	}

	l.Printf("PARAMS: %#v\n", c)
	l.Printf("CMD: %q\n", string(cmdparams.Bytes()))
	l.Printf("BODY: %q\n", string(body.Bytes()))

	// todo:
	// - call lookup on targetUserId?
	// - implement useSpool mechanics for gobler

	// allowList
	re, err := regexp.Compile(c.allowList)
	if err != nil {
		return err
	}
	if !re.Match([]byte(mp.TargetUser)) {
		// not in allowList
		return errors.New("not allowed to send mail to user")
	}

	// send:
	cmd := exec.Command(c.mailCmd, cmdparams.String(), mp.TargetUser)
	l.Printf("ExecCMD: %q %q\n", cmd.Path, cmd.Args)
	cmd.Stdin = &body
	//cmd.Stdin = bytes.NewBuffer([]byte{0x04})
	out, e := cmd.Output()
	if e != nil {
		return e
	}

	l.Println(string(out))

	return e
}
