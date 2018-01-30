package store

type ResultKey interface {
	ActsAsResultKey()
}

type Filter interface {
	IsMatch(broadcast *Broadcast) (bool, error)
	GetResultKey(broadcast *Broadcast) ResultKey
}

type Match struct {
	Broadcast      *Broadcast
	cachedDistance float64
}

type NeighbourQuery struct {
	Point
	Filters []Filter
	Matches map[ResultKey]Match
}

func (q *NeighbourQuery) Scan(broadcast *Broadcast) error {
	keys, err := anyMatch(q.Filters, broadcast)
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
		existingMatch, exists := q.Matches[key]
		// PERF: Need not calc dist if !exists.
		if !exists || dist < existingMatch.cachedDistance {
			q.Matches[key] = Match{broadcast, dist}
		}
	}
	return nil
}

func (q *NeighbourQuery) GetMatchingBroadcasts() []*Broadcast {
	broadcasts := make([]*Broadcast, 0)
	matched := make(map[int64]struct{})
	for _, match := range q.Matches {
		broadcastId := match.Broadcast.BroadcastId
		_, isMatched := matched[broadcastId]
		if !isMatched {
			broadcasts = append(broadcasts, match.Broadcast)
			matched[broadcastId] = struct{}{}
		}
	}
	return broadcasts
}

func anyMatch(filters []Filter, broadcast *Broadcast) ([]ResultKey, error) {
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
