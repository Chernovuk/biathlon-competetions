package statistics

import (
	"slices"
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
	ID                 int
	Status             string
	ScheduledStartTime time.Time
	FinishTime         time.Time
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

func New(c biathlon.Config) *Statistics {
	return &Statistics{
		competitorsInfo: make(map[int]Competitor),
		laps:            c.Laps,
		lapLen:          c.LapLen,
		penaltyLen:      c.PenaltyLen,
	}
}

func (s *Statistics) GetResults() []Result {
	resultingTable := make([]Result, 0, len(s.competitorsInfo))
	for _, competitor := range s.competitorsInfo {
		res := Result{
			CompetitorID: competitor.ID,
			TotalHits:    competitor.TotalHits,
			TotalShots:   competitor.TotalShots,
		}
		for _, lap := range competitor.LapsInfo {
			v := struct {
				duration time.Duration
				avgSpeed float64
			}{
				lap.Duration,
				lap.AvgSpeed,
			}
			res.LapsInfo = append(res.LapsInfo, v)
		}
		for _, penLap := range competitor.PenaltiesInfo {
			v := struct {
				duration time.Duration
				avgSpeed float64
			}{
				penLap.Duration,
				penLap.AvgSpeed,
			}
			res.PenaltyLapsInfo = append(res.PenaltyLapsInfo, v)
		}
		if competitor.FinishTime.IsZero() {
			res.Result = competitor.Status
		} else {
			res.Result = formatDuration(competitor.FinishTime.Sub(competitor.LapsInfo[0].StartTime))
		}
		resultingTable = append(resultingTable, res)
	}
	slices.SortFunc(resultingTable, cmp)

	return resultingTable
}

func cmp(a, b Result) int {
	if a.Result < b.Result {
		return -1
	} else if a.Result == b.Result {
		return 0
	} else {
		return 1
	}
}

func (s *Statistics) OnRegister(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.ID = e.CompetitorID

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnBeSheduled(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	scheduledStartTime := e.ExtraParams[0].(time.Time)
	stat.ScheduledStartTime = scheduledStartTime

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnStart(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	lapInfo := LapInfo{
		StartTime: e.TimeStamp,
	}
	stat.LapsInfo = append(stat.LapsInfo, lapInfo)

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnComeToFiringRange(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.TotalShots += 5

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnHitTarget(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.TotalHits++

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnEnterPenaltyLap(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	penaltyInfo := PenaltyLapInfo{
		EntryTime: e.TimeStamp,
	}
	stat.PenaltiesInfo = append(stat.PenaltiesInfo, penaltyInfo)

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnLeavePenaltyLap(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]

	currLap := len(stat.PenaltiesInfo) - 1
	penaltyInfo := stat.PenaltiesInfo[currLap]
	penaltyInfo.ExitTime = e.TimeStamp
	penaltyInfo.Duration = penaltyInfo.ExitTime.Sub(penaltyInfo.EntryTime)
	penaltyInfo.AvgSpeed = s.penaltyLen / penaltyInfo.Duration.Seconds()

	stat.PenaltiesInfo[currLap] = penaltyInfo

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnEndMainLap(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]

	currLap := len(stat.LapsInfo) - 1
	lapInfo := stat.LapsInfo[currLap]
	lapInfo.EndTime = e.TimeStamp
	lapInfo.Duration = lapInfo.EndTime.Sub(lapInfo.StartTime)
	lapInfo.AvgSpeed = s.lapLen / lapInfo.Duration.Seconds()

	stat.LapsInfo[currLap] = lapInfo

	if currLap < s.laps-1 {
		newLapInfo := LapInfo{
			StartTime: e.TimeStamp,
		}
		stat.LapsInfo = append(stat.LapsInfo, newLapInfo)
	}

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnBeUnableToContinue(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.Status = "NotFinished"

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnDisqualify(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]

	if len(stat.LapsInfo) > 0 {
		stat.Status = "Disqualified"
	} else {
		stat.Status = "NotFinished"
	}

	s.competitorsInfo[e.CompetitorID] = stat
}

func (s *Statistics) OnFinish(e biathlon.Event) {
	stat := s.competitorsInfo[e.CompetitorID]
	stat.FinishTime = e.TimeStamp

	s.competitorsInfo[e.CompetitorID] = stat
}
