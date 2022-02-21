package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pja237/goslmailer/internal/config"
	"github.com/pja237/goslmailer/internal/connectors"
	"github.com/pja237/goslmailer/internal/message"
	"github.com/pja237/goslmailer/internal/spool"
)

var lock sync.Mutex

type MsgList []message.MessagePack

type monitor struct {
	connector string
	spoolDir  string
}

type picker struct {
	connector string
	msgcount  map[string]int
}

type sender struct {
	connector string
	spoolDir  string
	conn      connectors.Connector
}

func NewPicker(c string) (*picker, error) {
	var p picker

	p.connector = c
	p.msgcount = map[string]int{}

	return &p, nil
}

func NewMonitor(c string, s string) (*monitor, error) {
	var m monitor

	if s != "" {
		m.connector = c
		m.spoolDir = s
	} else {
		return nil, errors.New("no spooldir, aborting")
	}

	return &m, nil
}

func NewSender(c string, sd string, cons *connectors.Connectors) (*sender, error) {
	var s sender

	s.connector = c     // connector name
	s.spoolDir = sd     // connector spooldir
	s.conn = (*cons)[c] // connector interface

	return &s, nil
}

func (s *sender) SenderWorker(psCh <-chan *spool.FileGob, psfCh chan<- *spool.FileGob, wg *sync.WaitGroup, l *log.Logger) error {

	defer wg.Done()

	l.Println("======================= Sender start ===========================================")
	for {
		msg := <-psCh
		l.Printf("SENDER %s: received %#v\n", s.connector, msg)

		// fetch gob mp
		// todo: error handling here needs more attention!
		sd, err := spool.NewSpool(s.spoolDir)
		if err != nil {
			l.Printf("SENDER %s: newspool returned error %s\n", s.connector, err)
			continue
		}
		mp, err := sd.FetchGob(msg.Filename)
		if err != nil {
			l.Printf("SENDER %s: fetchgob returned error %s\n", s.connector, err)
			continue
		}
		// useSpool == false when called from here, gob is already on disk!
		err = s.conn.SendMessage(mp, false, l)
		if err != nil {
			l.Printf("SENDER %s: connector.sendmessage() returned error %s\n", s.connector, err)
			// send it back to picker
			psfCh <- msg
		} else {
			// Send succeeded, delete gob
			lock.Lock()
			err = os.Remove(s.spoolDir + "/" + msg.Filename)
			if err != nil {
				l.Printf("SENDER %s: error removing file %s\n", s.connector, err)
			}
			lock.Unlock()
			l.Printf("SENDER %s: Gob deleted\n", s.connector)
		}
	}
	l.Println("======================= Sender end =============================================")
	return nil
}

func (p *picker) PickerWorker(mpCh <-chan *spool.SpooledGobs, psCh chan<- *spool.FileGob, psfCh <-chan *spool.FileGob, wg *sync.WaitGroup, l *log.Logger) error {

	var newgobs *spool.SpooledGobs
	var allGobs = spool.SpooledGobs{}

	defer wg.Done()

	l.Println("======================= Picker start ===========================================")
	// configurable picker/sender frequency
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case newgobs = <-mpCh:
			l.Printf("PICKER %s: Received gobs %#v\n", p.connector, newgobs)
			// iterate and increase the counter
			for k, v := range *newgobs {
				p.msgcount[v.User]++
				// append newgobs to allgobs
				allGobs[k] = v
			}
		case failedGob := <-psfCh:
			l.Printf("PICKER %s: Received FAILED gob %#v\n", p.connector, failedGob)
			// return to allGobs
			allGobs[failedGob.Filename] = *failedGob
		default:
			l.Printf("PICKER %s: allGobs: %#v\n", p.connector, allGobs)
			// HERE, call the Pick() and Send()
			for k, v := range allGobs {
				// pick first, send and delete
				l.Printf("PICK %s: SEND to Sender: %#v\n", p.connector, v)
				psCh <- &v
				delete(allGobs, k)
				break
			}
		}
		<-ticker
	}

	l.Println("======================= Picker end =============================================")
	return nil
}

