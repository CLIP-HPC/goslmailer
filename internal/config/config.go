/*
Package config implements the ConfigContainer structure and accompanying methods.
It holds the configuration data for all utilities.
Configuration file format is the same for all.
*/
package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type ConfigContainer struct {
	DebugConfig      bool                         `json:"debugconfig"`
	Logfile          string                       `json:"logfile"`
	Binpaths         map[string]string            `json:"binpaths"`
	DefaultConnector string                       `json:"defaultconnector"`
	Connectors       map[string]map[string]string `json:"connectors"`
	QosMap           map[string]uint64            `json:"qosmap"`
	//QosMap           map[uint64]string            `json:"qosmap"`
}

func NewConfigContainer() *ConfigContainer {
	return new(ConfigContainer)
}

// Read & unmarshall configuration from 'name' file into configContainer structure
func (cc *ConfigContainer) GetConfig(name string) error {
	f, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	// if HasSuffix(".toml") -> toml.Unmarshall
	// else json.Unmarshall
	if strings.HasSuffix(name, ".toml") {
		err = toml.Unmarshal(f, cc)
	} else {
		err = json.Unmarshal(f, cc)
	}

	if err != nil {
		return err
	}

	cc.testNsetBinPaths()

	return nil
}

func (cc *ConfigContainer) testNsetBinPaths() error {

	if cc.Binpaths == nil {
		cc.Binpaths = make(map[string]string)
	}

	// default paths
	defaultpaths := map[string]string{
		"sacct": "/usr/bin/sacct",
		"sstat": "/usr/bin/sstat",
	}

	for key, path := range defaultpaths {
		if val, exists := cc.Binpaths[key]; !exists || val == "" {
			cc.Binpaths[key] = path
		}
	}

	return nil
}

func (cc *ConfigContainer) DumpConfig(l *log.Logger) {
	if cc.DebugConfig {
		l.Printf("DUMP CONFIG:\n")
		l.Printf("CONFIGURATION: %#v\n", cc)
		l.Printf("CONFIGURATION logfile: %s\n", cc.Logfile)
		l.Printf("--------------------------------------------------------------------------------\n")
	}
}
