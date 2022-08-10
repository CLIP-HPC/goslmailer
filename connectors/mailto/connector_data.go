package mailto

const connectorName = "mailto"

type Connector struct {
	name          string
	mailCmd       string
	mailCmdParams string
	mailTemplate  string
	mailFormat    string
	allowList     string
	blockList     string
}

var connMailto *Connector = new(Connector)
