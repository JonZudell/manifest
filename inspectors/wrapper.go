package inspectors

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/blakewilliams/customs"
)

// Wrap wraps an inspector function to easily handle the conversion of STDIN to
// a `customs.Import` and STDOUT to `customs.Result` JSON.
func Wrap(name string, f func(entry *customs.Import, r *customs.Result) error) error {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("could not stat stdin: %w", err)
	}
	if (fi.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("stdin was not provided")
	}

	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("could not read error in '%s': %w", name, err)
	}

	if len(in) == 0 {
		return fmt.Errorf("no content was passed to stdin")
	}

	i := &customs.Import{}
	err = json.Unmarshal(in, i)
	if err != nil {
		return fmt.Errorf("failed to read import JSON: %w", err)
	}
	result := &customs.Result{Comments: make([]customs.Comment, 0)}

	err = f(i, result)
	if err != nil {
		return fmt.Errorf("failed to run inspector '%s': %w", name, err)
	}

	out, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal output for '%s': %w", name, err)
	}

	fmt.Fprint(os.Stdout, string(out))

	return nil
}
