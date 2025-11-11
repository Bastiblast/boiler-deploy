package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type EnvironmentDeleter struct {
	environments []string
	cursor       int
	storage      *storage.Storage
	confirming   bool
	confirmIndex int
	err          error
	success      bool
}

func NewEnvironmentDeleter() EnvironmentDeleter {
	stor := storage.NewStorage(".")
	envs, _ := stor.ListEnvironments()
	
	return EnvironmentDeleter{
		environments: envs,
		storage:      stor,
		confirming:   false,
		confirmIndex: 0,
	}
}

func (d EnvironmentDeleter) Init() tea.Cmd {
	return nil
}

func (d EnvironmentDeleter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if d.confirming {
				// Cancel confirmation
				d.confirming = false
				d.err = nil
				return d, nil
			}
			// Return to main menu
			return NewMainMenu(), nil
			
		case "up", "k":
			if d.confirming {
				if d.confirmIndex > 0 {
					d.confirmIndex--
				}
			} else {
				if d.cursor > 0 {
					d.cursor--
				}
			}
			
		case "down", "j":
			if d.confirming {
				if d.confirmIndex < 1 {
					d.confirmIndex++
				}
			} else {
				if d.cursor < len(d.environments)-1 {
					d.cursor++
				}
			}
			
		case "enter":
			if len(d.environments) == 0 {
				return NewMainMenu(), nil
			}
			
			if d.confirming {
				if d.confirmIndex == 0 {
					// Yes - delete
					envName := d.environments[d.cursor]
					if err := d.storage.DeleteEnvironment(envName); err != nil {
						d.err = fmt.Errorf("failed to delete: %v", err)
						d.confirming = false
						return d, nil
					}
					d.success = true
					// Return to main menu after 2 seconds
					return NewMainMenu(), nil
				} else {
					// No - cancel
					d.confirming = false
					d.err = nil
				}
			} else {
				// Show confirmation
				d.confirming = true
				d.confirmIndex = 1 // Default to "No"
			}
			
		case "q":
			if !d.confirming {
				return NewMainMenu(), nil
			}
		}
	}
	
	return d, nil
}

func (d EnvironmentDeleter) View() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("🗑️  Delete Environment"))
	b.WriteString("\n\n")
	
	if len(d.environments) == 0 {
		b.WriteString(errorStyle.Render("No environments found"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("[Enter/Esc] Back to menu"))
		return b.String()
	}
	
	if d.confirming {
		// Confirmation dialog
		envName := d.environments[d.cursor]
		b.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  Delete '%s'?", envName)))
		b.WriteString("\n\n")
		b.WriteString("This will permanently delete:\n")
		b.WriteString(fmt.Sprintf("  • inventory/%s/\n", envName))
		b.WriteString("  • All servers and configurations\n\n")
		b.WriteString("Are you sure?\n\n")
		
		// Confirmation choices
		choices := []string{"Yes, delete", "No, cancel"}
		for i, choice := range choices {
			cursor := "  "
			style := normalItemStyle
			if d.confirmIndex == i {
				cursor = "▶ "
				style = selectedItemStyle
			}
			b.WriteString(style.Render(cursor+choice) + "\n")
		}
		
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[↑↓] Select  [Enter] Confirm  [Esc] Cancel"))
	} else {
		// Environment list
		b.WriteString("Select environment to delete:\n\n")
		
		for i, env := range d.environments {
			cursor := "  "
			style := normalItemStyle
			if d.cursor == i {
				cursor = "▶ "
				style = selectedItemStyle
			}
			
			// Load server count
			envData, err := d.storage.LoadEnvironment(env)
			serverCount := "? servers"
			if err == nil {
				serverCount = fmt.Sprintf("%d servers", len(envData.Servers))
			}
			
			b.WriteString(style.Render(fmt.Sprintf("%s%s (%s)", cursor, env, serverCount)) + "\n")
		}
		
		b.WriteString("\n")
		
		if d.err != nil {
			b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", d.err)))
			b.WriteString("\n\n")
		}
		
		if d.success {
			b.WriteString(successStyle.Render("✓ Environment deleted successfully!"))
			b.WriteString("\n\n")
		}
		
		b.WriteString(helpStyle.Render("[↑↓/jk] Navigate  [Enter] Delete  [Esc/q] Back"))
	}
	
	return b.String()
}