func (m *monitor) MonitorWorker(ch chan<- *spool.SpooledGobs, wg *sync.WaitGroup, l *log.Logger) error {

	var oldList, newList, newFiles *spool.SpooledGobs
	oldList = &spool.SpooledGobs{}
	newFiles = &spool.SpooledGobs{}

	defer wg.Done()
	// configurable monitor timer
	ticker := time.Tick(10 * time.Second)

	l.Println("======================= Monitor start ==========================================")
	l.Printf("MON %s Starting\n", m.connector)
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
			l.Printf("MON %s: Failed on Getspooledgobslist(), error %s\n", m.connector, err)
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
		l.Printf("MON %s: Sending newFiles list: %#v\n", m.connector, newFiles)
		// send new-found files
		ch <- newFiles
		// oldlist=newlist
		oldList = newList
		// empty newfiles for the next iteration
		newFiles = &spool.SpooledGobs{}

		l.Printf("MON %s: Sleeping.\n", m.connector)
		//time.Sleep(5 * time.Second)
		//l.Printf("Time: %s\n", <-ticker)
		<-ticker
	}
	l.Printf("Exiting monitor routine %s\n", m.spoolDir)
	l.Println("======================= Monitor end ============================================")
	return nil
}

func main() {

	var (
		conns   = make(connectors.Connectors)
		logFile io.Writer
		wg      sync.WaitGroup
	)

	// read configuration
	// how to handle hardcoding config file?
	cfg := config.NewConfigContainer()
	err := cfg.GetConfig("/etc/slurm/goslmailer.conf")
	if err != nil {
		fmt.Printf("getConfig failed: %s", err)
		os.Exit(1)
	}

	// setup logger
	if cfg.Logfile != "" {
		logFile, err = os.OpenFile(cfg.Logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("ERROR: can not open configured log file. Exiting.")
			os.Exit(1)
		}
	} else {
		logFile = os.Stderr
	}
	log := log.New(logFile, "gobler:", log.Lshortfile|log.Ldate|log.Lmicroseconds)

	log.Println("======================= Gobler start ===========================================")
	cfg.DumpConfig(log)

	// populate map with configured referenced connectors
	conns.PopulateConnectors(cfg, log)

	// iterate and spin up monitor,picker and sender routines
	for con := range cfg.Connectors {
		spd, ok := cfg.Connectors[con]["spoolDir"]
		if ok {
			log.Printf("spoolDir exists: %s - %s\n", cfg.Connectors[con]["spoolDir"], spd)
			mpChan := make(chan *spool.SpooledGobs, 1)
			// configurable buffer size
			psChan := make(chan *spool.FileGob, 1)
			psChanFailed := make(chan *spool.FileGob, 1)
			mon, err := NewMonitor(con, spd)
			if err != nil {
				log.Println("Monitor inst failed")
			} else {
				log.Println("Monitor startup...")
				wg.Add(1)
				go mon.MonitorWorker(mpChan, &wg, log)
				log.Println("Monitor exit...")
			}
			pickr, err := NewPicker(con)
			if err != nil {
				log.Println("Picker inst failed")
			} else {
				log.Println("Picker startup...")
				wg.Add(1)
				go pickr.PickerWorker(mpChan, psChan, psChanFailed, &wg, log)
				log.Println("Monitor exit...")
			}
			sendr, err := NewSender(con, cfg.Connectors[con]["spoolDir"], &conns)
			if err != nil {
				log.Println("Sender inst failed")
			} else {
				log.Println("Sender startup...")
				wg.Add(1)
				go sendr.SenderWorker(psChan, psChanFailed, &wg, log)
				log.Println("Sender exit...")
			}

		} else {
			log.Printf("connector %s doesn't have spoolDir defined\n", con)
		}
	}

	log.Printf("Waiting for routines to finish...\n")
	wg.Wait()
	log.Printf("All routines finished, exiting main\n")

	log.Println("======================= Gobler end =============================================")
}
