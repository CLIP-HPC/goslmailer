package spool

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/CLIP-HPC/goslmailer/internal/message"
)

type depTest struct {
	dir       string
	mp        *message.MessagePack
	expectErr bool
}

type dTList []depTest

func TestDepositToSpool(t *testing.T) {
	// is this a smart thing to do in testing?
	name, err := os.MkdirTemp("", "goslTest-")
	if err != nil {
		t.Fatalf("ERROR making temp dir for test: %q\n", err)
	} else {
		t.Logf("TEMPDIR: %q\n", name)
		defer func() {
			e := os.RemoveAll(name)
			if e != nil {
				t.Fatalf("ERROR removing temp dir for test: %q\n", err)
			}
		}()
	}
	dts := dTList{
		{
			dir:       "",
			mp:        &message.MessagePack{},
			expectErr: true,
		},
		{
			dir:       "/for/sure/this/doesnt/exist",
			mp:        &message.MessagePack{},
			expectErr: true,
		},
		{
			dir:       name,
			mp:        &message.MessagePack{},
			expectErr: true,
		},
		{
			dir:       name,
			mp:        nil,
			expectErr: true,
		},
		{
			dir: name,
			mp: &message.MessagePack{
				Connector:  "conn",
				TargetUser: "pja",
			},
			expectErr: false,
		},
	}

	for i, v := range dts {
		t.Run(fmt.Sprintf("TEST %d: %q %v", i, v.dir, v.mp), func(t *testing.T) {
			err := DepositToSpool(v.dir, v.mp)
			switch {
			case v.expectErr && err == nil:
				t.Fatalf("FAILED test %q, deposit to %q with %q\n", i, v.dir, err)
			case !v.expectErr && err != nil:
				t.Fatalf("FAILED test %q, deposit to %q with %q\n", i, v.dir, err)
			}
		})
	}
}

func TestNewSpool(t *testing.T) {
	nsl := []struct {
		dir       string
		expecterr bool
	}{
		{
			dir:       "/tmp",
			expecterr: false,
		},
		{
			dir:       "/for/sure/this/doesnt/exist",
			expecterr: true,
		},
		{
			dir:       "",
			expecterr: true,
		},
	}

	for i, v := range nsl {
		teststr := fmt.Sprintf("TEST %d %s %v", i, v.dir, v.expecterr)
		t.Run(teststr, func(t *testing.T) {
			sp, err := NewSpool(v.dir)
			t.Logf("Got: %v\n", sp)
			switch {
			case v.expecterr && err == nil:
				t.Fatalf("FAIL: Expected err and got none")
			case !v.expecterr && err != nil:
				t.Fatalf("FAIL: Expected ok and got err: %q", err)
			case err != nil:
				// here we break the switch if we're in "expected" error, so we can't test sp value below
				break
			case v.dir != sp.spoolDir:
				t.Fatalf("FAIL: Expected spooldir=%q ,got %q\n", v.dir, sp.spoolDir)
			}
		})
	}
}

func TestGenFileName(t *testing.T) {
	testList := []struct {
		name      string
		sdir      string
		mp        *message.MessagePack
		wantstr   string
		expecterr bool
	}{
		{
			name: "test all ok",
			sdir: "/tmp",
			mp: &message.MessagePack{
				Connector:  "testCon",
				TargetUser: "pja",
			},
			wantstr:   "/tmp/testCon-pja-",
			expecterr: false,
		},
		{
			name: "test empty spooldir",
			sdir: "",
			mp: &message.MessagePack{
				Connector:  "testCon",
				TargetUser: "pja",
			},
			wantstr:   "/tmp/testCon-pja-",
			expecterr: true,
		},
		{
			name: "test missing targetuser",
			sdir: "/tmp",
			mp: &message.MessagePack{
				Connector: "testCon",
			},
			wantstr:   "/tmp/testCon-pja-",
			expecterr: true,
		},
		{
			name:      "test empty messagepack",
			sdir:      "/tmp",
			mp:        &message.MessagePack{},
			wantstr:   "/tmp/testCon-pja-",
			expecterr: true,
		},
		{
			name:      "test nil messagepack",
			sdir:      "/tmp",
			mp:        nil,
			wantstr:   "/tmp/testCon-pja-",
			expecterr: true,
		},
	}

	for _, test := range testList {
		t.Run(test.name, func(t *testing.T) {
			fn, err := genFileName(test.sdir, test.mp)
			switch {
			case test.expecterr && err == nil:
				t.Fatalf("FAIL: Expected err and got none")
			case !test.expecterr && err != nil:
				t.Fatalf("FAIL: Expected ok and got err: %q", err)
			case err != nil:
				break
			case !strings.HasPrefix(fn, test.wantstr):
				t.Fatalf("FAIL: %q doesn't have prefix %q", fn, test.wantstr)
			default:
				t.Logf("SUCCESS %q has prefix %q", fn, test.wantstr)
			}
		})
	}
}
