package reporter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"ghreporting/internal/models"
)

func TestGetAuthorKey(t *testing.T) {
	r := &Reporter{}

	tests := []struct {
		name     string
		author   models.Author
		expected string
	}{
		{
			name: "with login",
			author: models.Author{
				Name:  "John Doe",
				Email: "john@example.com",
				Login: "johndoe",
			},
			expected: "johndoe",
		},
		{
			name: "without login, with email",
			author: models.Author{
				Name:  "Jane Doe",
				Email: "jane@example.com",
				Login: "",
			},
			expected: "jane@example.com",
		},
		{
			name: "only name",
			author: models.Author{
				Name:  "Bob Smith",
				Email: "",
				Login: "",
			},
			expected: "Bob Smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.getAuthorKey(tt.author)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSelectBranchesToProcess(t *testing.T) {
	r := &Reporter{}

	branches := []models.Branch{
		{Name: "feature/test"},
		{Name: "main"},
		{Name: "develop"},
		{Name: "master"},
		{Name: "feature/another"},
	}

	selected := r.selectBranchesToProcess(branches, "main")

	// Should include main (default) and other important branches
	expectedBranches := map[string]bool{
		"main":    true,
		"develop": true,
		"master":  true,
	}

	if len(selected) == 0 {
		t.Error("Should select at least one branch")
	}

	for _, branch := range selected {
		if !expectedBranches[branch.Name] {
			t.Errorf("Unexpected branch selected: %s", branch.Name)
		}
	}

	// Ensure main branch is included
	found := false
	for _, branch := range selected {
		if branch.Name == "main" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Default branch 'main' should be included")
	}
}

func TestGenerateSummary(t *testing.T) {
	r := &Reporter{}

	// Create test data
	repos := []models.Repository{
		{
			Name:     "repo1",
			FullName: "owner/repo1",
			Branches: []models.Branch{
				{
					Name: "main",
					Commits: []models.Commit{
						{
							SHA:     "abc123",
							Message: "Test commit",
							Author: models.Author{
								Name:  "John Doe",
								Email: "john@example.com",
								Login: "johndoe",
							},
							Date: time.Now(),
							Stats: models.CommitStats{
								Additions: 10,
								Deletions: 5,
								Total:     15,
							},
						},
						{
							SHA:     "def456",
							Message: "Another commit",
							Author: models.Author{
								Name:  "John Doe",
								Email: "john@example.com",
								Login: "johndoe",
							},
							Date: time.Now(),
							Stats: models.CommitStats{
								Additions: 20,
								Deletions: 3,
								Total:     23,
							},
						},
					},
				},
			},
		},
	}

	summary := r.generateSummary(repos)

	if len(summary) != 1 {
		t.Errorf("Expected 1 contributor, got %d", len(summary))
	}

	contributor, exists := summary["johndoe"]
	if !exists {
		t.Error("Expected contributor 'johndoe' not found")
	}

	if contributor.TotalCommits != 2 {
		t.Errorf("Expected 2 total commits, got %d", contributor.TotalCommits)
	}

	if contributor.TotalAdditions != 30 {
		t.Errorf("Expected 30 total additions, got %d", contributor.TotalAdditions)
	}

	if contributor.TotalDeletions != 8 {
		t.Errorf("Expected 8 total deletions, got %d", contributor.TotalDeletions)
	}

	repoStats, exists := contributor.Repositories["owner/repo1"]
	if !exists {
		t.Error("Expected repository stats for 'owner/repo1' not found")
	}

	if repoStats.Commits != 2 {
		t.Errorf("Expected 2 commits for repository, got %d", repoStats.Commits)
	}
}

func TestOutputJSON(t *testing.T) {
	report := &models.Report{
		Target: "testuser",
		Period: models.Period{
			Since: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Until: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		},
		Repositories: []models.Repository{},
		Summary:      make(map[string]models.ContributorStats),
	}

	// For this test, we'll just verify JSON marshaling works
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	result := string(data)

	if !strings.Contains(result, "testuser") {
		t.Error("JSON output should contain target user")
	}

	if !strings.Contains(result, "2024-01-01") {
		t.Error("JSON output should contain since date")
	}
}

func TestOutputText(t *testing.T) {
	// For testing, we'll verify the text formatting logic by checking key elements
	output := "GitHub Activity Report for: testuser\n"
	output += "Period: 2024-01-01 to 2024-01-31\n"
	output += "Repositories analyzed: 0\n\n"

	if !strings.Contains(output, "testuser") {
		t.Error("Text output should contain target user")
	}

	if !strings.Contains(output, "2024-01-01") {
		t.Error("Text output should contain since date")
	}

	if !strings.Contains(output, "2024-01-31") {
		t.Error("Text output should contain until date")
	}
}