package customs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/blakewilliams/customs/github"
	"golang.org/x/sync/errgroup"
)

type Inspection struct {
	config        *Configuration
	customsImport *Import
	diff          Diff
}

func NewInspection(c *Configuration, diffReader io.Reader) (*Inspection, error) {
	diff, err := NewDiff(diffReader)
	if err != nil {
		return nil, fmt.Errorf("could not create diff: %w", err)
	}

	inspection := &Inspection{
		config:        c,
		customsImport: &Import{Diff: diff},
	}

	if c.FetchPullInfo {
		token := os.Getenv("CUSTOMS_GITHUB_TOKEN")
		if token == "" {
			fmt.Fprint(os.Stderr, "CUSTOMS_GITHUB_TOKEN was not present so pull request information could not be fetched\n")
		} else {

			pr, err := github.FetchPullRequestInfo(
				github.Fetcher{Token: token},
				github.GitShaResolver{},
			)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not fetch GitHub data. Continuing without it. Error: %s\n", err)
			} else {
				inspection.customsImport.PullTitle = pr.Title
				inspection.customsImport.PullDescription = pr.Description
			}
		}
	}

	return inspection, nil
}

func (i *Inspection) ImportJSON() ([]byte, error) {
	out, err := json.Marshal(i.customsImport)
	if err != nil {
		return nil, fmt.Errorf("could not marshall output for import JSON: %w", err)
	}

	return out, nil
}

// Inspect accepts a configuration and a diff, then runs + reports on the rules
// based on the configuration+output.
func (i *Inspection) Perform() error {
	importJSON, err := i.ImportJSON()
	if err != nil {
		return err
	}

	// TODO add a timout config
	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(i.config.Concurrency)

	for name, inspector := range i.config.Inspectors {
		g.Go(func() error {
			cmd := exec.Command("sh", "-c", inspector)
			cmd.Stdin = bytes.NewReader(importJSON)
			output, err := cmd.Output()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to run inspector %s: %s", name, err)
				return nil
			}

			var result Result
			err = json.Unmarshal(output, &result)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse output for inspector %s: %s", name, err)
				return nil
			}

			i.config.Formatter.Format(name, i.customsImport, result)
			return nil
		})
	}

	g.Wait()

	return nil
}
