package query_test

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"go.bilus.io/fencer/feature"
	"go.bilus.io/fencer/primitives"
	"go.bilus.io/fencer/query"
	// "testing"
)

type CountryID int

func (id CountryID) String() string {
	return strconv.Itoa(int(id))
}

type Country struct {
	ID           CountryID
	Name         string
	Population   int
	Change       float64
	Region       string
	BoundingRect *primitives.Rect
}

func (c *Country) Key() feature.Key {
	return feature.Key(c.ID)
}

func (c *Country) Contains(p primitives.Point) (bool, error) {
	return p.MinDist(c.BoundingRect) == 0, nil
}

func (c *Country) Bounds() *primitives.Rect {
	return c.BoundingRect
}

var countries = []*Country{
	{1, "Vatican City", 800, -0.011, "Europe", makeRect(bounds[0])},
	{2, "Tokelau", 1300, 0.014, "Polynesia", makeRect(bounds[1])},
	{3, "Niue", 1600, -0.004, "Polynesia", makeRect(bounds[2])},
	{4, "Tuvalu", 11200, 0.009, "Oceania", makeRect(bounds[3])},
	{5, "Nauru", 11300, 0.001, "Oceania", makeRect(bounds[4])},
	{6, "Poland", 38224, -0.001, "Europe", makeRect(bounds[5])},
	{7, "Ukraine", 44400, 0, "Europe", makeRect(bounds[6])},
}

// This example uses an example spatial feature implementation.
// See https://go.bilus.io/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_preconditionsUsingPredicates() {
	query := query.Build().Where(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return feature.(*Country).Population > 10000, nil
		}),
	).Query()
	for _, country := range countries {
		query.Scan(country)
	}
	fmt.Println("Countries with population > 10000:", len(query.Distinct()))
	// Output: Countries with population > 10000: 4
}

type PopulationGreaterThan struct {
	threshold int
}

func (p PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	return feature.(*Country).Population > p.threshold, nil
}

// This example uses an example spatial feature implementation.
// See https://go.bilus.io/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_preconditionsUsingStructs() {
	// This is how you implement struct conditions:
	//
	//   type PopulationGreaterThan struct {
	//   	  threshold int
	//   }

	//   func (p PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	//   	  return feature.(*Country).Population > p.threshold, nil
	//   }

	query := query.Build().Where(PopulationGreaterThan{10000}).Query()
	for _, country := range countries {
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
// See https://go.bilus.io/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_conjunction() {
	// Both preconditions must match.
	q := query.Build()
	q.Where(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return feature.(*Country).Population > 10000, nil
		}))
	q.Where(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return strings.HasPrefix(feature.(*Country).Name, "T"), nil
		}),
	)
	query := q.Query()
	for _, country := range countries {
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
	match.Replace(match.Feature.(*Country).Region)
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
			existingCountry := entry.Features[0].(*Country)
			currentCountry := match.Feature.(*Country)
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
// See https://go.bilus.io/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_groupingResults() {
	qb := query.Build()
	stream := qb.StreamTo(MostPopulated{})
	stream.Map(GroupByRegion{})
	query := qb.Query() // Map(MostPopulatedByRegion{}).Query()
	for _, country := range countries {
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
	country := match.Feature.(*Country)
	if country.Change < 0 {
		match.AddKey("declining")
	}
	return match, nil
}

// This example uses an example spatial feature implementation.
// See https://go.bilus.io/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_complexAggregation() {
	qb := query.Build()
	stream := qb.StreamTo(MostPopulated{})
	stream.Map(GroupByRegion{})
	stream.Map(DecliningPopulation{})
	query := qb.Query() // Map(MostPopulatedByRegion{}).Query()
	for _, country := range countries {
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

// Really, really rough boundaries generated by  following steps in: https://github.com/JamesChevalier/cities, importing into geojson.io and drawing bounding boxes around the polygons.

var bounds = [][]primitives.Point{
	// Vatican City
	{
		{12.44450569152832, 41.89978557507729},
		{12.459547519683836, 41.89978557507729},
		{12.459547519683836, 41.907946360630994},
		{12.44450569152832, 41.907946360630994},
	},
	// Tokelau
	{
		{-172.7874755859375, -9.66573839518868},
		{-170.947265625, -9.66573839518868},
		{-170.947265625, -8.303905908124174},
		{-172.7874755859375, -8.303905908124174},
	},
	// Niue
	{
		{-170.13702392578125, -19.265776189877485},
		{-169.5849609375, -19.265776189877485},
		{-169.5849609375, -18.818567424622376},
		{-170.13702392578125, -18.818567424622376},
	},
	// Tuvalu
	{
		{174.74853515625, -11.059820828563412},
		{180.296630859375, -11.059820828563412},
		{180.296630859375, -5.397273407690904},
		{174.74853515625, -5.397273407690904},
	},
	// Nauru
	{
		{166.79443359375, -0.6227752122036241},
		{167.07595825195312, -0.6227752122036241},
		{167.07595825195312, -0.4051174740026618},
		{166.79443359375, -0.4051174740026618},
	},
	// Poland
	{
		{14.04052734375, 48.922499263758255},
		{24.27978515625, 48.922499263758255},
		{24.27978515625, 54.99022172004893},
		{14.04052734375, 54.99022172004893},
	},
	// Ukraine
	{
		{21.665039062499996, 44.02442151965934},
		{40.341796875, 44.02442151965934},
		{40.341796875, 52.482780222078226},
		{21.665039062499996, 52.482780222078226},
	},
}

func makeRect(points []primitives.Point) *primitives.Rect {
	p := points[0]
	op := points[2]
	rect, err := primitives.NewRect(p, op[0]-p[0], op[1]-p[1])
	if err != nil {
		panic(err)
	}
	return rect
}

func printNamesSorted(features []feature.Feature) {
	names := make([]string, 0)
	for _, f := range features {
		names = append(names, f.(*Country).Name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}
