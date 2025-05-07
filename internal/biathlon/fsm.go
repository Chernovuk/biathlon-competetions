package biathlon

import (
	"fmt"
	"slices"
)

type callback func(e Event, c *Competitor) ([]Event, error)

type edge struct {
	src   competitorStatus
	dst   competitorStatus
	event eventType
	cb    callback
}

func cmp(a, b edge) int {
	if a.src < b.src {
		return -1
	}
	if a.src > b.src {
		return 1
	}

	if a.event < b.event {
		return -1
	}
	if a.event > b.event {
		return 1
	}

	return 0
}

type FiniteStateMachine struct {
	graph map[competitorStatus]map[eventType]competitorStatus
	edges []edge
}

func makeFSM(edges ...edge) FiniteStateMachine {
	edgesCopy := make([]edge, len(edges))
	copy(edgesCopy, edges)

	slices.SortFunc(edgesCopy, cmp)
	for i := 1; i < len(edgesCopy); i++ {
		if cmp(edges[i-1], edges[i]) == 0 {
			msg := fmt.Sprintf(
				"There has already been an edge with src: %v and ev: %v.",
				edges[i].src,
				edges[i].event,
			)
			panic(msg)
		}
	}

	return FiniteStateMachine{
		edges: edgesCopy,
	}
}

func (f FiniteStateMachine) LookupPath(
	src competitorStatus,
	ev eventType,
) (competitorStatus, callback, bool) {
	idx, ok := slices.BinarySearchFunc(f.edges, edge{src: src, event: ev}, cmp)
	if !ok {
		return competitorStatus(-1), nil, false
	}

	dst := f.edges[idx].dst
	cb := f.edges[idx].cb

	return dst, cb, ok
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
