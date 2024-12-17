package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Fetcher struct {
	Token string
}

type pullsForShaResponse struct {
	Number int `json:"number"`
}

func (p Fetcher) PullsForSha(owner, repo, sha string) ([]int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s/pulls", owner, repo, sha)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.Token)
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

func (p Fetcher) PullDetails(owner, repo string, number int) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.Token)
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
