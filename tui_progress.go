package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	processStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	progressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575"))
)

type progressModel struct {
	percent float64
	prog    progress.Model
}

func newProgressModel() progressModel {
	return progressModel{
		prog: progress.New(progress.WithDefaultGradient()),
	}
}

func (m progressModel) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return t
	})
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case time.Time:
		if m.percent < 1.0 {
			m.percent += 0.02
			return m, tickCmd()
		}
	}
	return m, nil
}

func (m progressModel) View() string {
	bar := progressBarStyle.Render(m.prog.View())
	return fmt.Sprintf("%s %s %.0f%%\n", processStyle.Render("Processing:"), bar, m.percent*100)
}
