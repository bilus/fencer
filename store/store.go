package store

import (
	"database/sql"
	"github.com/bilus/gogeos/geos"
	"github.com/bilus/rtreego"
	pq "github.com/mc2soft/pq-types"
	"github.com/paulmach/go.geo"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	"log"
)

type Point = rtreego.Point

type BroadcastStore struct {
	*rtreego.Rtree
}

func Load(db *sql.DB) (*BroadcastStore, error) {
	rt := BroadcastStore{rtreego.NewTree(2, 5, 20)}

	rows, err := db.Query("SELECT id, broadcast_type, baseline_data, ST_Extent(coverage_area::geometry)::box2d, ST_AsGeoJson(coverage_area::geometry) FROM broadcasts GROUP BY id")
	if err != nil {
		return nil, err
	}
	numSkipped := 0
	defer rows.Close()
	var id int64
	var broadcastType sql.NullString
	var baselineData sql.NullString
	var boundingBox pq.PostGISBox2D
	var geoJson sql.NullString
	for rows.Next() {
		if err := rows.Scan(&id, &broadcastType, &baselineData, &boundingBox, &geoJson); err != nil {
			return nil, err
		}
		if geoJson.Valid && broadcastType.Valid && baselineData.Valid {
			var covArea geom.T
			if err = geojson.Unmarshal([]byte(geoJson.String), &covArea); err != nil {
				return nil, err
			}
			if broadcast, err := NewBroadcast(id, broadcastType.String, baselineData.String, boundingBox, covArea); err != nil {
				// log.Printf("Skipping %v: %v", id, err)
			} else {
				rt.Insert(broadcast)
			}
		} else {
			numSkipped++
			// log.Printf("Skipping broadcast %v: missing data", id)
		}

	}
	log.Printf("Skipped: %v broadcasts", numSkipped)
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

type Singleton struct{}

type ClosestOutside struct {
	Point
}

func (s Singleton) ActsAsResultKey() {}

func (co ClosestOutside) IsMatch(broadcast *Broadcast) (bool, error) {
	dist, err := broadcast.MinDistance(co.Point)
	if err != nil {
		return false, err
	}
	return dist > 0, nil
}

func (ClosestOutside) GetResultKey(broadcast *Broadcast) ResultKey {
	return Singleton{}
}

func (rt *BroadcastStore) FindClosestBroadcasts(point Point) ([]*Broadcast, error) {
	bounds, err := geomBoundsAround(point, 1000)
	if err != nil {
		log.Fatal(err)
	}

	candidates := rt.SearchIntersect(bounds)
	if len(candidates) == 0 {
		return nil, nil
	}

	query := NeighbourQuery{point, ClosestOutside{point}, make(map[ResultKey]Match)}
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
