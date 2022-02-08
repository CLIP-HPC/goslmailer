package mailto

type Connector struct {
	name          string
	mailCmd       string
	mailCmdParams string
	mailTemplate  string
	allowList     string
	blockList     string
}
