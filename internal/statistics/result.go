package statistics

import (
	"fmt"
	"strings"
	"time"
)

type Result struct {
	Result       string
	CompetitorID int
	LapsInfo     []struct {
		duration time.Duration
		avgSpeed float64
	}
	PenaltyLapsInfo []struct {
		duration time.Duration
		avgSpeed float64
	}
	TotalHits  int
	TotalShots int
}

func (r Result) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s] %d ", r.Result, r.CompetitorID))

	sb.WriteString("[")
	for i, lap := range r.LapsInfo {
		if i > 0 {
			sb.WriteString(", ")
		}
		if lap.duration > 0 {
			sb.WriteString(fmt.Sprintf("{%s, %.3f}",
				formatDuration(lap.duration),
				lap.avgSpeed,
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
			if pl.duration > 0 {
				sb.WriteString(fmt.Sprintf("{%s, %.3f}",
					formatDuration(pl.duration),
					pl.avgSpeed,
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

// formatDuration prints a time.Duration as HH:MM:SS.sss.
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
