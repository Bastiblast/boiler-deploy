package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/ssh"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type ServerManager struct {
	environment *inventory.Environment
	cursor      int
	storage     *storage.Storage
	message     string
	messageType string // "success", "error", "info"
	testing     bool   // Is SSH test in progress
	testingAll  bool   // Is testing all servers
}

// SSH test result message
type sshTestResultMsg struct {
	index  int
	result ssh.TestResult
}

// Batch SSH test results
type sshTestAllResultsMsg struct {
	results []sshTestResultMsg
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
	case sshTestAllResultsMsg:
		// Handle batch SSH test results
		m.testing = false
		m.testingAll = false
		
		passed := 0
		failed := 0
		
		for _, result := range msg.results {
			if result.index >= 0 && result.index < len(m.environment.Servers) {
				m.environment.Servers[result.index].SSHTested = true
				if result.result.Success {
					m.environment.Servers[result.index].SSHStatus = "✓ " + result.result.Message
					passed++
				} else {
					m.environment.Servers[result.index].SSHStatus = "✗ " + result.result.Message
					failed++
				}
			}
		}
		
		if failed == 0 {
			m.message = fmt.Sprintf("✓ All servers tested successfully (%d/%d)", passed, passed+failed)
			m.messageType = "success"
		} else {
			m.message = fmt.Sprintf("⚠ Tests completed: %d passed, %d failed", passed, failed)
			m.messageType = "error"
		}
		
		return m, nil
	
	case sshTestResultMsg:
		// Handle single SSH test result
		m.testing = false
		if msg.index >= 0 && msg.index < len(m.environment.Servers) {
			m.environment.Servers[msg.index].SSHTested = true
			if msg.result.Success {
				m.environment.Servers[msg.index].SSHStatus = "✓ " + msg.result.Message
				m.message = fmt.Sprintf("✓ SSH test passed for '%s'", m.environment.Servers[msg.index].Name)
				m.messageType = "success"
			} else {
				m.environment.Servers[msg.index].SSHStatus = "✗ " + msg.result.Message
				m.message = fmt.Sprintf("✗ SSH test failed for '%s': %s", 
					m.environment.Servers[msg.index].Name, msg.result.Message)
				m.messageType = "error"
			}
		}
		return m, nil
	
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// Return to environment selector
			return NewEnvironmentSelector(), nil

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
		
		case "t":
			// Test SSH connection for selected server
			if m.cursor < len(m.environment.Servers) && !m.testing {
				m.testing = true
				m.message = "Testing SSH connection..."
				m.messageType = "info"
				
				server := m.environment.Servers[m.cursor]
				idx := m.cursor
				
				// Run SSH test asynchronously
				return m, func() tea.Msg {
					result := ssh.TestConnection(server.IP, server.Port, server.SSHUser, server.SSHKeyPath)
					return sshTestResultMsg{index: idx, result: result}
				}
			}
		
		case "T":
			// Test all servers
			if !m.testing && len(m.environment.Servers) > 0 {
				m.testing = true
				m.testingAll = true
				m.message = "Testing all SSH connections..."
				m.messageType = "info"
				
				// Capture environment servers for the closure
				servers := m.environment.Servers
				
				// Test all servers sequentially
				return m, func() tea.Msg {
					var results []sshTestResultMsg
					for i, server := range servers {
						result := ssh.TestConnection(server.IP, server.Port, server.SSHUser, server.SSHKeyPath)
						results = append(results, sshTestResultMsg{index: i, result: result})
					}
					return sshTestAllResultsMsg{results: results}
				}
			}
		}
	}

	return m, nil
}

func (m ServerManager) View() string {
	var b strings.Builder

	title := fmt.Sprintf("🖥️  Manage Environment: %s", m.environment.Name)
	if m.testing {
		if m.testingAll {
			title += " [Testing all servers...]"
		} else {
			title += " [Testing SSH...]"
		}
	}
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
			if server.SSHTested {
				status = server.SSHStatus
			}

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
	helpLine1 := "[a] Add  [e] Edit  [d] Delete  [t] Test SSH  [T] Test All"
	helpLine2 := "[s] Save  [g] Generate  [Esc] Back"
	b.WriteString(helpStyle.Render(helpLine1))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(helpLine2))

	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}

// Helper function to truncate strings
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
