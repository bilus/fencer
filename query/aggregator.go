package query

import (
	"fmt"

	"github.com/bilus/fencer/feature"
)

// Mapper transforms a match, updating Match.ResultKeys.
type Mapper[K feature.Key, F feature.Feature[K]] interface {
	Map(match *Match[K, F]) (*Match[K, F], error)
}

// Reducer updates result based on a match.
type Reducer[K feature.Key, F feature.Feature[K]] interface {
	Reduce(result *Result[K, F], match *Match[K, F]) error
}

// Aggregator is a map-reduce operation. It allows filtering and aggregation of results.
type Aggregator[K feature.Key, F feature.Feature[K]] interface {
	Mapper[K, F]
	Reducer[K, F]
}

// StreamAggregator is a an aggregator supporting a sequence of mappers.
type StreamAggregator[K feature.Key, F feature.Feature[K]] struct {
	Mappers []Mapper[K, F]
	Reducer[K, F]
}

// Map takes a match and sends it through a sequence of mappers.
func (stream StreamAggregator[K, F]) Map(match *Match[K, F]) (*Match[K, F], error) {
	var err error
	for _, mapper := range stream.Mappers {
		match, err = mapper.Map(match)
		if err != nil {
			return nil, err
		}
		if match == nil {
			return nil, fmt.Errorf("Internal error: nil match returned from mapper %T", mapper)
		}
	}
	return match, nil
}
