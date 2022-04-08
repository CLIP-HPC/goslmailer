package telegram

type Connector struct {
	name            string
	url             string
	token           string
	renderToFile    string
	spoolDir        string
	messageTemplate string
	useLookup       string
        format          string
}
