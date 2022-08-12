package main

import (
	"bytes"
	"log"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/connectors"
)

const con = "msteams"

type timesTests []struct {
	name    string
	t       string
	want    time.Duration
	wanterr bool
}

func TestGetConfTime(t *testing.T) {
	var tt = timesTests{
		{
			name:    "testEmpty",
			t:       "",
			want:    -1 * time.Second,
			wanterr: true,
		},
		{
			name:    "test1000ms",
			t:       "1000ms",
			want:    1 * time.Second,
			wanterr: false,
		},
		{
			name:    "test1s",
			t:       "1",
			want:    1 * time.Second,
			wanterr: false,
		},
		{
			name:    "testJunk",
			t:       "asd",
			want:    -1 * time.Second,
			wanterr: true,
		},
	}

	for k, v := range tt {
		t.Logf("Running test %d", k)
		t.Run(v.name, func(t *testing.T) {
			got, err := getConfTime(v.t)
			t.Logf("Test %q: GOT: %v WANT: %v WANTERR: %v", v.name, got, v.want, v.wanterr)
			switch {
			case !v.wanterr && err != nil:
				t.Fatalf("FAILED: test %q didn't want error and got one", v.name)
			case v.wanterr && err == nil:
				t.Fatalf("FAILED: test %q wanted error and got none", v.name)
			case v.want != got:
				t.Fatalf("FAILED: test %q wanted: %v and got: %v", v.name, v.want, got)
			}
		})
	}
	// todo
}

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
