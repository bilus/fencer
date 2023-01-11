package query

import (
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/primitives"
)

// Contains is an example condition. It accepts all features containing a given point.
type Contains[K feature.Key, F feature.Feature[K]] struct {
	primitives.Point
}

func (c Contains[K, F]) IsMatch(feature F) (bool, error) {
	return feature.Contains(c.Point)
}

// Pred is a predicate condition letting use a function instead of creating a structure
// implementing the Condition interface.
type Pred[K feature.Key, F feature.Feature[K]] func(feature F) (bool, error)

func (p Pred[K, F]) IsMatch(feature F) (bool, error) {
	return p(feature)
}
