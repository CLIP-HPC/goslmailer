package main

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/pja237/goslmailer/internal/spool"
)

// picker holds connector name string, msgcount map with {username:SpooledGobsCount} and pickerT polling period
type picker struct {
	connector string
	msgcount  map[string]int // map holding {username:SpooledGobsCount}
	pickerT   time.Duration
}

// NewPicker creates and initializes a new picker object with:
// c connector name and t polling time period parameters.
func NewPicker(c string, t time.Duration) (*picker, error) {
	var p picker

	p.connector = c
	p.msgcount = map[string]int{}
	p.pickerT = t

	return &p, nil
}

func (p *picker) PickNext(allGobs *spool.SpooledGobs) (*spool.FileGob, error) {

	var nextgob spool.FileGob

	if len(*allGobs) == 0 {
		return nil, errors.New("no gobs in spool")
	}

	// todo: here implement something meaningful
	for _, v := range *allGobs {
		nextgob = v
		break
	}

	return &nextgob, nil
}

func (p *picker) PickerWorker(mpCh <-chan *spool.SpooledGobs, psCh chan<- *spool.FileGob, psfCh <-chan *spool.FileGob, wg *sync.WaitGroup, l *log.Logger) error {

	var newgobs *spool.SpooledGobs
	var allGobs = spool.SpooledGobs{}

	defer wg.Done()

	l.Println("======================= Picker start ===========================================")
	ticker := time.Tick(p.pickerT)
	for {
		l.Printf("PICKER %s: Users msg count %v\n", p.connector, p.msgcount)
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
			p.msgcount[failedGob.User]++
		case <-ticker:
			l.Printf("PICKER %s: Chan status: mpCh = %d msgs psCh = %d / %d msgs, psfCh = %d / %d msgs \n", p.connector, len(mpCh), len(psCh), cap(psCh), len(psfCh), cap(psfCh))
		}
		if len(psCh) < cap(psCh) {
			//l.Printf("PICKER %s: allGobs: %#v\n", p.connector, allGobs)
			// HERE, call the Pick() and Send()
			nextGob, err := p.PickNext(&allGobs)
			if err == nil {
				l.Printf("PICKER %s: SEND to Sender: %#v\n", p.connector, nextGob)
				p.msgcount[nextGob.User]--
				psCh <- nextGob
				delete(allGobs, nextGob.Filename)
			}
		} else {
			l.Printf("psCh FULL!\n")
		}
	}

	l.Println("======================= Picker end =============================================")
	return nil
}
