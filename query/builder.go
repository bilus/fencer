package query

type builder struct {
	query Query
}

// Build returns a new query builder.
func Build() *builder {
	return &builder{New(nil, nil, nil)}
}

func (builder *builder) Query() Query {
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
