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

// Preconditions sets query pre-conditions.
func (builder *builder) Preconditions(conditions ...Condition) *builder {
	builder.query.Preconditions = conditions
	return builder
}

// Filters sets query filters.
func (builder *builder) Filters(filters ...Filter) *builder {
	builder.query.Filters = filters
	return builder
}

// Reducer sets query reducer.
func (builder *builder) Reducer(reducer Reducer) *builder {
	builder.query.Reducer = reducer
	return builder
}
