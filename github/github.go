package github

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var ErrNoPR = errors.New("no PR exists for current branch")
var originRegexp = regexp.MustCompile(`(?:https?://github\.com/|git@github\.com:)([^/]+)/([^\.]+)`)

type (
	// PullRequestFetcher is the interface for ultimately fetching the title and description of a Pull Request
	PullRequestFetcher interface {
		PullsForSha(owner, repo, sha string) ([]int, error)
		PullDetails(owner, repo string, number int) (*PullRequest, error)
	}

	// PullRequest represents a subset of GitHub Pull Request
	PullRequest struct {
		Title       string
		Description string
	}

	// ShaResolver is used to fetch the most relevant SHA
	ShaResolver interface {
		RelevantSha() (string, error)
	}

	// The default ShaResolver that uses local git
	GitShaResolver struct{}
)

func FetchPullRequestInfo(ghFetcher PullRequestFetcher, shaResolver ShaResolver) (*PullRequest, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	remoteURL := strings.TrimSpace(string(output))
	matches := originRegexp.FindStringSubmatch(remoteURL)
	if len(matches) != 3 {
		return nil, fmt.Errorf("could not parse owner and repo from remote URL")
	}
	owner := matches[1]
	repo := matches[2]
	sha, err := shaResolver.RelevantSha()
	if err != nil {
		return nil, fmt.Errorf("could not get latest pushed SHA: %w", err)
	}

	numbers, err := ghFetcher.PullsForSha(owner, repo, sha)
	if len(numbers) == 0 {
		return nil, ErrNoPR
	}

	pr, err := ghFetcher.PullDetails(owner, repo, numbers[0])
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (g GitShaResolver) RelevantSha() (string, error) {
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get the latest pushed SHA for the current branch
	shaCmd := exec.Command("git", "rev-parse", "origin/"+branch)
	shaOutput, err := shaCmd.Output()
	if err != nil {
		if strings.Contains(string(shaOutput), "unknown revision") {
			return "", ErrNoPR
		}
		return "", err
	}
	sha := strings.TrimSpace(string(shaOutput))

	return sha, nil
}
