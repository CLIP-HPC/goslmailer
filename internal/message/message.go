package message

import (
	"time"

	"github.com/pja237/goslmailer/internal/slurmjob"
)

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
