package main

import (
	"bytes"
	"log"
	"reflect"
	"testing"
)

type ic_test_case struct {
	name   string
	defcon string
	invocationContext
	want Receivers
}

func TestGenerateReceivers(t *testing.T) {

	wr := bytes.Buffer{}
	l := log.New(&wr, "Testing: ", log.Llongfile)

	for _, v := range ic_tc {
		t.Run(v.name, func(t *testing.T) {
			// func (ic *invocationContext) generateReceivers(defCon string, l *log.Logger) {
			v.invocationContext.generateReceivers(v.defcon, l)
			t.Logf("\nTest  : %s\nSent  : %q, %v\nGot   : %q\nExpect: %q\n", v.name, v.invocationContext.CmdParams, v.defcon, v.invocationContext.Receivers, v.want)
			if !reflect.DeepEqual(v.want, v.invocationContext.Receivers) {
				//t.Logf("\nTest  : %s\nSent  : %q, %v\nGot   : %q\nExpect: %q\n", v.name, v.invocationContext.CmdParams, v.defcon, v.invocationContext.Receivers, v.want)
				t.Fail()
			}
		})
	}

}
