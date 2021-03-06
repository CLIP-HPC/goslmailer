package telegram

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
	telebot "gopkg.in/telebot.v3"
)

func NewConnector(conf map[string]string) (*Connector, error) {
	// here we need some test if the connectors "minimal" configuration is satisfied, e.g. must have url at minimum
	c := Connector{
		name:            conf["name"],
		url:             conf["url"],
		token:           conf["token"],
		renderToFile:    conf["renderToFile"],
		spoolDir:        conf["spoolDir"],
		messageTemplate: conf["messageTemplate"],
		useLookup:       conf["useLookup"],
		format:          conf["format"],
	}
	// if renderToFile=="no" or "spool" then spoolDir must not be empty
	switch c.renderToFile {
	case "no", "spool":
		if c.spoolDir == "" {
			return nil, errors.New("telegram spoolDir must be defined, aborting")
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

	l.Println("................... sendToTelegram START ........................................")

	// debug purposes
	c.dumpConnector(l)

	// spin up new bot
	tb, err := telebot.NewBot(telebot.Settings{
		Token: c.token,
	})
	if err != nil {
		l.Fatal(err)
	}

	// lookup the end-system userid from the one sent by slurm (if lookup is set in "useLookup" config param)
	enduser, err := lookup.ExtLookupUser(mp.TargetUser, c.useLookup, l)
	if err != nil {
		l.Printf("Lookup failed for %s with %s\n", mp.TargetUser, err)
		return err
	}
	l.Printf("Looked up with %q %s -> %s\n", c.useLookup, mp.TargetUser, enduser)

	l.Printf("Sending to targetUserID: %s\n", enduser)

	// get chat ID which comes from --mail-user=telegram:cID switch
	cID, err := strconv.ParseInt(enduser, 10, 64)
	if err != nil {
		l.Printf("cID strconv failed %s", err)
		return err
	}

	chat, err := tb.ChatByID(cID)
	if err != nil {
		l.Printf("chatbyusername failed %s", err)
		return err
	}

	// don't render template when using spool
	if c.renderToFile != "spool" {
		// buffer to place rendered json in
		buffer = bytes.Buffer{}
		//err := c.telegramRenderTemplate(mp.JobContext, enduser, &buffer)
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
		//https://core.telegram.org/bots/api#formatting-options
		msg, err := tb.Send(chat, buffer.String(), c.format)
		if err != nil {
			l.Printf("bot.Send() Failed: %s\n", err)
			dts = true
			//return err
			e = err
		} else {
			l.Printf("bot.Send() successful, messageID: %d\n", msg.ID)
			dts = false
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

	l.Println("................... sendToTelegram END ..........................................")

	return e
}
