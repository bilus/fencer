package query

import (
	"go.bilus.io/fencer/feature"
)

// defaultFilter accepts all features.
type defaultFilter struct{}

func (defaultFilter) IsMatch(feature feature.Feature) (bool, error) {
	return true, nil
}

// defaultAggregator keeps one result per feature key.
type defaultAggregator struct{}

func (defaultAggregator) Map(match *Match) (*Match, error) {
	match.AddKey(match.Feature.Key())
	return match, nil
}

func (defaultAggregator) Reduce(result *Result, match *Match) error {
	for _, k := range match.ResultKeys {
		err := result.Update(k, func(entry *ResultEntry) error {
			entry.Features = []feature.Feature{match.Feature}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
