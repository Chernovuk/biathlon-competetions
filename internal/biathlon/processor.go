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

func (p *Processor) Start() {
	for {
		if len(p.eventsQueue) == 0 {
			e, ok := <-p.events
			if !ok {
				break
			}
			p.eventsQueue = append(p.eventsQueue, e)
		}
		e := p.eventsQueue[0]
		p.processEvent(e)
		p.eventsQueue = p.eventsQueue[1:]
	}
}

func (p *Processor) Handle(e eventType, handler EventHandler) {
	if _, ok := p.handlers[e]; ok {
		panic("da")
	}
	p.handlers[e] = handler
}

func (p *Processor) processEvent(e Event) {
	cID := e.CompetitorID
	competitor := p.competitors[cID]

	status, ok := p.fsm.LookupPath(competitor.Status, e.Type)
	if !ok {
		// panic("aaaaaa")
		return
	}

	switch e.Type {
	case BeSheduled:
		startTime, ok := e.ExtraParams[0].(time.Time)
		if !ok {
		}
		competitor.ScheduledStartTime = startTime
	case Start:
		competitor.ActualStartTime = e.TimeStamp

		start := competitor.ActualStartTime
		threshold := competitor.ScheduledStartTime.Add(time.Duration(p.config.StartDelta))

		if start.Before(threshold) {
			competitor.CurrentLap = 1
		} else {
			disqualify := Event{TimeStamp: e.TimeStamp, Type: Disqualify, CompetitorID: cID}
			p.eventsQueue = append(p.eventsQueue, disqualify)
		}
	case EndMainLap:
		if competitor.CurrentLap < p.config.Laps {
			competitor.CurrentLap++
		} else {
			finish := Event{TimeStamp: e.TimeStamp, Type: Finish, CompetitorID: cID}
			p.eventsQueue = append(p.eventsQueue, finish)
		}
	}

	handler, ok := p.handlers[e.Type]
	if ok {
		var err error
		err = handler(e)
		if err != nil {
		}
	}
	competitor.Status = status

	p.competitors[cID] = competitor
}

// func (p *Processor) trigger(c *Competitor, e Event) error {
// 	handler, ok := p.handlers[e.Type]
// 	if !ok {
// 	}

// 	var err error
// 	err = handler(e)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (p *Processor) processCompetitorState() {
// }
