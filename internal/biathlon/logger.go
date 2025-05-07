package biathlon

import (
	"fmt"
	"io"
	"time"
)

type Logger interface {
	Event(e Event)
	Error(time time.Time, err error)
}

type DefaultLogger struct {
	out io.Writer
}

func NewDefaultLogger(out io.Writer) *DefaultLogger {
	return &DefaultLogger{
		out: out,
	}
}

func (l *DefaultLogger) Error(time time.Time, err error) {
	ts := time.Format("15:04:05.000")
	msg := fmt.Sprintf("[%s] %s\n", ts, err.Error())

	if _, err := l.out.Write([]byte(msg)); err != nil {
		fmt.Printf("Logger error: %s\n", err.Error())
	}
}

func (l *DefaultLogger) Event(e Event) {
	msg := l.msgFromEvent(e)

	if _, err := l.out.Write([]byte(msg)); err != nil {
		fmt.Printf("Logger error: %s\n", err.Error())
	}
}

func (l *DefaultLogger) msgFromEvent(e Event) string {
	ts := e.TimeStamp.Format("15:04:05.000")
	switch e.Type {
	case Register:
		return fmt.Sprintf("[%s] The competitor(%d) registered\n", ts, e.CompetitorID)

	case BeSheduled:
		t := e.ExtraParams[0].(time.Time)
		return fmt.Sprintf("[%s] The start time for the competitor(%d) was set by a draw to %s\n",
			ts, e.CompetitorID, t.Format("15:04:05.000"))

	case ComeToStartLine:
		return fmt.Sprintf("[%s] The competitor(%d) is on the start line\n", ts, e.CompetitorID)

	case Start:
		return fmt.Sprintf("[%s] The competitor(%d) has started\n", ts, e.CompetitorID)

	case ComeToFiringRange:
		line := e.ExtraParams[0].(int)
		return fmt.Sprintf(
			"[%s] The competitor(%d) is on the firing range(%d)\n",
			ts,
			e.CompetitorID,
			line,
		)

	case HitTarget:
		target := e.ExtraParams[0].(int)
		return fmt.Sprintf(
			"[%s] The target(%d) has been hit by competitor(%d)\n",
			ts,
			target,
			e.CompetitorID,
		)

	case LeaveFiringRange:
		return fmt.Sprintf("[%s] The competitor(%d) left the firing range\n", ts, e.CompetitorID)

	case EnterPenaltyLap:
		return fmt.Sprintf("[%s] The competitor(%d) entered the penalty laps\n", ts, e.CompetitorID)

	case LeavePenaltyLap:
		return fmt.Sprintf("[%s] The competitor(%d) left the penalty laps\n", ts, e.CompetitorID)

	case EndMainLap:
		return fmt.Sprintf("[%s] The competitor(%d) ended the main lap\n", ts, e.CompetitorID)

	case BeUnableToContinue:
		comment := e.ExtraParams[0].(string)
		return fmt.Sprintf(
			"[%s] The competitor(%d) can`t continue: %s\n",
			ts,
			e.CompetitorID,
			comment,
		)

	case Disqualify:
		return fmt.Sprintf("[%s] The competitor(%d) is disqualified\n", ts, e.CompetitorID)

	case Finish:
		return fmt.Sprintf("[%s] The competitor(%d) has finished\n", ts, e.CompetitorID)

	default:
		return fmt.Sprintf("[%s] Unknown event(%d) for competitor(%d)\n", ts, e.Type, e.CompetitorID)
	}
}
