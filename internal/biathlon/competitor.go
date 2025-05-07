package biathlon

import "time"

type LapInfo struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	AvgSpeed  float64
}

type PenaltyLapInfo struct {
	EntryTime time.Time
	ExitTime  time.Time
	Duration  time.Duration
	AvgSpeed  float64
}

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

type Competitor struct {
	ID                 int
	Status             competitorStatus
	ScheduledStartTime time.Time
	ActualStartTime    time.Time
	FinishTime         time.Time
	CurrentLap         int
	TotalHits          int
	TotalShots         int
	LapsInfo           []LapInfo
	PenaltiesInfo      []PenaltyLapInfo
	DNFComment         string
	LastEventTime      time.Time
	RangeEntryTime     time.Time
	PenaltyEntryTime   time.Time
	HitsThisRange      [5]bool
	VisitedRanges      []bool
}

type CompetitorState struct {
	ID                 int
	Status             competitorStatus
	ScheduledStartTime time.Time
	ActualStartTime    time.Time
	VisitedRanges      []bool
	HitsThisRange      [5]bool
}
