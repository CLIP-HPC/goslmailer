package main

import (
	"bytes"
	"log"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/pja237/goslmailer/internal/config"
	"github.com/pja237/goslmailer/internal/connectors"
)

const con = "msteams"

func TestConmonGoRoutines(t *testing.T) {
	var (
		conns = make(connectors.Connectors)
		wg    sync.WaitGroup
	)

	wr := bytes.Buffer{}
	l := log.New(&wr, "Testing: ", log.Llongfile)

	cfg := config.NewConfigContainer()
	err := cfg.GetConfig("../../test_data/config_test/gobler.conf")
	if err != nil {
		t.Fatalf("MAIN: getConfig(gobconfig) failed: %s", err)
	}
	ns, err := strconv.Atoi(cfg.Connectors[con]["numSenders"])
	if err != nil {
		t.Fatalf("Atoi(numSenders) failed: %s\n", err)
	}
	expected := 4 + ns

	cm, err := NewConMon(con, cfg.Connectors[con], l)
	if err != nil {
		t.Fatalf("MAIN: NewConMon(%s) failed with: %s\n", con, err)
	}

	err = cm.SpinUp(conns, &wg, l)
	if err != nil {
		t.Fatalf("MAIN: SpinUp(%s) failed with: %s\n", con, err)
	}

	t.Logf("Num goroutines test: got: %d, expected %d (test,main,monitor,picker,%dx sender)\n", runtime.NumGoroutine(), expected, ns)
	if runtime.NumGoroutine() != expected {
		t.Fatal("numGoroutines test failed.")
	} else {
		t.Log("numGoroutines test OK.")
	}

}
