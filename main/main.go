package main

import (
	"database/sql"
	"github.com/bilus/fencer/countries"
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/query"
	"github.com/bilus/fencer/store"
	"github.com/bilus/rtreego"
	_ "github.com/jackc/pgx/stdlib"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	db, err := sql.Open("postgres", "user=postgres password=mysecretpassword host=localhost port=5432 dbname=broadcasts sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return
	}

	err = runExperiment(db)
	if err != nil {
		log.Fatal(err)
	}
}

func runExperiment(db *sql.DB) error {
	bs, numSkipped, err := store.LoadFromSQL(db)
	if err != nil {
		return err
	}
	if numSkipped > 0 {
		log.Printf("Skipped: %v broadcasts due to errors or missing data", numSkipped)
	}
	radius := 58000.0
	isoCountryCode, err := countries.GccToIso("DE0")
	if err != nil {
		return err
	}
	country := store.Country(isoCountryCode)

	// point := store.Point{-74.0059413, 40.71DB27837} // New York
	// freqs := []Freq{1520, 1310}
	// results, err := bs.FindClosestBroadcasts(point, radius, MatchFreqs{freqs})

	// NEED TESTS!

	point := rtreego.Point{13.4, 52.52} // Berlin
	dabs := []DAB{
		{country, "10C6"},
		{country, "10F2"},
	}
	rdss := []RDS{
		{country, "D3D8", 101000},
		{country, "D3D9", 98400},
	}
	types := []store.BroadcastType{"analog"}
	results, err := bs.Find(
		point, radius,
		[]query.Condition{MatchBroadcastTypes{types}},
		[]query.Filter{MatchRDSs{rdss}, MatchDABs{dabs}},
		MinDistanceReducer{point},
	)
	if err != nil {
		return err
	}

	for _, result := range results {
		println(result.Key().(store.BroadcastId))
	}
	log.Printf("%v result(s).", len(results))
	return nil
}
