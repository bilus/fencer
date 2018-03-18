package query

import (
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/primitives"
)

type Contains struct {
	primitives.Point
}

func (c Contains) IsMatch(feature feature.Feature) (bool, error) {
	return feature.Contains(c.Point)
}

type Pred func(feature feature.Feature) (bool, error)

func (p Pred) IsMatch(feature feature.Feature) (bool, error) {
	return p(feature)
}
