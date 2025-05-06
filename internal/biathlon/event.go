package biathlon

import (
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

type Event struct {
	TimeStamp    time.Time
	Type         eventType
	CompetitorID int
	ExtraParams  []any // Should it be a slice? There's always only one extraParams
}

func ParseEvent(eventLine string) (Event, error) {
	rawEvent := strings.Split(eventLine, " ")
	// Need to check the len of it

	event := Event{}
	timeStamp, err := parseEventTime(rawEvent[0])
	if err != nil {
		return Event{}, err
	}
	event.TimeStamp = timeStamp

	eventID, err := strconv.Atoi(rawEvent[1])
	if err != nil {
		return Event{}, err
	}
	event.Type = eventType(eventID)

	competitorID, err := strconv.Atoi(rawEvent[2])
	if err != nil {
		return Event{}, err
	}
	event.CompetitorID = competitorID

	switch event.Type {
	case BeSheduled:
		timeStamp, err := parseEventTime(rawEvent[3])
		if err != nil {
			return Event{}, err
		}
		event.ExtraParams = append(event.ExtraParams, timeStamp)
	case ComeToFiringRange:
		firingRange, err := strconv.Atoi(rawEvent[3])
		if err != nil {
			return Event{}, err
		}
		event.ExtraParams = append(event.ExtraParams, firingRange)
	case HitTarget:
		target, err := strconv.Atoi(rawEvent[3])
		if err != nil {
			return Event{}, err
		}

		if target < 1 || target > 5 {
			return Event{}, fmt.Errorf("Wrong parameter, there're only targets 1 to 5")
		}
		event.ExtraParams = append(event.ExtraParams, target)
	case BeUnableToContinue:
		comment := rawEvent[3]
		event.ExtraParams = append(event.ExtraParams, comment)
	}

	return event, nil
}

func parseEventTime(rawTime string) (time.Time, error) {
	trimmedtime := strings.Trim(rawTime, `[]`)
	if trimmedtime == "" || trimmedtime == "null" {
		return time.Time{}, fmt.Errorf("wrong time stamp")
	}

	t, err := time.Parse(time.TimeOnly, trimmedtime)
	if err != nil {
		return time.Time{}, fmt.Errorf("wrong formatting of time stamp")
	}
	return t, nil
}
