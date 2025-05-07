package biathlon

import (
	"errors"
	"log"
	"time"
)

var ErrWrongEventsSequence = errors.New("impossible sequence of events")

type EventHandler func(e Event)

type Processor struct {
	events      <-chan Event
	eventsQueue []Event
	competitors map[int]Competitor

	fsm FiniteStateMachine

	handlers map[eventType]EventHandler

	config Config
}

func NewProcessor(conf Config, events <-chan Event) *Processor {
	fsm = makeFSM(
		[]edge{
			{src: Unknown, dst: Registered, event: Register},

			{
				src:   Registered,
				dst:   Scheduled,
				event: BeSheduled,
				cb: func(e Event, c *Competitor) ([]Event, error) {
					startTime := e.ExtraParams[0].(time.Time)
					c.ScheduledStartTime = startTime

					return []Event{}, nil
				},
			},
			{src: Registered, dst: NotStarted, event: Disqualify},
			{src: Registered, dst: CannotContinue, event: BeUnableToContinue}, // ???

			{
				src:   Scheduled,
				dst:   OnStartLine,
				event: ComeToStartLine,
				cb: func(e Event, c *Competitor) ([]Event, error) {
					scheduledTime := c.ScheduledStartTime
					threshold := scheduledTime.Add(time.Duration(conf.StartDelta))
					if e.TimeStamp.After(threshold) {
						disqualify := Event{
							TimeStamp: e.TimeStamp, Type: Disqualify, CompetitorID: e.CompetitorID,
						}
						return []Event{disqualify}, nil
					}

					return []Event{}, nil
				},
			},
			{src: Scheduled, dst: NotStarted, event: Disqualify},
			{src: Scheduled, dst: CannotContinue, event: BeUnableToContinue}, // ???

			{
				src:   OnStartLine,
				dst:   OnMainLap,
				event: Start,
				cb: func(e Event, c *Competitor) ([]Event, error) {
					scheduledTime := c.ScheduledStartTime
					threshold := scheduledTime.Add(time.Duration(conf.StartDelta))
					if e.TimeStamp.After(threshold) {
						disqualify := Event{
							TimeStamp: e.TimeStamp, Type: Disqualify, CompetitorID: e.CompetitorID,
						}
						return []Event{disqualify}, nil
					} else {
						c.ActualStartTime = e.TimeStamp
						c.CurrentLap = 1
					}

					return []Event{}, nil
				},
			},
			{src: OnStartLine, dst: NotStarted, event: Disqualify},
			{src: OnStartLine, dst: CannotContinue, event: BeUnableToContinue}, // ???

			{
				src:   OnMainLap,
				dst:   OnRange,
				event: ComeToFiringRange,
				cb: func(e Event, c *Competitor) ([]Event, error) {
					firingRange := e.ExtraParams[0].(int)
					if c.VisitedRanges[firingRange-1] {
						disqualify := Event{
							TimeStamp: e.TimeStamp, Type: Disqualify, CompetitorID: e.CompetitorID,
						}
						return []Event{disqualify}, nil
					} else {
						c.VisitedRanges[firingRange-1] = true
					}

					return []Event{}, nil
				},
			},
			{src: OnMainLap, dst: OnPenaltyLap, event: EnterPenaltyLap},
			{
				src:   OnMainLap,
				dst:   OnMainLap,
				event: EndMainLap,
				cb: func(e Event, c *Competitor) ([]Event, error) {
					if c.CurrentLap < conf.Laps {
						c.CurrentLap++
					} else {
						finish := Event{TimeStamp: e.TimeStamp, Type: Finish, CompetitorID: e.CompetitorID}
						return []Event{finish}, nil
					}

					return []Event{}, nil
				},
			},
			{
				src:   OnMainLap,
				dst:   Finished,
				event: Finish,
				cb: func(e Event, c *Competitor) ([]Event, error) {
					for _, visited := range c.VisitedRanges {
						if !visited {
							disqualify := Event{
								TimeStamp: e.TimeStamp, Type: Disqualify, CompetitorID: e.CompetitorID,
							}
							return []Event{disqualify}, nil
						}
					}
					return []Event{}, nil
				},
			},
			{src: OnMainLap, dst: Disqualified, event: Disqualify},
			{src: OnMainLap, dst: CannotContinue, event: BeUnableToContinue}, // ???

			{
				src:   OnRange,
				dst:   OnRange,
				event: HitTarget,
				cb: func(e Event, c *Competitor) ([]Event, error) {
					target := e.ExtraParams[0].(int)
					if c.HitsThisRange[target-1] {
						return []Event{}, ErrWrongEventsSequence
					} else {
						c.HitsThisRange[target-1] = true
					}

					return []Event{}, nil
				},
			},
			{src: OnRange, dst: OnMainLap, event: LeaveFiringRange},
			{src: OnRange, dst: Disqualified, event: Disqualify},
			{src: OnRange, dst: CannotContinue, event: BeUnableToContinue}, // ???

			{src: OnPenaltyLap, dst: OnMainLap, event: LeavePenaltyLap},
			{src: OnPenaltyLap, dst: CannotContinue, event: BeUnableToContinue}, // ???

			{src: Finished, dst: Disqualified, event: Disqualify},
		}...,
	)

	return &Processor{
		events:      events,
		competitors: make(map[int]Competitor),
		fsm:         fsm,
		config:      conf,
		handlers:    make(map[eventType]EventHandler),
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
		err := p.processEvent(e)
		if err != nil {
			log.Printf("%v", err.Error())
		} else {
			logEvent(e)
		}

		p.eventsQueue = p.eventsQueue[1:]
	}
}

func (p *Processor) Handle(e eventType, handler EventHandler) {
	if _, ok := p.handlers[e]; ok {
		panic("da")
	}
	p.handlers[e] = handler
}

func (p *Processor) processEvent(e Event) error {
	cID := e.CompetitorID
	competitor := p.competitors[cID]

	status, cb, ok := p.fsm.LookupPath(competitor.Status, e.Type)
	if !ok {
		return ErrWrongEventsSequence
	}

	if cb != nil {
		generatedEvents, err := cb(e, &competitor)
		if err != nil {
			return err
		}
		p.eventsQueue = append(p.eventsQueue, generatedEvents...)
	}

	competitor.Status = status

	handler, ok := p.handlers[e.Type]
	if ok {
		handler(e)
	}

	p.competitors[cID] = competitor
	return nil
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
