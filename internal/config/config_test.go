package config_test

import (
	"os"
	"testing"

	"github.com/CLIP-HPC/goslmailer/internal/config"
)

const testDir = "../../test_data/config_test"

func TestConfig(t *testing.T) {
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("ERROR: can not read config test directory %s\n", err)
	}
	for _, f := range files {
		t.Run(f.Name(), func(t *testing.T) {
			t.Logf("Processing file %s\n", f.Name())
			c := config.NewConfigContainer()
			e := c.GetConfig(testDir + "/" + f.Name())
			if _, ok := c.Connectors["msteams"]; e != nil || !ok {
				t.Fatalf("Test %s failed.\n", f.Name())
			}
		})
	}
}
