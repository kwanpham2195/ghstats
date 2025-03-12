# Github Contributor Stats Exporter

This tool fetches GitHub contributor statistics for repositories and exports the data as CSV. It provides an interactive interface for easy configuration and execution.

## Features

- Interactive TUI (Terminal User Interface)
- Date range selection with smart defaults
- GitHub token handling with environment variable support
- CSV export of contributor statistics
- Support for multiple repositories processing
- Progress indicator during data fetching

## Prerequisites

- Go must be installed
- GitHub fine-grained token with required permissions:
  - For public repositories: minimal scopes are sufficient
  - For private repositories: `metadata` - `read` scope is required
- Token can be provided via:
  - Environment variable: `GITHUB_TOKEN`
  - Interactive prompt during execution

## Installation

You can install this tool using one of these methods:

From source:

```bash
git clone https://github.com/kwanpham2195/ghstats.git
cd ghstats
go run .
```

Via go install:

```bash
go install github.com/kwanpham2195/ghstats@latest
```

## Usage

1. Prepare a repositories file (default: `repos.txt`) containing target repositories:

```
owner1/repo1
owner2/repo2
organization/repo3
```

2. Run the application:

```bash
ghstats
```

3. Follow the interactive prompts:

   - Start Date (default: start of current month)
   - End Date (default: today)
   - GitHub Token (skipped if GITHUB_TOKEN environment variable exists)
   - Output File Path (default: output.csv)
   - Repositories File Path (default: repos.txt)

4. The tool will process each repository and generate a CSV file containing:
   - Repository name
   - Contributor username
   - Number of additions
   - Number of deletions
   - Number of commits
   - Date range

## Example Output

The generated CSV file will look like this:

```csv
Repository,Contributor,Additions,Deletions,Commits,StartDate,EndDate
owner1/repo1,user1,150,50,10,2024-03-01,2024-03-25
owner1/repo1,user2,300,100,15,2024-03-01,2024-03-25
owner2/repo2,user3,200,75,8,2024-03-01,2024-03-25
```
