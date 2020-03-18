package index

import (
	"fmt"
	"math"

	"github.com/bilus/rtreego"
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/primitives"
	"github.com/bilus/fencer/query"
)

type ErrFeatureNotFound struct {
	Key feature.Key
}

func (err ErrFeatureNotFound) Error() string {
	return fmt.Sprintf("Feature not found (key = %q)", err.Key)
}

type rtreegoFeatureAdapter struct {
	key feature.Key
	bounds *rtreego.Rect
}

func newRtreegoFeatureAdapter(feature feature.Feature) rtreegoFeatureAdapter {
	return rtreegoFeatureAdapter{
		key: feature.Key(),
		bounds: (*rtreego.Rect)(feature.Bounds()),
	}
}


func (a rtreegoFeatureAdapter) Bounds() *rtreego.Rect {
	return a.bounds
}

type featureByKey map[feature.Key]feature.Feature

// Index allows finding features by bounding box and custom queries.
// It is NOT thread-safe.
type Index struct {
	*rtreego.Rtree
	featureByKey
}

// Creates a new index containing features.
func New(features []feature.Feature) (*Index, error) {
	index := Index{rtreego.NewTree(2, 5, 20), make(featureByKey)}
	for _, f := range features {
		if err := index.Insert(f); err != nil {
			return nil, err
		}
	}
	return &index, nil
}

// Insert adds a feature to the index.
func (index *Index) Insert(f feature.Feature) error {
	index.Rtree.Insert(newRtreegoFeatureAdapter(f))
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
	ok = index.Rtree.DeleteWithComparator(newRtreegoFeatureAdapter(feature),
		func (l rtreego.Spatial, r rtreego.Spatial) bool {
			lf, ok := l.(rtreegoFeatureAdapter)
			if !ok {
				panic("Internal error in Index.Delete")
			}
			rf, ok := r.(rtreegoFeatureAdapter)
			if !ok {
				panic("Internal error in Index.Delete")
			}
			return lf.key == rf.key
		})
	if !ok {
		return ErrFeatureNotFound{Key: key}
	}
	return nil
}

// Update updates a feature (either its bounding rectangle or properties).
func (index *Index) Update(f feature.Feature) error {
	existing, ok := index.featureByKey[f.Key()]
	if !ok {
		return ErrFeatureNotFound{Key: f.Key()}
	}
	if !existing.Bounds().Equal(f.Bounds()) {
		// Bounds changed, re-insert.
		if err := index.Delete(existing.Key()); err != nil {
			return err
		}

		return index.Insert(f)
	}
	// Bounds haven't changed
	index.featureByKey[f.Key()] = f
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
		feature, err := index.lookupOne(candidate.(rtreegoFeatureAdapter).key)
		if err != nil {
			return nil, err
		}
		err = query.Scan(feature)
		if err != nil {
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

// lookupOne returns one feature based on its key or error.
func (index *Index) lookupOne(key feature.Key) (feature.Feature, error) {
	f, ok := index.featureByKey[key]
	if ok {
		return f, nil

	} else {
		return nil, ErrFeatureNotFound{Key: key}
	}

}
