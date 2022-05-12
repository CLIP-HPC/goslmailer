package logger

import (
	"errors"
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
		lf, err = os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, errors.New("can not open configured log file")
		}
	} else {
		lf = os.Stderr
	}
	l = log.New(lf, app+":", log.Lshortfile|log.Ldate|log.Lmicroseconds)

	return l, nil

}
