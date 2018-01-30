package store

import (
	"database/sql"
	"github.com/bilus/gogeos/geos"
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

func FilterContaining(point *geos.Geometry) rtreego.Filter {
	return func(results []rtreego.Spatial, object rtreego.Spatial) (refuse, abort bool) {
		if object.(*Broadcast).Contains(point) {
			return false, false
		} else {
			return true, false
		}
	}
}

func (store *BroadcastStore) FindBroadcasts(point Point) ([]rtreego.Spatial, error) {
	geosPoint, err := geos.NewPoint(geos.NewCoord(point[0], point[1]))
	if err != nil {
		return nil, err
	}

	p, err := rtreego.NewRect(rtreego.Point(point), []float64{0.01, 0.01})
	if err != nil {
		return nil, err
	}
	broadcasts := store.SearchIntersect(p, FilterContaining(geosPoint))
	return broadcasts, nil
}

func (store *BroadcastStore) FindClosestBroadcasts(point Point, radiusMeters float64, filters []Filter) ([]*Broadcast, error) {
	bounds, err := geomBoundsAround(point, radiusMeters)
	if err != nil {
		log.Fatal(err)
	}

	candidates := store.SearchIntersect(bounds)
	if len(candidates) == 0 {
		return nil, nil
	}

	query := NeighbourQuery{point, filters, make(map[ResultKey]Match)}
	for _, candidate := range candidates {
		err := query.Scan(candidate.(*Broadcast))
		if err != nil {
			return nil, err
		}
	}
	return query.GetMatchingBroadcasts(), nil
}

func geomBoundsAround(point Point, radiusMeters float64) (*rtreego.Rect, error) {
	bound := geo.NewGeoBoundAroundPoint(geo.NewPoint(point[0], point[1]), radiusMeters)
	tl := bound.SouthWest()
	return rtreego.NewRect(rtreego.Point{tl[0], tl[1]}, []float64{bound.Width(), bound.Height()})
}
