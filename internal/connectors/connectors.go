package connectors

import (
	"log"

	"github.com/pja237/goslmailer/connectors/mailto"
	"github.com/pja237/goslmailer/connectors/msteams"
	"github.com/pja237/goslmailer/connectors/telegram"
	"github.com/pja237/goslmailer/internal/config"
	"github.com/pja237/goslmailer/internal/message"
)

type Connector interface {
	//SendMessage(mp *message.MessagePack, useSpool bool, l *log.Logger) error
	SendMessage(*message.MessagePack, bool, *log.Logger) error
}

type Connectors map[string]Connector

// Populate the map 'connectors' with connectors specified in config file and their instance from package.
func (c *Connectors) PopulateConnectors(conf *config.ConfigContainer, l *log.Logger) error {
	// Iterate through map of connectors from config file.
	for k, v := range conf.Connectors {
		switch k {
		case "mailto":
			// For each recognized, call the connectorpkg.NewConnector() and...
			// todo: make this a little bit less ugly...
			con, err := mailto.NewConnector(v)
			if err != nil {
				l.Printf("Problem: %q with %s connector configuration. Ignoring.\n", err, k)
				break
			}
			l.Printf("%s connector configured.\n", k)
			// ...asign its return object value to the connectors map.
			(*c)[k] = con
		case "msteams":
			// For each recognized, call the connectorpkg.NewConnector() and...
			con, err := msteams.NewConnector(v)
			if err != nil {
				l.Printf("Problem: %q with %s connector configuration. Ignoring.\n", err, k)
				break
			}
			l.Printf("%s connector configured.\n", k)
			// ...asign its return object value to the connectors map.
			(*c)[k] = con
		case "telegram":
			// For each recognized, call the connectorpkg.NewConnector() and...
			con, err := telegram.NewConnector(v)
			if err != nil {
				l.Printf("Problem: %q with %s connector configuration. Ignoring.\n", err, k)
				break
			}
			l.Printf("%s connector configured.\n", k)
			// ...asign its return object value to the connectors map.
			(*c)[k] = con
		default:
			l.Printf("Unsupported connector found. Ignoring %#v : %#v\n", k, v)
		}
	}
	return nil
}
