package customs

import (
	"encoding/json"
	"fmt"
	"io"
)

type Inspection struct {
	config Configuration
	diff   Diff
}

func NewInspection(c Configuration, diffReader io.Reader) (*Inspection, error) {
	diff, err := NewDiff(diffReader)
	if err != nil {
		return nil, fmt.Errorf("could not create diff: %w", err)
	}
	return &Inspection{
		config: c,
		diff:   diff,
	}, nil
}

func (i *Inspection) ImportJSON() ([]byte, error) {
	customsImport := &Import{Diff: i.diff}
	out, err := json.Marshal(customsImport)
	if err != nil {
		return nil, fmt.Errorf("could not marshall output for import JSON: %w", err)
	}

	return out, nil
}

// Inspect accepts a configuration and a diff, then runs + reports on the rules
// based on the configuration+output.
func (i *Inspection) Perform() error {
	return nil
}
