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
	if len(builder.query.Filters) == 0 {
		builder.Filter(defaultFilter{})
	}
	if builder.query.Reducer == nil {
		builder.Reducer(defaultReducer{})
	}
	if len(builder.query.Aggregators) == 0 {
		builder.Aggregate(defaultAggregator{})
	}
	builder.query.matches = make(map[ResultKey]*Match)
	return builder.query
}

// Precondition adds a precondition to the query.
func (builder *builder) Precondition(condition Condition) *builder {
	builder.query.Preconditions = append(builder.query.Preconditions, condition)
	return builder
}

// Filter adds a query filter.
func (builder *builder) Filter(filter Filter) *builder {
	builder.query.Filters = append(builder.query.Filters, filter)
	return builder
}

// Reducer sets query reducer.
func (builder *builder) Reducer(reducer Reducer) *builder {
	builder.query.Reducer = reducer
	return builder
}

func (builder *builder) Aggregate(aggregator Aggregator) *builder {
	builder.query.Aggregators = append(builder.query.Aggregators, aggregator)
	return builder
}
