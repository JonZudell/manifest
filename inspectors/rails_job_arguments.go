package inspectors

import (
	"regexp"
	"strings"

	"github.com/blakewilliams/customs"
)

var performRegex = regexp.MustCompile(`def\s+perform\((.*)\)`)

func RailsJobArguments(entry *customs.Import, r *customs.Result) error {
	for fileName, file := range entry.Diff.Files {
		if !strings.HasSuffix(fileName, "_job.rb") {
			continue
		}

		for _, l := range file.Right {
			if performRegex.MatchString(l.Content) {
				r.WarnLine(fileName, l.LineNo, `You have modified an ActiveRecord job's arguments. In order to avoid job failures please read and follow X documentation.`)
			}
		}
	}

	return nil
}
