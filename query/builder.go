package query

type builder struct {
	query Query
}

// Build returns a new query builder.
func Build() *builder {
	return &builder{}
}

func (builder *builder) Query() Query {
	if len(builder.query.Preconditions) == 0 {
		builder.Precondition(defaultFilter{})
	}
	if len(builder.query.Aggregators) == 0 {
		builder.Aggregate(defaultAggregator{})
	}
	builder.query.results = &Result{
		entries: make(map[ResultKey]*ResultEntry),
	}
	return builder.query
}

// Precondition adds a precondition to the query.
func (builder *builder) Precondition(condition Condition) *builder {
	builder.query.Preconditions = append(builder.query.Preconditions, condition)
	return builder
}

func (builder *builder) Aggregate(aggregator Aggregator) *builder {
	builder.query.Aggregators = append(builder.query.Aggregators, aggregator)
	return builder
}
