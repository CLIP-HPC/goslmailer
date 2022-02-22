package main

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/pja237/goslmailer/internal/connectors"
	"github.com/pja237/goslmailer/internal/spool"
)

type conMon struct {
	conn     string
	spoolDir string
	monitorT time.Duration
	pickerT  time.Duration
}

const (
	monitorTdefault = 10
	pickerTdefault  = 2
)

func NewConMon(con string, conCfg map[string]string) (*conMon, error) {
	var cm conMon

	cm.conn = con
	cm.spoolDir = conCfg["spoolDir"]
	if e, ok := conCfg["monitorT"]; ok {
		T, err := strconv.ParseUint(e, 10, 64)
		if err != nil {
			return nil, err
		}
		cm.monitorT = time.Duration(T)
	} else {
		cm.monitorT = time.Duration(monitorTdefault)
	}
	if e, ok := conCfg["pickerT"]; ok {
		T, err := strconv.ParseUint(e, 10, 64)
		if err != nil {
			return nil, err
		}
		cm.pickerT = time.Duration(T)
	} else {
		cm.pickerT = time.Duration(pickerTdefault)
	}

	return &cm, nil
}

// SpinUp start 3 goroutines: monitor, picker and sender for a connector (has "spoolDir" attribute in .conf)
func (cm *conMon) SpinUp(conns connectors.Connectors, wg *sync.WaitGroup, log *log.Logger) error {

	mpChan := make(chan *spool.SpooledGobs, 1)
	// make configurable buffer size
	psChan := make(chan *spool.FileGob, 1)
	psChanFailed := make(chan *spool.FileGob, 1)

	// spin-up
	mon, err := NewMonitor(cm.conn, cm.spoolDir, cm.monitorT)
	if err != nil {
		log.Printf("Monitor %s inst FAILED\n", cm.conn)
	} else {
		log.Printf("Monitor %s startup...\n", cm.conn)
		wg.Add(1)
		go mon.MonitorWorker(mpChan, wg, log)
	}
	pickr, err := NewPicker(cm.conn, cm.pickerT)
	if err != nil {
		log.Printf("Picker %s inst FAILED\n", cm.conn)
	} else {
		log.Printf("Picker %s startup...\n", cm.conn)
		wg.Add(1)
		go pickr.PickerWorker(mpChan, psChan, psChanFailed, wg, log)
	}
	sendr, err := NewSender(cm.conn, cm.spoolDir, &conns)
	if err != nil {
		log.Printf("Sender %s inst failed\n", cm.conn)
	} else {
		log.Printf("Sender %s startup...\n", cm.conn)
		wg.Add(1)
		go sendr.SenderWorker(psChan, psChanFailed, wg, log)
		log.Println("Sender exit...")
	}

	return nil
}
