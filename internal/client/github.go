package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"

	"ghreporting/internal/models"
)

// GitHubClient wraps the GitHub API client
type GitHubClient struct {
	client *github.Client
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient(token string) *GitHubClient {
	var client *github.Client
	
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &GitHubClient{client: client}
}

// ListRepositories retrieves all repositories for a user or organization
func (gc *GitHubClient) ListRepositories(ctx context.Context, target string) ([]models.Repository, error) {
	var allRepos []*github.Repository
	orgOpt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// Try as organization first, then as user
	for {
		repos, resp, err := gc.client.Repositories.ListByOrg(ctx, target, orgOpt)
		if err != nil {
			// If org fails, try as user
			log.Printf("Failed to list org repositories, trying as user: %v", err)
			return gc.listUserRepositories(ctx, target)
		}

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		orgOpt.Page = resp.NextPage
	}

	return gc.convertRepositories(allRepos), nil
}

func (gc *GitHubClient) listUserRepositories(ctx context.Context, target string) ([]models.Repository, error) {
	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := gc.client.Repositories.List(ctx, target, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list user repositories: %w", err)
		}

		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return gc.convertRepositories(allRepos), nil
}

// ListBranches retrieves all branches for a repository
func (gc *GitHubClient) ListBranches(ctx context.Context, owner, repo string) ([]models.Branch, error) {
	var allBranches []*github.Branch
	opt := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		branches, resp, err := gc.client.Repositories.ListBranches(ctx, owner, repo, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches for %s/%s: %w", owner, repo, err)
		}

		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var result []models.Branch
	for _, branch := range allBranches {
		result = append(result, models.Branch{
			Name: branch.GetName(),
			SHA:  branch.GetCommit().GetSHA(),
		})
	}

	return result, nil
}

// ListCommits retrieves commits for a repository branch within a time range
func (gc *GitHubClient) ListCommits(ctx context.Context, owner, repo, branch string, since, until time.Time) ([]models.Commit, error) {
	var allCommits []*github.RepositoryCommit
	opt := &github.CommitsListOptions{
		SHA:   branch,
		Since: since,
		Until: until,
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		commits, resp, err := gc.client.Repositories.ListCommits(ctx, owner, repo, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list commits for %s/%s@%s: %w", owner, repo, branch, err)
		}

		allCommits = append(allCommits, commits...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	var result []models.Commit
	for _, commit := range allCommits {
		// Get detailed commit information with stats
		detailedCommit, _, err := gc.client.Repositories.GetCommit(ctx, owner, repo, commit.GetSHA(), nil)
		if err != nil {
			log.Printf("Warning: failed to get detailed commit info for %s: %v", commit.GetSHA(), err)
			continue
		}

		author := models.Author{
			Name:  commit.GetCommit().GetAuthor().GetName(),
			Email: commit.GetCommit().GetAuthor().GetEmail(),
		}
		if commit.GetAuthor() != nil {
			author.Login = commit.GetAuthor().GetLogin()
		}

		result = append(result, models.Commit{
			SHA:     commit.GetSHA(),
			Message: commit.GetCommit().GetMessage(),
			Author:  author,
			Date:    commit.GetCommit().GetAuthor().GetDate().Time,
			Stats: models.CommitStats{
				Additions: detailedCommit.GetStats().GetAdditions(),
				Deletions: detailedCommit.GetStats().GetDeletions(),
				Total:     detailedCommit.GetStats().GetTotal(),
			},
		})
	}

	return result, nil
}

func (gc *GitHubClient) convertRepositories(repos []*github.Repository) []models.Repository {
	var result []models.Repository
	for _, repo := range repos {
		if repo.GetArchived() {
			continue // Skip archived repositories
		}

		result = append(result, models.Repository{
			Name:        repo.GetName(),
			FullName:    repo.GetFullName(),
			URL:         repo.GetHTMLURL(),
			DefaultBranch: repo.GetDefaultBranch(),
		})
	}
	return result
}