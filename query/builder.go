package query

type queryBuilder struct {
	query Query
}

// Build returns a new query builder.
func Build() *queryBuilder {
	return &queryBuilder{}
}

// Query returns a complete constructed query.
func (builder *queryBuilder) Query() Query {
	if len(builder.query.Conditions) == 0 {
		builder.Where(defaultFilter{})
	}
	if len(builder.query.Aggregators) == 0 {
		builder.Aggregate(defaultAggregator{})
	}
	builder.query.results = &Result{
		entries: make(map[ResultKey]*ResultEntry),
	}
	return builder.query
}

// Where adds a filter to the query. Multiple filters act as a logical AND.
func (builder *queryBuilder) Where(condition Condition) *queryBuilder {
	builder.query.Conditions = append(builder.query.Conditions, condition)
	return builder
}

// Aggregate adds a new aggregator.
func (builder *queryBuilder) Aggregate(aggregator Aggregator) *queryBuilder {
	builder.query.Aggregators = append(builder.query.Aggregators, aggregator)
	return builder
}

// StreamTo creates a new aggregator stream and returns its builder.
func (builder *queryBuilder) StreamTo(reducer Reducer) *streamBuilder {
	stream := &StreamAggregator{
		Reducer: reducer,
	}
	builder.query.Aggregators = append(builder.query.Aggregators, stream)
	return &streamBuilder{stream}
}

type streamBuilder struct {
	stream *StreamAggregator
}

// Map adds a new mapper to the stream; mappers form a sequence with each consecutive mappers
// transforming match received from the previous one.
func (builder *streamBuilder) Map(mapper Mapper) *streamBuilder {
	builder.stream.Mappers = append(builder.stream.Mappers, mapper)
	return builder
}
