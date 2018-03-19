package query

import (
	"github.com/bilus/fencer/feature"
)

type ResultKey interface{}

type Result struct {
	Matches []*Match
	Meta    interface{}
}

func NewResult(matches ...*Match) *Result {
	return &Result{
		Matches: matches,
	}
}

func (result *Result) Merge(match *Match) error {
	result.Matches = append(result.Matches, match)
	return nil
}

func (result *Result) Replace(match *Match) error {
	result.Matches = match.ToSlice()
	return nil
}

type Condition interface {
	IsMatch(feature feature.Feature) (bool, error)
}

type Distinctor interface {
	DistinctKey(feature feature.Feature) ResultKey
}

type Reducer interface {
	Reduce(result *Result, match *Match) error
}

type Filter interface {
	Condition
	Distinctor
}

type Aggregator interface {
	Map(feature feature.Feature) ([]*Match, error)
	Reducer
}

type Match struct {
	ResultKey
	Feature feature.Feature
	Meta    interface{}
}

func NewMatch(resultKey ResultKey, feature feature.Feature) *Match {
	return &Match{
		ResultKey: resultKey,
		Feature:   feature,
	}
}

func (match *Match) ToSlice() []*Match {
	return []*Match{match}
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
	results     map[ResultKey]*Result
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
			result := q.results[match.ResultKey]
			if result == nil {
				result = NewResult()
				q.results[match.ResultKey] = result
			}
			if err := aggregator.Reduce(result, match); err != nil {
				return err
			}
		}
	}

	return nil
}

// Distinct returns distinct features matching the query.
func (q *Query) Distinct() []feature.Feature {
	features := make([]feature.Feature, 0)
	matched := make(map[feature.Key]struct{})
	for _, result := range q.results {
		for _, match := range result.Matches {
			key := match.Feature.Key()
			_, isMatched := matched[key]
			if !isMatched {
				features = append(features, match.Feature)
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
