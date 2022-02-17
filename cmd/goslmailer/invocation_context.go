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
			case 1:
				if targetsSplit[0] != "" {
					ic.Receivers = append(ic.Receivers, struct {
						scheme string
						target string
					}{
						// receivers with unspecified connector scheme get global config key "DefaultConnector" set here:
						scheme: defCon,
						target: targetsSplit[0],
					})
				} else {
					l.Printf("generateReceivers: target %d = %#v is an empty receiver, ignoring!\n", i, targetsSplit)
				}
			case 2:
				if targetsSplit[1] != "" && targetsSplit[0] != "" {
					ic.Receivers = append(ic.Receivers, struct {
						scheme string
						target string
					}{
						scheme: targetsSplit[0],
						target: targetsSplit[1],
					})
				} else {
					l.Printf("generateReceivers: target %d = %#v is an empty receiver, ignoring!\n", i, targetsSplit)
				}
			default:
				l.Printf("generateReceivers: IGNORING! unrecognized target string: %s\n", t)
			}
		}
	}
}
