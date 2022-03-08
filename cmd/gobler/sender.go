package main

import (
	"log"
	"os"
	"sync"

	"github.com/pja237/goslmailer/internal/connectors"
	"github.com/pja237/goslmailer/internal/spool"
)

type sender struct {
	connector string
	spoolDir  string
	conn      connectors.Connector
	num       int
}

func NewSender(c string, sd string, cons *connectors.Connectors, num int) (*sender, error) {
	var s sender

	s.connector = c     // connector name
	s.spoolDir = sd     // connector spooldir
	s.conn = (*cons)[c] // connector interface
	s.num = num         // sender number

	return &s, nil
}

func (s *sender) SenderWorker(psCh <-chan *spool.FileGob, psfCh chan<- *spool.FileGob, wg *sync.WaitGroup, l *log.Logger) error {

	defer wg.Done()

	l.Println("======================= Sender start ===========================================")
	for {
		msg := <-psCh
		l.Printf("SENDER %s#%d: received %#v\n", s.connector, s.num, msg)

		// fetch gob mp
		// todo: error handling here needs more attention!
		sd, err := spool.NewSpool(s.spoolDir)
		if err != nil {
			l.Printf("SENDER %s#%d: newspool returned error %s\n", s.connector, s.num, err)
			continue
		}
		mp, err := sd.FetchGob(msg.Filename, l)
		if err != nil {
			l.Printf("SENDER %s#%d: fetchgob returned error %s\n", s.connector, s.num, err)
			continue
		}
		// useSpool == false when called from here, gob is already on disk!
		err = s.conn.SendMessage(mp, false, l)
		if err != nil {
			l.Printf("SENDER %s#%d: connector.sendmessage() returned error %s\n", s.connector, s.num, err)
			// send it back to picker
			psfCh <- msg
		} else {
			// Send succeeded, delete gob
			lock.Lock()
			err = os.Remove(s.spoolDir + "/" + msg.Filename)
			if err != nil {
				l.Printf("SENDER %s#%d: error removing file %s\n", s.connector, s.num, err)
				// todo: unlock and return error? or leave this logged and proceed?
			} else {
				l.Printf("SENDER %s#%d: Gob deleted\n", s.connector, s.num)
			}
			lock.Unlock()
		}
	}
	l.Println("======================= Sender end =============================================")
	return nil
}
