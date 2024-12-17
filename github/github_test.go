package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeFetcher struct {
	pullsForShaNumbers []int
	pullsForShaError   error

	pullDetailsPullInfo *PullRequest
	pullDetailsError    error
}

func (f *fakeFetcher) PullsForSha(owner, repo string, number string) ([]int, error) {
	return f.pullsForShaNumbers, f.pullsForShaError
}

func (f *fakeFetcher) PullDetails(owner, repo string, number int) (*PullRequest, error) {
	return f.pullDetailsPullInfo, f.pullDetailsError
}

type fakeGit struct {
	relevantShaResult string
	relevantShaError  error
}

func (f *fakeGit) RelevantSha() (string, error) {
	return f.relevantShaResult, f.relevantShaError
}

func TestFetchPullRequest(t *testing.T) {
	fetcher := &fakeFetcher{
		pullsForShaNumbers: []int{1},
		pullDetailsPullInfo: &PullRequest{
			Title:       "I want to believe",
			Description: "wow!",
		},
	}

	shaResolver := &fakeGit{relevantShaResult: "8f6a7cd5d54e889173834ec10d7755a536cd0dbf"}

	pullDetails, err := FetchPullRequestInfo(fetcher, shaResolver)
	require.NoError(t, err)

	require.Equal(t, "I want to believe", pullDetails.Title)
	require.Equal(t, "wow!", pullDetails.Description)
}

func TestFetchPullRequest_PermeatesErrNoPR(t *testing.T) {
	validFetcher := &fakeFetcher{
		pullsForShaNumbers: []int{1},
		pullDetailsPullInfo: &PullRequest{
			Title:       "I want to believe",
			Description: "wow!",
		},
	}
	validShaResolver := &fakeGit{relevantShaResult: "8f6a7cd5d54e889173834ec10d7755a536cd0dbf"}

	t.Run("sha resolver with no upstream", func(t *testing.T) {
		shaResolver := &fakeGit{relevantShaError: ErrNoPR}

		_, err := FetchPullRequestInfo(validFetcher, shaResolver)
		require.ErrorIs(t, err, ErrNoPR)
	})

	t.Run("pulls for sha with no results", func(t *testing.T) {
		fetcher := &fakeFetcher{
			pullsForShaNumbers: []int{},
			pullDetailsPullInfo: &PullRequest{
				Title:       "I want to believe",
				Description: "wow!",
			},
		}
		_, err := FetchPullRequestInfo(fetcher, validShaResolver)
		require.ErrorIs(t, err, ErrNoPR)
	})
}
