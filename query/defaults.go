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
