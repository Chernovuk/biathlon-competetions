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

