package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	progressPadding      = 2
	defaultTaskDuration  = 25 * time.Minute
	defaultBreakDuration = 5 * time.Minute
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type tickMsg time.Time

type errMsg error

type model struct {
	progress progress.Model

	timeTotalSec   int64
	timeElapsedSec int64
	isTimerRunning bool
	quitting       bool

	err error
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

var startStopKeys = key.NewBinding(
	key.WithKeys("s"),
	key.WithHelp("", "press s to start/stop timer"),
)

func initialModel() model {
	return model{progress: progress.New(progress.WithDefaultGradient())}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		if key.Matches(msg, startStopKeys) {
			if !m.isTimerRunning {
				m.timeTotalSec = int64(defaultTaskDuration.Seconds())
				m.timeElapsedSec = 0
			}
			m.isTimerRunning = !m.isTimerRunning
			if m.isTimerRunning {
				return m, tickCmd()
			}

			return m, m.progress.SetPercent(0.0)
		}
		return m, nil
	case tickMsg:
		if m.progress.Percent() == 1.0 {
			return m, tea.Quit
		}

		if !m.isTimerRunning {
			return m, nil
		}

		m.timeElapsedSec++

		var cmd tea.Cmd
		if m.timeElapsedSec >= m.timeTotalSec {
			cmd = m.progress.SetPercent(1.0)
		} else {
			cmd = m.progress.SetPercent(float64(m.timeElapsedSec) / float64(m.timeTotalSec))
		}

		return m, tea.Batch(tickCmd(), cmd)

		// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case errMsg:
		m.err = msg
		return m, nil

	default:
		return m, nil
	}
}

func (m model) View() string {
	screen := "\n"
	pad := strings.Repeat(" ", progressPadding)

	if m.isTimerRunning {
		screen += pad + m.progress.View() + "\n"
		screen += pad + fmt.Sprintf("%ds / %ds", m.timeElapsedSec, m.timeTotalSec)
	}

	screen += "\n\n"

	screen += "\n" + helpStyle(startStopKeys.Help().Desc)
	screen += "\n" + helpStyle(quitKeys.Help().Desc)

	return screen
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
