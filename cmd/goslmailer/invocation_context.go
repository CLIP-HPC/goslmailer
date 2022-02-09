package main

import (
	"flag"
	"log"
	"strings"
)

type CmdParams struct {
	Subject string
	Other   []string
}

type Receivers []struct {
	scheme string
	target string
}

// Holder of:
// 1. command line parameters (via log package)
// 2. receivers: from parsed command line (1.) comming from: --mail-user userx,mailto:usery@domain,skype:userid via (*invocationContext) generateReceivers() method
type invocationContext struct {
	CmdParams
	Receivers
}

func (ic *invocationContext) getCMDLine() {
	flag.StringVar(&ic.CmdParams.Subject, "s", "Default Blank Subject", "e-mail subject")
	flag.Parse()
	ic.CmdParams.Other = flag.Args()
}

func (ic *invocationContext) dumpCMDLine(l *log.Logger) {
	l.Println("Parsing CMDLine:")
	l.Printf("CMD subject: %#v\n", ic.CmdParams.Subject)
	l.Printf("CMD others: %#v\n", ic.CmdParams.Other)
	l.Println("--------------------------------------------------------------------------------")
}

func (ic *invocationContext) dumpReceivers(l *log.Logger) {
	l.Println("DUMP RECEIVERS:")
	l.Printf("Receivers: %#v\n", ic.Receivers)
	l.Printf("invocationContext: %#v\n", ic)
	l.Println("--------------------------------------------------------------------------------")
}

// populate ic.Receivers (scheme:target) from ic.CmdParams.Other using defCon (defaultconnector) config parameter for undefined schemes
func (ic *invocationContext) generateReceivers(defCon string, l *log.Logger) {
	for _, v := range ic.CmdParams.Other {
		targets := strings.Split(v, ",")
		for i, t := range targets {
			targetsSplit := strings.Split(t, ":")
			l.Printf("generateReceivers: target %d = %#v\n", i, targetsSplit)
			switch len(targetsSplit) {
			// todo: move the lookup part of the code to the connectors, and allow every connector to specify its lookup function
			case 1:
				ic.Receivers = append(ic.Receivers, struct {
					scheme string
					target string
				}{
					// receivers with unspecified connector scheme get global config key "DefaultConnector" set here:
					scheme: defCon,
					// Lookup is now moved to connector package, remove comment later.
					//target: lookup.ExtLookupUser(targetsSplit[0], defCon),
					target: targetsSplit[0],
				})
			case 2:
				ic.Receivers = append(ic.Receivers, struct {
					scheme string
					target string
				}{
					scheme: targetsSplit[0],
					// Lookup is now moved to connector package, remove comment later.
					//target: lookup.ExtLookupUser(targetsSplit[1], targetsSplit[0]),
					target: targetsSplit[1],
				})
			default:
				l.Printf("generateReceivers: IGNORING! unrecognized target string: %s\n", t)
			}
		}
	}
}
