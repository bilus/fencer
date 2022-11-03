package index

import (
	"fmt"
	"math"

	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/primitives"
	"github.com/bilus/fencer/query"
	"github.com/tidwall/rtree"
	"github.com/zyedidia/generic"
	"github.com/zyedidia/generic/hashset"
)

type ErrFeatureNotFound struct {
	Key feature.Key
}

func (err ErrFeatureNotFound) Error() string {
	return fmt.Sprintf("Feature not found (key = %q)", err.Key)
}

type featureByKey map[feature.Key]feature.Feature

// Index allows finding features by bounding box and custom queries.
// It is NOT thread-safe.
type Index struct {
	rtree rtree.RTreeG[feature.Feature]
	featureByKey
}

// Creates a new index containing features.
func New(features []feature.Feature) (*Index, error) {
	index := Index{featureByKey: make(featureByKey)}
	for _, f := range features {
		if err := index.Insert(f); err != nil {
			return nil, err
		}
	}
	return &index, nil
}

// Insert adds a feature to the index.
func (index *Index) Insert(f feature.Feature) error {
	bounds := f.Bounds()
	index.rtree.Insert(bounds.Min, bounds.Max, f)
	index.featureByKey[f.Key()] = f
	return nil
}

// Delete removes a feature by its key.
func (index *Index) Delete(key feature.Key) error {
	feature, ok := index.featureByKey[key]
	if !ok {
		return ErrFeatureNotFound{Key: key}
	}
	delete(index.featureByKey, key)

	bounds := feature.Bounds()
	index.rtree.Delete(bounds.Min, bounds.Max, feature)
	return nil
}

// Update updates a feature (either its bounding rectangle or properties).
func (index *Index) Update(f feature.Feature) error {
	if err := index.Delete(f.Key()); err != nil {
		return err
	}
	return index.Insert(f)
}

// FindContaining returns features containing the given point.
func (index *Index) FindContaining(point primitives.Point) ([]feature.Feature, error) {
	size := math.SmallestNonzeroFloat64
	maxX := point[0] + size
	maxY := point[1] + size
	bounds := primitives.Rect{
		Min: point,
		Max: primitives.Point{maxX, maxY},
	}
	return index.Query(
		&bounds,
		query.Build().Where(query.Contains{Point: point}).Query(),
	)
}

// Intersect returns features whose bounding boxes intersect the given bounding box.
func (index *Index) Intersect(bounds *primitives.Rect) ([]feature.Feature, error) {
	return index.Query(bounds, query.Build().Query())
}

// Query returns features with bounding boxes intersecting the specified bounding box and matching the provided query.
func (index *Index) Query(bounds *primitives.Rect, query query.Query) ([]feature.Feature, error) {
	candidates := make([]feature.Feature, 0)
	index.rtree.Search(bounds.Min, bounds.Max, func(min, max primitives.Point, f feature.Feature) bool {
		candidates = append(candidates, f)
		return true
	})
	if len(candidates) == 0 {
		return nil, nil
	}
	for _, feature := range candidates {
		if err := query.Scan(feature); err != nil {
			return nil, err
		}
	}
	return query.Distinct(), nil
}

// Lookup returns a feature based on its key. It returns a slice containing one result or an empty slice if there's no match.
func (index *Index) Lookup(key feature.Key) ([]feature.Feature, error) {
	f, ok := index.featureByKey[key]
	if ok && f != nil {
		return []feature.Feature{f}, nil

	} else {
		return nil, nil
	}
}

// Keys returns a set containing all keys.
func (index *Index) Keys() *hashset.Set[feature.Key] {
	keys := hashset.New(uint64(len(index.featureByKey)),
		func(l feature.Key, r feature.Key) bool { return l == r },
		func(k feature.Key) uint64 { return generic.HashString(k.String()) },
	)
	for key := range index.featureByKey {
		keys.Put(key)
	}
	return keys
}

// Size returns the number of features in the index.
func (index *Index) Size() int {
	return len(index.featureByKey)
}

// lookupOne returns one feature based on its key or error.
func (index *Index) lookupOne(key feature.Key) (feature.Feature, error) {
	f, ok := index.featureByKey[key]
	if ok {
		return f, nil

	} else {
		return nil, ErrFeatureNotFound{Key: key}
	}

}
