package index

import (
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/primitives"
	"github.com/bilus/fencer/query"
	"github.com/bilus/rtreego"
	"github.com/paulmach/go.geo"
)

type featureByID map[feature.Key]feature.Feature

type Index struct {
	*rtreego.Rtree
	featureByID
}

func New(features []feature.Feature) (*Index, error) {
	index := Index{rtreego.NewTree(2, 5, 20), make(featureByID)}
	for _, f := range features {
		if err := index.Insert(f); err != nil {
			return nil, err
		}
	}
	return &index, nil
}

// Inserts add a feature to the index.
func (index *Index) Insert(f feature.Feature) error {
	index.Rtree.Insert(f)
	index.featureByID[f.Key()] = f
	return nil
}

// Find returns a slice of features intersecting a square around a given `point` covering the `distance` radius.
// - point, distance - the range of a spatial query
// - preconditions - a conjunction of predicates, set to nil to pass all features matching the spatial query
// - filters - a disjunction of predicates filtering the result from predicates + definition of distinctness, set to nil to pass all features
// - reducer - defines how to combine results together (a reduce function), set to nil to simply store it under one or more keys returned by the filters
func (index *Index) Find(point primitives.Point, distance float64, preconditions []query.Condition, filters []query.Filter, reducer query.Reducer) ([]feature.Feature, error) {
	bounds, err := geomBoundsAround(point, distance)
	if err != nil {
		return nil, err
	}

	candidates := index.SearchIntersect(bounds)
	if len(candidates) == 0 {
		return nil, nil
	}
	query := query.New(preconditions, filters, reducer)
	for _, candidate := range candidates {
		err := query.Scan(candidate.(feature.Feature))
		if err != nil {
			return nil, err
		}
	}
	return query.MatchingFeatures(), nil
}

// Lookup returns a feature based on its key. It returns a slice containing one result or an empty slice if there's no match.
func (index *Index) Lookup(key feature.Key) ([]feature.Feature, error) {
	f := index.featureByID[key]
	if f != nil {
		return []feature.Feature{f}, nil

	} else {
		return nil, nil
	}

}

func geomBoundsAround(point primitives.Point, distance float64) (*rtreego.Rect, error) {
	bound := geo.NewGeoBoundAroundPoint(geo.NewPoint(point[0], point[1]), distance)
	tl := bound.SouthWest()
	return rtreego.NewRect(rtreego.Point{tl[0], tl[1]}, []float64{bound.Width(), bound.Height()})
}
