# GitHub Reporting Tool - Example Usage

This document provides examples of how to use the GitHub Reporting Tool.

## Prerequisites

1. **Build the tool:**
   ```bash
   task build
   # or
   go build -o ghreporting .
   ```

2. **Set up GitHub Token (recommended):**
   ```bash
   export GITHUB_TOKEN="your_personal_access_token_here"
   ```

## Basic Usage Examples

### 1. Generate Report for a User
```bash
# Generate report for the last 30 days (default)
./ghreporting -target octocat

# Generate report for a specific date range
./ghreporting -target octocat -since 2024-01-01 -until 2024-01-31

# Generate report with custom output file
./ghreporting -target octocat -output report.txt
```

### 2. Generate Report for an Organization
```bash
# Generate report for an organization
./ghreporting -target github -since 2024-10-01 -until 2024-10-31

# Generate report in JSON format
./ghreporting -target github -format json -output github_report.json

# Generate report in CSV format for spreadsheet analysis
./ghreporting -target github -format csv -output github_report.csv
```

### 3. Advanced Usage
```bash
# Generate comprehensive report with all output formats
./ghreporting -target myorg -since 2024-01-01 -until 2024-12-31 -output yearly_report.txt
./ghreporting -target myorg -since 2024-01-01 -until 2024-12-31 -format json -output yearly_report.json
./ghreporting -target myorg -since 2024-01-01 -until 2024-12-31 -format csv -output yearly_report.csv

# Generate report for the last week
./ghreporting -target username -since $(date -d '7 days ago' +%Y-%m-%d) -until $(date +%Y-%m-%d)

# Generate report for last month
./ghreporting -target username -since $(date -d 'last month' +%Y-%m-01) -until $(date -d 'last month' +%Y-%m-31)
```

## Sample Output Analysis

### Text Report Interpretation
When you run the tool with text output, you'll see:

1. **Header Information:**
   - Target (user/organization name)
   - Analysis period
   - Number of repositories processed

2. **Contributor Summary:**
   - Contributors sorted by total code changes (additions + deletions)
   - For each contributor:
     - Name, GitHub username, and email
     - Total commits, additions, and deletions
     - Top repositories they contributed to

### JSON Report Structure
The JSON output provides programmatic access to all data:
- `target`: The analyzed user/organization
- `period`: Time range for analysis
- `repositories`: Detailed repository and commit information
- `summary`: Aggregated contributor statistics

### CSV Report Usage
The CSV format is ideal for:
- Importing into spreadsheet applications
- Creating charts and graphs
- Further data analysis and processing
- Integration with other reporting tools

## Tips for Large Organizations

When analyzing large organizations:

1. **Use authentication** to avoid rate limits:
   ```bash
   export GITHUB_TOKEN="your_token"
   ```

2. **Analyze smaller time periods** to reduce API calls:
   ```bash
   ./ghreporting -target large-org -since 2024-10-01 -until 2024-10-07
   ```

3. **Monitor rate limits** and spread requests over time if needed

4. **Save results** in multiple formats for different use cases:
   ```bash
   # Executive summary
   ./ghreporting -target org -format text -output summary.txt
   
   # Detailed analysis
   ./ghreporting -target org -format json -output detailed.json
   
   # Spreadsheet analysis  
   ./ghreporting -target org -format csv -output analysis.csv
   ```

## Troubleshooting

### Rate Limit Issues
```
Error: GET https://api.github.com/...: 403 []
```
**Solution:** Set up a GitHub Personal Access Token

### Authentication Issues
```
Error: failed to list repositories: 404 Not Found
```
**Solutions:**
- Check if the user/organization name is correct
- Verify your token has appropriate permissions
- Ensure the target has public repositories or you have access

### No Data Returned
**Possible causes:**
- No commits in the specified time period
- All repositories are private and inaccessible
- Organization has no public repositories

**Solutions:**
- Extend the time period
- Use a token with appropriate permissions
- Verify the target exists and has repositories

## Integration Examples

### Shell Script Integration
```bash
#!/bin/bash
# weekly_report.sh - Generate weekly reports

CURRENT_DATE=$(date +%Y-%m-%d)
WEEK_AGO=$(date -d '7 days ago' +%Y-%m-%d)

for org in "myorg1" "myorg2" "myorg3"; do
    echo "Generating report for $org..."
    ./ghreporting -target "$org" -since "$WEEK_AGO" -until "$CURRENT_DATE" \
                  -output "weekly_report_${org}_$(date +%Y%m%d).txt"
done
```

### CI/CD Pipeline Integration
```yaml
# .github/workflows/weekly-report.yml
name: Weekly GitHub Activity Report

on:
  schedule:
    - cron: '0 9 * * 1'  # Every Monday at 9 AM

jobs:
  generate-report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Build reporting tool
        run: go build -o ghreporting .
      - name: Generate weekly report
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          WEEK_AGO=$(date -d '7 days ago' +%Y-%m-%d)
          TODAY=$(date +%Y-%m-%d)
          ./ghreporting -target myorg -since $WEEK_AGO -until $TODAY \
                        -output weekly_report.txt
      - name: Upload report
        uses: actions/upload-artifact@v2
        with:
          name: weekly-report
          path: weekly_report.txt
```