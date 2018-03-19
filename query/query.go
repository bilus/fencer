package query

import (
	"fmt"
	"github.com/bilus/fencer/feature"
)

type Condition interface {
	IsMatch(feature feature.Feature) (bool, error)
}

type Reducer interface {
	Reduce(result *Result, match *Match) error
}

type Mapper interface {
	Map(match *Match) (*Match, error)
}

type Aggregator interface {
	Mapper
	Reducer
}

type StreamAggregator struct {
	Mappers []Mapper
	Reducer
}

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

func NewPipeline(reducer Reducer, mappers ...Mapper) StreamAggregator {
	return StreamAggregator{
		Mappers: mappers,
		Reducer: reducer,
	}
}

type Match struct {
	ResultKeys []ResultKey
	Feature    feature.Feature
	Meta       interface{}
}

func (match *Match) AddKey(resultKey interface{}) {
	match.ResultKeys = append(match.ResultKeys, resultKey)
}

func (match *Match) Replace(resultKey interface{}) {
	match.ResultKeys = []ResultKey{resultKey}
}

// Query holds configuration of a query pipeline for narrowing down spatial
// search results.
type Query struct {
	Conditions  []Condition  // Logical conjunction (AND).
	Aggregators []Aggregator // Logical disjunction (OR).
	results     *Result
}

// Scan sends a feature through the query pipeline, first rejecting it unless
// all preconditions (conjunction step) match, then applying each of the filters
// and finally performing a reduce step to update query results.
func (q *Query) Scan(feature feature.Feature) error {
	isMatch, err := allMatch(q.Conditions, feature)
	if err != nil {
		return err
	}
	if !isMatch {
		return nil
	}
	match := &Match{
		Feature: feature,
	}
	for _, aggregator := range q.Aggregators {
		match, err := aggregator.Map(match)
		if err != nil {
			return err
		}
		if match == nil {
			return fmt.Errorf("Internal error: nil match returned from aggregator %T", aggregator)
		}
		if len(match.ResultKeys) == 0 {
			// Rejected by Map.
			continue
		}
		if err := aggregator.Reduce(q.results, match); err != nil {
			return err
		}
	}

	return nil
}

// Distinct returns distinct features matching the query.
func (q *Query) Distinct() []feature.Feature {
	return q.results.distinct()
}

func allMatch(conditions []Condition, feature feature.Feature) (bool, error) {
	for _, condition := range conditions {
		match, err := condition.IsMatch(feature)
		if err != nil {
			return false, err
		}
		if !match {
			return false, nil
		}
	}
	return true, nil
}
