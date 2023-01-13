package slack

import "log"

const connectorName = "slack"

type Connector struct {
	name            string // optional
	allowedChannels string // list of comma-seperated (no spaces) allowed channel IDs (ex: "C04B9GLV9MZ,C04B9GNTAKD,C04BC3UB42G")
	token           string // slack api token of the bot
	messageTemplate string // path to template file

	renderToFile string // renderToFile can be: "yes", "no", "spool"
	spoolDir     string // where to place spooled messages
	useLookup    string // string passed to lookup.ExtLookupUser() which determines which lookup function to call
}

func (c *Connector) dumpConnector(l *log.Logger) {
	l.Printf("slack.dumpConnector: name: %q\n", c.name)
	l.Printf("slack.dumpConnector: messageTemplate: %q\n", c.messageTemplate)
	l.Printf("slack.dumpConnector: allowedChannels: %q\n", c.allowedChannels)
	l.Printf("slack.dumpConnector: token: PRESENT\n")
	l.Printf("slack.dumpConnector: renderToFile: %q\n", c.renderToFile)
	l.Printf("slack.dumpConnector: spoolDir: %q\n", c.spoolDir)
	l.Printf("slack.dumpConnector: useLookup: %q\n", c.useLookup)
	l.Println("................................................................................")

}

var connSlack *Connector = new(Connector)
