package main

import (
	"database/sql"
	"github.com/bilus/fencer/countries"
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/query"
	"github.com/bilus/fencer/store"
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

type MinDistanceReducer struct {
	store.Point
}

func (r MinDistanceReducer) Reduce(matches map[query.ResultKey]query.Match, keys []query.ResultKey, feature feature.Feature) error {
	dist, err := feature.(*store.Broadcast).MinDistance(r.Point)
	if err != nil {
		return err
	}
	newMatch := query.Match{feature, dist}
	for _, key := range keys {
		existingMatch, exists := matches[key]
		if !exists || dist < existingMatch.Cache.(float64) {
			matches[key] = newMatch
		}
	}
	return nil
}

type TakeAllReducer struct{}

func (TakeAllReducer) Reduce(matches map[query.ResultKey]query.Match, keys []query.ResultKey, feature feature.Feature) error {
	newMatch := query.Match{feature, struct{}{}}
	for _, key := range keys {
		matches[key] = newMatch
	}
	return nil
}

type BroadcastType store.BroadcastType   // Only so we can use it as ResultKey.
func (s BroadcastType) ActsAsResultKey() {}

type MatchBroadcastTypes struct {
	BroadcastTypes []store.BroadcastType
}

func (co MatchBroadcastTypes) IsMatch(feature feature.Feature) (bool, error) {
	broadcast := feature.(*store.Broadcast)
	for _, broadcastType := range co.BroadcastTypes {
		if broadcastType == broadcast.BroadcastType {
			return true, nil
		}
	}
	return false, nil
}

type Freq store.Freq          // Only so we can use it as ResultKey.
func (Freq) ActsAsResultKey() {}

type MatchFreqs struct {
	Frequencies []store.Freq
}

func (f MatchFreqs) IsMatch(feature feature.Feature) (bool, error) {
	broadcast := feature.(*store.Broadcast)
	if broadcast.Freq == nil {
		return false, nil
	}
	for _, freq := range f.Frequencies {
		if freq == *broadcast.Freq {
			return true, nil
		}
	}
	return false, nil
}

func (f MatchFreqs) DistinctKey(feature feature.Feature) query.ResultKey {
	broadcast := feature.(*store.Broadcast)
	// Safe to dereference, DistinctKey never gets called if IsMatch returns false.
	return Freq(*broadcast.Freq)
}

type DAB struct {
	Country store.Country
	Eid     store.Eid
}

type MatchDABs struct {
	RDSs []DAB
}

func (DAB) ActsAsResultKey() {}

func (filter MatchDABs) IsMatch(feature feature.Feature) (bool, error) {
	broadcast := feature.(*store.Broadcast)
	if broadcast.Eid == nil || broadcast.Country == nil {
		return false, nil
	}
	for _, dab := range filter.RDSs {
		if dab.Country.Equals(*broadcast.Country) && dab.Eid.Equals(*broadcast.Eid) {
			return true, nil
		}
	}
	return false, nil
}

func (MatchDABs) DistinctKey(feature feature.Feature) query.ResultKey {
	broadcast := feature.(*store.Broadcast)
	// Safe to dereference, DistinctKey never gets called if IsMatch returns false.
	return DAB{*broadcast.Country, *broadcast.Eid}
}

type RDS struct {
	Country store.Country
	PiCode  store.PiCode
	Freq    store.Freq
}

type MatchRDSs struct {
	RDSs []RDS
}

func (RDS) ActsAsResultKey() {}

func (filter MatchRDSs) IsMatch(feature feature.Feature) (bool, error) {
	broadcast := feature.(*store.Broadcast)
	if broadcast.Country == nil || broadcast.PiCode == nil || broadcast.Freq == nil {
		return false, nil
	}
	for _, rds := range filter.RDSs {
		if rds.Country.Equals(*broadcast.Country) && rds.PiCode.Equals(*broadcast.PiCode) && rds.Freq == *broadcast.Freq {
			return true, nil
		}
	}
	return false, nil
}

func (MatchRDSs) DistinctKey(feature feature.Feature) query.ResultKey {
	broadcast := feature.(*store.Broadcast)
	// Safe to dereference, DistinctKey never gets called if IsMatch returns false.
	return RDS{*broadcast.Country, *broadcast.PiCode, *broadcast.Freq}
}

func runExperiment(db *sql.DB) error {
	bs, err := store.LoadFromSQL(db)
	if err != nil {
		return err
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

	point := store.Point{13.4, 52.52} // Berlin
	dabs := []DAB{
		{country, "10C6"},
		{country, "10F2"},
	}
	rdss := []RDS{
		{country, "D3D8", 101000},
		{country, "D3D9", 98400},
	}
	types := []store.BroadcastType{"analog"}
	results, err := bs.FindClosestBroadcasts(
		point, radius,
		[]query.Condition{MatchBroadcastTypes{types}},
		[]query.Filter{MatchRDSs{rdss}, MatchDABs{dabs}},
		MinDistanceReducer{point},
	)
	if err != nil {
		return err
	}

	for _, result := range results {
		println(result.BroadcastId)
	}
	log.Printf("%v result(s).", len(results))
	return nil
}
