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

func (defaultResultKey) ActsAsResultKey() {}

type defaultReducer struct{}

func (defaultReducer) Reduce(matches map[ResultKey]Match, keys []ResultKey, feature feature.Feature) error {
	newMatch := NewMatch(feature)
	for _, key := range keys {
		matches[key] = newMatch
	}
	return nil
}
