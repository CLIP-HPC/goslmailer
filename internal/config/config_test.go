package config

import (
	"os"
	"testing"
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
			c := NewConfigContainer()
			e := c.GetConfig(testDir + "/" + f.Name())
			if _, ok := c.Connectors["msteams"]; e != nil || !ok {
				t.Fatalf("Test %s failed with: %s\n", f.Name(), e)
			}
			if v, ok := c.Connectors["msteams"]["url"]; !ok || v != "https://msteams/webhook/url" {
				t.Fatalf("Test %s failed finding connectors.msteams.url\n", f.Name())
			}
		})
	}
}

func TestSetBinpaths(t *testing.T) {
	cc := []ConfigContainer{
		{
			Binpaths: map[string]string{
				"sacct": "/usr/bin/sacct",
				"sstat": "blabla",
			},
		},
		{
			Binpaths: map[string]string{
				"sacct": "",
				"sstat": "blabla1",
			},
		},
		{
			Binpaths: map[string]string{
				"sstat": "blabla1",
			},
		},
		{},
	}
	// todo: add []results and test both

	for i, v := range cc {
		t.Logf("Running test %d\n", i)
		t.Logf("PRE: %v\n", v.Binpaths)
		err := v.testNsetBinPaths()
		t.Logf("POST: %v\n", v.Binpaths)
		if err != nil {
			t.Fatalf("Test %d failed with: %s\n", i, err)
		}
		if v.Binpaths["sacct"] == "/usr/bin/sacct" {
			t.Logf("SUCCESS")
		} else {
			t.Fatalf("FAILED: WANT: %q GOT: %q\n", "/usr/bin/sacct", v.Binpaths["sacct"])
		}
	}
}
