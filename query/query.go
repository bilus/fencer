package query

import (
	"github.com/bilus/fencer/feature"
)

type ResultKey interface{}

type Condition interface {
	IsMatch(feature feature.Feature) (bool, error)
}

type Distinctor interface {
	DistinctKey(feature feature.Feature) ResultKey
}

type Reducer interface {
	Reduce(matches map[ResultKey]*Match, match *Match) error
}

type Filter interface {
	Condition
	Distinctor
}

type Aggregator interface {
	Map(feature feature.Feature) ([]*Match, error)
	Reduce(matches map[ResultKey]*Match, match *Match) error
}

type Match struct {
	ResultKey
	Features []feature.Feature
	Cache    interface{}
}

func NewMatch(resultKey ResultKey, features ...feature.Feature) *Match {
	return &Match{
		ResultKey: resultKey,
		Features:  features,
	}
}

func (match *Match) Merge(feature feature.Feature) error {
	match.Features = append(match.Features, feature)
	return nil
}

type Mapper interface {
	Map(feature feature.Feature) ([]*Match, error)
}

// Query is a nearest neighour query returning features matching the filters,
// that are closest to the Point, at most one Match per ResultKey.
//
// Precedence:
// Precondition0 AND Precondition1 AND ... PreconditionN AND (Filter0 OR Filter1 OR ... FilterN)
type Query struct {
	// Mapper
	Preconditions []Condition // Logical conjunction.
	Filters       []Filter    // Logical disjunction.
	Reducer
	Aggregators []Aggregator
	matches     map[ResultKey]*Match
}

// Scan sends a feature through the query pipeline, first rejecting it unless
// all preconditions (conjunction step) match, then applying each of the filters
// and finally performing a reduce step to update query results.
func (q *Query) Scan(feature feature.Feature) error {
	isMatch, err := allMatch(q.Preconditions, feature)
	if err != nil {
		return err
	}
	if !isMatch {
		return nil
	}
	for _, aggregator := range q.Aggregators {
		matches, err := aggregator.Map(feature)
		if err != nil {
			return err
		}
		for _, match := range matches {
			if err := aggregator.Reduce(q.matches, match); err != nil {
				return err
			}
		}
	}

	// matches, err := filter(q.Filters, feature)
	// if err != nil {
	// 	return err
	// }
	// if len(matches) == 0 {
	// 	return nil
	// }
	// for _, match := range matches {
	// 	if err := q.Reducer.Reduce(q.matches, match); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

// Distinct returns distinct features matching the query.
func (q *Query) Distinct() []feature.Feature {
	features := make([]feature.Feature, 0)
	matched := make(map[feature.Key]struct{})
	for _, match := range q.matches {
		for _, feature := range match.Features {
			key := feature.Key()
			_, isMatched := matched[key]
			if !isMatched {
				features = append(features, feature)
				matched[key] = struct{}{}
			}
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

// type FilterMapper struct {
// 	Filter
// }

// func (m FilterMapper) Map(feature feature.Feature) ([]Match, error) {

// }

func filter(filters []Filter, feature feature.Feature) ([]*Match, error) {
	matchMap := make(map[ResultKey]*Match)
	for _, filter := range filters {
		isMatch, err := filter.IsMatch(feature)
		if err != nil {
			return nil, err
		}
		if isMatch {
			key := filter.DistinctKey(feature)
			match := matchMap[key]
			if match == nil {
				matchMap[key] = NewMatch(key, feature)
			} else {
				match.Merge(feature)
			}
		}
	}
	matches := make([]*Match, 0, len(matchMap))
	for key, match := range matchMap {
		// We are not setting the key ^ for performance reasons.
		match.ResultKey = key
		matches = append(matches, match)
	}
	return matches, nil
}

// +Change reducer interface so it works with Matches containing keys.
// Wrap Filter in FilterMapper so it emits matches.
// Wrap preconditions in PreconditionMapper so it emits matches.
// Think.
