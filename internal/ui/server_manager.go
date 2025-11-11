package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type ServerManager struct {
	environment      *inventory.Environment
	cursor           int
	storage          *storage.Storage
	message          string
	messageType      string // "success", "error", "info"
	confirmingDelete bool
	confirmIndex     int
}

func NewServerManager(env *inventory.Environment) ServerManager {
	return ServerManager{
		environment:      env,
		cursor:           0,
		storage:          storage.NewStorage("."),
		confirmingDelete: false,
		confirmIndex:     0,
	}
}

func (m ServerManager) Init() tea.Cmd {
	return nil
}

func (m ServerManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.confirmingDelete {
				// Cancel confirmation
				m.confirmingDelete = false
				m.message = ""
				return m, nil
			}
			// Return to main menu
			return NewMainMenu(), nil

		case "up", "k":
			if m.confirmingDelete {
				if m.confirmIndex > 0 {
					m.confirmIndex--
				}
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.confirmingDelete {
				if m.confirmIndex < 1 {
					m.confirmIndex++
				}
			} else {
				if m.cursor < len(m.environment.Servers) {
					m.cursor++
				}
			}

		case "a":
			if !m.confirmingDelete {
				// Add new server
				return NewServerForm(m.environment), nil
			}

		case "e":
			if !m.confirmingDelete {
				// Edit selected server
				if m.cursor < len(m.environment.Servers) {
					return NewServerForm(m.environment, &m.environment.Servers[m.cursor]), nil
				}
			}

		case "d":
			if !m.confirmingDelete {
				// Delete selected server
				if m.cursor < len(m.environment.Servers) && len(m.environment.Servers) > 0 {
					serverName := m.environment.Servers[m.cursor].Name
					// Remove server
					m.environment.Servers = append(
						m.environment.Servers[:m.cursor],
						m.environment.Servers[m.cursor+1:]...,
					)
					
					// Save
					if err := m.storage.SaveEnvironment(*m.environment); err != nil {
						m.message = fmt.Sprintf("Failed to save: %v", err)
						m.messageType = "error"
					} else {
						m.message = fmt.Sprintf("✓ Server '%s' deleted", serverName)
						m.messageType = "success"
					}
					
					// Adjust cursor
					if m.cursor >= len(m.environment.Servers) && m.cursor > 0 {
						m.cursor--
					}
				}
			}
		
		case "x":
			if !m.confirmingDelete {
				// Delete environment
				m.confirmingDelete = true
				m.confirmIndex = 1 // Default to "No"
			}
		
		case "enter":
			if m.confirmingDelete {
				if m.confirmIndex == 0 {
					// Yes - delete environment
					envName := m.environment.Name
					if err := m.storage.DeleteEnvironment(envName); err != nil {
						m.message = fmt.Sprintf("Failed to delete: %v", err)
						m.messageType = "error"
						m.confirmingDelete = false
						return m, nil
					}
					// Return to main menu after successful deletion
					return NewMainMenu(), nil
				} else {
					// No - cancel
					m.confirmingDelete = false
					m.message = ""
				}
			}

		case "s":
			if !m.confirmingDelete {
				// Save environment
				if err := m.storage.SaveEnvironment(*m.environment); err != nil {
					m.message = fmt.Sprintf("Failed to save: %v", err)
					m.messageType = "error"
				} else {
					m.message = "✓ Environment saved"
					m.messageType = "success"
				}
			}

		case "g":
			if !m.confirmingDelete {
				// Generate and show YAML
				generator := inventory.NewGenerator()
				summary := generator.GenerateEnvironmentSummary(*m.environment)
				m.message = summary
				m.messageType = "info"
			}
		}
	}

	return m, nil
}

func (m ServerManager) View() string {
	var b strings.Builder

	title := fmt.Sprintf("🖥️  Manage Environment: %s", m.environment.Name)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")
	
	// Show confirmation dialog if deleting environment
	if m.confirmingDelete {
		b.WriteString(errorStyle.Render(fmt.Sprintf("⚠️  Delete environment '%s'?", m.environment.Name)))
		b.WriteString("\n\n")
		b.WriteString("This will permanently delete:\n")
		b.WriteString(fmt.Sprintf("  • inventory/%s/\n", m.environment.Name))
		b.WriteString(fmt.Sprintf("  • All %d server(s) and configurations\n\n", len(m.environment.Servers)))
		b.WriteString("Are you sure?\n\n")
		
		// Confirmation choices
		choices := []string{"Yes, delete environment", "No, cancel"}
		for i, choice := range choices {
			cursor := "  "
			style := normalItemStyle
			if m.confirmIndex == i {
				cursor = "▶ "
				style = selectedItemStyle
			}
			b.WriteString(style.Render(cursor+choice) + "\n")
		}
		
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("[↑↓] Select  [Enter] Confirm  [Esc] Cancel"))
		return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
	}

	// Services enabled
	b.WriteString("Services: ")
	services := []string{}
	if m.environment.Services.Web {
		services = append(services, "Web")
	}
	if m.environment.Services.Database {
		services = append(services, "Database")
	}
	if m.environment.Services.Monitoring {
		services = append(services, "Monitoring")
	}
	b.WriteString(strings.Join(services, ", "))
	b.WriteString("\n\n")

	// Servers list
	if len(m.environment.Servers) == 0 {
		b.WriteString("No servers configured yet.\n")
		b.WriteString("Press 'a' to add a server.\n\n")
	} else {
		b.WriteString(fmt.Sprintf("Servers (%d):\n\n", len(m.environment.Servers)))

		// Table header
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
		b.WriteString(headerStyle.Render("  Name                IP              Port    Type    Status"))
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", 70))
		b.WriteString("\n")

		// Table rows
		for i, server := range m.environment.Servers {
			cursor := "  "
			style := normalItemStyle
			if m.cursor == i {
				cursor = "▶ "
				style = selectedItemStyle
			}

			status := "⚠ Not tested"
			// TODO: Add SSH test status here

			row := fmt.Sprintf("%s%-18s %-15s %-7d %-7s %s",
				cursor,
				truncate(server.Name, 18),
				server.IP,
				server.AppPort,
				server.Type,
				status,
			)
			b.WriteString(style.Render(row))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	// Message display
	if m.message != "" {
		b.WriteString("\n")
		switch m.messageType {
		case "success":
			b.WriteString(successStyle.Render(m.message))
		case "error":
			b.WriteString(errorStyle.Render(m.message))
		case "info":
			b.WriteString(infoBoxStyle.Render(m.message))
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	helpText := "[a] Add  [e] Edit  [d] Delete Server  [x] Delete Environment  [s] Save  [g] Generate  [Esc] Back"
	b.WriteString(helpStyle.Render(helpText))

	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}

// Helper function to truncate strings
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
