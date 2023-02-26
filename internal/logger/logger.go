package logger

import (
	"log"
	"os"
)

func SetupLogger(logfile string, app string) (*log.Logger, error) {
	var (
		l   *log.Logger
		lf  *os.File
		err error
	)

	if logfile != "" {
		// https://github.com/CLIP-HPC/goslmailer/issues/31
		// e.g. If log file is not writable, this returns err & causes panic in main.
		// We'll still return err, but also back off to stderr and log it.
		lf, err = os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			lf = os.Stderr
			//return nil, errors.New("can not open configured log file")
		}
	} else {
		lf = os.Stderr
	}
	l = log.New(lf, app+":", log.Lshortfile|log.Ldate|log.Lmicroseconds)

	return l, err

}
