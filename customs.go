package customs

import (
	"fmt"
	"io"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
)

// Entry is the struct passed to plugins that provide the info necessary to
// apply rules. It includes information about the PR if present, and the diff.
type Entry struct {
	// PullTitle is the title of the pull request if present.
	PullTitle string
	// PullDescription is the description of the pull request, if present
	PullDescription string
	// PullProvided is true if the pull request is provided.
	PullProvided bool

	// Diff is the parsed changes for this diff
	Diff Diff `json:"diff"`
}

// Diff represents the provided diff
type Diff struct {
	// ChangedFiles is a list of files that have been changed. It does not
	// include deleted, renamed, or new files.
	ChangedFiles []string `json:"changed"`
	// DeletedFiles is a list of files that have been deleted.
	DeletedFiles []string `json:"deleted"`
	// RenamedFiles is a list of files that have been renamed.
	RenamedFiles []string `json:"renamed"`
	// NewFiles is a list of files that have been added.
	NewFiles []string `json:"new"`

	// Files is a mapping of file names to the file contents
	Files map[string]File `json:"files"`
}

// File is represents a single file in a diff
type File struct {
	IsNew     bool `json:"new"`
	IsDeleted bool `json:"deleted"`
	IsRenamed bool `json:"renamed"`

	// TODO this HAS to include line numbers
	Left  []Line `json:"left"`
	Right []Line `json:"right"`
}

// Line represents a change (add/delete) in a diff
type Line struct {
	Number  uint
	Content string
}

// NewDiff returns a new diff that can be used by plugins
func NewDiff(f io.Reader) (Diff, error) {
	files, _, err := gitdiff.Parse(f)

	if err != nil {
		return Diff{}, fmt.Errorf("failed to parse git diff: %w", err)
	}

	diff := &Diff{
		ChangedFiles: make([]string, 0),
		DeletedFiles: make([]string, 0),
		RenamedFiles: make([]string, 0),
		NewFiles:     make([]string, 0),
		Files:        make(map[string]File, len(files)),
	}

	for _, file := range files {
		leftLines := make([]Line, 0)
		rightLines := make([]Line, 0)

		for _, fragment := range file.TextFragments {
			leftStart := fragment.OldPosition
			rightStart := fragment.NewPosition
			fmt.Println(fragment.OldLines)

			for _, line := range fragment.Lines {
				switch line.Op {
				case gitdiff.OpDelete:
					leftLines = append(leftLines, Line{
						Number:  uint(leftStart),
						Content: line.Line,
					})
					leftStart++
				case gitdiff.OpAdd:
					rightLines = append(rightLines, Line{
						Number:  uint(rightStart),
						Content: line.Line,
					})
					rightStart++
				default:
					leftStart++
					rightStart++
				}
			}
		}

		// Add the file to the mapping
		name := file.OldName
		if name == "" {
			name = file.NewName
		}
		diff.Files[name] = File{
			IsNew:     file.IsNew,
			IsDeleted: file.IsDelete,
			IsRenamed: file.IsRename,
			Left:      leftLines,
			Right:     rightLines,
		}

		if file.IsNew {
			diff.NewFiles = append(diff.NewFiles, file.NewName)
		} else if file.IsDelete {
			diff.DeletedFiles = append(diff.DeletedFiles, file.OldName)
		} else if file.IsRename {
			diff.RenamedFiles = append(diff.RenamedFiles, file.OldName)
		} else {
			diff.ChangedFiles = append(diff.ChangedFiles, file.OldName)
		}

	}

	return *diff, nil
}
