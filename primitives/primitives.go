// Package primitives contains basic spatial types.
package primitives

import (
	"github.com/bilus/rtreego"
)

// Rect represents a bounding rectangle.
type Rect rtreego.Rect

func NewRect(p Point, lengths ...float64) (*Rect, error) {
	r, err := rtreego.NewRect(rtreego.Point(p), lengths)
	if err != nil {
		return nil, err
	}
	return (*Rect)(r), nil
}

func (r *Rect) Equal(other *Rect) bool {
	return ((*rtreego.Rect)(r)).Equal((*rtreego.Rect)(other))
}

// Point represents a point.
type Point rtreego.Point

// MinDist returns a minimum distance from the point to a rectangle.
func (p Point) MinDist(r *Rect) float64 {
	return rtreego.Point(p).MinDist((*rtreego.Rect)(r))
}
