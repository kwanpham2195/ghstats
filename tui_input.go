package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputState int

const (
	inputStart inputState = iota
	inputEnd
	inputToken
	inputOutput
	inputRepos
	inputDone
)

type inputModel struct {
	state inputState

	start  textinput.Model
	end    textinput.Model
	token  textinput.Model
	output textinput.Model
	repos  textinput.Model

	startDate    string
	endDate      string
	githubToken  string
	outputFormat string
	outputPath   string
	reposPath    string
}

func getDefaultDates() (string, string) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return startOfMonth.Format("2006-01-02"), now.Format("2006-01-02")
}

func newInputModel() inputModel {
	m := inputModel{
		state:  inputStart,
		start:  textinput.New(),
		end:    textinput.New(),
		token:  textinput.New(),
		output: textinput.New(),
		repos:  textinput.New(),
	}
	m.start.Placeholder = "YYYY-MM-DD"
	m.start.Prompt = "Start Date: "
	m.end.Placeholder = "YYYY-MM-DD"
	m.end.Prompt = "End Date: "
	m.token.Placeholder = "github_pat_..."
	m.token.Prompt = "GitHub Token: "
	m.output.Placeholder = "output.csv"
	m.output.Prompt = "Output File Path: "
	m.repos.Placeholder = "repos.txt"
	m.repos.Prompt = "Repositories File Path: "
	// Set proper focus:
	m.start.Focus()
	m.end.Blur()
	m.token.Blur()
	m.output.Blur()
	m.repos.Blur()
	return m
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			switch m.state {
			case inputStart:
				m.startDate = m.start.Value()
				if m.startDate == "" {
					defaultStart, _ := getDefaultDates()
					m.startDate = defaultStart
				}
				m.state = inputEnd
				m.start.Blur()
				m.end.Focus()
			case inputEnd:
				m.endDate = m.end.Value()
				if m.endDate == "" {
					_, defaultEnd := getDefaultDates()
					m.endDate = defaultEnd
				}
				if os.Getenv("GITHUB_TOKEN") == "" {
					m.state = inputToken
					m.end.Blur()
					m.token.Focus()
				} else {
					m.state = inputOutput
					m.end.Blur()
					m.output.Focus()
				}
			case inputToken:
				m.githubToken = m.token.Value()
				m.state = inputOutput
				m.token.Blur()
				m.output.Focus()
			case inputOutput:
				m.outputPath = m.output.Value()
				if m.outputPath == "" {
					m.outputPath = "output.csv"
				}
				m.state = inputRepos
				m.output.Blur()
				m.repos.Focus()
			case inputRepos:
				m.reposPath = m.repos.Value()
				if m.reposPath == "" {
					m.reposPath = "repos.txt"
				}
				m.state = inputDone
				m.repos.Blur()
				m.outputFormat = "csv"
				return m, tea.Quit
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	switch m.state {
	case inputStart:
		m.start, cmd = m.start.Update(msg)
	case inputEnd:
		m.end, cmd = m.end.Update(msg)
	case inputToken:
		m.token, cmd = m.token.Update(msg)
	case inputOutput:
		m.output, cmd = m.output.Update(msg)
	case inputRepos:
		m.repos, cmd = m.repos.Update(msg)
	}
	return m, cmd
}

func (m inputModel) View() string {
	if m.state == inputDone {
		var sb strings.Builder
		defaultStart, defaultEnd := getDefaultDates()

		sb.WriteString("Inputs received:\n")
		sb.WriteString(fmt.Sprintf("Start Date: %s\n", getValueOrDefault(m.startDate, defaultStart+" (start of month)")))
		sb.WriteString(fmt.Sprintf("End Date: %s\n", getValueOrDefault(m.endDate, defaultEnd+" (today)")))
		if m.githubToken != "" {
			sb.WriteString("GitHub Token: [provided]\n")
		}
		sb.WriteString(fmt.Sprintf("Output File: %s\n", getValueOrDefault(m.outputPath, "output.csv")))
		sb.WriteString(fmt.Sprintf("Repositories File: %s\n", getValueOrDefault(m.reposPath, "repos.txt")))
		sb.WriteString("\nStart processing...")
		return sb.String()
	}

	s := "Enter parameters for processing:\n\n"
	switch m.state {
	case inputStart:
		s += m.start.View()
	case inputEnd:
		s += m.end.View()
	case inputToken:
		s += m.token.View()
	case inputOutput:
		s += m.output.View()
	case inputRepos:
		s += m.repos.View()
	}
	s += "\n\n(Press Enter for default value and continue)"
	return s
}

func getValueOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
