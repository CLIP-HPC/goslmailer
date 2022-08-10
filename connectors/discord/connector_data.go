package discord

import "log"

type Connector struct {
	name            string
	triggerString   string
	token           string
	renderToFile    string
	spoolDir        string
	messageTemplate string
	useLookup       string
	format          string
}

func (c *Connector) dumpConnector(l *log.Logger) {
	l.Printf("discord.dumpConnector: name: %q\n", c.name)
	l.Printf("discord.dumpConnector: triggerstring: %q\n", c.triggerString)
	l.Printf("discord.dumpConnector: token: PRESENT\n")
	l.Printf("discord.dumpConnector: renderToFile: %q\n", c.renderToFile)
	l.Printf("discord.dumpConnector: spoolDir: %q\n", c.spoolDir)
	l.Printf("discord.dumpConnector: messageTemplate: %q\n", c.messageTemplate)
	l.Printf("discord.dumpConnector: useLookup: %q\n", c.useLookup)
	l.Printf("discord.dumpConnector: format: %q\n", c.format)
	l.Println("................................................................................")

}
