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

func Load(db *sql.DB) (*BroadcastStore, error) {
	rt := BroadcastStore{rtreego.NewTree(2, 5, 20)}

	rows, err := db.Query(LoadBroadcastsQuery)
	if err != nil {
		return nil, err
	}
	numSkipped := 0
	defer rows.Close()
	for rows.Next() {
		if broadcast, err := NewBroadcastFromRow(rows); err != nil {
			numSkipped++
		} else {
			rt.Insert(broadcast)
		}
	}
	log.Printf("Skipped: %v broadcasts due to errors or missing data", numSkipped)
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &rt, nil
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

func (rt *BroadcastStore) FindBroadcasts(point Point) ([]rtreego.Spatial, error) {
	geosPoint, err := geos.NewPoint(geos.NewCoord(point[0], point[1]))
	if err != nil {
		return nil, err
	}

	p, err := rtreego.NewRect(rtreego.Point(point), []float64{0.01, 0.01})
	if err != nil {
		return nil, err
	}
	broadcasts := rt.SearchIntersect(p, FilterContaining(geosPoint))
	return broadcasts, nil
}

func (rt *BroadcastStore) FindClosestBroadcasts(point Point, radiusMeters float64, filter Filter) ([]*Broadcast, error) {
	bounds, err := geomBoundsAround(point, radiusMeters)
	if err != nil {
		log.Fatal(err)
	}

	candidates := rt.SearchIntersect(bounds)
	if len(candidates) == 0 {
		return nil, nil
	}

	query := NeighbourQuery{point, filter, make(map[ResultKey]Match)}
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
