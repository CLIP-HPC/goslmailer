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
	maxMsgPU  int
}

// NewMonitor creates and initializes a new monitor object with:
// c connector name, s spooldir location (from config file), t polling time period and mpu is MaximumMessagesPerUser (from conf.)
func NewMonitor(c string, s string, t time.Duration, mpu int) (*monitor, error) {
	var m monitor

	if s != "" {
		m.connector = c
		m.spoolDir = s
		m.monitorT = t
		m.maxMsgPU = mpu
	} else {
		return nil, errors.New("no spooldir, aborting")
	}

	return &m, nil
}

// if trimming logic works ok from picker, remove this and also remove maxMsgPU from monitor struct...
//
//func (m *monitor) trimExcessiveMsgs(newFiles *spool.SpooledGobs, mpu int, l *log.Logger) error {
//	uc := make(map[string]int)
//
//	for f, fg := range *newFiles {
//		uc[fg.User]++
//		if uc[fg.User] > mpu {
//			lock.Lock()
//			err := os.Remove(m.spoolDir + "/" + fg.Filename)
//			if err != nil {
//				l.Printf("MONITOR %s: error removing file %s\n", m.connector, err)
//			}
//			lock.Unlock()
//			l.Printf("MONITOR %s: Gob %s deleted\n", m.connector, f)
//			uc[fg.User]--
//			delete(*newFiles, f)
//		}
//	}
//	l.Printf("UserCount map == %#v\n", uc)
//	return nil
//}

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
			} else {
				// exists in old, do nothing
			}
		}

		// todo: decide if here we do the purge of newFiles for messages above maxMsgPU, or in picker?
		//m.trimExcessiveMsgs(newFiles, m.maxMsgPU, l)

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
