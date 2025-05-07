package biathlon

import (
	"errors"
	"os"
	"time"
)

var (
	ErrWrongEventsSequence = errors.New("impossible sequence of events")
	ErrInvalidParamValue   = errors.New("invalid parameter value")
)

type EventHandler func(e Event)

type Processor struct {
	events      <-chan Event
	eventsQueue []Event
	competitors map[int]CompetitorState

	fsm FSM

	handlers map[eventType]EventHandler

	config Config

	log Logger
}

func NewProcessor(conf Config, events <-chan Event) *Processor {
	return &Processor{
		events:      events,
		competitors: make(map[int]CompetitorState),
		fsm:         initBiathlonFSM(conf),
		config:      conf,
		handlers:    make(map[eventType]EventHandler),
		log:         NewDefaultLogger(os.Stdout),
	}
}

func (p *Processor) SetLogger(log Logger) {
	p.log = log
}

func (p *Processor) Start() {
	var lastTime time.Time

	for {
		if len(p.eventsQueue) == 0 {
			e, ok := <-p.events
			if !ok {
				break
			}
			p.eventsQueue = append(p.eventsQueue, e)
		}
		e := p.eventsQueue[0]
		p.eventsQueue = p.eventsQueue[1:]

		lastTime = e.TimeStamp

		if err := p.processEvent(e); err != nil {
			p.log.Error(e.TimeStamp, err)
		} else {
			p.log.Event(e)
		}
	}
	p.finalize(lastTime)
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

func (p *Processor) finalize(lastTime time.Time) {
	for cID, c := range p.competitors {
		status, cb, ok := p.fsm.LookupPath(c.Status, Disqualify)
		if !ok || c.Status == Finished {
			continue
		}

		disqualify := Event{TimeStamp: lastTime, Type: Disqualify, CompetitorID: cID}
		if cb != nil {
			_, err := cb(disqualify, &c)
			if err != nil {
				p.log.Error(lastTime, err)
				continue
			}
		}
		c.Status = status

		handler, ok := p.handlers[Disqualify]
		if ok {
			handler(disqualify)
		}

		p.log.Event(disqualify)

		p.competitors[cID] = c
	}
}

func initBiathlonFSM(conf Config) FSM {
	return NewFSM(
		[]Edge{
			{
				Src: Unknown,
				Dst: Registered, Event: Register,
				Cb: func(_ Event, c *CompetitorState) ([]Event, error) {
					c.VisitedRanges = make([]bool, conf.FiringLines)
					return []Event{}, nil
				},
			},

			{
				Src:   Registered,
				Dst:   Scheduled,
				Event: BeSheduled,
				Cb: func(e Event, c *CompetitorState) ([]Event, error) {
					startTime := e.ExtraParams[0].(time.Time)
					c.ScheduledStartTime = startTime

					return []Event{}, nil
				},
			},
			{Src: Registered, Dst: NotStarted, Event: Disqualify},
			{Src: Registered, Dst: CannotContinue, Event: BeUnableToContinue}, // ???

			{
				Src:   Scheduled,
				Dst:   OnStartLine,
				Event: ComeToStartLine,
				Cb: func(e Event, c *CompetitorState) ([]Event, error) {
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
			{Src: Scheduled, Dst: NotStarted, Event: Disqualify},
			{Src: Scheduled, Dst: CannotContinue, Event: BeUnableToContinue}, // ???

			{
				Src:   OnStartLine,
				Dst:   OnMainLap,
				Event: Start,
				Cb: func(e Event, c *CompetitorState) ([]Event, error) {
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
			{Src: OnStartLine, Dst: NotStarted, Event: Disqualify},
			{Src: OnStartLine, Dst: CannotContinue, Event: BeUnableToContinue}, // ???

			{
				Src:   OnMainLap,
				Dst:   OnRange,
				Event: ComeToFiringRange,
				Cb: func(e Event, c *CompetitorState) ([]Event, error) {
					firingRange := e.ExtraParams[0].(int)
					if firingRange < 1 || firingRange > conf.FiringLines {
						return []Event{}, ErrInvalidParamValue
					}
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
			{Src: OnMainLap, Dst: OnPenaltyLap, Event: EnterPenaltyLap},
			{
				Src:   OnMainLap,
				Dst:   OnMainLap,
				Event: EndMainLap,
				Cb: func(e Event, c *CompetitorState) ([]Event, error) {
					if c.CurrentLap < conf.Laps {
						c.CurrentLap++
						c.HitsThisRange = [5]bool{}
					} else {
						finish := Event{TimeStamp: e.TimeStamp, Type: Finish, CompetitorID: e.CompetitorID}
						return []Event{finish}, nil
					}

					return []Event{}, nil
				},
			},
			{
				Src:   OnMainLap,
				Dst:   Finished,
				Event: Finish,
				Cb: func(e Event, c *CompetitorState) ([]Event, error) {
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
			{Src: OnMainLap, Dst: Disqualified, Event: Disqualify},
			{Src: OnMainLap, Dst: CannotContinue, Event: BeUnableToContinue}, // ???

			{
				Src:   OnRange,
				Dst:   OnRange,
				Event: HitTarget,
				Cb: func(e Event, c *CompetitorState) ([]Event, error) {
					target := e.ExtraParams[0].(int)
					if target < 1 || target > 5 {
						return []Event{}, ErrInvalidParamValue
					}
					if c.HitsThisRange[target-1] {
						return []Event{}, ErrWrongEventsSequence
					} else {
						c.HitsThisRange[target-1] = true
					}

					return []Event{}, nil
				},
			},
			{Src: OnRange, Dst: OnMainLap, Event: LeaveFiringRange},
			{Src: OnRange, Dst: Disqualified, Event: Disqualify},
			{Src: OnRange, Dst: CannotContinue, Event: BeUnableToContinue}, // ???

			{Src: OnPenaltyLap, Dst: OnMainLap, Event: LeavePenaltyLap},
			{Src: OnPenaltyLap, Dst: Disqualified, Event: Disqualify},
			{Src: OnPenaltyLap, Dst: CannotContinue, Event: BeUnableToContinue}, // ???

			{Src: Finished, Dst: Disqualified, Event: Disqualify},
		}...,
	)
}
