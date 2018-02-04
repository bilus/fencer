package store

import (
	"database/sql"
	"github.com/bilus/fencer/query"
	"github.com/bilus/rtreego"
	"github.com/paulmach/go.geo"
	"log"
)

type Point = rtreego.Point

type BroadcastStore struct {
	*rtreego.Rtree
}

func LoadFromSQL(db *sql.DB) (*BroadcastStore, error) {
	broadcasts, numSkipped, err := LoadBroadcastsFromSQL(db)
	if err != nil {
		return nil, err
	}
	log.Printf("Skipped: %v broadcasts due to errors or missing data", numSkipped)
	return New(broadcasts)
}

func New(broadcasts []*Broadcast) (*BroadcastStore, error) {
	store := BroadcastStore{rtreego.NewTree(2, 5, 20)}
	for _, broadcast := range broadcasts {
		store.Insert(broadcast)
	}
	return &store, nil
}

type MatchContaining struct {
	Point
}

func (mc MatchContaining) IsMatch(broadcast *Broadcast) (bool, error) {
	return broadcast.Contains(mc.Point), nil
}

func (store *BroadcastStore) FindBroadcasts(point Point) ([]*Broadcast, error) {
	p, err := rtreego.NewRect(rtreego.Point(point), []float64{0.01, 0.01})
	if err != nil {
		return nil, err
	}
	candidates := store.SearchIntersect(p)
	if len(candidates) == 0 {
		return nil, nil
	}
	log.Println("candidates =", len(candidates), point)

	// TODO: Refactor using NewQuery. But we first need to put the
	// code to select the nearest broadcast into a Filter or Aggregator.
	// Otherwise, it generates unnecessary overhead.
	// query := NewQuery(point, conditions, nil)
	// conditions := []Condition{MatchContaining{point}}
	// for _, candidate := range candidates {
	// 	err := query.Scan(candidate.(*Broadcast))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	// return query.MatchingFeatures(), nil

	condition := MatchContaining{point}
	broadcasts := make([]*Broadcast, 0, len(candidates))
	for _, candidate := range candidates {
		broadcast := candidate.(*Broadcast)
		match, err := condition.IsMatch(broadcast)
		if err != nil {
			return nil, err
		}
		if match {
			broadcasts = append(broadcasts, broadcast)
		}

	}
	return broadcasts, nil
}

func (store *BroadcastStore) FindClosestBroadcasts(point Point, radiusMeters float64, preconditions []query.Condition, filters []query.Filter, reducer query.Reducer) ([]*Broadcast, error) {
	bounds, err := geomBoundsAround(point, radiusMeters)
	if err != nil {
		log.Fatal(err)
	}

	candidates := store.SearchIntersect(bounds)
	if len(candidates) == 0 {
		return nil, nil
	}

	query := query.New(preconditions, filters, reducer)
	for _, candidate := range candidates {
		err := query.Scan(candidate.(*Broadcast))
		if err != nil {
			return nil, err
		}
	}
	features := query.MatchingFeatures()
	broadcasts := make([]*Broadcast, len(features))
	for i, feature := range features {
		broadcasts[i] = feature.(*Broadcast)
	}
	return broadcasts, nil
}

func geomBoundsAround(point Point, radiusMeters float64) (*rtreego.Rect, error) {
	bound := geo.NewGeoBoundAroundPoint(geo.NewPoint(point[0], point[1]), radiusMeters)
	tl := bound.SouthWest()
	return rtreego.NewRect(rtreego.Point{tl[0], tl[1]}, []float64{bound.Width(), bound.Height()})
}
