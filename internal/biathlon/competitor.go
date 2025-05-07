package biathlon

import "time"

type competitorStatus int

const (
	Unknown competitorStatus = iota
	Registered
	Scheduled
	OnStartLine
	OnMainLap
	OnRange
	OnPenaltyLap
	Finished
	NotStarted
	Disqualified
	CannotContinue
)

type CompetitorState struct {
	ID                 int
	Status             competitorStatus
	CurrentLap         int
	ScheduledStartTime time.Time
	ActualStartTime    time.Time
	VisitedRanges      []bool
	HitsThisRange      [5]bool
}
