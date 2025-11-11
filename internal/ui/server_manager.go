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
	environment *inventory.Environment
	cursor      int
	storage     *storage.Storage
	message     string
	messageType string // "success", "error", "info"
}

func NewServerManager(env *inventory.Environment) ServerManager {
	return ServerManager{
		environment: env,
		cursor:      0,
		storage:     storage.NewStorage("."),
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
			// Return to main menu
			return NewMainMenu(), nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.environment.Servers) {
				m.cursor++
			}

		case "a":
			// Add new server
			return NewServerForm(m.environment), nil

		case "e":
			// Edit selected server
			if m.cursor < len(m.environment.Servers) {
				return NewServerForm(m.environment, &m.environment.Servers[m.cursor]), nil
			}

		case "d":
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

		case "s":
			// Save environment
			if err := m.storage.SaveEnvironment(*m.environment); err != nil {
				m.message = fmt.Sprintf("Failed to save: %v", err)
				m.messageType = "error"
			} else {
				m.message = "✓ Environment saved"
				m.messageType = "success"
			}

		case "g":
			// Generate and show YAML
			generator := inventory.NewGenerator()
			summary := generator.GenerateEnvironmentSummary(*m.environment)
			m.message = summary
			m.messageType = "info"
		}
	}

	return m, nil
}

func (m ServerManager) View() string {
	var b strings.Builder

	title := fmt.Sprintf("🖥️  Manage Environment: %s", m.environment.Name)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

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
	helpText := "[a] Add  [e] Edit  [d] Delete  [s] Save  [g] Generate Summary  [Esc] Back"
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
