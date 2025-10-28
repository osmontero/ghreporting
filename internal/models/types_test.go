package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRepositoryMarshalJSON(t *testing.T) {
	repo := Repository{
		Name:        "test-repo",
		FullName:    "owner/test-repo",
		URL:         "https://github.com/owner/test-repo",
		DefaultBranch: "main",
		Branches: []Branch{
			{
				Name: "main",
				SHA:  "abc123",
			},
		},
	}

	data, err := json.Marshal(repo)
	if err != nil {
		t.Fatalf("Failed to marshal repository: %v", err)
	}

	var unmarshaled Repository
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal repository: %v", err)
	}

	if unmarshaled.Name != repo.Name {
		t.Errorf("Expected name %s, got %s", repo.Name, unmarshaled.Name)
	}
	if unmarshaled.FullName != repo.FullName {
		t.Errorf("Expected full name %s, got %s", repo.FullName, unmarshaled.FullName)
	}
}

func TestCommitStatsCalculation(t *testing.T) {
	stats := CommitStats{
		Additions: 100,
		Deletions: 50,
		Total:     150,
	}

	if stats.Total != stats.Additions+stats.Deletions {
		t.Errorf("Total should equal additions + deletions, got %d", stats.Total)
	}
}

func TestContributorStatsAggregation(t *testing.T) {
	stats := ContributorStats{
		Name:           "John Doe",
		Email:          "john@example.com",
		Login:          "johndoe",
		TotalCommits:   0,
		TotalAdditions: 0,
		TotalDeletions: 0,
		Repositories:   make(map[string]RepositoryStats),
	}

	// Add stats for two repositories
	stats.Repositories["repo1"] = RepositoryStats{
		Commits:   5,
		Additions: 100,
		Deletions: 20,
	}
	stats.Repositories["repo2"] = RepositoryStats{
		Commits:   3,
		Additions: 50,
		Deletions: 10,
	}

	// Calculate totals
	for _, repoStats := range stats.Repositories {
		stats.TotalCommits += repoStats.Commits
		stats.TotalAdditions += repoStats.Additions
		stats.TotalDeletions += repoStats.Deletions
	}

	expectedCommits := 8
	expectedAdditions := 150
	expectedDeletions := 30

	if stats.TotalCommits != expectedCommits {
		t.Errorf("Expected %d total commits, got %d", expectedCommits, stats.TotalCommits)
	}
	if stats.TotalAdditions != expectedAdditions {
		t.Errorf("Expected %d total additions, got %d", expectedAdditions, stats.TotalAdditions)
	}
	if stats.TotalDeletions != expectedDeletions {
		t.Errorf("Expected %d total deletions, got %d", expectedDeletions, stats.TotalDeletions)
	}
}

func TestPeriodValidation(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	period := Period{
		Since: yesterday,
		Until: now,
	}

	if period.Until.Before(period.Since) {
		t.Error("Period until should be after since")
	}

	duration := period.Until.Sub(period.Since)
	expectedDuration := 24 * time.Hour
	if duration < expectedDuration-time.Minute || duration > expectedDuration+time.Minute {
		t.Errorf("Expected duration around %v, got %v", expectedDuration, duration)
	}
}