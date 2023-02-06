package query

import "github.com/bilus/fencer/feature"

type QueryBuilder[K feature.Key, F feature.Feature[K]] struct {
	query Query[K, F]
}

// Build returns a new query builder.
func Build[K feature.Key, F feature.Feature[K]]() *QueryBuilder[K, F] {
	return &QueryBuilder[K, F]{}
}

// Query returns a complete constructed query.
func (builder *QueryBuilder[K, F]) Query() Query[K, F] {
	if len(builder.query.Conditions) == 0 {
		builder.Where(defaultFilter[K, F]{})
	}
	if len(builder.query.Aggregators) == 0 {
		builder.Aggregate(defaultAggregator[K, F]{})
	}
	builder.query.results = &Result[K, F]{
		entries: make(map[ResultKey]*ResultEntry[K, F]),
	}
	return builder.query
}

// Where adds a filter to the query. Multiple filters act as a logical AND.
func (builder *QueryBuilder[K, F]) Where(condition Condition[K, F]) *QueryBuilder[K, F] {
	builder.query.Conditions = append(builder.query.Conditions, condition)
	return builder
}

// Aggregate adds a new aggregator.
func (builder *QueryBuilder[K, F]) Aggregate(aggregator Aggregator[K, F]) *QueryBuilder[K, F] {
	builder.query.Aggregators = append(builder.query.Aggregators, aggregator)
	return builder
}

// StreamTo creates a new aggregator stream and returns its builder.
func (builder *QueryBuilder[K, F]) StreamTo(reducer Reducer[K, F]) *streamBuilder[K, F] {
	stream := &StreamAggregator[K, F]{
		Reducer: reducer,
	}
	builder.query.Aggregators = append(builder.query.Aggregators, stream)
	return &streamBuilder[K, F]{stream}
}

type streamBuilder[K feature.Key, F feature.Feature[K]] struct {
	stream *StreamAggregator[K, F]
}

// Map adds a new mapper to the stream; mappers form a sequence with each consecutive mappers
// transforming match received from the previous one.
func (builder *streamBuilder[K, F]) Map(mapper Mapper[K, F]) *streamBuilder[K, F] {
	builder.stream.Mappers = append(builder.stream.Mappers, mapper)
	return builder
}
