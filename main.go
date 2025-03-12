package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
	p := tea.NewProgram(newInputModel())
	model, err := p.StartReturningModel()
	if err != nil {
		slog.Error("TUI input error", "err", err)
		os.Exit(1)
	}
	inp := model.(inputModel)
	if inp.githubToken != "" {
		os.Setenv("GITHUB_TOKEN", inp.githubToken)
	}
	runProcessing(inp.startDate, inp.endDate, inp.outputPath, inp.reposPath)
}

func runProcessing(startDate, endDate, out, reposPath string) {
	// Read repositories file:
	data, err := os.ReadFile(reposPath)
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

	// Setup output generation and HTTP client:
	parsedStart, parsedEnd := parseDates(startDate, endDate)
	outputFile, err := os.Create("output.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)
	writer.Write([]string{"Repository", "Contributor", "Additions", "Deletions", "Commits", "StartDate", "EndDate"})
	defer writer.Flush()
	client := &http.Client{}

	processing := newProcessingModel(repos, parsedStart, parsedEnd, out, client, writer)
	p := tea.NewProgram(processing)
	if err := p.Start(); err != nil {
		// Error running processing TUI
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
			time.Sleep(1 * time.Second)
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

		writer.Write([]string{
			repo,
			contributor.Author.Login,
			fmt.Sprintf("%d", totalAdditions),
			fmt.Sprintf("%d", totalDeletions),
			fmt.Sprintf("%d", totalCommits),
			start.Format("2006-01-02"),
			end.Format("2006-01-02"),
		})
	}
}

type (
	repoProcessedMsg string
	TickMsg          time.Time
)

func processRepo(repo string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(2 * time.Second)
		return repoProcessedMsg(repo)
	}
}

type processingModel struct {
	repos   []string
	current int
	spinner spinner.Model
	done    bool
	message string
	start   time.Time
	end     time.Time
	out     string
	client  *http.Client
	writer  *csv.Writer
}

func newProcessingModel(repos []string, start, end time.Time, out string, client *http.Client, writer *csv.Writer) processingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return processingModel{
		repos:   repos,
		current: 0,
		spinner: s,
		done:    false,
		start:   start,
		end:     end,
		out:     out,
		client:  client,
		writer:  writer,
	}
}

func (m processingModel) Init() tea.Cmd {
	if len(m.repos) > 0 {
		return tea.Batch(m.spinner.Tick, m.processCurrentRepo())
	}
	return m.spinner.Tick
}

func (m processingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case repoProcessedMsg:
		m.message = fmt.Sprintf("Processed repository: %s", msg)
		m.current++
		if m.current < len(m.repos) {
			return m, tea.Batch(tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return TickMsg(t)
			}), m.processCurrentRepo())
		} else {
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m processingModel) View() string {
	if m.done {
		return "All repositories processed.\n"
	}
	currentRepo := ""
	if m.current < len(m.repos) {
		currentRepo = m.repos[m.current]
	}
	return fmt.Sprintf("Processing repository %d/%d: %s\n%s", m.current+1, len(m.repos), currentRepo, m.spinner.View())
}

func (m processingModel) processCurrentRepo() tea.Cmd {
	return func() tea.Msg {
		repo := m.repos[m.current]
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Invalid repository format: %s\n", repo)
			return repoProcessedMsg(repo)
		}
		owner, repoName := parts[0], parts[1]
		stats, err := fetchContributorStats(m.client, owner, repoName, os.Getenv("GITHUB_TOKEN"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching stats for %s: %v\n", repo, err)
			return repoProcessedMsg(repo)
		}
		processStats(stats, repo, m.start, m.end, m.writer, m.out)
		return repoProcessedMsg(repo)
	}
}
