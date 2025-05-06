package biathlon

import (
	"bufio"
	"log"
	"os"
)

type EventListener struct {
	events chan Event
	file   *os.File
}

func NewEventListener(filepath string) (EventListener, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return EventListener{}, err
	}

	return EventListener{
		events: make(chan Event),
		file:   f,
	}, nil
}

func (l EventListener) Events() <-chan Event {
	return l.events
}

func (l EventListener) Start() {
	scanner := bufio.NewScanner(l.file)
	for scanner.Scan() {
		event, err := ParseEvent(scanner.Text())
		if err != nil {
			log.Printf("Trash: %v", err)
		}

		l.events <- event
	}
	close(l.events)

	if err := l.file.Close(); err != nil {
		log.Fatalf("%v", err)
	}
}
