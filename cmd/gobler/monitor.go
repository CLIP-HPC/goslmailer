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
	// configurable monitor timer
	ticker := time.Tick(m.monitorT * time.Second)

	l.Println("======================= Monitor start ==========================================")
	l.Printf("MONITOR %s Starting\n", m.connector)
	sp, err := spool.NewSpool(m.spoolDir)
	if err != nil {
		return err
	}
	for {
		lock.Lock()
		// get new list of files
		newList, err = sp.GetSpooledGobsList()
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
			} else {
				// exists in old, do nothing
			}
		}
		l.Printf("MONITOR %s: Sending newFiles list: %#v\n", m.connector, newFiles)
		// send new-found files
		ch <- newFiles
		// oldlist=newlist
		oldList = newList
		// empty newfiles for the next iteration
		newFiles = &spool.SpooledGobs{}

		//l.Printf("MONITOR %s: Sleeping.\n", m.connector)
		//time.Sleep(5 * time.Second)
		//l.Printf("Time: %s\n", <-ticker)
		<-ticker
	}
	l.Printf("Exiting monitor routine %s\n", m.spoolDir)
	l.Println("======================= Monitor end ============================================")
	return nil
}
