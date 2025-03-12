package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
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
	bar := m.prog.View()
	return fmt.Sprintf("Processing: %s %.0f%%\n", bar, m.percent*100)
}
