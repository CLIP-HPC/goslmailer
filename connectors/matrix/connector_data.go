package matrix

const connectorName = "matrix"

type Connector struct {
	username   string
	token      string
	homeserver string
	template   string
}

var connMatrix *Connector = new(Connector)
