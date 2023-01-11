package query

import (
	"github.com/bilus/fencer/feature"
)

// defaultFilter accepts all features.
type defaultFilter[K feature.Key, F feature.Feature[K]] struct{}

func (defaultFilter[K, F]) IsMatch(feature F) (bool, error) {
	return true, nil
}

// defaultAggregator keeps one result per feature key.
type defaultAggregator[K feature.Key, F feature.Feature[K]] struct{}

func (defaultAggregator[K, F]) Map(match *Match[K, F]) (*Match[K, F], error) {
	match.AddKey(match.Feature.Key())
	return match, nil
}

func (defaultAggregator[K, F]) Reduce(result *Result[K, F], match *Match[K, F]) error {
	for _, k := range match.ResultKeys {
		err := result.Update(k, func(entry *ResultEntry[K, F]) error {
			entry.Features = []F{match.Feature}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
