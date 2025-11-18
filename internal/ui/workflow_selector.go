package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type WorkflowSelector struct {
	environments []string
	cursor       int
	storage      *storage.Storage
	err          error
}

func NewWorkflowSelector() WorkflowSelector {
	stor := storage.NewStorage(".")
	envs, err := stor.ListEnvironments()
	
	return WorkflowSelector{
		environments: envs,
		cursor:       0,
		storage:      stor,
		err:          err,
	}
}

func (w WorkflowSelector) Init() tea.Cmd {
	return nil
}

func (w WorkflowSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			// Return to main menu
			return NewMainMenu(), nil

		case "up", "k":
			if w.cursor > 0 {
				w.cursor--
			}

		case "down", "j":
			if w.cursor < len(w.environments)-1 {
				w.cursor++
			}

		case "enter":
			if len(w.environments) > 0 {
				selectedEnv := w.environments[w.cursor]
				// Create WorkflowView with selected environment
				wv, err := NewWorkflowViewWithEnv(selectedEnv)
				if err != nil {
					w.err = err
					return w, nil
				}
				return wv, wv.Init()
			}
		}
	}

	return w, nil
}

func (w WorkflowSelector) View() string {
	s := titleStyle.Render("📋 Select Environment for Inventory Work")
	s += "\n\n"

	if w.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %v", w.err))
		s += "\n\n"
		s += helpStyle.Render("[Esc] Back to menu")
		return lipgloss.NewStyle().Margin(1, 2).Render(s)
	}

	if len(w.environments) == 0 {
		s += "No environments found.\n"
		s += "Create one first from the main menu.\n\n"
		s += helpStyle.Render("[Esc] Back to menu")
		return lipgloss.NewStyle().Margin(1, 2).Render(s)
	}

	s += "Select an environment to validate inventory, provision and deploy:\n\n"

	// List environments
	for i, env := range w.environments {
		cursor := "  "
		style := normalItemStyle
		if w.cursor == i {
			cursor = "▶ "
			style = selectedItemStyle
		}
		
		// Show environment details
		envData, err := w.storage.LoadEnvironment(env)
		displayText := env
		if err == nil {
			displayText = fmt.Sprintf("%s (%d servers)", env, len(envData.Servers))
		}
		
		s += style.Render(cursor+displayText) + "\n"
	}

	s += "\n"
	s += helpStyle.Render("[↑↓/jk] Navigate  [Enter] Select  [Esc] Back to menu")

	return lipgloss.NewStyle().Margin(1, 2).Render(s)
}
