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
