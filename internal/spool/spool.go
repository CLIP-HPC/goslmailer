package spool

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/CLIP-HPC/goslmailer/internal/message"
)

type spool struct {
	spoolDir string
}

// DepositToSpool is a wrapper around spool.NewSpool and *spool.DepositGob
func DepositToSpool(dir string, m *message.MessagePack) error {

	s, err := NewSpool(dir)
	if err != nil {
		return err
	}
	err = s.DepositGob(m)
	if err != nil {
		return err
	}
	return nil
}

func NewSpool(dir string) (*spool, error) {
	var gd = new(spool)
	// test if dir exists and we can write into it
	fi, err := os.Stat(dir)
	switch {
	case err != nil:
		return nil, err
	case !fi.IsDir():
		return nil, errors.New("ERROR: Gob directory is not a directory")
		// todo: missing writability test
	}
	gd.spoolDir = dir
	return gd, nil
}

func (s *spool) DepositGob(m *message.MessagePack) error {
	// todo: replace with a function
	fn := s.spoolDir + "/" + m.Connector + "-" + m.TargetUser + "-" + strconv.FormatInt(m.TimeStamp.UnixNano(), 10) + ".gob"
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	genc := gob.NewEncoder(f)
	err = genc.Encode(*m)
	if err != nil {
		return err
	}
	// todo: proper logging
	fmt.Println("Deposit gob OK!")

	return nil
}

func (s *spool) FetchGob(fileName string, l *log.Logger) (*message.MessagePack, error) {
	var mp = new(message.MessagePack)

	f, err := os.Open(s.spoolDir + "/" + fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	genc := gob.NewDecoder(f)
	err = genc.Decode(mp)
	if err != nil {
		l.Println(err)
		return nil, err
	}
	//l.Printf("Fetch gob OK! Gob timestamp: %s\n", mp.TimeStamp.String())

	return mp, nil
}
