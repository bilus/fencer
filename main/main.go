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

type DAB struct {
	Country string
	Eid     string
}

type MatchDabs struct {
	RDSs []DAB
}

func (DAB) ActsAsResultKey() {}

func (filter MatchDabs) IsMatch(broadcast *store.Broadcast) (bool, error) {
	if broadcast.Eid == nil || broadcast.Country == nil {
		return false, nil
	}
	// log.Printf("Broadcast id=%v Eid=%v country=%v", broadcast.BroadcastId, *broadcast.Eid, *broadcast.Country)
	for _, dab := range filter.RDSs {
		if dab.Country == *broadcast.Country && dab.Eid == *broadcast.Eid {
			return true, nil
		}
	}
	return false, nil
}

func (MatchDabs) GetResultKey(broadcast *store.Broadcast) store.ResultKey {
	// Safe to dereference, GetResultKey never gets called if IsMatch returns false.
	return DAB{*broadcast.Country, *broadcast.Eid}
}

type RDS struct {
	Country string
	PiCode  string
	Freq    Freq
}

type MatchRDSs struct {
	RDSs []RDS
}

func (RDS) ActsAsResultKey() {}

func (filter MatchRDSs) IsMatch(broadcast *store.Broadcast) (bool, error) {
	if broadcast.Country == nil || broadcast.PiCode == nil || broadcast.Freq == nil {
		return false, nil
	}
	for _, rds := range filter.RDSs {
		if rds.Country == *broadcast.Country && rds.PiCode == *broadcast.PiCode && rds.Freq == Freq(*broadcast.Freq) {
			return true, nil
		}
	}
	return false, nil
}

func (MatchRDSs) GetResultKey(broadcast *store.Broadcast) store.ResultKey {
	// Safe to dereference, GetResultKey never gets called if IsMatch returns false.
	return RDS{*broadcast.Country, *broadcast.PiCode, Freq(*broadcast.Freq)}
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

	// point := store.Point{-74.0059413, 40.71DB27837} // New York
	// freqs := []Freq{1520, 1310}
	// results, err := bs.FindClosestBroadcasts(point, radius, MatchAnyFrequency{freqs})

	// point := store.Point{13.4, 52.52} // Berlin
	// dabs := []DAB{
	// 	{isoCountryCode, "10C6"},
	// 	{isoCountryCode, "10F2"},
	// }
	// results, err := bs.FindClosestBroadcasts(point, radius, MatchDABs{dabs})
	// if err != nil {
	// 	return err
	// }

	point := store.Point{13.4, 52.52} // Berlin
	rdss := []RDS{
		{isoCountryCode, "D3D8", 101000},
		{isoCountryCode, "D3D9", 98400},
	}
	results, err := bs.FindClosestBroadcasts(point, radius, MatchRDSs{rdss})
	if err != nil {
		return err
	}

	for _, result := range results {
		println(result.BroadcastId)
	}
	log.Printf("%v result(s).", len(results))
	return nil
}
