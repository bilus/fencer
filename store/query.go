package store

// import (
// 	"github.com/bilus/rtreego"
// )

type ResultKey interface {
	ActsAsResultKey()
}

type Condition interface {
	IsMatch(broadcast *Broadcast) (bool, error)
}

// Possible optimization:
// func RTreeGoFilter(cond Condition) rtreego.Filter {
// 	return func(results []rtreego.Spatial, object rtreego.Spatial) (refuse, abort bool) {
// 		var err error
// 		refuse, err = cond.IsMatch(object.(*Broadcast))
// 		if err != nil {
// 			panic(err)
// 		}
// 		abort = false
// 		return
// 	}
// }

type Aggregation interface {
	GetResultKey(broadcast *Broadcast) ResultKey
}

type Filter interface {
	Condition
	Aggregation
}

type Match struct {
	Broadcast      *Broadcast
	cachedDistance float64
}

// NeighbourQuery is a nearest neighour query returning broadcasts matching the filters,
// that are closest to the Point, at most one Match per ResultKey.
//
// Filters' precedence:
//
// ConjFilter0 AND ConjFilter1 AND ... ConjFilterN AND (DisjFilter0 OR DisjFilter1 OR ... DisjFilterN)
type NeighbourQuery struct {
	Point
	Preconditions []Condition // Pre-conditions forming a logical conjunction. NOTE: Take precedence over Filters.
	Filters       []Filter    // Logical disjunction.
	matches       map[ResultKey]Match
}

func NewNeighbourQuery(point Point, preconditions []Condition, filters []Filter) NeighbourQuery {
	return NeighbourQuery{point, preconditions, filters, make(map[ResultKey]Match)}
}

func (q *NeighbourQuery) Scan(broadcast *Broadcast) error {
	match, err := allMatch(q.Preconditions, broadcast)
	if err != nil {
		return err
	}
	if !match {
		return nil
	}
	keys, err := filter(q.Filters, broadcast)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	dist, err := broadcast.MinDistance(q.Point)
	if err != nil {
		return err
	}
	for _, key := range keys {
		existingMatch, exists := q.matches[key]
		if !exists || dist < existingMatch.cachedDistance {
			q.matches[key] = Match{broadcast, dist}
		}
	}
	return nil
}

func (q *NeighbourQuery) GetMatchingBroadcasts() []*Broadcast {
	broadcasts := make([]*Broadcast, 0)
	matched := make(map[int64]struct{})
	for _, match := range q.matches {
		broadcastId := match.Broadcast.BroadcastId
		_, isMatched := matched[broadcastId]
		if !isMatched {
			broadcasts = append(broadcasts, match.Broadcast)
			matched[broadcastId] = struct{}{}
		}
	}
	return broadcasts
}

func allMatch(conditions []Condition, broadcast *Broadcast) (bool, error) {
	for _, condition := range conditions {
		match, err := condition.IsMatch(broadcast)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}

func filter(filters []Filter, broadcast *Broadcast) ([]ResultKey, error) {
	keys := make([]ResultKey, 0)
	for _, filter := range filters {
		match, err := filter.IsMatch(broadcast)
		if err != nil {
			return nil, err
		}
		if match {
			keys = append(keys, filter.GetResultKey(broadcast))
		}
	}
	return keys, nil
}
