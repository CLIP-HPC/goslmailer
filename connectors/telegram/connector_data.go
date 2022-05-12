package telegram

import "log"

type Connector struct {
	name            string
	url             string
	token           string
	renderToFile    string
	spoolDir        string
	messageTemplate string
	useLookup       string
	format          string
}

func (c *Connector) dumpConnector(l *log.Logger) {
	l.Printf("telegram.dumpConnector: name: %q\n", c.name)
	l.Printf("telegram.dumpConnector: url: %q\n", c.url)
	l.Printf("telegram.dumpConnector: token: %q\n", c.token)
	l.Printf("telegram.dumpConnector: renderToFile: %q\n", c.renderToFile)
	l.Printf("telegram.dumpConnector: spoolDir: %q\n", c.spoolDir)
	l.Printf("telegram.dumpConnector: messageTemplate: %q\n", c.messageTemplate)
	l.Printf("telegram.dumpConnector: useLookup: %q\n", c.useLookup)
	l.Printf("telegram.dumpConnector: format: %q\n", c.format)
	l.Println("................................................................................")

}
