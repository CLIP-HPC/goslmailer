package mattermost

import "log"

const connectorName = "mattermost"

type Connector struct {
	name            string
	serverUrl       string
	wsUrl           string
	triggerString   string
	token           string
	renderToFile    string
	spoolDir        string
	messageTemplate string
	useLookup       string
	format          string
}

func (c *Connector) dumpConnector(l *log.Logger) {
	l.Printf("mattermost.dumpConnector: name: %q\n", c.name)
	l.Printf("mattermost.dumpConnector: serverUrl: %q\n", c.serverUrl)
	l.Printf("mattermost.dumpConnector: wsUrl: %q\n", c.wsUrl)
	l.Printf("mattermost.dumpConnector: triggerString: %qn", c.triggerString)
	l.Printf("mattermost.dumpConnector: token: ***\n")
	l.Printf("mattermost.dumpConnector: renderToFile: %q\n", c.renderToFile)
	l.Printf("mattermost.dumpConnector: spoolDir: %q\n", c.spoolDir)
	l.Printf("mattermost.dumpConnector: messageTemplate: %q\n", c.messageTemplate)
	l.Printf("mattermost.dumpConnector: useLookup: %q\n", c.useLookup)
	l.Printf("mattermost.dumpConnector: format: %q\n", c.format)
	l.Println("................................................................................")

}

var connmattermost *Connector = new(Connector)
