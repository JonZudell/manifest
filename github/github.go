package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

var ErrNoPR = errors.New("no PR exists for current branch")

type (
	Client interface {
		DetailsForPull(owner, repo string, number int) (*PullRequest, error)
		PullRequestIDsForSha(owner, repo, sha string) ([]int, error)
	}

	defaultClient struct {
		token      string
		HttpClient *http.Client
	}

	// PullRequestFetcher is the interface for ultimately fetching the title and description of a Pull Request
	PullRequestFetcher interface {
		PullsForSha(owner, repo, sha string) ([]int, error)
		PullDetails(owner, repo string, number int) (*PullRequest, error)
	}

	// PullRequest represents a subset of GitHub Pull Request
	PullRequest struct {
		ID    uint
		Title string
		Body  string
	}

	// ShaResolver is used to fetch the most relevant SHA
	ShaResolver interface {
		RelevantSha() (string, error)
	}

	// The default ShaResolver that uses local git
	GitShaResolver struct{}
)

func NewClient(token string) Client {
	return defaultClient{
		token:      token,
		HttpClient: http.DefaultClient,
	}
}

func (c defaultClient) DetailsForPull(owner, repo string, number int) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.groot-preview+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	pullRequest := &PullRequest{}
	if err := json.Unmarshal(body, &pullRequest); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return pullRequest, nil
}

func (c defaultClient) PullRequestIDsForSha(owner, repo string, sha string) ([]int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s/pulls", owner, repo, sha)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.groot-preview+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	type pullsForShaResponse struct {
		Number int `json:"number"`
	}

	var pullRequests []pullsForShaResponse
	if err := json.Unmarshal(body, &pullRequests); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	numbers := make([]int, len(pullRequests))
	for i, pull := range pullRequests {
		numbers[i] = pull.Number
	}

	return numbers, nil
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
