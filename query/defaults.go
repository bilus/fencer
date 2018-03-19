package query

import (
	"github.com/bilus/fencer/feature"
)

type defaultFilter struct{}

func (defaultFilter) IsMatch(feature feature.Feature) (bool, error) {
	return true, nil
}

type defaultAggregator struct{}

func (defaultAggregator) Map(match *Match) (*Match, error) {
	match.AddKey(match.Feature.Key())
	return match, nil
}

func (defaultAggregator) Reduce(result *Result, match *Match) error {
	result.Replace(match)
	return nil
}
