package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type MainMenu struct {
	choices     []string
	cursor      int
	environments []string
}

func NewMainMenu() MainMenu {
	// Load actual environments from storage
	stor := storage.NewStorage(".")
	envList, _ := stor.ListEnvironments()
	
	envDisplay := []string{}
	for _, env := range envList {
		// Load to count servers
		envData, err := stor.LoadEnvironment(env)
		if err == nil {
			envDisplay = append(envDisplay, fmt.Sprintf("%s (%d servers)", env, len(envData.Servers)))
		} else {
			envDisplay = append(envDisplay, env)
		}
	}
	
	return MainMenu{
		choices: []string{
			"Create new environment",
			"Manage existing environment",
			"Validate all inventories",
			"Quit",
		},
		environments: envDisplay,
	}
}

func (m MainMenu) Init() tea.Cmd {
	return nil
}

func (m MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter":
			// Handle selection
			switch m.cursor {
			case 0:
				// Create new environment
				return NewEnvironmentForm(), nil
			case 1:
				// Manage environment
				return NewEnvironmentSelector(), nil
			case 2:
				// Validate - TODO
				return m, nil
			case 3:
				// Quit
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m MainMenu) View() string {
	s := titleStyle.Render("🔧 Ansible Inventory Manager v1.0")
	s += "\n\n"

	// Existing environments
	s += "📁 Existing environments:\n"
	for _, env := range m.environments {
		s += fmt.Sprintf("   • %s\n", env)
	}
	s += "\n"

	// Menu choices
	menuBox := ""
	for i, choice := range m.choices {
		cursor := "  "
		style := normalItemStyle
		if m.cursor == i {
			cursor = "▶ "
			style = selectedItemStyle
		}
		menuBox += style.Render(cursor+choice) + "\n"
	}

	s += infoBoxStyle.Render(menuBox)
	s += "\n"
	s += helpStyle.Render("[↑↓/jk] Navigate  [Enter] Select  [q] Quit")

	return lipgloss.NewStyle().Margin(1, 2).Render(s)
}
