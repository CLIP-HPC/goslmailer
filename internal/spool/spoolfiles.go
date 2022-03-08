package spool

import (
	"log"
	"os"
	"strings"
	"time"
)

type FileGob struct {
	//filename fs.DirEntry
	Filename  string
	User      string
	Connector string
	TimeStamp time.Time
}

//type SpooledGobs map[filename]FileGob
type SpooledGobs map[string]FileGob

func (s *spool) GetSpooledGobsList(l *log.Logger) (*SpooledGobs, error) {
	var sg SpooledGobs = SpooledGobs{}

	de, err := os.ReadDir(s.spoolDir)
	if err != nil {
		return nil, err
	}
	for _, v := range de {
		//l.Printf("FILE: %q is regular file: %v\n", v.Name(), v.Type().IsRegular())
		if v.Type().IsRegular() && strings.HasSuffix(v.Name(), ".gob") {
			//l.Printf("FILE: %q is a GOB\n", v.Name())
			mp, err := s.FetchGob(v.Name(), l)
			if err != nil {
				l.Printf("FAILED to read %q gob file: %s\n", v.Name(), err)
			} else {
				sg[v.Name()] = FileGob{
					Filename:  v.Name(),
					User:      mp.TargetUser,
					Connector: mp.Connector,
					TimeStamp: mp.TimeStamp,
				}
			}

		}
	}
	return &sg, nil
}
