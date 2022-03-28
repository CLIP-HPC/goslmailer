package main

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/pja237/goslmailer/internal/spool"
)

type monitor struct {
	connector string
	spoolDir  string
	monitorT  time.Duration
}

// NewMonitor creates and initializes a new monitor object with:
// c connector name, s spooldir location (from config file), t polling time period
func NewMonitor(c string, s string, t time.Duration) (*monitor, error) {
	var m monitor

	if s != "" {
		m.connector = c
		m.spoolDir = s
		m.monitorT = t
	} else {
		return nil, errors.New("no spooldir, aborting")
	}

	return &m, nil
}

func (m *monitor) MonitorWorker(ch chan<- *spool.SpooledGobs, wg *sync.WaitGroup, l *log.Logger) error {

	var oldList, newList, newFiles *spool.SpooledGobs
	oldList = &spool.SpooledGobs{}
	newFiles = &spool.SpooledGobs{}

	defer wg.Done()
	ticker := time.Tick(m.monitorT)

	l.Println("======================= Monitor start ==========================================")
	l.Printf("MONITOR %s Starting\n", m.connector)
	sp, err := spool.NewSpool(m.spoolDir)
	if err != nil {
		return err
	}
	for {
		lock.Lock()
		// get new list of files
		newList, err = sp.GetSpooledGobsList(l)
		lock.Unlock()
		if err != nil {
			l.Printf("MONITOR %s: Failed on Getspooledgobslist(), error %s\n", m.connector, err)
			return err
		}
		// iterate over newlist and each file that doesn't exist in old, put into newfiles to be sent to the Picker
		for k, v := range *newList {
			if _, ok := (*oldList)[k]; !ok {
				// doesn't
				(*newFiles)[k] = v
			}
		}

		// send new-found files
		l.Printf("MONITOR %s: Sent %d gobs\n", m.connector, len(*newFiles))
		ch <- newFiles
		oldList = newList
		newFiles = &spool.SpooledGobs{}

		<-ticker
	}
	l.Printf("Exiting monitor routine %s\n", m.spoolDir)
	l.Println("======================= Monitor end ============================================")
	return nil
}
