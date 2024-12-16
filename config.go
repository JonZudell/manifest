package customs

// Formatter is used to output inspection results. For example, you could have
// an stdout formatter for local development and a GitHub formatter to post
// results to a Pull Request.
type Formatter interface {
	Format(source string, i *Import, r Result) error
}

type Configuration struct {
	// ConcurrentInspections is the number of inspections to run concurrently.
	ConcurrentInspections int
	// Formatter is used to output the customs.Result
	Formatter Formatter
}
