package message

import (
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/slurmjob"
)

// MessagePack is the central data structure that holds all the data about the message that is currently being processed.
// It is used to pass the "message" and its "metadata" between all of the components of the system, e.g. main->connector->spooler->gobler->sender etc.
type MessagePack struct {
	Connector  string
	TargetUser string
	JobContext *slurmjob.JobContext
	TimeStamp  time.Time
}

// NewMsgPack returns the instantiated message.MessagePack structure
func NewMsgPack(connectorName string, targetUser string, jobContext *slurmjob.JobContext) (*MessagePack, error) {
	var m = new(MessagePack)
	m.Connector = connectorName
	m.TargetUser = targetUser
	m.JobContext = jobContext
	m.TimeStamp = time.Now()
	return m, nil
}
