package mailto

type Connector struct {
	name          string
	mailCmd       string
	mailCmdParams string
	mailTemplate  string
	mailFormat    string
	allowList     string
	blockList     string
}
