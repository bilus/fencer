package query

import (
	"fmt"
)

// Mapper transforms a match, updating Match.ResultKeys.
type Mapper interface {
	Map(match *Match) (*Match, error)
}

// Reducer updates result based on a match.
type Reducer interface {
	Reduce(result *Result, match *Match) error
}

// Aggregator is a map-reduce operation. It allows filtering and aggregation of results.
type Aggregator interface {
	Mapper
	Reducer
}

// StreamAggregator is a an aggregator supporting a sequence of mappers.
type StreamAggregator struct {
	Mappers []Mapper
	Reducer
}

// Map takes a match and sends it through a sequence of mappers.
func (stream StreamAggregator) Map(match *Match) (*Match, error) {
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
