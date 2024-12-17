package inspectors

import (
	"strings"

	"github.com/blakewilliams/customs"
)

func PullBody(entry *customs.Import, r *customs.Result) error {
	// Fail closed
	if !entry.PullProvided {
		return nil
	}

	if strings.TrimSpace(entry.PullDescription) == "" {
		r.Warn("It looks like your pull request is empty! Please provide a description of your changes.")
	}

	return nil
}
