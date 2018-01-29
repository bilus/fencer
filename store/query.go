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
	Filter
	Matches map[ResultKey]Match
}

func (q *NeighbourQuery) Scan(broadcast *Broadcast) error {
	isMatch, err := q.Filter.IsMatch(broadcast)
	if err != nil {
		return err
	}
	if !isMatch {
		return nil
	}
	dist, err := broadcast.MinDistance(q.Point)
	if err != nil {
		return err
	}
	key := q.Filter.GetResultKey(broadcast)
	existingMatch, exists := q.Matches[key]
	if !exists || dist < existingMatch.cachedDistance {
		q.Matches[key] = Match{broadcast, dist}
	}
	return nil
}

// TODO: Needless coying if broadcast instances.
func (q *NeighbourQuery) GetMatchingBroadcasts() []*Broadcast {
	broadcasts := make([]*Broadcast, 0)
	for _, match := range q.Matches {
		broadcasts = append(broadcasts, match.Broadcast)
	}
	return broadcasts
}
