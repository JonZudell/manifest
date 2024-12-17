package inspectors

import (
	"strings"

	"github.com/blakewilliams/manifest"
)

func PullBody(entry *manifest.Import, r *manifest.Result) error {
	if entry.PullTitle == "" && entry.PullDescription == "" && entry.Strict {
		r.Failure = "No pull request description provided"
	}

	if strings.TrimSpace(entry.PullDescription) == "" {
		r.Error("It looks like your pull request description is empty! Please provide a description of your changes.")
	}

	// for testing purposes
	r.WarnLine("inspectors/pull_body.go", "RIGHT", 20, "This is a warning")

	return nil
}
