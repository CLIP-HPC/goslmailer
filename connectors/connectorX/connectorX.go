package connectorX

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/lookup"
	"github.com/CLIP-HPC/goslmailer/internal/message"
	"github.com/CLIP-HPC/goslmailer/internal/renderer"
	"github.com/CLIP-HPC/goslmailer/internal/spool"
)

// init registers the new connector with the connectors package
func init() {
	connectors.Register(connectorName, connConnectorX)
}

// ConfigConnector fills out the package Connector structure with values from config file.
// Recommended to also do sanity checking of config values here. Mandatory.
// Makes connectorX.Connector type satisfy the connectors.Connector interface.
func (c *Connector) ConfigConnector(conf map[string]string) error {
	// Fill out the Connector structure with values from config file
	c.name = conf["name"]
	c.addr = conf["addr"]
	c.port = conf["port"]
	c.templateFile = conf["templateFile"]
	c.renderToFile = conf["renderToFile"]
	c.spoolDir = conf["spoolDir"]
	c.useLookup = conf["useLookup"]

	// Here you can do sanity checking and defaulting if needed.
	// e.g.
	// if renderToFile=="no" or "spool" then spoolDir must not be empty
	switch c.renderToFile {
	case "no", "spool":
		if c.spoolDir == "" {
			return errors.New("spoolDir must be defined, aborting")
		}
	}
	return nil
}

// SendMessage is the main method of the connector code, it usually does something like:
// lookup.ExtLookupUser() (optional), renderer.RenderTemplate() (optional), code to deliver msg (mandatory) and call to spool.DepositToSpool() to spool it (optional).
// useSpool bool is used to distinguish whether it's called from goslmailer (true, allow use of spooling) or gobler sender (false, must not use spool.DepositToSpool())
// Note: spooling code is completely OPTIONAL and connector doesn't have to implement it (see: mailto connector)
// Returns error if encountered.
// Makes connectorX.Connector type satisfy the connectors.Connector interface used in the main loop of goslmailer and goblers sender goroutine.
// Mandatory.
func (c *Connector) SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error {

	var (
		e       error = nil
		outFile string
		dts     bool = false // DumpToSpool
		buffer  bytes.Buffer
	)

	l.Println("................... sendToconnectorX START ........................................")

	// LOOKUP. optional.
	// mp.TargetUser contains the raw string coming from slurm (or from: --mail-user=connector:TargetUser)
	// If there is a need to do a special lookup on that string to get the enduser id, implement lookup function in the lookup package and call from here.
	// EXAMPLE use case: slurm sends us a linux username, but to make msteams connector work we need to get the UserPrincipalName AD attribute to deliver the message.
	// So, we configure sssd to place UPN in GECOS field and implement the GECOS lookup function to do the translation.
	// Else, if what slurm supplies is ok, you can ignore this part, or leave this in and don't specify the useLookup in which case enduser==mp.TargetUser
	// Optional.
	enduser, err := lookup.ExtLookupUser(mp.TargetUser, c.useLookup, l)
	if err != nil {
		l.Printf("Lookup failed for %s with %s\n", mp.TargetUser, err)
		return err
	}
	l.Printf("Looked up with %q %s -> %s\n", c.useLookup, mp.TargetUser, enduser)

	l.Printf("Sending to targetUserID: %s\n", enduser)
	// EOLOOKUP.

	// debug purposes
	c.dumpConnector(l)

	// BOILERPLATE MAIN
	// The following part of the code can be reused for practically any connector.
	// It includes the spooling functionality conditional code paths (can be ripped out).

	// If we don't want to spool, i.e. we want to render to file or send, then call the renderer and prep the message.
	// (if needed, i can imagine all sorts of weird connectors that send/print raw data from message.MessagePack
	if c.renderToFile != "spool" {
		// buffer to place rendered json in
		buffer = bytes.Buffer{}
		// call the renderer with:
		// 1. the path to template file,
		// 2. "text" - use go package text/template (other possible value: "HTML" to use html/template)
		// 3. message.MessagePack.JobContext structure (used in rendering as data)
		// 4. enduser string (target user to send the message to, also used in rendering as data)
		// 5. buffer to hold the rendered message
		err := renderer.RenderTemplate(c.templateFile, "text", mp.JobContext, enduser, &buffer)
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
			err := spool.DepositToSpool(c.spoolDir, mp)
			if err != nil {
				l.Printf("DepositToSpool Failed!\n")
				return err
			}
		}
	default:
		// HERE be dragons.
		// So, we're not spooling for later, not rendering to local files, meaning here, we put in the code, or call the functions to do the message delivery.
		// Below is an inline example that includes BACKOFF CODE!
		// IF our attempt to send failed, we could set the dts (DumpToSpool) bool to trigger spooling after all
		// (if this connector supports it of course, if not, then just return err)
		con, err := net.Dial("tcp", c.addr+":"+c.port)
		if err != nil {
			return err
		}
		defer con.Close()

		n, err := con.Write(buffer.Bytes())
		if err != nil {
			l.Printf("net.Con.Write() Failed!\n")
			dts = true
			e = err
		} else {
			l.Printf("SENT %d bytes from buffer\n", n)
		}
	}

	// BACKOFF code, sending failed, we set dts to true and if we're allowed to spool (again, NO from gobler) then we spool.
	// Optional.
	if dts && useSpool {
		l.Printf("Backing off to spool.\n")
		err := spool.DepositToSpool(c.spoolDir, mp)
		if err != nil {
			l.Printf("DepositToSpool Failed!\n")
			return err
		}
	}

	l.Println("................... sendToconnectorX END ..........................................")

	return e
}
