# GitHub Reporting Tool

A command-line tool for generating comprehensive reports on GitHub repository activity, including commit statistics and code change metrics per contributor across all repositories of a user or organization.

## Features

- **Repository Analysis**: Scans all repositories for a given GitHub user or organization
- **Branch Coverage**: Analyzes commits across important branches (main, master, develop, etc.) or all branches with `-all-branches` flag
- **Time-based Filtering**: Generates reports for specific date ranges
- **Contributor Metrics**: Tracks code additions, deletions, and commit counts per contributor
- **Multiple Output Formats**: Supports text, JSON, and CSV output formats
- **Rate Limit Handling**: Includes concurrent processing with appropriate rate limiting

## Installation

### Prerequisites
- Go 1.21 or later
- [Task](https://taskfile.dev/installation/) (optional, for using Taskfile)
- GitHub Personal Access Token (optional but recommended for higher rate limits)

### Build from Source
```bash
git clone https://github.com/osmontero/ghreporting.git
cd ghreporting

# Using Task (recommended)
task build

# Or using Go directly
go mod tidy
go build -o ghreporting
```

### Available Tasks
If you have [Task](https://taskfile.dev/) installed, you can use these convenient commands:

```bash
task                # List all available tasks
task build          # Build the application
task test           # Run all tests
task clean          # Remove build artifacts
task install        # Install dependencies
task run            # Run with example parameters
task all            # Install, build and test
task prod-install   # Install to /usr/local/bin
task lint           # Run Go linter and formatter
task dev            # Development mode with file watching
task check          # Run all checks (lint, test, build)
```

## Usage

### Basic Usage
```bash
# Generate report for a GitHub user
./ghreporting -target username

# Generate report for a GitHub organization  
./ghreporting -target orgname

# Specify a custom date range (last 30 days by default)
./ghreporting -target username -since 2024-01-01 -until 2024-01-31

# Use GitHub token for higher rate limits
./ghreporting -target username -token YOUR_GITHUB_TOKEN

# Or set environment variable
export GITHUB_TOKEN=YOUR_GITHUB_TOKEN
./ghreporting -target username

# Analyze all branches (not just important ones like main, master, develop)
./ghreporting -target username -all-branches
```

### Output Options
```bash
# Save report to file
./ghreporting -target username -output report.txt

# Generate JSON output
./ghreporting -target username -format json -output report.json

# Generate CSV output for spreadsheet analysis
./ghreporting -target username -format csv -output report.csv
```

### Branch Analysis Options
```bash
# Analyze all branches (default: only important branches like main, master, develop)
./ghreporting -target username -all-branches

# Combine with other options
./ghreporting -target username -all-branches -format json -output full_analysis.json

# For comprehensive analysis with date range
./ghreporting -target username -all-branches -since 2024-01-01 -until 2024-01-31
```

**Note**: Using `-all-branches` will significantly increase API calls as it analyzes every branch in every repository. This may hit rate limits faster, especially for organizations with many repositories and branches. Consider using a GitHub token for higher rate limits.

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-target` | GitHub organization or user (required) | - |
| `-token` | GitHub personal access token | Uses `GITHUB_TOKEN` env var |
| `-since` | Start date for analysis (YYYY-MM-DD) | 30 days ago |
| `-until` | End date for analysis (YYYY-MM-DD) | Current date |
| `-format` | Output format: `text`, `json`, `csv` | `text` |
| `-output` | Output file path | stdout |
| `-all-branches` | Analyze all branches instead of just important ones | `false` |

## GitHub Token Setup

For better rate limits and access to private repositories, create a GitHub Personal Access Token:

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select appropriate scopes:
   - `repo` (for private repositories)
   - `read:org` (for organization repositories)
   - `read:user` (for user information)
4. Set the token as environment variable:
   ```bash
   export GITHUB_TOKEN=your_token_here
   ```

## Sample Output

### Text Format
```
GitHub Activity Report for: myorg
Period: 2024-01-01 to 2024-01-31
Repositories analyzed: 15

CONTRIBUTOR SUMMARY
==================

John Doe (@johndoe)
  Email: john@example.com
  Total Commits: 45
  Total Additions: 1,250
  Total Deletions: 340
  Repositories: 8
  Top Repositories:
    - myorg/backend-api: 15 commits (+450/-120)
    - myorg/frontend-app: 12 commits (+380/-95)
    - myorg/data-pipeline: 8 commits (+220/-65)
    - myorg/config-management: 5 commits (+150/-40)
    - myorg/documentation: 5 commits (+50/-20)

Jane Smith (@janesmith)
  Email: jane@example.com
  Total Commits: 32
  Total Additions: 890
  Total Deletions: 245
  Repositories: 6
  Top Repositories:
    - myorg/frontend-app: 18 commits (+520/-140)
    - myorg/mobile-app: 8 commits (+280/-75)
    - myorg/shared-components: 4 commits (+60/-20)
    - myorg/testing-framework: 2 commits (+30/-10)
```

### JSON Format
```json
{
  "target": "myorg",
  "period": {
    "since": "2024-01-01T00:00:00Z",
    "until": "2024-01-31T23:59:59Z"
  },
  "repositories": [
    {
      "name": "backend-api",
      "full_name": "myorg/backend-api",
      "url": "https://github.com/myorg/backend-api",
      "default_branch": "main",
      "branches": [
        {
          "name": "main",
          "sha": "abc123...",
          "commits": [...]
        }
      ]
    }
  ],
  "summary": {
    "johndoe": {
      "name": "John Doe",
      "email": "john@example.com", 
      "login": "johndoe",
      "total_commits": 45,
      "total_additions": 1250,
      "total_deletions": 340,
      "repositories": {
        "myorg/backend-api": {
          "commits": 15,
          "additions": 450,
          "deletions": 120
        }
      }
    }
  }
}
```

### CSV Format
```csv
Author,Login,Email,Repository,Commits,Additions,Deletions
John Doe,johndoe,john@example.com,myorg/backend-api,15,450,120
John Doe,johndoe,john@example.com,myorg/frontend-app,12,380,95
Jane Smith,janesmith,jane@example.com,myorg/frontend-app,18,520,140
```

## Rate Limits and Performance

- **Without token**: ~60 requests/hour (GitHub's unauthenticated limit)
- **With token**: ~5,000 requests/hour (GitHub's authenticated limit)  
- **Concurrent processing**: Uses worker pools to process repositories in parallel
- **Branch selection**: By default, focuses on main branches (main, master, develop) to optimize API usage. Use `-all-branches` to analyze all branches (may increase API calls significantly)

## Error Handling

The tool gracefully handles common scenarios:
- **Rate limit exceeded**: Provides clear error message suggesting token usage
- **Private repositories**: Skips inaccessible repos and continues with available ones  
- **Network issues**: Retries failed requests and logs warnings
- **Invalid dates**: Validates date formats and provides helpful error messages

## Limitations

- By default, processes only important branches to manage API rate limits (use `-all-branches` to analyze all branches)
- Skips archived repositories
- Large organizations may require multiple runs due to rate limits
- Commit statistics are based on GitHub's diff calculations

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Commit your changes (`git commit -am 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues, questions, or contributions, please visit the [GitHub repository](https://github.com/osmontero/ghreporting) and create an issue.
