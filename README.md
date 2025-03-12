# Github Contributor Stats Exporters

This tool fetches GitHub contributor statistics for repositories listed in a file and exports the data as CSV or JSON.

## Prerequisites

- Go must be installed.
- Export the environment variable `GITHUB_TOKEN` in your shell (e.g. `export GITHUB_TOKEN=your_token_here`).
- Ensure that the GitHub token has the required permissions: for public repositories, minimal scopes are sufficient; for private repositories, grant the token `metadata` - `read`
  permissions.

## Installation

You can install this tool from source or directly via go install.

From source:

```bash
git clone https://github.com/kwanpham2195/ghstats.git
cd ghstats
go build -o ghstats main.go
```

Or use go install:

```bash
go install github.com/kwanpham2195/ghstats@latest
```

## Usage

To run the application, execute the following command:

```bash
ghstats --repos-file=repos.txt --start=2025-03-01 --end=2025-03-30 --output=csv
```

- `--repos-file`: Path to the file containing repositories (one repo per line, formatted as "owner/repo").
- `--start`: Start date in YYYY-MM-DD format.
- `--end`: End date in YYYY-MM-DD format.
- `--output`: Output format ("csv" or "json").

Enjoy!
