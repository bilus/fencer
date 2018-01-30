package main

import (
	"database/sql"
	"github.com/bilus/fencer/countries"
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

type Singleton struct{}

type ClosestOutside struct {
	store.Point
}

func (s Singleton) ActsAsResultKey() {}

func (co ClosestOutside) IsMatch(broadcast *store.Broadcast) (bool, error) {
	dist, err := broadcast.MinDistance(co.Point)
	if err != nil {
		return false, err
	}
	return dist > 0, nil
}

func (ClosestOutside) GetResultKey(broadcast *store.Broadcast) store.ResultKey {
	return Singleton{}
}

type Freq int64

type MatchAnyFrequency struct {
	Frequencies []Freq
}

func (Freq) ActsAsResultKey() {}

func (f MatchAnyFrequency) IsMatch(broadcast *store.Broadcast) (bool, error) {
	if broadcast.Freq == nil {
		return false, nil
	}
	for _, freq := range f.Frequencies {
		if freq == Freq(*broadcast.Freq) {
			return true, nil
		}
	}
	return false, nil
}

func (f MatchAnyFrequency) GetResultKey(broadcast *store.Broadcast) store.ResultKey {
	// Safe to dereference, GetResultKey never gets called if IsMatch returns false.
	return Freq(*broadcast.Freq)
}

type Dab struct {
	gcc string
	eid string
}

type MatchDabs struct {
	Dabs []Dab
}

func (Dab) ActsAsResultKey() {}

func (filter MatchDabs) IsMatch(broadcast *store.Broadcast) (bool, error) {
	if broadcast.Eid == nil || broadcast.Country == nil {
		return false, nil
	}
	log.Printf("Broadcast id=%v eid=%v country=%v", broadcast.BroadcastId, *broadcast.Eid, *broadcast.Country)
	for _, dab := range filter.Dabs {
		isoCountryCode, err := countries.GccToIso(dab.gcc)
		if err != nil {
			// TODO: Need more robust error handling.
			log.Printf("Problem handling gcc %v searching for broadcasts: %v", dab.gcc, err)
			continue
		}
		if isoCountryCode == *broadcast.Country && dab.eid == *broadcast.Eid {
			return true, nil
		}
	}
	return false, nil
}

func (MatchDabs) GetResultKey(broadcast *store.Broadcast) store.ResultKey {
	// Safe to dereference, GetResultKey never gets called if IsMatch returns false.
	return Dab{*broadcast.Country, *broadcast.Eid}
}

func runExperiment(db *sql.DB) error {
	bs, err := store.Load(db)
	if err != nil {
		return err
	}
	radius := 58000.0
	// point := store.Point{-74.0059413, 40.7127837} // New York
	// freqs := []Freq{1520, 1310}
	point := store.Point{13.4, 52.52} // Berlin
	dabs := []Dab{
		{"DE0", "10C6"},
		{"DE0", "10F2"},
	}
	// results, err := bs.FindClosestBroadcasts(point, radius, MatchAnyFrequency{freqs})

	results, err := bs.FindClosestBroadcasts(point, radius, MatchDabs{dabs})
	if err != nil {
		return err
	}

	for _, result := range results {
		println(result.BroadcastId)
	}
	log.Printf("%v result(s).", len(results))
	return nil
}
