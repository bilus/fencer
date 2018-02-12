// dev package contains development-related helper functions.
package dev

import (
	"github.com/paulmach/go.geo"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// RectToGeoJson converts a rectangle to its geojson representation.
func BoundToGeoJson(bound *geo.Bound) (string, error) {
	se := bound.SouthEast()
	ne := bound.NorthEast()
	nw := bound.NorthWest()
	sw := bound.SouthWest()
	poly := geom.NewPolygon(geom.XY)
	ring := geom.NewLinearRing(geom.XY)
	ring.MustSetCoords([]geom.Coord{{se[0], se[1]}, {ne[0], ne[1]}, {nw[0], nw[1]}, {sw[0], sw[1]}, {se[0], se[1]}})
	poly.Push(ring)
	b, err := geojson.Marshal(poly)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
