package flagship

// VisitorOptions represents the visitor options of the Flagship SDK
type VisitorOptions struct {
	IsAuthenticated bool
}

// VisitorOptionBuilder is a func type to set options to the VisitorOptions.
type VisitorOptionBuilder func(*VisitorOptions)

// WithAuthenticated sets the is authenticated options of the visitor
func WithAuthenticated(isAuthenticated bool) VisitorOptionBuilder {
	return func(f *VisitorOptions) {
		f.IsAuthenticated = isAuthenticated
	}
}

// BuildVisitorOptions fill out the FlagshipOption struct from option builders
func (f *VisitorOptions) BuildVisitorOptions(visitorOptions ...VisitorOptionBuilder) {
	// extract options
	for _, opt := range visitorOptions {
		opt(f)
	}
}
