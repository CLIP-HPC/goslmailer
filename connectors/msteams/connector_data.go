package msteams

type Connector struct {
	name string
	url  string
	// renderToFile can be: "yes", "no", "spool" <- to chain with "throttler"
	renderToFile         string
	spoolDir             string
	adaptiveCardTemplate string
	useLookup            string
}
