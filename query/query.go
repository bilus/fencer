package query

import (
	"github.com/bilus/fencer/feature"
)

type ResultKey interface {
	ActsAsResultKey()
}

type Condition interface {
	IsMatch(feature feature.Feature) (bool, error)
}

type Distinctor interface {
	DistinctKey(feature feature.Feature) ResultKey
}

type Reducer interface {
	Reduce(matches map[ResultKey]Match, keys []ResultKey, feature feature.Feature) error
}

type Filter interface {
	Condition
	Distinctor
}

type Match struct {
	Feature feature.Feature
	Cache   interface{}
}

// Query is a nearest neighour query returning features matching the filters,
// that are closest to the Point, at most one Match per ResultKey.
//
// Precedence:
// Precondition0 AND Precondition1 AND ... PreconditionN AND (Filter0 OR Filter1 OR ... FilterN)
type Query struct {
	Preconditions []Condition // Logical conjunction.
	Filters       []Filter    // Logical disjunction.
	Reducer
	matches map[ResultKey]Match
}

func New(preconditions []Condition, filters []Filter, reducer Reducer) Query {
	if len(preconditions) == 0 {
		preconditions = []Condition{defaultFilter{}}
	}
	if len(filters) == 0 {
		filters = []Filter{defaultFilter{}}
	}
	if reducer == nil {
		reducer = defaultReducer{}
	}
	return Query{preconditions, filters, reducer, make(map[ResultKey]Match)}
}

func (q *Query) Scan(feature feature.Feature) error {
	match, err := allMatch(q.Preconditions, feature)
	if err != nil {
		return err
	}
	if !match {
		return nil
	}

	keys, err := filter(q.Filters, feature)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return q.Reducer.Reduce(q.matches, keys, feature)
}

func (q *Query) MatchingFeatures() []feature.Feature {
	features := make([]feature.Feature, 0)
	matched := make(map[feature.Key]struct{})
	for _, match := range q.matches {
		key := match.Feature.Key()
		_, isMatched := matched[key]
		if !isMatched {
			features = append(features, match.Feature)
			matched[key] = struct{}{}
		}
	}
	return features
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

func filter(filters []Filter, feature feature.Feature) ([]ResultKey, error) {
	keys := make([]ResultKey, 0)
	for _, filter := range filters {
		match, err := filter.IsMatch(feature)
		if err != nil {
			return nil, err
		}
		if match {
			keys = append(keys, filter.DistinctKey(feature))
		}
	}
	return keys, nil
}
