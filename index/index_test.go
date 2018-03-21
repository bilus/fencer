package index_test

import (
	"fmt"
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/geo"
	"github.com/bilus/fencer/index"
	"github.com/bilus/fencer/primitives"
	"github.com/bilus/fencer/query"
	"github.com/bilus/fencer/test/fixtures"
)

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_Lookup() {
	index, _ := index.New([]feature.Feature{&fixtures.Wroclaw, &fixtures.Szczecin})
	results, _ := index.Lookup(fixtures.CityID("wrocław"))
	fmt.Println(len(results), "match")
	// Output: 1 match
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_FindContaining() {
	index, _ := index.New([]feature.Feature{&fixtures.Wroclaw, &fixtures.Szczecin})
	location := primitives.Point{14.499678611755371, 53.41209631751399}
	results, _ := index.FindContaining(location)
	fmt.Println(len(results), "result:", results[0].(*fixtures.City).Name)
	// Output: 1 result: Szczecin
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_Intersect() {
	index, _ := index.New([]feature.Feature{&fixtures.Wroclaw, &fixtures.Szczecin})
	location := primitives.Point{14.499678611755371, 53.41209631751399}
	// A 1000kmx1000km bounding rectangle around the location so we match both cities.
	radius := 500000.0
	bounds, _ := geo.NewBoundsAround(location, radius)
	results, _ := index.Intersect(bounds)
	fmt.Println(len(results), "results")
	// Output: 2 results
}

type PopulationGreaterThan struct {
	Threshold int
}

func (c PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	return feature.(*fixtures.City).Population > c.Threshold, nil
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_Query_preconditions() {
	index, _ := index.New([]feature.Feature{&fixtures.Wroclaw, &fixtures.Szczecin})
	location := primitives.Point{14.499678611755371, 53.41209631751399}
	// A 1000kmx1000km bounding rectangle around the location so we match both cities.
	radius := 500000.0
	bounds, _ := geo.NewBoundsAround(location, radius)
	results, _ := index.Query(bounds, query.Build().Where(PopulationGreaterThan{500000}).Query())
	fmt.Println(len(results), "result")
	// Output: 1 result
}
