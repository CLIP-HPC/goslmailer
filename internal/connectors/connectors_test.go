package connectors_test

import (
	"bytes"
	"log"
	"testing"

	_ "github.com/CLIP-HPC/goslmailer/connectors/discord"
	_ "github.com/CLIP-HPC/goslmailer/connectors/mailto"
	_ "github.com/CLIP-HPC/goslmailer/connectors/matrix"
	_ "github.com/CLIP-HPC/goslmailer/connectors/msteams"
	_ "github.com/CLIP-HPC/goslmailer/connectors/telegram"
	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/connectors"
)

var connectorsExpected = []string{"msteams", "mailto"}
var connectorsExpectedNot = []string{"textfile"}

func TestPopulateConnectors(t *testing.T) {

	wr := bytes.Buffer{}
	l := log.New(&wr, "Testing: ", log.Llongfile)

	cfg := config.NewConfigContainer()
	err := cfg.GetConfig("../../test_data/config_test/gobler.conf")
	if err != nil {
		t.Fatalf("MAIN: getConfig(gobconfig) failed: %s", err)
	}

	err = connectors.ConMap.PopulateConnectors(cfg, l)
	if err != nil {
		t.Fatalf("conns.PopulateConnectors() FAILED with %s\n", err)
	}

	t.Run("connectorsExpected", func(t *testing.T) {
		for _, v := range connectorsExpected {
			t.Logf("Testing for connector %s", v)
			if _, ok := connectors.ConMap[v]; !ok {
				t.Fatalf("Connector %s not configured!", v)
			} else {
				t.Logf("FOUND... good!\n")
			}
		}
	})
	t.Run("connectorsExpectedNot", func(t *testing.T) {
		for _, v := range connectorsExpectedNot {
			t.Logf("Testing for connector %s", v)
			if _, ok := connectors.ConMap[v]; ok {
				t.Fatalf("Connector %s configured but must NOT be!", v)
			} else {
				t.Logf("NOT FOUND... good!\n")
			}
		}
	})
}
