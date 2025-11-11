package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type EnvironmentSelector struct {
	environments []string
	cursor       int
	storage      *storage.Storage
	err          error
	selected     string
}

func NewEnvironmentSelector() EnvironmentSelector {
	stor := storage.NewStorage(".")
	envs, err := stor.ListEnvironments()
	
	return EnvironmentSelector{
		environments: envs,
		cursor:       0,
		storage:      stor,
		err:          err,
	}
}

func (e EnvironmentSelector) Init() tea.Cmd {
	return nil
}

func (e EnvironmentSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			// Return to main menu
			return NewMainMenu(), nil

		case "up", "k":
			if e.cursor > 0 {
				e.cursor--
			}

		case "down", "j":
			if e.cursor < len(e.environments)-1 {
				e.cursor++
			}

		case "enter":
			if len(e.environments) > 0 {
				e.selected = e.environments[e.cursor]
				// Load environment and go to server manager
				env, err := e.storage.LoadEnvironment(e.selected)
				if err != nil {
					e.err = err
					return e, nil
				}
				return NewServerManager(env), nil
			}
		}
	}

	return e, nil
}

func (e EnvironmentSelector) View() string {
	s := titleStyle.Render("📁 Select Environment to Manage")
	s += "\n\n"

	if e.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %v", e.err))
		s += "\n\n"
		s += helpStyle.Render("[Esc] Back to menu")
		return lipgloss.NewStyle().Margin(1, 2).Render(s)
	}

	if len(e.environments) == 0 {
		s += "No environments found.\n"
		s += "Create one first from the main menu.\n\n"
		s += helpStyle.Render("[Esc] Back to menu")
		return lipgloss.NewStyle().Margin(1, 2).Render(s)
	}

	// List environments
	for i, env := range e.environments {
		cursor := "  "
		style := normalItemStyle
		if e.cursor == i {
			cursor = "▶ "
			style = selectedItemStyle
		}
		s += style.Render(cursor+env) + "\n"
	}

	s += "\n"
	s += helpStyle.Render("[↑↓/jk] Navigate  [Enter] Select  [Esc] Back")

	return lipgloss.NewStyle().Margin(1, 2).Render(s)
}
