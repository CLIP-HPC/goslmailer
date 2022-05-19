/*
connectorX is a sample boilerplate connector. It opens a TCP connection to the specified address:port and sends a rendered template string.
It is an example of how to use all of the 'extra' capabilities: lookups, rendertofile and spooling.
You can start from here and replace the parts of code needed/not needed to implement a new connector.

MUST: Besides the connector package code, in order to "register" the connector, you MUST add its 'case block' to connectors package Connectors.PupulateConnectors().
HERE: ../../internal/connectors/connectors.go:25
*/
package connectorX

import "log"

// Connector structure contains configuration data read in from config file with connectorX.NewConnector().
// Populate this structure with the configuration variables a new connector needs
type Connector struct {
	name         string // optional
	addr         string // hostname/ip to connect to
	port         string // port to connect to
	templateFile string // template file

	// these 3 are optional, if the connector won't use lookups and/or spooling capabilities (gobler service) but send directly from goslmailer they can be completely removed
	renderToFile string // renderToFile can be: "yes", "no", "spool"
	spoolDir     string // where to place spooled messages
	useLookup    string // string passed to lookup.ExtLookupUser() which determines which lookup function to call
}

// dumpConnector logs the connector configuration read from config file
func (c *Connector) dumpConnector(l *log.Logger) {
	l.Printf("connectorX.dumpConnector: name: %q\n", c.name)
	l.Printf("connectorX.dumpConnector: addr: %q\n", c.addr)
	l.Printf("connectorX.dumpConnector: port: %q\n", c.port)
	l.Printf("connectorX.dumpConnector: renderToFile: %q\n", c.renderToFile)
	l.Printf("connectorX.dumpConnector: spoolDir: %q\n", c.spoolDir)
	l.Printf("connectorX.dumpConnector: useLookup: %q\n", c.useLookup)
	l.Println("................................................................................")

}
