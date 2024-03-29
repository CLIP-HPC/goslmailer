package msteams

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/lookup"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"
	"github.com/CLIP-HPC/goslmailer/internal/spool"
)

func init() {
	connectors.Register(connectorName, connMsteams)
}

func (c *Connector) ConfigConnector(conf map[string]string) error {

	c.name = conf["name"]
	c.url = conf["url"]
	c.renderToFile = conf["renderToFile"]
	c.spoolDir = conf["spoolDir"]
	c.adaptiveCardTemplate = conf["adaptiveCardTemplate"]
	c.useLookup = conf["useLookup"]

	// if renderToFile=="no" or "spool" then spoolDir must not be empty
	switch c.renderToFile {
	case "no", "spool":
		if c.spoolDir == "" {
			return errors.New("spoolDir must be defined, aborting")
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

	l.Println("................... sendToMSTeams START ........................................")

	// lookup the end-system userid from the one sent by slurm (if lookup is set in "useLookup" config param)
	enduser, err := lookup.ExtLookupUser(mp.TargetUser, c.useLookup, l)
	if err != nil {
		l.Printf("Lookup failed for %s with %s\n", mp.TargetUser, err)
		return err
	}
	l.Printf("Looked up with %q %s -> %s\n", c.useLookup, mp.TargetUser, enduser)

	l.Printf("Sending to targetUserID: %s\n", enduser)

	// debug purposes
	c.dumpConnector(l)

	// don't render template when using spool
	if c.renderToFile != "spool" {
		// buffer to place rendered json in
		buffer = bytes.Buffer{}
		//err := c.msteamsRenderCardTemplate(mp.JobContext, enduser, &buffer)
		err := renderer.RenderTemplate(c.adaptiveCardTemplate, "text", mp.JobContext, enduser, &buffer)
		if err != nil {
			return err
		}
	}

	// this can be: "yes", "spool", anythingelse
	switch c.renderToFile {
	case "yes":
		// render json template to a file in working directory - debug purposes

		// prepare outfile name
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
		// deposit GOB to spoolDir if allowed
		if useSpool {
			err := spool.DepositToSpool(c.spoolDir, mp)
			if err != nil {
				l.Printf("DepositToSpool Failed!\n")
				return err
			}
		}
	default:
		// handle here "too many requests" 4xx and place the rendered message to spool dir to be picked up later by the "throttler"
		resp, err := http.Post(c.url, "application/json", &buffer)
		if err != nil {
			l.Printf("http.Post Failed!\n")
			dts = true
			//return err
			e = err
		} else {
			l.Printf("RESPONSE Status: %s\n", resp.Status)
			switch resp.StatusCode {
			case 429:
				l.Printf("429 received.\n")
				dts = true
			default:
				l.Printf("Send OK!\n")
			}
		}
	}

	// either http.Post failed, or it got 429, save mp to spool if we're allowed (not allowed when called from gobler, to prevent gobs multiplying)
	if dts && useSpool {
		l.Printf("Backing off to spool.\n")
		err := spool.DepositToSpool(c.spoolDir, mp)
		if err != nil {
			l.Printf("DepositToSpool Failed!\n")
			return err
		}
	}

	l.Println("................... sendToMSTeams END ..........................................")

	return e
}
