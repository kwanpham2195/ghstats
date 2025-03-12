package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type ContributorStats struct {
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Weeks []struct {
		Week      int64 `json:"w"`
		Additions int   `json:"a"`
		Deletions int   `json:"d"`
		Commits   int   `json:"c"`
	} `json:"weeks"`
}

func main() {
	reposFile := flag.String("repos-file", "repos.txt", "Path to repository list file, one repo per line (required)")
	startDate := flag.String("start", "", "Start date in YYYY-MM-DD format (default: 1 month ago)")
	endDate := flag.String("end", "", "End date in YYYY-MM-DD format (default: today)")
	output := flag.String("output", "csv", "Output format (csv or json)")
	flag.Parse()
	slog.Info("Program started")

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		if err := godotenv.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading .env file: %v\n", err)
			os.Exit(1)
		}
		slog.Info("Loaded .env file successfully")
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		flag.Usage()
		os.Exit(1)
	} else {
		slog.Info("GITHUB_TOKEN read from environment")
	}

	data, err := os.ReadFile(*reposFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading repository file: %v\n", err)
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")
	var repos []string
	for _, line := range lines {
		repo := strings.TrimSpace(line)
		if repo != "" {
			repos = append(repos, repo)
		}
	}
	slog.Info("Loaded repositories", "count", len(repos))
	start, end := parseDates(*startDate, *endDate)
	client := &http.Client{}

	var writer *csv.Writer
	if *output == "csv" {
		outputFile, err := os.Create("output.csv")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()
		writer = csv.NewWriter(outputFile)
		writer.Write([]string{"Repository", "Contributor", "Additions", "Deletions", "Commits", "StartDate", "EndDate"})
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	defer writer.Flush()

	for _, repo := range repos {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Invalid repository format: %s\n", repo)
			continue
		}
		owner, repoName := parts[0], parts[1]
		slog.Info("Fetching contributor stats", "repository", repo)

		stats, err := fetchContributorStats(client, owner, repoName, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching stats for %s: %v\n", repo, err)
			continue
		}

		processStats(stats, repo, start, end, writer, *output)
	}
}

func parseDates(startStr, endStr string) (time.Time, time.Time) {
	loc := time.UTC
	now := time.Now().In(loc)
	var start time.Time

	if endStr == "" {
		endStr = now.Format("2006-01-02")
	}
	end, _ := time.ParseInLocation("2006-01-02", endStr, loc)
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	if startStr == "" {
		start = end.AddDate(0, -1, 0)
	} else {
		start, _ = time.ParseInLocation("2006-01-02", startStr, loc)
	}
	return start.UTC(), end.UTC()
}

func fetchContributorStats(client *http.Client, owner, repo, token string) ([]ContributorStats, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/stats/contributors", owner, repo)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")

	for {
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			var stats []ContributorStats
			if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
				return nil, err
			}
			return stats, nil
		case http.StatusAccepted:
			slog.Info("GitHub API accepted request, waiting for stats to be generated...")
			time.Sleep(2 * time.Second)
			continue
		default:
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}
}

func processStats(stats []ContributorStats, repo string, start, end time.Time, writer *csv.Writer, format string) {
	for _, contributor := range stats {
		var totalAdditions, totalDeletions, totalCommits int

		for _, week := range contributor.Weeks {
			weekStart := time.Unix(week.Week, 0).UTC()
			weekEnd := weekStart.Add(7 * 24 * time.Hour)

			if weekStart.After(end) || weekEnd.Before(start) {
				continue
			}

			totalAdditions += week.Additions
			totalDeletions += week.Deletions
			totalCommits += week.Commits
		}

		if totalAdditions == 0 && totalDeletions == 0 && totalCommits == 0 {
			continue
		}

		switch format {
		case "csv":
			writer.Write([]string{
				repo,
				contributor.Author.Login,
				fmt.Sprintf("%d", totalAdditions),
				fmt.Sprintf("%d", totalDeletions),
				fmt.Sprintf("%d", totalCommits),
				start.Format("2006-01-02"),
				end.Format("2006-01-02"),
			})
		case "json":
			json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"repository":  repo,
				"contributor": contributor.Author.Login,
				"additions":   totalAdditions,
				"deletions":   totalDeletions,
				"commits":     totalCommits,
				"start_date":  start.Format("2006-01-02"),
				"end_date":    end.Format("2006-01-02"),
			})
		}
	}
}
