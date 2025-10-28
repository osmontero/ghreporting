package reporter

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"ghreporting/internal/client"
	"ghreporting/internal/models"
)

// Reporter handles report generation
type Reporter struct {
	client *client.GitHubClient
}

// NewReporter creates a new reporter instance
func NewReporter(client *client.GitHubClient) *Reporter {
	return &Reporter{client: client}
}

// GenerateReport generates a comprehensive report for the given target
func (r *Reporter) GenerateReport(ctx context.Context, target string, since, until time.Time) (*models.Report, error) {
	log.Printf("Generating report for %s from %s to %s", target, since.Format("2006-01-02"), until.Format("2006-01-02"))

	// Get all repositories
	repos, err := r.client.ListRepositories(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	log.Printf("Found %d repositories", len(repos))

	// Process repositories concurrently
	reposChan := make(chan models.Repository, len(repos))
	resultsChan := make(chan models.Repository, len(repos))
	errorsChan := make(chan error, len(repos))

	// Worker pool for processing repositories
	const maxWorkers = 10
	var wg sync.WaitGroup

	for i := 0; i < maxWorkers && i < len(repos); i++ {
		wg.Add(1)
		go r.processRepositoryWorker(ctx, reposChan, resultsChan, errorsChan, since, until, &wg)
	}

	// Send repositories to workers
	go func() {
		for _, repo := range repos {
			reposChan <- repo
		}
		close(reposChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	var processedRepos []models.Repository
	var errors []error

	// Collect results and errors
	done := false
	for !done {
		select {
		case repo, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else {
				processedRepos = append(processedRepos, repo)
			}
		case err, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
			} else {
				errors = append(errors, err)
			}
		}
		done = resultsChan == nil && errorsChan == nil
	}

	// Log errors but continue
	for _, err := range errors {
		log.Printf("Warning: %v", err)
	}

	log.Printf("Successfully processed %d repositories", len(processedRepos))

	// Generate summary statistics
	summary := r.generateSummary(processedRepos)

	return &models.Report{
		Target:      target,
		Period:      models.Period{Since: since, Until: until},
		Repositories: processedRepos,
		Summary:     summary,
	}, nil
}

func (r *Reporter) processRepositoryWorker(ctx context.Context, reposChan <-chan models.Repository, resultsChan chan<- models.Repository, errorsChan chan<- error, since, until time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	for repo := range reposChan {
		processedRepo, err := r.processRepository(ctx, repo, since, until)
		if err != nil {
			errorsChan <- fmt.Errorf("repository %s: %w", repo.FullName, err)
			continue
		}
		resultsChan <- processedRepo
	}
}

func (r *Reporter) processRepository(ctx context.Context, repo models.Repository, since, until time.Time) (models.Repository, error) {
	log.Printf("Processing repository: %s", repo.FullName)

	// Parse owner and repo from full name
	parts := strings.Split(repo.FullName, "/")
	if len(parts) != 2 {
		return repo, fmt.Errorf("invalid repository name format: %s", repo.FullName)
	}
	owner, repoName := parts[0], parts[1]

	// Get branches
	branches, err := r.client.ListBranches(ctx, owner, repoName)
	if err != nil {
		return repo, err
	}

	// Process only a subset of important branches to avoid rate limits
	branchesToProcess := r.selectBranchesToProcess(branches, repo.DefaultBranch)

	var processedBranches []models.Branch
	for _, branch := range branchesToProcess {
		commits, err := r.client.ListCommits(ctx, owner, repoName, branch.Name, since, until)
		if err != nil {
			log.Printf("Warning: failed to get commits for %s@%s: %v", repo.FullName, branch.Name, err)
			continue
		}

		branch.Commits = commits
		processedBranches = append(processedBranches, branch)
		log.Printf("  Branch %s: %d commits", branch.Name, len(commits))
	}

	repo.Branches = processedBranches
	return repo, nil
}

func (r *Reporter) selectBranchesToProcess(branches []models.Branch, defaultBranch string) []models.Branch {
	// Always include default branch
	var selected []models.Branch
	branchMap := make(map[string]bool)

	for _, branch := range branches {
		if branch.Name == defaultBranch {
			selected = append(selected, branch)
			branchMap[branch.Name] = true
			break
		}
	}

	// Add other important branches (main, master, develop, dev) if not already included
	importantBranches := []string{"main", "master", "develop", "dev", "staging", "production"}
	for _, importantBranch := range importantBranches {
		if !branchMap[importantBranch] {
			for _, branch := range branches {
				if branch.Name == importantBranch {
					selected = append(selected, branch)
					branchMap[branch.Name] = true
					break
				}
			}
		}
	}

	return selected
}

func (r *Reporter) generateSummary(repos []models.Repository) map[string]models.ContributorStats {
	summary := make(map[string]models.ContributorStats)

	for _, repo := range repos {
		for _, branch := range repo.Branches {
			for _, commit := range branch.Commits {
				authorKey := r.getAuthorKey(commit.Author)

				stats, exists := summary[authorKey]
				if !exists {
					stats = models.ContributorStats{
						Name:         commit.Author.Name,
						Email:        commit.Author.Email,
						Login:        commit.Author.Login,
						Repositories: make(map[string]models.RepositoryStats),
					}
				}

				// Update global stats
				stats.TotalCommits++
				stats.TotalAdditions += commit.Stats.Additions
				stats.TotalDeletions += commit.Stats.Deletions

				// Update repository-specific stats
				repoStats := stats.Repositories[repo.FullName]
				repoStats.Commits++
				repoStats.Additions += commit.Stats.Additions
				repoStats.Deletions += commit.Stats.Deletions
				stats.Repositories[repo.FullName] = repoStats

				summary[authorKey] = stats
			}
		}
	}

	return summary
}

func (r *Reporter) getAuthorKey(author models.Author) string {
	// Prioritize login, then email, then name for consistency
	if author.Login != "" {
		return author.Login
	}
	if author.Email != "" {
		return author.Email
	}
	return author.Name
}

// OutputReport outputs the report in the specified format
func (r *Reporter) OutputReport(report *models.Report, outputFile, format string) error {
	switch format {
	case "json":
		return r.outputJSON(report, outputFile)
	case "csv":
		return r.outputCSV(report, outputFile)
	case "text":
		return r.outputText(report, outputFile)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (r *Reporter) outputJSON(report *models.Report, outputFile string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if outputFile == "" {
		fmt.Print(string(data))
	} else {
		return os.WriteFile(outputFile, data, 0644)
	}
	return nil
}

func (r *Reporter) outputCSV(report *models.Report, outputFile string) error {
	var output *os.File = os.Stdout
	if outputFile != "" {
		var err error
		output, err = os.Create(outputFile)
		if err != nil {
			return err
		}
		defer output.Close()
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write header
	header := []string{"Author", "Login", "Email", "Repository", "Commits", "Additions", "Deletions"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Sort contributors by total contributions
	var contributors []string
	for contributor := range report.Summary {
		contributors = append(contributors, contributor)
	}
	sort.Slice(contributors, func(i, j int) bool {
		a := report.Summary[contributors[i]]
		b := report.Summary[contributors[j]]
		return (a.TotalAdditions + a.TotalDeletions) > (b.TotalAdditions + b.TotalDeletions)
	})

	// Write data
	for _, contributor := range contributors {
		stats := report.Summary[contributor]
		for repoName, repoStats := range stats.Repositories {
			record := []string{
				stats.Name,
				stats.Login,
				stats.Email,
				repoName,
				fmt.Sprintf("%d", repoStats.Commits),
				fmt.Sprintf("%d", repoStats.Additions),
				fmt.Sprintf("%d", repoStats.Deletions),
			}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Reporter) outputText(report *models.Report, outputFile string) error {
	var output *os.File = os.Stdout
	if outputFile != "" {
		var err error
		output, err = os.Create(outputFile)
		if err != nil {
			return err
		}
		defer output.Close()
	}

	// Print header
	fmt.Fprintf(output, "GitHub Activity Report for: %s\n", report.Target)
	fmt.Fprintf(output, "Period: %s to %s\n", report.Period.Since.Format("2006-01-02"), report.Period.Until.Format("2006-01-02"))
	fmt.Fprintf(output, "Repositories analyzed: %d\n\n", len(report.Repositories))

	// Sort contributors by total contributions
	var contributors []string
	for contributor := range report.Summary {
		contributors = append(contributors, contributor)
	}
	sort.Slice(contributors, func(i, j int) bool {
		a := report.Summary[contributors[i]]
		b := report.Summary[contributors[j]]
		return (a.TotalAdditions + a.TotalDeletions) > (b.TotalAdditions + b.TotalDeletions)
	})

	// Print summary
	fmt.Fprintf(output, "CONTRIBUTOR SUMMARY\n")
	fmt.Fprintf(output, "==================\n\n")

	for _, contributor := range contributors {
		stats := report.Summary[contributor]
		fmt.Fprintf(output, "%s", stats.Name)
		if stats.Login != "" {
			fmt.Fprintf(output, " (@%s)", stats.Login)
		}
		fmt.Fprintf(output, "\n")
		if stats.Email != "" {
			fmt.Fprintf(output, "  Email: %s\n", stats.Email)
		}
		fmt.Fprintf(output, "  Total Commits: %d\n", stats.TotalCommits)
		fmt.Fprintf(output, "  Total Additions: %d\n", stats.TotalAdditions)
		fmt.Fprintf(output, "  Total Deletions: %d\n", stats.TotalDeletions)
		fmt.Fprintf(output, "  Repositories: %d\n", len(stats.Repositories))
		
		// Show top repositories for this contributor
		var repoNames []string
		for repoName := range stats.Repositories {
			repoNames = append(repoNames, repoName)
		}
		sort.Slice(repoNames, func(i, j int) bool {
			a := stats.Repositories[repoNames[i]]
			b := stats.Repositories[repoNames[j]]
			return (a.Additions + a.Deletions) > (b.Additions + b.Deletions)
		})

		fmt.Fprintf(output, "  Top Repositories:\n")
		for i, repoName := range repoNames {
			if i >= 5 { // Show top 5 repositories
				break
			}
			repoStats := stats.Repositories[repoName]
			fmt.Fprintf(output, "    - %s: %d commits (+%d/-%d)\n", 
				repoName, repoStats.Commits, repoStats.Additions, repoStats.Deletions)
		}
		fmt.Fprintf(output, "\n")
	}

	return nil
}