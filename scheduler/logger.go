package scheduler

import (
	"fmt"
	"io"
	"os"
)

type Logger interface {
	Info(msg string, v ...any)
	Error(err error, msg string, v ...any)
	EntryAdded(*Entry)
	EntryRemoved(*Entry)
	EntryUpdated(*Entry)
}

func PrintlnLogger() Logger {
	return &printlnLogger{infoW: os.Stdout, errW: os.Stderr}
}

type printlnLogger struct {
	infoW io.Writer
	errW  io.Writer
}

func (p *printlnLogger) Info(msg string, v ...any) {
	m := msg
	for i := 0; i < len(v); i += 2 {
		m = m + fmt.Sprintf("  %v: %v", v[i], v[i+1])
	}
	fmt.Fprintln(p.infoW, m)
}

func (p *printlnLogger) Error(err error, msg string, v ...any) {
	fmt.Fprintln(p.errW, err.Error())
	if msg != "" {
		m := msg
		for i := 0; i < len(v); i += 2 {
			m = m + fmt.Sprintf("  %v: %v", v[i], v[i+1])
		}
		fmt.Fprintln(p.infoW, m)
	}
}

func (p *printlnLogger) EntryAdded(*Entry) {}

func (p *printlnLogger) EntryRemoved(*Entry) {}

func (p *printlnLogger) EntryUpdated(*Entry) {}
