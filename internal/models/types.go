package models

import "time"

// Repository represents a GitHub repository
type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	URL         string `json:"url"`
	DefaultBranch string `json:"default_branch"`
	Branches    []Branch `json:"branches"`
}

// Branch represents a repository branch
type Branch struct {
	Name   string   `json:"name"`
	SHA    string   `json:"sha"`
	Commits []Commit `json:"commits"`
}

// Commit represents a commit with change statistics
type Commit struct {
	SHA       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    Author    `json:"author"`
	Date      time.Time `json:"date"`
	Stats     CommitStats `json:"stats"`
}

// Author represents a commit author
type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Login string `json:"login"` // GitHub username
}

// CommitStats represents code changes in a commit
type CommitStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Total     int `json:"total"`
}

// Report represents the final generated report
type Report struct {
	Target      string                    `json:"target"`
	Period      Period                    `json:"period"`
	Repositories []Repository             `json:"repositories"`
	Summary     map[string]ContributorStats `json:"summary"`
}

// Period represents the time range for the report
type Period struct {
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

// ContributorStats aggregates statistics per contributor
type ContributorStats struct {
	Name         string                    `json:"name"`
	Email        string                    `json:"email"`
	Login        string                    `json:"login"`
	TotalCommits int                       `json:"total_commits"`
	TotalAdditions int                     `json:"total_additions"`
	TotalDeletions int                     `json:"total_deletions"`
	Repositories map[string]RepositoryStats `json:"repositories"`
}

// RepositoryStats represents contributor stats per repository
type RepositoryStats struct {
	Commits   int `json:"commits"`
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
}