package store

import (
	"database/sql"
	"errors"
	"fmt"
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

type Broadcast struct {
	BroadcastId      int64
	BroadcastType    string
	BaselineData     string
	bounds           *rtreego.Rect
	coverageArea     []*geos.PGeometry
	combinedCoverage *geos.Geometry
}

func NewBroadcast(id int64, broadcastType string, baselineData string, bounds pq.PostGISBox2D, coverageArea geom.T) (*Broadcast, error) {
	multiPoly := coverageArea.(*geom.MultiPolygon)
	preparedCoverageAreaGeometries := make([]*geos.PGeometry, multiPoly.NumPolygons())
	coverageAreaGeometries := make([]*geos.Geometry, multiPoly.NumPolygons())
	for i := 0; i < multiPoly.NumPolygons(); i++ {
		geometry := polygonToGeometry(multiPoly.Polygon(i))
		coverageAreaGeometries[i] = geometry
		preparedCoverageAreaGeometries[i] = geometry.Prepare()
	}
	combinedCoverage, err := unionGeometries(coverageAreaGeometries)
	if err != nil {
		return nil, err
	}

	rtBounds, err := rtreego.NewRect(
		rtreego.Point{bounds.Min.Lon, bounds.Min.Lat},
		lengths(bounds),
	)
	if err != nil {
		return nil, err
	}
	return &Broadcast{
		id,
		broadcastType,
		baselineData,
		rtBounds,
		preparedCoverageAreaGeometries,
		combinedCoverage,
	}, nil
}

func unionGeometries(geometries []*geos.Geometry) (*geos.Geometry, error) {
	if len(geometries) == 0 {
		return nil, errors.New("No geometries")
	}
	result := geometries[0]
	var err error
	for _, geometry := range geometries[1:] {
		result, err = result.Union(geometry)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (b *Broadcast) Bounds() *rtreego.Rect {
	return b.bounds
}

func lengths(bounds pq.PostGISBox2D) []float64 {
	return []float64{
		bounds.Max.Lon - bounds.Min.Lon,
		bounds.Max.Lat - bounds.Min.Lat,
	}
}

func (b *Broadcast) Contains(point *geos.Geometry) bool {
	for _, geometry := range b.coverageArea {
		inter, err := geometry.Covers(point) // TODO: Prepare before contains
		if err != nil {
			log.Println(geometry)
			log.Println(point)
			panic(fmt.Sprintf("Ooops: %v", err))
		}
		if inter {
			return true
		}
	}
	return false
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

type Query struct {
	point Point
}

func (rt *BroadcastStore) FindClosestBroadcasts(point Point) ([]rtreego.Spatial, error) {
	bounds, err := geomBoundsAround(point, 1000)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(bounds.PointCoord(0))
	log.Println(bounds.PointCoord(1))
	log.Println(bounds.PointCoord(0) + bounds.LengthsCoord(0))
	log.Println(bounds.PointCoord(1) + bounds.LengthsCoord(1))

	candidates := rt.SearchIntersect(bounds)
	if len(candidates) == 0 {
		return nil, nil
	}
	broadcast := candidates[0]
	dist, err := trueDistance(point, broadcast)
	if err != nil {
		return nil, err
	}
	for _, candidate := range candidates[1:] {
		if point.MinDist(candidate.Bounds()) > dist {
			continue
		}
		candDist, err := trueDistance(point, candidate)
		if err != nil {
			return nil, err
		}
		if candDist < dist && candDist > 0 {
			dist = candDist
			broadcast = candidate
		}
	}

	// broadcasts := rt.NearestNeighbors(10, rtreego.Point(point), Query{point}.MinDistance(1000))
	return []rtreego.Spatial{broadcast}, nil
}

// func (rt *BroadcastStore) FindClosestBroadcastsApprox(point Point) ([]rtreego.Spatial, error) {
// 	return rt.NearestNeighbors(1, rtreego.Point(point), Query{point}.NonZeroDistance()), nil
// }

func trueDistance(point Point, broadcast rtreego.Spatial) (float64, error) {
	geosPoint, err := geos.NewPoint(geos.NewCoord(point[0], point[1]))
	if err != nil {
		return -1, err
	}

	minDist, err := geosPoint.Distance(broadcast.(*Broadcast).combinedCoverage)
	if err != nil {
		return -1, err
	}
	return minDist, nil
}

func geomBoundsAround(point Point, radiusMeters float64) (*rtreego.Rect, error) {
	bound := geo.NewGeoBoundAroundPoint(geo.NewPoint(point[0], point[1]), radiusMeters)
	tl := bound.SouthWest()
	return rtreego.NewRect(rtreego.Point{tl[0], tl[1]}, []float64{bound.Width(), bound.Height()})
}

func polygonToGeometry(geofence *geom.Polygon) *geos.Geometry {
	// Convert the outer shell to geos format.
	shell := geofence.LinearRing(0).Coords()
	shellGeos := geomToGeosCoords(shell)

	// TODO: Holes!
	// Convert each hole to geos format.
	// numHoles := geofence.NumLinearRings() - 1
	// holes := make([][]geos.Coord, numHoles)
	// for i := 0; i < numHoles; i++ {
	// 	holes[i] = geomToGeosCoords(geofence.LinearRing(i).Coords())
	// }

	return geos.Must(geos.NewPolygon(shellGeos)) //, holes...))
}

func geomToGeosCoord(coord geom.Coord) geos.Coord {
	return geos.Coord{
		X: coord.X(),
		Y: coord.Y(),
	}
}

func geomToGeosCoords(coords []geom.Coord) []geos.Coord {
	out := make([]geos.Coord, len(coords))
	for i := 0; i < len(coords); i++ {
		out[i] = geomToGeosCoord(coords[i])
	}
	return out
}
