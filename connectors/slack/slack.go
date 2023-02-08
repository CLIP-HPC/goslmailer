package slack

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
	"github.com/eritikass/githubmarkdownconvertergo"
	"github.com/slack-go/slack"
)

func init() {
	connectors.Register(connectorName, connSlack)
}

func (c *Connector) ConfigConnector(conf map[string]string) error {
	// Fill out the Connector structure with values from config file
	c.token = conf["token"]
	c.messageTemplate = conf["messageTemplate"]
	c.renderToFile = conf["renderToFile"]
	c.spoolDir = conf["spoolDir"]
	c.useLookup = conf["useLookup"]

	switch {
	// slack token must be present
	case c.token == "":
		return errors.New("slack bot token must be defined, aborting")
	// if renderToFile=="no" or "spool" then spoolDir must not be empty
	case c.renderToFile == "no" || c.renderToFile == "spool":
		if c.spoolDir == "" {
			return errors.New("slack spoolDir must be defined, aborting")
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

	l.Println("................... sendToSlack START ........................................")
	l.Print("MessagePack: ", mp)

	// Create a new Slack sesison using the provided bot token
	api := slack.New(c.token)

	enduser, err := lookup.ExtLookupUser(mp.TargetUser, c.useLookup, l)
	if err != nil {
		l.Printf("Lookup failed for %s with %s\n", mp.TargetUser, err)
		return err
	}
	l.Printf("Looked up with %q %s -> %s\n", c.useLookup, mp.TargetUser, enduser)

	// Get the correct enduser to send to
	// The bot should not be added to channels it may not send messages in.
	if c.renderToFile != "spool" {
		// buffer to place rendered json in
		buffer = bytes.Buffer{}
		err := renderer.RenderTemplate(c.messageTemplate, "", mp.JobContext, enduser, &buffer)
		if err != nil {
			return err
		}
	}

	// this can be: "yes", "spool", anythingelse
	switch c.renderToFile {
	case "yes":
		// render json template to a file in working directory - debug purposes
		// Optional. But can be extremely useful.
		t := strconv.FormatInt(time.Now().UnixNano(), 10)
		l.Printf("Time: %s\n", t)
		outFile = "rendered-" + mp.JobContext.SLURM_JOB_ID + "-" + enduser + "-" + t + ".json"
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
		// deposit GOB to spoolDir if allowed (can be: YES from goslmailer, NO from gobler, since it's already spooled)
		if useSpool {
			l.Printf(c.spoolDir)
			err := spool.DepositToSpool(c.spoolDir, mp)
			if err != nil {
				l.Printf("DepositToSpool Failed!\n")
				return err
			}
		}
	default:
		l.Printf("Sending to channelID or userID: %s\n", enduser)

		markdown := githubmarkdownconvertergo.Slack(buffer.String(), githubmarkdownconvertergo.SlackConvertOptions{Headlines: true})
		// markdown := strings.ReplaceAll(buffer.String(), "**", "*")
		mdBlock := slack.NewTextBlockObject("mrkdwn", markdown, false, false)
		sectionBlock := slack.NewSectionBlock(mdBlock, nil, nil)
		options := slack.MsgOptionBlocks(sectionBlock)
		_, _, _, err := api.SendMessage(enduser, options)
		if err != nil {
			l.Println("PostMessage error: ", err)
			dts = true
		}
	}

	// BACKOFF code, sending failed, we set dts to true and if we're allowed to spool (again, NO from gobler) then we spool.
	if dts && useSpool {
		l.Printf("Backing off to spool.\n")
		err := spool.DepositToSpool(c.spoolDir, mp)
		if err != nil {
			l.Printf("DepositToSpool Failed!\n")
			return err
		}
	}

	l.Println("................... sendToSlack END ..........................................")

	return e
}
