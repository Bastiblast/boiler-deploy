package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type EnvironmentSelector struct {
	environments     []string
	cursor           int
	storage          *storage.Storage
	err              error
	selected         string
	confirmingDelete bool
	confirmIndex     int
}

func NewEnvironmentSelector() EnvironmentSelector {
	stor := storage.NewStorage(".")
	envs, err := stor.ListEnvironments()
	
	return EnvironmentSelector{
		environments:     envs,
		cursor:           0,
		storage:          stor,
		err:              err,
		confirmingDelete: false,
		confirmIndex:     0,
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
			if e.confirmingDelete {
				// Cancel confirmation
				e.confirmingDelete = false
				e.err = nil
				return e, nil
			}
			// Return to main menu
			return NewMainMenu(), nil

		case "up", "k":
			if e.confirmingDelete {
				if e.confirmIndex > 0 {
					e.confirmIndex--
				}
			} else {
				if e.cursor > 0 {
					e.cursor--
				}
			}

		case "down", "j":
			if e.confirmingDelete {
				if e.confirmIndex < 1 {
					e.confirmIndex++
				}
			} else {
				if e.cursor < len(e.environments)-1 {
					e.cursor++
				}
			}

		case "enter":
			if e.confirmingDelete {
				if e.confirmIndex == 0 {
					// Yes - delete environment
					envName := e.environments[e.cursor]
					if err := e.storage.DeleteEnvironment(envName); err != nil {
						e.err = fmt.Errorf("failed to delete: %v", err)
						e.confirmingDelete = false
						return e, nil
					}
					// Refresh environment list
					envs, err := e.storage.ListEnvironments()
					e.environments = envs
					e.err = err
					e.confirmingDelete = false
					// Adjust cursor
					if e.cursor >= len(e.environments) && e.cursor > 0 {
						e.cursor--
					}
					return e, nil
				} else {
					// No - cancel
					e.confirmingDelete = false
				}
			} else {
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

		case "d", "x":
			if !e.confirmingDelete && len(e.environments) > 0 {
				// Delete environment
				e.confirmingDelete = true
				e.confirmIndex = 1 // Default to "No"
			}
		}
	}

	return e, nil
}

func (e EnvironmentSelector) View() string {
	s := titleStyle.Render("📁 Select Environment to Manage")
	s += "\n\n"

	if e.confirmingDelete && len(e.environments) > 0 {
		envName := e.environments[e.cursor]
		s += errorStyle.Render(fmt.Sprintf("⚠️  Delete environment '%s'?", envName))
		s += "\n\n"
		s += "This will permanently delete:\n"
		s += fmt.Sprintf("  • inventory/%s/\n", envName)
		s += "  • All server configurations\n\n"
		s += "Are you sure?\n\n"
		
		// Confirmation choices
		choices := []string{"Yes, delete environment", "No, cancel"}
		for i, choice := range choices {
			cursor := "  "
			style := normalItemStyle
			if e.confirmIndex == i {
				cursor = "▶ "
				style = selectedItemStyle
			}
			s += style.Render(cursor+choice) + "\n"
		}
		
		s += "\n"
		s += helpStyle.Render("[↑↓] Select  [Enter] Confirm  [Esc] Cancel")
		return lipgloss.NewStyle().Margin(1, 2).Render(s)
	}

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
	s += helpStyle.Render("[↑↓/jk] Navigate  [Enter] Select  [d/x] Delete  [Esc] Back")

	return lipgloss.NewStyle().Margin(1, 2).Render(s)
}
