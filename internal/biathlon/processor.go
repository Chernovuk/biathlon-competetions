package biathlon

import (
	"time"
)

type EventHandler func(e Event) error

type Processor struct {
	events      <-chan Event
	eventsQueue []Event
	competitors map[int]Competitor

	fsm FiniteStateMachine

	handlers map[eventType]EventHandler

	config Config
}

func NewProcessor(conf Config, events <-chan Event) *Processor {
	return &Processor{
		events:      events,
		competitors: make(map[int]Competitor),
		fsm:         FSM(),
		config:      conf,
	}
}
