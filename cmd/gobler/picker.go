package main

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pja237/goslmailer/internal/spool"
)

// picker holds connector name string, msgcount map with {username:SpooledGobsCount} and pickerT polling period
type picker struct {
	connector string
	spoolDir  string
	msgcount  map[string]int // map holding {username:SpooledGobsCount}
	pickerT   time.Duration
	maxMsgPU  int
}

// NewPicker creates and initializes a new picker object with:
// c connector name and t polling time period parameters.
func NewPicker(c string, s string, t time.Duration, mpu int) (*picker, error) {
	var p picker

	p.connector = c
	p.spoolDir = s
	p.msgcount = map[string]int{}
	p.pickerT = t
	p.maxMsgPU = mpu

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

func (p *picker) trimExcessiveMsgs(allGobs *spool.SpooledGobs, mpu int, l *log.Logger) error {
	for f, fg := range *allGobs {
		if p.msgcount[fg.User] > mpu {
			lock.Lock()
			err := os.Remove(p.spoolDir + "/" + fg.Filename)
			if err != nil {
				l.Printf("PICKER %s: error removing file %s\n", p.connector, err)
			} else {
				l.Printf("PICKER %s: Gob %s deleted\n", p.connector, f)
				p.msgcount[fg.User]--
				delete(*allGobs, f)
			}
			lock.Unlock()
		}
	}
	return nil
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
			l.Printf("PICKER %s: Received %d gobs.\n", p.connector, len(*newgobs))
			// iterate and increase the counter
			for k, v := range *newgobs {
				p.msgcount[v.User]++
				// append newgobs to allgobs
				allGobs[k] = v
			}
			// todo: or call trimExcessiveMsgs here?
			p.trimExcessiveMsgs(&allGobs, p.maxMsgPU, l)
		case failedGob := <-psfCh:
			l.Printf("PICKER %s: Received FAILED gob %#v\n", p.connector, failedGob)
			// return to allGobs
			allGobs[failedGob.Filename] = *failedGob
			p.msgcount[failedGob.User]++
		case <-ticker:
			l.Printf("PICKER %s on TICK: Chan status: mpCh = %d msgs psCh = %d / %d msgs, psfCh = %d / %d msgs \n", p.connector, len(mpCh), len(psCh), cap(psCh), len(psfCh), cap(psfCh))
			if len(psCh) < cap(psCh) {
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
	}

	l.Println("======================= Picker end =============================================")
	return nil
}
