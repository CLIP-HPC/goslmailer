package mattermost

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/lookup"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"
	"github.com/CLIP-HPC/goslmailer/internal/spool"
	"github.com/mattermost/mattermost-server/v5/model"
)

func init() {
	connectors.Register(connectorName, connmattermost)
}

func (c *Connector) ConfigConnector(conf map[string]string) error {

	c.name = conf["name"]
	c.serverUrl = conf["serverUrl"]
	c.wsUrl = conf["wsUrl"]
	c.triggerString = conf["triggerString"]
	c.token = conf["token"]
	c.renderToFile = conf["renderToFile"]
	c.spoolDir = conf["spoolDir"]
	c.messageTemplate = conf["messageTemplate"]
	c.useLookup = conf["useLookup"]
	c.format = conf["format"]

	// if renderToFile=="no" or "spool" then spoolDir must not be empty
	switch c.renderToFile {
	case "no", "spool":
		if c.spoolDir == "" {
			return errors.New("mattermost spoolDir must be defined, aborting")
		}
	}
	return nil
}

func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {

	var (
		e       error = nil
		outFile string
		dts     bool = false // DumpToSpool
		buffer  bytes.Buffer
	)

	l.Println("................... sendTomattermost START ........................................")

	// debug purposes
	c.dumpConnector(l)

	// lookup the end-system userid from the one sent by slurm (if lookup is set in "useLookup" config param)
	enduser, err := lookup.ExtLookupUser(mp.TargetUser, c.useLookup, l)
	if err != nil {
		l.Printf("Lookup failed for %s with %s\n", mp.TargetUser, err)
		return err
	}
	l.Printf("Looked up with %q %s -> %s\n", c.useLookup, mp.TargetUser, enduser)

	l.Printf("Sending to targetUserID: %s\n", enduser)

	// don't render template when using spool
	if c.renderToFile != "spool" {
		// buffer to place rendered json in
		buffer = bytes.Buffer{}
		err := renderer.RenderTemplate(c.messageTemplate, c.format, mp.JobContext, enduser, &buffer)
		if err != nil {
			return err
		}
	}

	// this can be: "yes", "spool", anythingelse
	switch c.renderToFile {
	case "yes":
		// render markdown template to a file in working directory - debug purposes
		// prepare outfile name
		t := strconv.FormatInt(time.Now().UnixNano(), 10)
		l.Printf("Time: %s\n", t)
		outFile = "rendered-" + mp.JobContext.SLURM_JOB_ID + "-" + enduser + "-" + t + ".md"
		res, err := io.ReadAll(&buffer)
		if err != nil {
			return err
		}
		err = os.WriteFile(outFile, res, 0644)
		if err != nil {
			return err
		}
		l.Printf("Send successful to file: %s\n", outFile)
	case "spool":
		// deposit GOB to spoolDir if allowed
		if useSpool {
			err := spool.DepositToSpool(c.spoolDir, mp)
			if err != nil {
				l.Printf("DepositToSpool Failed!\n")
				return err
			}
		}
	default:
		// Send message via mattermost

		client := model.NewAPIv4Client(c.serverUrl)
		client.SetOAuthToken(c.token)
		l.Printf("\nclient: %#v\n", client)

		resPost := model.Post{}
		resPost.ChannelId = enduser
		resPost.Message = buffer.String()
		if _, r := client.CreatePost(&resPost); r.Error == nil {
			l.Printf("Post response to chan: %s successfull!\n", resPost.ChannelId)
		} else {
			l.Printf("Post response FAILED with: %#v\n", r)
			dts = true
		}
	}

	// save mp to spool if we're allowed (not allowed when called from gobler, to prevent gobs multiplying)
	if dts && useSpool {
		l.Printf("Backing off to spool.\n")
		err := spool.DepositToSpool(c.spoolDir, mp)
		if err != nil {
			l.Printf("DepositToSpool Failed!\n")
			return err
		}
	}

	l.Println("................... sendTomattermost END ..........................................")

	return e
}
