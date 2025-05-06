package statistics

import (
	"fmt"
	"strings"
	"time"

	"github.com/Chernovuk/biathlon-competetions/internal/biathlon"
)

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

type Competitor struct {
	// CompetitorID  int
	NotStarted         bool
	NotFinished        bool
	ScheduledStartTime time.Time
	TotalHits          int
	TotalShots         int
	LapsInfo           []LapInfo
	PenaltiesInfo      []PenaltyLapInfo
}

type Statistics struct {
	laps            int
	lapLen          float64
	penaltyLen      float64
	competitorsInfo map[int]Competitor
}

type Result struct {
	Result          string
	CompetitorID    int
	LapsInfo        []LapInfo
	PenaltyLapsInfo []PenaltyLapInfo
	TotalHits       int
	TotalShots      int
}

func (r Result) String() string {
	var sb strings.Builder

	// [Result] CompetitorID
	sb.WriteString(fmt.Sprintf("[%s] %d ", r.Result, r.CompetitorID))

	// Laps: always bracketed list
	sb.WriteString("[")
	for i, lap := range r.LapsInfo {
		if i > 0 {
			sb.WriteString(", ")
		}
		if lap.Duration > 0 {
			sb.WriteString(fmt.Sprintf("{%s, %.3f}",
				formatDuration(lap.Duration),
				lap.AvgSpeed,
			))
		} else {
			sb.WriteString("{,}")
		}
	}
	sb.WriteString("]")

	if n := len(r.PenaltyLapsInfo); n > 0 {
		sb.WriteString(" [")
		for i, pl := range r.PenaltyLapsInfo {
			if i > 0 {
				sb.WriteString(", ")
			}
			if pl.Duration > 0 {
				sb.WriteString(fmt.Sprintf("{%s, %.3f}",
					formatDuration(pl.Duration),
					pl.AvgSpeed,
				))
			} else {
				sb.WriteString("{,}")
			}
		}
		sb.WriteString("]")
	}

	sb.WriteString(fmt.Sprintf(" %d/%d", r.TotalHits, r.TotalShots))

	return sb.String()
}

// formatDuration prints a time.Duration as HH:MM:SS.mmm
func formatDuration(d time.Duration) string {
	h := int(d / time.Hour)
	d -= time.Duration(h) * time.Hour
	m := int(d / time.Minute)
	d -= time.Duration(m) * time.Minute
	s := int(d / time.Second)
	d -= time.Duration(s) * time.Second
	ms := int(d / time.Millisecond)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func New(c biathlon.Config) *Statistics {
	return &Statistics{
		competitorsInfo: make(map[int]Competitor),
		laps:            c.Laps,
		lapLen:          c.LapLen,
		penaltyLen:      c.PenaltyLen,
	}
}

func (s *Statistics) OnBeSheduled(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]
	scheduledStartTime, ok := e.ExtraParams[0].(time.Time)
	if !ok {
		return fmt.Errorf("o")
	}
	stat.ScheduledStartTime = scheduledStartTime

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnStart(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]
	lapInfo := LapInfo{
		StartTime: e.TimeStamp,
	}
	stat.LapsInfo = append(stat.LapsInfo, lapInfo)

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnComeToFiringRange(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.TotalShots += 5

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnHitTarget(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.TotalHits++

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnEnterPenaltyLap(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]
	penaltyInfo := PenaltyLapInfo{
		EntryTime: e.TimeStamp,
	}
	stat.PenaltiesInfo = append(stat.PenaltiesInfo, penaltyInfo)

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnLeavePenaltyLap(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]

	currLap := len(stat.PenaltiesInfo) - 1
	penaltyInfo := stat.PenaltiesInfo[currLap]
	penaltyInfo.ExitTime = e.TimeStamp
	penaltyInfo.Duration = penaltyInfo.ExitTime.Sub(penaltyInfo.ExitTime)
	penaltyInfo.AvgSpeed = s.penaltyLen / penaltyInfo.Duration.Seconds()

	stat.PenaltiesInfo[currLap] = penaltyInfo

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnEndMainLap(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]

	currLap := len(stat.LapsInfo) - 1
	lapInfo := stat.LapsInfo[currLap]
	lapInfo.EndTime = e.TimeStamp
	lapInfo.Duration = lapInfo.EndTime.Sub(lapInfo.EndTime)
	lapInfo.AvgSpeed = s.lapLen / lapInfo.Duration.Seconds()

	stat.LapsInfo[currLap] = lapInfo

	if currLap < s.laps {
		newLapInfo := LapInfo{
			StartTime: e.TimeStamp,
		}
		stat.LapsInfo = append(stat.LapsInfo, newLapInfo)
	}

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnBeUnableToContinue(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.NotFinished = true

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}

func (s *Statistics) OnDisqualify(e biathlon.Event) error {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.NotStarted = true

	s.competitorsInfo[e.CompetitorID] = stat
	return nil
}
