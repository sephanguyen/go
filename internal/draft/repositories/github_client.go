package repositories

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v41/github"
)

type GithubClient struct {
	cl *github.Client
}

func NewClient(appID int64, installationID int64, secret []byte) *GithubClient {
	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, appID, installationID, secret)
	if err != nil {
		panic(fmt.Errorf("ghinstallation.New %w", err))
	}

	return &GithubClient{
		cl: github.NewClient(&http.Client{Transport: itr}),
	}
}
func (g *GithubClient) CreateCommitStatus(ctx context.Context, owner, repo, branch string, status *github.RepoStatus) error {
	_, _, err := g.cl.Repositories.CreateStatus(ctx, owner, repo, branch, status)
	if err != nil {
		return fmt.Errorf("Repositories.CreateStatus %w", err)
	}
	return nil
}

func (g *GithubClient) GetLastCommit(ctx context.Context, owner, repo string, prnumer int) (*github.RepositoryCommit, error) {
	commits, _, err := g.cl.PullRequests.ListCommits(ctx, owner, repo, prnumer, &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("PullRequests.ListCommits %w", err)
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("unknown error, pr %d has no commit", prnumer)
	}
	return commits[len(commits)-1], nil
}
