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

func (defaultReducer) Reduce(matches map[ResultKey]*Match, match *Match) error {
	matches[match.ResultKey] = match
	return nil
}

type defaultAggregator struct {
	defaultReducer
}

func (defaultAggregator) Map(feature feature.Feature) ([]*Match, error) {
	return []*Match{NewMatch(feature.Key(), feature)}, nil
}
