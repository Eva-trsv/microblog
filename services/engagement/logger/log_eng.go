package logger

import (
	"log/slog"
	"os"
	"sync"
)

type Event struct {
	Message string
	Fields  map[string]any
}

type Logger struct {
	ch      chan Event
	slogger *slog.Logger
	wg      sync.WaitGroup
}

func NewLogger(buffer int) *Logger {
	handler := slog.NewTextHandler(os.Stdout, nil)

	log := &Logger{
		ch:      make(chan Event, buffer),
		slogger: slog.New(handler),
	}

	log.wg.Add(1)
	go log.listen()

	return log
}

func (l *Logger) listen() {
	defer l.wg.Done()

	for ev := range l.ch {
		if ev.Fields != nil {
			l.slogger.Info(ev.Message, ev.mapToArgs()...)
		} else {
			l.slogger.Info(ev.Message)
		}
	}
}

func (l *Logger) Log(message string, fields map[string]any) {
	copied := make(map[string]any)
	for k, v := range fields {
		copied[k] = v
	}

	l.ch <- Event{
		Message: message,
		Fields:  copied,
	}

}

func (l *Logger) Close() {
	close(l.ch)
	l.wg.Wait()
}

func (e Event) mapToArgs() []any {
	args := make([]any, 0, len(e.Fields)*2)
	for k, v := range e.Fields {
		args = append(args, k, v)
	}
	return args
}
