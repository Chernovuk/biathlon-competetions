package biathlon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type eventType int

const (
	Register eventType = iota + 1
	BeSheduled
	ComeToStartLine
	Start
	ComeToFiringRange
	HitTarget
	LeaveFiringRange
	EnterPenaltyLap
	LeavePenaltyLap
	EndMainLap
	BeUnableToContinue
)

const (
	Disqualify eventType = 32
	Finish     eventType = 33
)

var ErrWrongEventFormat = errors.New(
	"wrong event format: [HH:MM:SS.sss] EventID CompetitorID ExtraParams... required",
)

type Event struct {
	TimeStamp    time.Time
	Type         eventType
	CompetitorID int
	ExtraParams  []any // Should it be a slice? There's always only one extraParams
}

func ParseEvent(eventLine string) (Event, error) {
	rawEvent := strings.Split(eventLine, " ")

	if len(rawEvent) < 3 {
		return Event{}, fmt.Errorf(
			"%d out of at least 3 params are passed: %w",
			len(rawEvent),
			ErrWrongEventFormat,
		)
	}

	e := Event{}
	timeStamp, err := parseEventTime(rawEvent[0])
	if err != nil {
		return Event{}, fmt.Errorf("%w: %w", err, ErrWrongEventFormat)
	}
	e.TimeStamp = timeStamp

	eventID, err := strconv.Atoi(rawEvent[1])
	if err != nil {
		return Event{}, fmt.Errorf("%w: %w", err, ErrWrongEventFormat)
	}
	e.Type = eventType(eventID)

	competitorID, err := strconv.Atoi(rawEvent[2])
	if err != nil {
		return Event{}, fmt.Errorf("%w: %w", err, ErrWrongEventFormat)
	}
	e.CompetitorID = competitorID

	switch e.Type {
	case BeSheduled:
		if len(rawEvent) != 4 {
			return Event{}, fmt.Errorf("%d event requires 4th param as start time", BeSheduled)
		}
		timeStamp, err := parseEventTime(rawEvent[3])
		if err != nil {
			return Event{}, err
		}
		e.ExtraParams = append(e.ExtraParams, timeStamp)
	case ComeToFiringRange:
		if len(rawEvent) != 4 {
			return Event{}, fmt.Errorf(
				"%d event requires 4th param as range number",
				ComeToFiringRange,
			)
		}
		firingRange, err := strconv.Atoi(rawEvent[3])
		if err != nil {
			return Event{}, err
		}
		e.ExtraParams = append(e.ExtraParams, firingRange)
	case HitTarget:
		if len(rawEvent) != 4 {
			return Event{}, fmt.Errorf("%d event requires 4th param as target", HitTarget)
		}
		target, err := strconv.Atoi(rawEvent[3])
		if err != nil {
			return Event{}, err
		}

		if target < 1 || target > 5 {
			return Event{}, fmt.Errorf("wrong parameter, there're only targets 1 to 5")
		}
		e.ExtraParams = append(e.ExtraParams, target)
	case BeUnableToContinue:
		if len(rawEvent) != 4 {
			return Event{}, fmt.Errorf("%d event requires 4th param as comment", BeUnableToContinue)
		}
		comment := rawEvent[3]
		e.ExtraParams = append(e.ExtraParams, comment)
	}

	return e, nil
}

func parseEventTime(rawTime string) (time.Time, error) {
	trimmedtime := strings.Trim(rawTime, `[]`)
	t, err := time.Parse(time.TimeOnly, trimmedtime)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
