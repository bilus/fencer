package index

import (
	"math"

	"github.com/bilus/rtreego"
	"go.bilus.io/fencer/feature"
	"go.bilus.io/fencer/primitives"
	"go.bilus.io/fencer/query"
)

type rtreegoFeatureAdapter struct {
	feature.Feature
}

func (a rtreegoFeatureAdapter) Bounds() *rtreego.Rect {
	return (*rtreego.Rect)(a.Feature.Bounds())
}

type featureByID map[feature.Key]feature.Feature

type Index struct {
	*rtreego.Rtree
	featureByID
}

// Creates a new index containing features.
func New(features []feature.Feature) (*Index, error) {
	index := Index{rtreego.NewTree(2, 5, 20), make(featureByID)}
	for _, f := range features {
		if err := index.Insert(f); err != nil {
			return nil, err
		}
	}
	return &index, nil
}

// Inserts adds a feature to the index.
func (index *Index) Insert(f feature.Feature) error {
	index.Rtree.Insert(rtreegoFeatureAdapter{f})
	index.featureByID[f.Key()] = f
	return nil
}

// FindContaining returns features containing the given point.
func (index *Index) FindContaining(point primitives.Point) ([]feature.Feature, error) {
	size := math.SmallestNonzeroFloat64
	bounds, err := primitives.NewRect(point, size, size)
	if err != nil {
		return nil, err
	}
	return index.Query(
		bounds,
		query.Build().Where(query.Contains{point}).Query(),
	)
}

// Intersect returns features whose bounding boxes intersect the given bounding box.
func (index *Index) Intersect(bounds *primitives.Rect) ([]feature.Feature, error) {
	return index.Query(bounds, query.Build().Query())
}

// Query returns features with bounding boxes intersecting the specified bounding box and matching the provided query.
func (index *Index) Query(bounds *primitives.Rect, query query.Query) ([]feature.Feature, error) {
	candidates := index.SearchIntersect((*rtreego.Rect)(bounds))
	if len(candidates) == 0 {
		return nil, nil
	}
	for _, candidate := range candidates {
		err := query.Scan(candidate.(rtreegoFeatureAdapter).Feature)
		if err != nil {
			return nil, err
		}
	}
	return query.Distinct(), nil
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
