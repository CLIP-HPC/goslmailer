package main

import (
	"log"

	"github.com/pja237/goslmailer/connectors/mailto"
	"github.com/pja237/goslmailer/connectors/msteams"
	"github.com/pja237/goslmailer/internal/slurmjob"
)

type connector interface {
	SendMessage(*slurmjob.JobContext, string, *log.Logger) error
}

type connectors map[string]connector

// Populate the map 'connectors' with connectors specified in config file and their instance from package.
func (c *connectors) populateConnectors(conf *configContainer, l *log.Logger) error {
	// Iterate through map of connectors from config file.
	for k, v := range conf.Connectors {
		switch k {
		case "mailto":
			// For each recognized, call the connectorpkg.NewConnector() and...
			// todo: make this a little bit less ugly...
			con, err := mailto.NewConnector(v)
			if err != nil {
				l.Printf("Problem with %s connector configuration. Ignoring.\n", k)
				break
			}
			l.Printf("%s connector configured.\n", k)
			// ...asign its return object value to the connectors map.
			(*c)[k] = con
		case "msteams":
			// For each recognized, call the connectorpkg.NewConnector() and...
			con, err := msteams.NewConnector(v)
			if err != nil {
				l.Printf("Problem with %s connector configuration. Ignoring.\n", k)
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
