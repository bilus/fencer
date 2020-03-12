package query

import (
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/primitives"
)

// Contains is an example condition. It accepts all features containing a given point.
type Contains struct {
	primitives.Point
}

func (c Contains) IsMatch(feature feature.Feature) (bool, error) {
	return feature.Contains(c.Point)
}

// Pred is a predicate condition letting use a function instead of creating a structure
// implementing the Condition interface.
type Pred func(feature feature.Feature) (bool, error)

func (p Pred) IsMatch(feature feature.Feature) (bool, error) {
	return p(feature)
}
