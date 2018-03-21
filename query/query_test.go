package query_test

import (
	"fmt"
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/query"
	"github.com/bilus/fencer/test/fixtures"
	"sort"
	"strings"
	// "testing"
)

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_preconditionsUsingPredicates() {
	query := query.Build().Where(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return feature.(*fixtures.Country).Population > 10000, nil
		}),
	).Query()
	for _, country := range fixtures.Countries {
		query.Scan(country)
	}
	fmt.Println("Countries with population > 10000:", len(query.Distinct()))
	// Output: Countries with population > 10000: 4
}

type PopulationGreaterThan struct {
	threshold int
}

func (p PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	return feature.(*fixtures.Country).Population > p.threshold, nil
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_preconditionsUsingStructs() {
	// This is how you implement struct conditions:
	//
	//   type PopulationGreaterThan struct {
	//   	  threshold int
	//   }

	//   func (p PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	//   	  return feature.(*fixtures.Country).Population > p.threshold, nil
	//   }

	query := query.Build().Where(PopulationGreaterThan{10000}).Query()
	for _, country := range fixtures.Countries {
		query.Scan(country)
	}
	fmt.Println("Countries with population > 10000:", len(query.Distinct()))
	printNamesSorted(query.Distinct())
	// Output:
	// Countries with population > 10000: 4
	// Nauru
	// Poland
	// Tuvalu
	// Ukraine
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_conjunction() {
	// Both preconditions must match.
	q := query.Build()
	q.Where(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return feature.(*fixtures.Country).Population > 10000, nil
		}))
	q.Where(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return strings.HasPrefix(feature.(*fixtures.Country).Name, "T"), nil
		}),
	)
	query := q.Query()
	for _, country := range fixtures.Countries {
		query.Scan(country)
	}
	fmt.Println("Countries with population > 10000 with name beginning with T:", len(query.Distinct()))
	printNamesSorted(query.Distinct())
	// Output:
	// Countries with population > 10000 with name beginning with T: 1
	// Tuvalu
}

type GroupByRegion struct{}

func (GroupByRegion) Map(match *query.Match) (*query.Match, error) {
	match.Replace(match.Feature.(*fixtures.Country).Region)
	return match, nil
}

type MostPopulated struct{}

func (MostPopulated) Reduce(result *query.Result, match *query.Match) error {
	for _, key := range match.ResultKeys {
		err := result.Update(key, func(entry *query.ResultEntry) error {
			if len(entry.Features) == 0 {
				entry.Features = []feature.Feature{match.Feature}
				return nil
			}

			// In production you'd probably want to use something more robust :>
			existingCountry := entry.Features[0].(*fixtures.Country)
			currentCountry := match.Feature.(*fixtures.Country)
			if existingCountry.Population < currentCountry.Population {
				entry.Features = []feature.Feature{match.Feature}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_groupingResults() {
	qb := query.Build()
	stream := qb.StreamTo(MostPopulated{})
	stream.Map(GroupByRegion{})
	query := qb.Query() // Map(MostPopulatedByRegion{}).Query()
	for _, country := range fixtures.Countries {
		query.Scan(country)
	}
	fmt.Println("Most populated countries per region:", len(query.Distinct()))
	printNamesSorted(query.Distinct())
	// Output:
	// Most populated countries per region: 3
	// Nauru
	// Niue
	// Ukraine
}

type DecliningPopulation struct{}

func (DecliningPopulation) Map(match *query.Match) (*query.Match, error) {
	country := match.Feature.(*fixtures.Country)
	if country.Change < 0 {
		match.AddKey("declining")
	}
	return match, nil
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_complexAggregation() {
	qb := query.Build()
	stream := qb.StreamTo(MostPopulated{})
	stream.Map(GroupByRegion{})
	stream.Map(DecliningPopulation{})
	query := qb.Query() // Map(MostPopulatedByRegion{}).Query()
	for _, country := range fixtures.Countries {
		query.Scan(country)
	}
	// Output will include Poland which isn't the largest country in its region
	// but it's the largest one with declining population.
	fmt.Println("Most populated countries per region and most populated country with declining population:", len(query.Distinct()))
	printNamesSorted(query.Distinct())
	// Output:
	// Most populated countries per region and most populated country with declining population: 4
	// Nauru
	// Niue
	// Poland
	// Ukraine
}

func printNamesSorted(features []feature.Feature) {
	names := make([]string, 0)
	for _, f := range features {
		names = append(names, f.(*fixtures.Country).Name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}
