package biathlon

import (
	"log"
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
		handlers:    make(map[eventType]EventHandler),
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
		logEvent(e)
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
		startTime := e.ExtraParams[0].(time.Time)
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

func logEvent(e Event) {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	ts := e.TimeStamp.Format("15:04:05.000")
	switch e.Type {
	case Register:
		log.Printf("[%s] The competitor(%d) registered", ts, e.CompetitorID)

	case BeSheduled:
		t := e.ExtraParams[0].(time.Time)
		log.Printf("[%s] The start time for the competitor(%d) was set by a draw to %s",
			ts, e.CompetitorID, t.Format("15:04:05.000"))

	case ComeToStartLine:
		log.Printf("[%s] The competitor(%d) is on the start line", ts, e.CompetitorID)

	case Start:
		log.Printf("[%s] The competitor(%d) has started", ts, e.CompetitorID)

	case ComeToFiringRange:
		line := e.ExtraParams[0].(int)
		log.Printf("[%s] The competitor(%d) is on the firing range(%d)", ts, e.CompetitorID, line)

	case HitTarget:
		target := e.ExtraParams[0].(int)
		log.Printf("[%s] The target(%d) has been hit by competitor(%d)", ts, target, e.CompetitorID)

	case LeaveFiringRange:
		log.Printf("[%s] The competitor(%d) left the firing range", ts, e.CompetitorID)

	case EnterPenaltyLap:
		log.Printf("[%s] The competitor(%d) entered the penalty laps", ts, e.CompetitorID)

	case LeavePenaltyLap:
		log.Printf("[%s] The competitor(%d) left the penalty laps", ts, e.CompetitorID)

	case EndMainLap:
		log.Printf("[%s] The competitor(%d) ended the main lap", ts, e.CompetitorID)

	case BeUnableToContinue:
		comment := e.ExtraParams[0].(string)
		log.Printf("[%s] The competitor(%d) can`t continue: %s", ts, e.CompetitorID, comment)

	case Disqualify:
		log.Printf("[%s] The competitor(%d) is disqualified", ts, e.CompetitorID)

	case Finish:
		log.Printf("[%s] The competitor(%d) has finished", ts, e.CompetitorID)

	default:
		log.Printf("[%s] Unknown event(%d) for competitor(%d)", ts, e.Type, e.CompetitorID)
	}
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
