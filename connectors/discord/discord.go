package discord

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/lookup"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"
	"github.com/CLIP-HPC/goslmailer/internal/spool"
	"github.com/bwmarrin/discordgo"
)

func NewConnector(conf map[string]string) (*Connector, error) {
	// here we need some test if the connectors "minimal" configuration is satisfied, e.g. must have url at minimum
	c := Connector{
		name:            conf["name"],
		triggerString:   conf["triggerString"],
		token:           conf["token"],
		renderToFile:    conf["renderToFile"],
		spoolDir:        conf["spoolDir"],
		messageTemplate: conf["messageTemplate"],
		useLookup:       conf["useLookup"],
		format:          conf["format"],
	}

	switch {
	// token must be present
	case c.token == "":
		return nil, errors.New("discord bot token must be defined, aborting")
	// if renderToFile=="no" or "spool" then spoolDir must not be empty
	case c.renderToFile == "no" || c.renderToFile == "spool":
		if c.spoolDir == "" {
			return nil, errors.New("discord spoolDir must be defined, aborting")
		}

	}

	return &c, nil
}

func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {

	var (
		e       error = nil
		outFile string
		dts     bool = false // DumpToSpool
		buffer  bytes.Buffer
	)

	l.Println("................... sendTodiscord START ........................................")

	// debug purposes
	c.dumpConnector(l)

	// spin up new bot
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + c.token)
	if err != nil {
		l.Println("error creating Discord session,", err)
		return err
	}

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
		//err := c.discordRenderTemplate(mp.JobContext, enduser, &buffer)
		err := renderer.RenderTemplate(c.messageTemplate, c.format, mp.JobContext, enduser, &buffer)
		if err != nil {
			return err
		}
	}

	// this can be: "yes", "spool", anythingelse
	switch c.renderToFile {
	case "yes":
		// render template to a file in working directory - debug purposes
		// prepare outfile name
		t := strconv.FormatInt(time.Now().UnixNano(), 10)
		l.Printf("Time: %s\n", t)
		outFile = "rendered-" + mp.JobContext.SLURM_JOB_ID + "-" + enduser + "-" + t + ".msg"
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
		// Then we send the message through the channel we created.
		//_, err = dg.ChannelMessageSend(enduser, "A successfull message at "+time.Now().String())
		_, err = dg.ChannelMessageSend(enduser, buffer.String())
		if err != nil {
			l.Printf("error sending DM message: %s\n", err)
			dts = true
		} else {
			l.Printf("bot.Send() successful\n")
			dts = false
		}

		dg.Close()
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

	l.Println("................... sendTodiscord END ..........................................")

	return e
}
