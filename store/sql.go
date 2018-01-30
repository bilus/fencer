package store

import (
	"database/sql"
	"errors"
	pq "github.com/mc2soft/pq-types"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var MissingDataError = errors.New("Missing broadcast data")

var LoadBroadcastsQuery = "SELECT id, broadcast_type, baseline_data, ST_Extent(coverage_area::geometry)::box2d, ST_AsGeoJson(coverage_area::geometry), freq, country, eid, pi_code FROM broadcasts GROUP BY id"

func NewBroadcastFromRow(rows *sql.Rows) (*Broadcast, error) {
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
	if geoJson.Valid && broadcastType.Valid && baselineData.Valid {
		var covArea geom.T
		if err := geojson.Unmarshal([]byte(geoJson.String), &covArea); err != nil {
			return nil, err
		}
		if broadcast, err := NewBroadcast(id, broadcastType.String, baselineData.String,
			optionalInt64(freq), optionalString(eid), optionalString(country), optionalString(piCode),
			boundingBox, covArea); err != nil {
			return nil, err
		} else {
			return broadcast, nil
		}
	} else {
		return nil, MissingDataError
	}
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
