package spool

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/pja237/goslmailer/internal/message"
)

type spool struct {
	spoolDir string
}

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
	}
	gd.spoolDir = dir
	return gd, nil
}

func (s *spool) DepositGob(m *message.MessagePack) error {
	fn := s.spoolDir + "/" + m.Connector + "-" + m.TargetUser + "-" + strconv.FormatInt(m.TimeStamp.UnixNano(), 10) + ".gob"
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	genc := gob.NewEncoder(f)
	err = genc.Encode(*m)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Deposit gob OK!")

	return nil
}

func (s *spool) FetchGob(fileName string) (*message.MessagePack, error) {
	var mp = new(message.MessagePack)

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	genc := gob.NewDecoder(f)
	err = genc.Decode(mp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("Fetch gob OK! Gob timestamp: %s\n", mp.TimeStamp.String())

	return mp, nil
}
