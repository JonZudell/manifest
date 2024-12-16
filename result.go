package customs

// Result is the result of a rule being run against a diff. Customs uses the
// result to determine if the PR passes and where to comment if configured to.
type Result struct {
	Pass     bool      `json:"pass"`
	Error    string    `json:"error"`
	Comments []Comment `json:"comments"`
}

type Severity string

const (
	// SeverityInfo does not fail the build and does not emhphasize the message.
	SeverityInfo Severity = "Info"
	// SeverityWarn does not fail the build, but emphasizes caution.
	SeverityWarn Severity = "Warn"
	// SeverityError fails the build
	SeverityError Severity = "Error"
)

// Comment is a comment that can be left on a PR or left as a warning in the
// terminal.
type Comment struct {
	// The file to comment on. Leave blank if the comment should be made in the
	// PR.
	File string `json:"file"`
	// The line to comment on. Leave blank alongside the File field to comment
	// top-level.
	Line uint `json:"line"`
	// The text to include in your comment.
	Text string `json:"text"`
	// Severity of the comment. Defaults to Info.
	Severity Severity `json:"severity"`
}

func (r *Result) Warn(message string) {
	r.Comments = append(r.Comments, Comment{
		Text:     message,
		Severity: SeverityWarn,
	})
}

func (r *Result) WarnLine(file string, line uint, message string) {
	r.Comments = append(r.Comments, Comment{
		File:     file,
		Line:     line,
		Text:     message,
		Severity: SeverityWarn,
	})
}
