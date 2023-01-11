package query

import (
	"fmt"

	"github.com/bilus/fencer/feature"
)

type Condition[K feature.Key, F feature.Feature[K]] interface {
	IsMatch(feature F) (bool, error)
}

type Match[K feature.Key, F feature.Feature[K]] struct {
	ResultKeys []ResultKey
	Feature    F
	Meta       interface{}
}

func (match *Match[K, F]) AddKey(resultKey ResultKey) {
	match.ResultKeys = append(match.ResultKeys, resultKey)
}

func (match *Match[K, F]) ReplaceKeys(resultKeys ...ResultKey) {
	match.ResultKeys = resultKeys
}

// Query holds configuration of a query pipeline for narrowing down spatial
// search results.
type Query[K feature.Key, F feature.Feature[K]] struct {
	Conditions  []Condition[K, F]  // Logical conjunction (AND).
	Aggregators []Aggregator[K, F] // Logical disjunction (OR).
	results     *Result[K, F]
}

// Scan sends a feature through the query pipeline, first rejecting it unless
// all preconditions (conjunction step) match, then applying each of the filters
// and finally performing a reduce step to update query results.
func (q *Query[K, F]) Scan(feature F) error {
	isMatch, err := allMatch[K, F](q.Conditions, feature)
	if err != nil {
		return err
	}
	if !isMatch {
		return nil
	}
	match := &Match[K, F]{
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
func (q *Query[K, F]) Distinct() []F {
	return q.results.distinct()
}

func allMatch[K feature.Key, F feature.Feature[K]](conditions []Condition[K, F], feature F) (bool, error) {
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
