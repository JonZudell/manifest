package prettyformat

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"

	"github.com/blakewilliams/customs"
)

type Formatter struct {
	out io.Writer
}

var warnColor = color.New(color.FgYellow, color.Bold).FprintFunc()
var errorColor = color.New(color.FgRed, color.Bold).FprintFunc()
var infoColor = color.New(color.FgBlue, color.Bold).FprintFunc()

func New(out io.Writer) *Formatter {
	return &Formatter{out: out}
}

func (s *Formatter) Format(source string, i *customs.Import, r customs.Result) error {
	for _, comment := range r.Comments {
		switch comment.Severity {
		case customs.SeverityError:
			errorColor(s.out, "== Error: %s", source)
			if comment.File != "" && comment.Line != 0 {
				errorColor(s.out, "%s:%s", comment.File, comment.Line)
			}
		case customs.SeverityWarn:
			warnColor(s.out, "== Warning: %s", source)
			if comment.File != "" && comment.Line != 0 {
				warnColor(s.out, "%s:%s", comment.File, comment.Line)
			}
		case customs.SeverityInfo:
			warnColor(s.out, "== Info: %s", source)
			if comment.File != "" && comment.Line != 0 {
				infoColor(s.out, "%s:%s", comment.File, comment.Line)
			}
		}

		for _, line := range strings.Split(comment.Text, "\n") {
			fmt.Fprintf(s.out, ">  %s", line)
		}

		fmt.Printf("\n\n")
	}

	return nil
}
