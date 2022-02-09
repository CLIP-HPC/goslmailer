package msteams

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/pja237/goslmailer/internal/lookup"
	"github.com/pja237/goslmailer/internal/slurmjob"
)

func NewConnector(conf map[string]string) (*Connector, error) {
	// here we need some test if the connectors "minimal" configuration is satisfied, e.g. must have url at minimum
	c := Connector{
		name:                 conf["name"],
		url:                  conf["url"],
		renderToFile:         conf["renderToFile"],
		spoolDir:             conf["spoolDir"],
		adaptiveCardTemplate: conf["adaptiveCardTemplate"],
		useLookup:            conf["useLookup"],
	}
	return &c, nil
}

func (c *Connector) msteamsRenderCardTemplate(j *slurmjob.JobContext, userid string, buf *bytes.Buffer) error {

	var x = struct {
		Job    slurmjob.JobContext
		UserID string
	}{
		*j,
		userid,
	}

	f, err := os.ReadFile(c.adaptiveCardTemplate)
	if err != nil {
		return err
	}
	t := template.Must(template.New("AdaptiveCard").Parse(string(f)))
	err = t.Execute(buf, x)
	if err != nil {
		return err
	}
	return nil
}

func (c *Connector) dumpConnector(l *log.Logger) {
	l.Printf("msteams.dumpConnector: name: %s\n", c.name)
	l.Printf("msteams.dumpConnector: url: %s\n", c.url)
	l.Printf("msteams.dumpConnector: renderToFile: %s\n", c.renderToFile)
	l.Printf("msteams.dumpConnector: spoolDir: %s\n", c.spoolDir)
	l.Printf("msteams.dumpConnector: adaptiveCardTemplate: %s\n", c.adaptiveCardTemplate)
	l.Printf("msteams.dumpConnector: useLookup: %s\n", c.useLookup)
	l.Println("................................................................................")

}

func (c *Connector) SendMessage(j *slurmjob.JobContext, targetUserId string, l *log.Logger) error {

	var (
		e       error
		outFile string
	)

	l.Println("................... sendToMSTeams START ........................................")

	// lookup the end-system userid from the one sent by slurm (if lookup is set in "useLookup" config param)
	enduser := lookup.ExtLookupUser(targetUserId, c.useLookup)
	l.Printf("Looked up %s -> %s\n", targetUserId, enduser)

	// prepare outfile name
	t := strconv.FormatInt(time.Now().UnixNano(), 10)
	l.Printf("MsTeams time: %s\n", t)
	outFile = "rendered-" + j.SLURM_JOB_ID + "-" + enduser + "-" + t + ".json"

	l.Printf("MsTeams sending to targetUserID: %s\n", enduser)

	// debug purposes
	c.dumpConnector(l)

	// here we can put some logic, e.g.
	// if job==fail, send red card
	// else if job==begin, send green card
	// else if job==end, send green card with jobinfo
	// else blabla
	// or do it in template

	// buffer to place rendered json in
	buffer := bytes.Buffer{}
	err := c.msteamsRenderCardTemplate(j, enduser, &buffer)
	if err != nil {
		return err
	}

	// this can be: "yes", "spool", anythingelse
	switch c.renderToFile {
	case "yes":
		res, err := io.ReadAll(&buffer)
		e = err
		os.WriteFile(outFile, res, 0644)
		l.Printf("MsTeams send to file: %s\n", outFile)
	case "spool":
		res, err := io.ReadAll(&buffer)
		e = err
		os.WriteFile(c.spoolDir+"/"+outFile, res, 0644)
		l.Printf("MsTeams send to spool-file: %s\n", c.spoolDir+"/"+outFile)
	default:
		// handle here "too many requests" 4xx and place the rendered message to spool dir to be picked up later by the "throttler"
		resp, err := http.Post(c.url, "application/json", &buffer)
		e = err
		l.Printf("MsTeams RESPONSE Status: %s\n", resp.Status)
		l.Printf("MsTeams RESPONSE Proto: %s\n", resp.Proto)
	}

	l.Println("................... sendToMSTeams END ..........................................")

	return e
}
