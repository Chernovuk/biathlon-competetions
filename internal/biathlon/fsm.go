package biathlon

import (
	"fmt"
	"slices"
)

type Callback func(e Event, c *CompetitorState) ([]Event, error)

type Edge struct {
	Src   competitorStatus
	Dst   competitorStatus
	Event eventType
	Cb    Callback
}

func CmpEdges(a, b Edge) int {
	if a.Src < b.Src {
		return -1
	}
	if a.Src > b.Src {
		return 1
	}

	if a.Event < b.Event {
		return -1
	}
	if a.Event > b.Event {
		return 1
	}

	return 0
}

type FSM struct {
	edges []Edge
}

func NewFSM(edges ...Edge) FSM {
	edgesCopy := make([]Edge, len(edges))
	copy(edgesCopy, edges)

	slices.SortFunc(edgesCopy, CmpEdges)
	for i := 1; i < len(edgesCopy); i++ {
		if CmpEdges(edges[i-1], edges[i]) == 0 {
			msg := fmt.Sprintf(
				"There has already been an edge with src: %v and ev: %v.",
				edges[i].Src,
				edges[i].Event,
			)
			panic(msg)
		}
	}

	return FSM{
		edges: edgesCopy,
	}
}

func (f FSM) LookupPath(
	src competitorStatus,
	ev eventType,
) (competitorStatus, Callback, bool) {
	idx, ok := slices.BinarySearchFunc(f.edges, Edge{Src: src, Event: ev}, CmpEdges)
	if !ok {
		return competitorStatus(-1), nil, false
	}

	dst := f.edges[idx].Dst
	cb := f.edges[idx].Cb

	return dst, cb, ok
}
