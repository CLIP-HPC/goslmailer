package connectors

import (
	"log"

	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/message"
)

type Connector interface {
	ConfigConnector(conf map[string]string) error
	SendMessage(*message.MessagePack, bool, *log.Logger) error
}

type Connectors map[string]Connector

var ConMap Connectors = Connectors{}

func Register(conName string, conStruct Connector) error {

	if _, ok := ConMap[conName]; !ok {
		log.Printf("Initializing connector: %s\n", conName)
		ConMap[conName] = conStruct
	} else {
		log.Printf("Connector %s already initialized.\n", conName)
	}

	return nil
}

// Populate the map 'connectors' with connectors specified in config file and their instance from package.
// Every newly developed connector must have a case block added here.
func (c *Connectors) PopulateConnectors(conf *config.ConfigContainer, l *log.Logger) error {

	for k, v := range conf.Connectors {
		// test if connector from config is registered in conMap
		if _, ok := (*c)[k]; !ok {
			l.Printf("ERROR: %q connector not initialized, skipping...\n", k)
			continue
		}
		// l.Printf("Unsupported connector found. Ignoring %#v : %#v\n", k, v)
		// if it is, try to configure it
		l.Printf("CONFIGURING: %s with: %#v\n", k, v)
		if err := (*c)[k].ConfigConnector(v); err != nil {
			// config failed, log and remove from map
			l.Printf("ERROR: %q with %s connector configuration. Ignoring.\n", err, k)
			delete(*c, k)
		} else {
			// config successfull, log and do nothing.
			l.Printf("SUCCESS: %s connector configured.\n", k)
		}
	}

	return nil
}
