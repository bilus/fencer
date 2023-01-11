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

type ErrFeatureNotFound[K feature.Key] struct {
	Key K
}

func (err ErrFeatureNotFound[K]) Error() string {
	return fmt.Sprintf("Feature not found (key = %q)", err.Key.String())
}

type featureByKey[K feature.Key, F feature.Feature[K]] map[K]F

// Index allows finding features by bounding box and custom queries.
// It is NOT thread-safe.
type Index[K feature.Key, F feature.Feature[K]] struct {
	rtree rtree.RTreeG[F]
	featureByKey[K, F]
}

// Creates a new index containing features.
func New[K feature.Key, F feature.Feature[K]](features []F) (*Index[K, F], error) {
	index := Index[K, F]{featureByKey: make(featureByKey[K, F])}
	for _, f := range features {
		if err := index.Insert(f); err != nil {
			return nil, err
		}
	}
	return &index, nil
}

// Insert adds a feature to the index.
func (index *Index[K, F]) Insert(f F) error {
	bounds := f.Bounds()
	index.rtree.Insert(bounds.Min, bounds.Max, f)
	index.featureByKey[f.Key()] = f
	return nil
}

// Delete removes a feature by its key.
func (index *Index[K, F]) Delete(key K) error {
	feature, ok := index.featureByKey[key]
	if !ok {
		return ErrFeatureNotFound[K]{Key: key}
	}
	delete(index.featureByKey, key)

	bounds := feature.Bounds()
	index.rtree.Delete(bounds.Min, bounds.Max, feature)
	return nil
}

// Update updates a feature (either its bounding rectangle or properties).
func (index *Index[K, F]) Update(f F) error {
	if err := index.Delete(f.Key()); err != nil {
		return err
	}
	return index.Insert(f)
}

// FindContaining returns features containing the given point.
func (index *Index[K, F]) FindContaining(point primitives.Point) ([]F, error) {
	size := math.SmallestNonzeroFloat64
	maxX := point[0] + size
	maxY := point[1] + size
	bounds := primitives.Rect{
		Min: point,
		Max: primitives.Point{maxX, maxY},
	}
	return index.Query(
		&bounds,
		query.Build[K, F]().Where(query.Contains[K, F]{Point: point}).Query(),
	)
}

// Intersect returns features whose bounding boxes intersect the given bounding box.
func (index *Index[K, F]) Intersect(bounds *primitives.Rect) ([]F, error) {
	return index.Query(bounds, query.Build[K, F]().Query())
}

// Query returns features with bounding boxes intersecting the specified bounding box and matching the provided query.
func (index *Index[K, F]) Query(bounds *primitives.Rect, query query.Query[K, F]) ([]F, error) {
	candidates := make([]F, 0)
	index.rtree.Search(bounds.Min, bounds.Max, func(min, max primitives.Point, f F) bool {
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
func (index *Index[K, F]) Lookup(key K) ([]F, error) {
	f, ok := index.featureByKey[key]
	if ok {
		return []F{f}, nil
	} else {
		return nil, nil
	}
}

// Keys returns a set containing all keys.
func (index *Index[K, F]) Keys() *hashset.Set[K] {
	keys := hashset.New(uint64(len(index.featureByKey)),
		func(l K, r K) bool { return l == r },
		func(k K) uint64 { return generic.HashString(k.String()) },
	)
	for key := range index.featureByKey {
		keys.Put(key)
	}
	return keys
}

// Size returns the number of features in the index.
func (index *Index[K, F]) Size() int {
	return len(index.featureByKey)
}

// lookupOne returns one feature based on its key or error.
func (index *Index[K, F]) lookupOne(key K) (F, error) {
	f, ok := index.featureByKey[key]
	if ok {
		return f, nil
	} else {
		return f, ErrFeatureNotFound[K]{Key: key}
	}
}
