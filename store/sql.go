package store

import (
	"database/sql"
	"errors"
	"github.com/bilus/fencer/feature"
	pq "github.com/mc2soft/pq-types"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var MissingDataError = errors.New("Missing broadcast data")

var loadBroadcastsQuery = "SELECT id, broadcast_type, baseline_data, ST_Extent(coverage_area::geometry)::box2d, ST_AsGeoJson(coverage_area::geometry), freq, country, eid, pi_code FROM broadcasts GROUP BY id"

func LoadBroadcastsFromSQL(db *sql.DB) ([]feature.Feature, int, error) {
	broadcasts := make([]feature.Feature, 0)
	rows, err := db.Query(loadBroadcastsQuery)
	if err != nil {
		return nil, 0, err
	}
	numSkipped := 0
	defer rows.Close()
	for rows.Next() {
		if broadcast, err := newBroadcastFromRow(rows); err != nil {
			numSkipped++
		} else {
			broadcasts = append(broadcasts, broadcast)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return broadcasts, numSkipped, nil
}

func newBroadcastFromRow(rows *sql.Rows) (feature.Feature, error) {
	var id int64
	var broadcastType sql.NullString
	var baselineData sql.NullString
	var boundingBox pq.PostGISBox2D
	var geoJson sql.NullString
	var freq sql.NullInt64
	var country sql.NullString
	var eid sql.NullString
	var piCode sql.NullString

	if err := rows.Scan(&id, &broadcastType, &baselineData, &boundingBox, &geoJson, &freq, &country, &eid, &piCode); err != nil {
		return nil, err
	}
	if !geoJson.Valid || !broadcastType.Valid || !baselineData.Valid {
		return nil, MissingDataError
	}
	var covArea geom.T
	if err := geojson.Unmarshal([]byte(geoJson.String), &covArea); err != nil {
		return nil, err
	}
	broadcast, err := NewBroadcast(
		BroadcastId(id),
		BroadcastType(broadcastType.String),
		baselineData.String,
		(*Freq)(optionalInt64(freq)),
		(*Eid)(optionalString(eid)),
		(*Country)(optionalString(country)),
		(*PiCode)(optionalString(piCode)),
		boundingBox,
		covArea)
	if err != nil {
		return nil, err
	}
	return broadcast, nil
}

func optionalInt64(i sql.NullInt64) *int64 {
	if i.Valid {
		return &i.Int64
	}
	return nil
}

func optionalString(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}
