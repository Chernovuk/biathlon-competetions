package biathlon

import (
	"fmt"
)

type FiniteStateMachine struct {
	graph map[competitorStatus]map[eventType]competitorStatus
}

type edge struct {
	src   competitorStatus
	dst   competitorStatus
	event eventType
}

func makeFSM(edges ...edge) FiniteStateMachine {
	graph := make(map[competitorStatus]map[eventType]competitorStatus)

	for _, edge := range edges {
		if _, ok := graph[edge.src]; !ok {
			graph[edge.src] = make(map[eventType]competitorStatus)
		}
		if _, ok := graph[edge.src][edge.event]; ok {
			msg := fmt.Sprintf(
				"There has already been an edge with src: %v and ev: %v.",
				edge.src,
				edge.event,
			)
			panic(msg)
		}
		graph[edge.src][edge.event] = edge.dst
	}

	return FiniteStateMachine{
		graph: graph,
	}
}

func (f FiniteStateMachine) LookupPath(src competitorStatus, ev eventType) (competitorStatus, bool) {
	paths, ok := f.graph[src]
	if !ok {
		return competitorStatus(-1), false
	}
	dst, ok := paths[ev]

	return dst, ok
}

var fsm = makeFSM(
	[]edge{
		{src: Unknown, dst: Registered, event: Register},

		{src: Registered, dst: Scheduled, event: BeSheduled},
		{src: Registered, dst: NotStarted, event: Disqualify},
		{src: Registered, dst: CannotContinue, event: BeUnableToContinue},

		{src: Scheduled, dst: OnStartLine, event: ComeToStartLine},
		{src: Scheduled, dst: NotStarted, event: Disqualify},
		{src: Scheduled, dst: CannotContinue, event: BeUnableToContinue},

		{src: OnStartLine, dst: OnMainLap, event: Start},
		{src: OnStartLine, dst: NotStarted, event: Disqualify},
		{src: OnStartLine, dst: CannotContinue, event: BeUnableToContinue},

		{src: OnMainLap, dst: OnRange, event: ComeToFiringRange},
		{src: OnMainLap, dst: OnPenaltyLap, event: EnterPenaltyLap},
		{src: OnMainLap, dst: OnMainLap, event: EndMainLap},
		{src: OnMainLap, dst: Finished, event: Finish},
		{src: OnMainLap, dst: CannotContinue, event: BeUnableToContinue},

		{src: OnRange, dst: OnRange, event: HitTarget},
		{src: OnRange, dst: OnMainLap, event: LeaveFiringRange},
		{src: OnRange, dst: CannotContinue, event: BeUnableToContinue},

		{src: OnPenaltyLap, dst: OnMainLap, event: LeavePenaltyLap},
		{src: OnPenaltyLap, dst: CannotContinue, event: BeUnableToContinue},
	}...,
)

func FSM() FiniteStateMachine {
	return fsm
}
