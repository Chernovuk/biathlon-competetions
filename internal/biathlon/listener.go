package biathlon

import (
	"bufio"
	"log"
	"os"
)

type EventListener struct {
	events chan Event
	file   *os.File

	log *log.Logger
}

func NewEventListener(filepath string) (*EventListener, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return &EventListener{}, err
	}

	return &EventListener{
		events: make(chan Event),
		file:   f, log: log.New(os.Stdout, "Listener: ", log.Ltime),
	}, nil
}

func (l *EventListener) Events() <-chan Event {
	return l.events
}

func (l *EventListener) SetLogger(log *log.Logger) {
	l.log = log
}

func (l *EventListener) Start() {
	scanner := bufio.NewScanner(l.file)
	for scanner.Scan() {
		event, err := ParseEvent(scanner.Text())
		if err != nil {
			l.log.Println(err)
		}

		l.events <- event
	}
	close(l.events)

	if err := l.file.Close(); err != nil {
		l.log.Fatalln(err)
	}
}
