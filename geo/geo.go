package geo

import (
	"github.com/bilus/fencer/primitives"
	"github.com/paulmach/go.geo"
)

// NewBoundsAroundPoint creates a new bounding rectangle given a center point,
// and a distance from the center point in meters.
func NewBoundsAround(point primitives.Point, radius float64) (*primitives.Rect, error) {
	bound := geo.NewGeoBoundAroundPoint(geo.NewPoint(point[0], point[1]), radius)
	tl := bound.SouthWest()
	return primitives.NewRect(
		primitives.Point{tl[0], tl[1]},
		bound.Width(),
		bound.Height(),
	)
}
