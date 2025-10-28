package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"ghreporting/internal/client"
	"ghreporting/internal/reporter"
)

func main() {
	var (
		orgUser    = flag.String("target", "", "GitHub organization or user (required)")
		token      = flag.String("token", "", "GitHub token (optional, can use GITHUB_TOKEN env var)")
		since      = flag.String("since", "", "Start date (YYYY-MM-DD) for commit analysis (default: 30 days ago)")
		until      = flag.String("until", "", "End date (YYYY-MM-DD) for commit analysis (default: now)")
		outputFile = flag.String("output", "", "Output file path (default: stdout)")
		format     = flag.String("format", "text", "Output format: text, json, csv")
	)
	flag.Parse()

	if *orgUser == "" {
		fmt.Fprintf(os.Stderr, "Error: -target parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Get token from flag or environment
	ghToken := *token
	if ghToken == "" {
		ghToken = os.Getenv("GITHUB_TOKEN")
	}

	// Parse dates
	var sinceTime, untilTime time.Time
	var err error

	if *since != "" {
		sinceTime, err = time.Parse("2006-01-02", *since)
		if err != nil {
			log.Fatalf("Invalid since date format: %v", err)
		}
	} else {
		sinceTime = time.Now().AddDate(0, 0, -30) // Default to 30 days ago
	}

	if *until != "" {
		untilTime, err = time.Parse("2006-01-02", *until)
		if err != nil {
			log.Fatalf("Invalid until date format: %v", err)
		}
	} else {
		untilTime = time.Now()
	}

	// Create GitHub client
	ghClient := client.NewGitHubClient(ghToken)

	// Create reporter
	rep := reporter.NewReporter(ghClient)

	// Generate report
	ctx := context.Background()
	report, err := rep.GenerateReport(ctx, *orgUser, sinceTime, untilTime)
	if err != nil {
		log.Fatalf("Error generating report: %v", err)
	}

	// Output report
	if err := rep.OutputReport(report, *outputFile, *format); err != nil {
		log.Fatalf("Error outputting report: %v", err)
	}
}