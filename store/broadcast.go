package store

import (
	"errors"
	"fmt"
	"github.com/bilus/gogeos/geos"
	"github.com/bilus/rtreego"
	pq "github.com/mc2soft/pq-types"
	geom "github.com/twpayne/go-geom"
	"log"
)

type Broadcast struct {
	BroadcastId      int64
	BroadcastType    string
	BaselineData     string
	Freq             *int64
	bounds           *rtreego.Rect
	coverageArea     []*geos.PGeometry
	combinedCoverage *geos.Geometry
}

func NewBroadcast(id int64, broadcastType string, baselineData string, freq *int64, bounds pq.PostGISBox2D, coverageArea geom.T) (*Broadcast, error) {
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
		freq,
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

func (broadcast *Broadcast) MinDistance(point Point) (float64, error) {
	geosPoint, err := geos.NewPoint(geos.NewCoord(point[0], point[1]))
	if err != nil {
		return -1, err
	}

	minDist, err := geosPoint.Distance(broadcast.combinedCoverage)
	if err != nil {
		return -1, err
	}
	return minDist, nil
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
