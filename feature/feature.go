package feature

import (
	"github.com/bilus/fencer/primitives"
)

// Key uniquely identifies a feature.
type Key interface {
	String() string
}

// Feature represents a spatial object.
type Feature interface {
	Bounds() *primitives.Rect
	Contains(point primitives.Point) (bool, error)
	Key() Key
}
