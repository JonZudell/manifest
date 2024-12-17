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

	return inspection, nil
}

func (i *Inspection) PopulatePullDetails(gh github.Client, owner, repo, sha string) error {
	numbers, err := gh.PullRequestIDsForSha(owner, repo, sha)
	if err != nil {
		return err
	}

	if len(numbers) == 0 {
		return github.ErrNoPR
	}

	pr, err := gh.DetailsForPull(owner, repo, numbers[0])
	if err != nil {
		return err
	}

	i.customsImport.PullTitle = pr.Title
	i.customsImport.PullDescription = pr.Body
	i.customsImport.PullProvided = true

	return nil
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

	success := true
	for name, inspector := range i.config.Inspectors {
		g.Go(func() error {
			cmd := exec.Command("sh", "-c", inspector)
			cmd.Stdin = bytes.NewReader(importJSON)
			output, err := cmd.Output()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to run inspector %s: %s\n", name, err)
				success = false
				return nil
			}

			var result Result
			err = json.Unmarshal(output, &result)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse output for inspector %s: %s\n", name, err)
				success = false
				return nil
			}

			if success {
				for _, comment := range result.Comments {
					if comment.Severity != SeverityInfo {
						success = false
						break
					}
				}
			}

			return i.config.Formatter.Format(name, i.customsImport, result)
		})
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("one or more rules failed")
	}

	return nil
}
