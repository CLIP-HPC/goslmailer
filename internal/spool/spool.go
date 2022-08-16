package spool

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
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

// NewSpool instantiates a spool structure with a dir path to the directory where gobs will be deposited (spooled)
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

// DepositGob takes a MessagePack and saves it to a gob file in the spool directory.
func (s *spool) DepositGob(m *message.MessagePack) error {

	// if we got empty messagepack, error!
	if m == nil {
		return errors.New("got nil MessagePack")
	}

	// generate gob file name
	fn, err := genFileName(s.spoolDir, m)
	if err != nil {
		return err
	}

	// create gob file
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	// deposit mp in gob
	genc := gob.NewEncoder(f)
	err = genc.Encode(*m)
	if err != nil {
		return err
	}

	// todo: proper logging
	fmt.Println("Deposit gob OK!")

	return nil
}

// genFileName generates gob full path-filename. Format: spooldir/connector-user-timestamp.gob
func genFileName(dir string, m *message.MessagePack) (string, error) {

	switch {
	case dir == "":
		return "", errors.New("got empty spooldir")
	case m == nil:
		return "", errors.New("got nil messagepack")
	case m.Connector == "":
		return "", errors.New("got empty connector")
	case m.TargetUser == "":
		return "", errors.New("got empty targetuser")
	}

	return dir + "/" + m.Connector + "-" + m.TargetUser + "-" + strconv.FormatInt(m.TimeStamp.UnixNano(), 10) + ".gob", nil

}

// FetchGob takes the gob filename, prepends the spooldir to the name, opens it and returns the decoded MessagePack structure.
// todo: test
func (s *spool) FetchGob(fileName string, l *log.Logger) (*message.MessagePack, error) {

	f, err := os.Open(s.spoolDir + "/" + fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mp, err := decodeGob(f, l)
	if err != nil {
		return nil, err
	}

	return mp, nil
}

// decodeGob take io.Reader and returns its decoded content into a MessagePack structure
// todo: test
func decodeGob(r io.Reader, l *log.Logger) (*message.MessagePack, error) {
	var mp = new(message.MessagePack)

	genc := gob.NewDecoder(r)
	err := genc.Decode(mp)
	if err != nil {
		l.Println(err)
		return nil, err
	}

	return mp, nil
}
