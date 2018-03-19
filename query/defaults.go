package query

import (
	"github.com/bilus/fencer/feature"
)

type defaultFilter struct{}

func (defaultFilter) IsMatch(feature feature.Feature) (bool, error) {
	return true, nil
}

func (defaultFilter) DistinctKey(feature feature.Feature) ResultKey {
	return defaultResultKey{feature.Key()}
}

type defaultResultKey struct {
	feature.Key
}

type defaultReducer struct{}

func (defaultReducer) Reduce(result *Result, match *Match) error {
	result.Replace(match)
	return nil
}

type defaultAggregator struct {
	defaultReducer
}

func (defaultAggregator) Map(feature feature.Feature) ([]*Match, error) {
	return NewMatch(feature.Key(), feature).ToSlice(), nil
}
