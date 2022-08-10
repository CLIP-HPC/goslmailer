package msteams

import "log"

const connectorName = "msteams"

type Connector struct {
	name string
	url  string
	// renderToFile can be: "yes", "no", "spool" <- to chain with "throttler"
	renderToFile         string
	spoolDir             string
	adaptiveCardTemplate string
	useLookup            string
}

func (c *Connector) dumpConnector(l *log.Logger) {
	l.Printf("msteams.dumpConnector: name: %q\n", c.name)
	l.Printf("msteams.dumpConnector: url: %q\n", c.url)
	l.Printf("msteams.dumpConnector: renderToFile: %q\n", c.renderToFile)
	l.Printf("msteams.dumpConnector: spoolDir: %q\n", c.spoolDir)
	l.Printf("msteams.dumpConnector: adaptiveCardTemplate: %q\n", c.adaptiveCardTemplate)
	l.Printf("msteams.dumpConnector: useLookup: %q\n", c.useLookup)
	l.Println("................................................................................")

}

var connMsteams *Connector = new(Connector)
