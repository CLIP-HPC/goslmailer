package main

import (
	"errors"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/pja237/goslmailer/internal/spool"
)

// picker holds connector name string, msgcount map with {username:SpooledGobsCount} and pickerT polling period
type picker struct {
	connector    string
	spoolDir     string
	msgcount     map[string]int    // map holding {username:SpooledGobsCount}
	deletedcount map[string]uint32 // map holding {username:DeletedGobsCount}
	pickerT      time.Duration
	maxMsgPU     int
}

// NewPicker creates and initializes a new picker object with:
// c - connector name, s - spool directory, t - polling time period and mpu - MaxMessagesPerUser threshold.
// Holds internal counter map msgcount which counts number of messages currently in system from every user.
func NewPicker(c string, s string, t time.Duration, mpu int) (*picker, error) {
	var p picker

	p.connector = c
	p.spoolDir = s
	p.msgcount = map[string]int{}
	p.deletedcount = map[string]uint32{}
	p.pickerT = t
	p.maxMsgPU = mpu

	return &p, nil
}

// PickNext() is called every picker.pickerT time period to return next message to be sent to sender.
// We pick the oldest message in the queue to return and send.
func (p *picker) PickNext(allGobs *spool.SpooledGobs, l *log.Logger) (*spool.FileGob, error) {

	var sortedGobs []spool.FileGob
	var nextgob spool.FileGob

	if len(*allGobs) == 0 {
		return nil, errors.New("no gobs in spool")
	}

	for _, fg := range *allGobs {
		sortedGobs = append(sortedGobs, fg)
	}

	sortGobsTimeFwd(sortedGobs)
	nextgob = sortedGobs[0]
	l.Printf("PICKER %s: picked gob: %q\n", p.connector, nextgob)

	return &nextgob, nil
}

func sortGobsTimeFwd(toSort []spool.FileGob) {
	sort.Slice(toSort, func(i, j int) bool {
		if toSort[i].TimeStamp.Before(toSort[j].TimeStamp) {
			return true
		} else {
			return false
		}
	})
}

func sortGobsTimeReverse(toSort []spool.FileGob) {
	sort.Slice(toSort, func(i, j int) bool {
		if toSort[i].TimeStamp.After(toSort[j].TimeStamp) {
			return true
		} else {
			return false
		}
	})
}

// trimExcessiveMsgs is called after new gobs list is received from monitor.
// Deletes all gobs for every user above the MaxMsgPerUser limit.
// Modifies count of deleted gobs in picker.deletedcount["username"] and picker.messagecount["username"].
func (p *picker) trimExcessiveMsgs(allGobs *spool.SpooledGobs, mpu int, l *log.Logger) error {

	var sortedGobs []spool.FileGob

	for _, fg := range *allGobs {
		sortedGobs = append(sortedGobs, fg)
	}
	//l.Printf("PICKER %s: allgobs: %#v\n", p.connector, allGobs)

	sortGobsTimeReverse(sortedGobs)

	for _, v := range sortedGobs {
		if p.msgcount[v.User] > mpu {
			lock.Lock()
			err := os.Remove(p.spoolDir + "/" + v.Filename)
			if err != nil {
				l.Printf("PICKER %s: error removing file %s\n", p.connector, err)
			} else {
				l.Printf("PICKER %s: Gob %s deleted\n", p.connector, v.Filename)
				p.deletedcount[v.User]++
				p.msgcount[v.User]--
				delete(*allGobs, v.Filename)
			}
			lock.Unlock()
		}
	}

	// Old way, unsorted random deletion. Now we sort before, and delete later messages that are overflowing. Remove later.
	//for f, fg := range *allGobs {
	//	if p.msgcount[fg.User] > mpu {
	//		lock.Lock()
	//		err := os.Remove(p.spoolDir + "/" + fg.Filename)
	//		if err != nil {
	//			l.Printf("PICKER %s: error removing file %s\n", p.connector, err)
	//		} else {
	//			l.Printf("PICKER %s: Gob %s deleted\n", p.connector, f)
	//			p.deletedcount[fg.User]++
	//			p.msgcount[fg.User]--
	//			delete(*allGobs, f)
	//		}
	//		lock.Unlock()
	//	}
	//}
	return nil
}

func (p *picker) PickerWorker(mpCh <-chan *spool.SpooledGobs, psCh chan<- psGob, psfCh <-chan psGob, wg *sync.WaitGroup, l *log.Logger) error {

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
			// iterate and increase the counter of messages per user
			for k, v := range *newgobs {
				p.msgcount[v.User]++
				// append newgobs to allgobs
				allGobs[k] = v
			}
			p.trimExcessiveMsgs(&allGobs, p.maxMsgPU, l)
		case failedGob := <-psfCh:
			l.Printf("PICKER %s: Received FAILED gob %#v\n", p.connector, failedGob)
			// return to allGobs, modify counters accordingly
			allGobs[failedGob.fileGob.Filename] = *failedGob.fileGob
			p.msgcount[failedGob.fileGob.User]++
			p.deletedcount[failedGob.fileGob.User] = failedGob.deletedCount
		case <-ticker:
			l.Printf("PICKER %s on TICK: Chan status: mpCh = %d msgs psCh = %d / %d msgs, psfCh = %d / %d msgs \n", p.connector, len(mpCh), len(psCh), cap(psCh), len(psfCh), cap(psfCh))
			if len(psCh) < cap(psCh) {
				// HERE, call the Pick() and Send()
				nextGob, err := p.PickNext(&allGobs, l)
				if err == nil {
					// package gob and deletedcounter in psGob to be sent to sender routine
					psG := psGob{
						fileGob:      nextGob,
						deletedCount: p.deletedcount[nextGob.User],
					}
					l.Printf("PICKER %s: SEND to Sender: %#v\n", p.connector, psG)
					p.msgcount[nextGob.User]--
					psCh <- psG
					delete(allGobs, nextGob.Filename)
					// we have sent the warning message, now we "zero" it
					delete(p.deletedcount, nextGob.User)
				} else {
					l.Printf("PICKER %s: %s\n", p.connector, err)
				}
			} else {
				l.Printf("psCh FULL!\n")
			}
		}
	}

	l.Println("======================= Picker end =============================================")
	return nil
}
