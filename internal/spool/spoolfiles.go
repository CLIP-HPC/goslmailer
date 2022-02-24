package spool

import (
	"fmt"
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

func (s *spool) GetSpooledGobsList() (*SpooledGobs, error) {
	var sg SpooledGobs = SpooledGobs{}

	de, err := os.ReadDir(s.spoolDir)
	if err != nil {
		return nil, err
	}
	for _, v := range de {
		//fmt.Printf("FILE: %q is regular file: %v\n", v.Name(), v.Type().IsRegular())
		if v.Type().IsRegular() && strings.HasSuffix(v.Name(), ".gob") {
			//fmt.Printf("FILE: %q is a GOB\n", v.Name())
			mp, err := s.FetchGob(v.Name())
			if err != nil {
				// todo: logging
				fmt.Printf("FAILED to read %q gob file: %s\n", v.Name(), err)
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
	// todo: remove this, was used in devlopment, or use logging, or not, too much info
	for _, v := range sg {
		// todo: logging
		fmt.Printf("FILE: %q\n", v.Filename)
	}
	return &sg, nil
}
