package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/connectors"
	"github.com/CLIP-HPC/goslmailer/internal/spool"
)

type conMon struct {
	conn           string
	spoolDir       string
	monitorT       time.Duration
	pickerT        time.Duration
	pickSendBufLen int
	numSenders     int
	maxMsgPU       int
}

const (
	monitorTdefault   = 10
	pickerTdefault    = 2
	psBufLenDefault   = 1
	numSendersDefault = 1
	maxMsgPUDefault   = 10
)

type psGob struct {
	fileGob      *spool.FileGob
	deletedCount uint32
}

// getConfTime converts config string to time.Duration value.
// If string is suffixed with "ms", return miliseconds, else seconds.
func getConfTime(e string) (time.Duration, error) {
	var milis bool = false

	if ms := strings.TrimSuffix(e, "ms"); ms != e {
		milis = true
		e = ms
	}

	T, err := strconv.ParseUint(e, 10, 64)
	if err != nil {
		return -1 * time.Second, errors.New("problem converting time from config to uint")
	}

	if milis {
		return time.Duration(T) * time.Millisecond, nil
	} else {
		return time.Duration(T) * time.Second, nil
	}
}

func NewConMon(con string, conCfg map[string]string, l *log.Logger) (*conMon, error) {
	var (
		cm  conMon
		err error
	)

	cm.conn = con
	cm.spoolDir = conCfg["spoolDir"]

	psbl, err := strconv.Atoi(conCfg["psBufLen"])
	if err != nil {
		// return nil, errors.New("psBufLen is not integer")
		// todo: no need to be so agressive, let's do default... or should we abort so the user knows he made a mistake?
		cm.pickSendBufLen = psBufLenDefault
	} else {
		cm.pickSendBufLen = psbl
	}

	ns, err := strconv.Atoi(conCfg["numSenders"])
	if err != nil {
		//return nil, errors.New("numSenders is not integer")
		cm.numSenders = numSendersDefault
	} else {
		cm.numSenders = ns
	}

	mpu, err := strconv.Atoi(conCfg["maxMsgPU"])
	if err != nil {
		//return nil, errors.New("maxNewMsgPU is not integer")
		cm.maxMsgPU = maxMsgPUDefault
	} else {
		cm.maxMsgPU = mpu
	}

	// if monitorT is specified...
	if e, ok := conCfg["monitorT"]; ok {
		cm.monitorT, err = getConfTime(e)
		if err != nil {
			return nil, err
		}
	} else {
		// nothing specified in config, use default seconds
		cm.monitorT = time.Duration(monitorTdefault) * time.Second
	}

	// if pickerT is specified...
	if e, ok := conCfg["pickerT"]; ok {
		cm.pickerT, err = getConfTime(e)
		if err != nil {
			return nil, err
		}
	} else {
		// nothing specified in config, use default seconds
		cm.pickerT = time.Duration(pickerTdefault) * time.Second
	}
	l.Printf("CM setup: %#v\n", cm)
	return &cm, nil
}

// SpinUp start 3 goroutines: monitor, picker and sender for a connector (each that has "spoolDir" attribute in .conf)
func (cm *conMon) SpinUp(conns connectors.Connectors, wg *sync.WaitGroup, l *log.Logger) error {

	mpChan := make(chan *spool.SpooledGobs, 1)
	psChan := make(chan psGob, cm.pickSendBufLen)
	psChanFailed := make(chan psGob, cm.pickSendBufLen)

	mon, err := NewMonitor(cm.conn, cm.spoolDir, cm.monitorT)
	if err != nil {
		l.Printf("Monitor %s inst FAILED\n", cm.conn)
	} else {
		l.Printf("Monitor %s startup...\n", cm.conn)
		wg.Add(1)
		go mon.MonitorWorker(mpChan, wg, l)
	}

	pickr, err := NewPicker(cm.conn, cm.spoolDir, cm.pickerT, cm.maxMsgPU)
	if err != nil {
		l.Printf("Picker %s inst FAILED\n", cm.conn)
	} else {
		l.Printf("Picker %s startup...\n", cm.conn)
		wg.Add(1)
		go pickr.PickerWorker(mpChan, psChan, psChanFailed, wg, l)
	}

	for i := 1; i <= cm.numSenders; i++ {
		sendr, err := NewSender(cm.conn, cm.spoolDir, &conns, i)
		if err != nil {
			l.Printf("Sender %d - %s inst failed\n", i, cm.conn)
		} else {
			l.Printf("Sender %d - %s startup...\n", i, cm.conn)
			wg.Add(1)
			go sendr.SenderWorker(psChan, psChanFailed, wg, l)
		}
	}

	return nil
}
