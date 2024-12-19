package githubformat

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blakewilliams/manifest"
	"github.com/blakewilliams/manifest/github"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeGitHubClient struct {
	mock.Mock
}

var _ GitHubClient = (*fakeGitHubClient)(nil)

func (f *fakeGitHubClient) Comment(number int, comment string) error {
	args := f.Called(number, comment)
	return args.Error(0)
}

func (f *fakeGitHubClient) FileComment(fc github.NewFileComment) error {
	args := f.Called(fc)
	return args.Error(0)
}

func TestFormat_FileComment(t *testing.T) {
	i := &manifest.Import{
		PullNumber: 1,
	}

	result := manifest.Result{
		Comments: []manifest.Comment{
			{
				Text:     "Test comment",
				Severity: manifest.SeverityError,
				File:     "test.go",
				Line:     10,
				Side:     "RIGHT",
			},
			{
				Text:     "Test comment 2",
				Severity: manifest.SeverityInfo,
			},
		},
	}

	client := &fakeGitHubClient{}
	client.On("FileComment", mock.MatchedBy(func(fc github.NewFileComment) bool {
		return fc.Number == 1 &&
			fc.File == "test.go" &&
			fc.Line == 10 &&
			fc.Side == "RIGHT" &&
			strings.Contains(fc.Text, "Test comment") &&
			strings.Contains(fc.Text, "> [!CAUTION]")
	})).Return(nil)

	client.On("Comment", 1, mock.MatchedBy(func(comment string) bool {
		return strings.Contains(comment, "Test comment 2") &&
			strings.Contains(comment, "> [!TIP]")
	})).Return(nil)

	formatter := New(client, 1, "abc123")
	err := formatter.Format("test", i, result)
	require.NoError(t, err)

	client.AssertExpectations(t)
}

func TestFormat_CommentError(t *testing.T) {
	i := &manifest.Import{
		PullNumber: 1,
	}

	result := manifest.Result{
		Comments: []manifest.Comment{
			{
				Text:     "Test comment",
				Severity: manifest.SeverityError,
				File:     "test.go",
				Line:     10,
				Side:     "RIGHT",
			},
		},
	}

	client := &fakeGitHubClient{}
	client.On("FileComment", mock.Anything).Return(fmt.Errorf("comment error"))

	formatter := New(client, 1, "abc123")
	err := formatter.Format("test", i, result)

	require.Error(t, err)
	require.Equal(t, "comment error", err.Error())

	client.AssertExpectations(t)
}