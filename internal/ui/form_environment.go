package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type EnvironmentForm struct {
	inputs      []textinput.Model
	focusIndex  int
	checkboxes  map[string]bool
	monoServer  bool
	err         error
	validator   *inventory.Validator
	storage     *storage.Storage
	done        bool
	environment *inventory.Environment
}

func NewEnvironmentForm() EnvironmentForm {
	// Initialize text inputs - name + optional mono IP
	inputs := make([]textinput.Model, 2)
	
	// Environment name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "production"
	inputs[0].Focus()
	inputs[0].Width = 40
	
	// Mono server IP (only shown if mono_server is checked)
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "192.168.1.10"
	inputs[1].Width = 20
	
	// Get current working directory for storage
	stor := storage.NewStorage(".")
	
	return EnvironmentForm{
		inputs: inputs,
		checkboxes: map[string]bool{
			"web":        true,
			"database":   false,
			"monitoring": false,
		},
		monoServer: false,
		validator:  inventory.NewValidator(),
		storage:    stor,
	}
}

func (f EnvironmentForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f EnvironmentForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// Return to main menu
			return NewMainMenu(), nil
			
		case "tab", "down":
			f.focusIndex++
			// Calculate max index: inputs + mono checkbox + 3 service checkboxes
			maxIndex := len(f.inputs) + 3
			if f.monoServer {
				maxIndex++ // Add 1 for IP input
			}
			if f.focusIndex > maxIndex {
				f.focusIndex = 0
			}
			return f, f.updateFocus()
			
		case "shift+tab", "up":
			f.focusIndex--
			if f.focusIndex < 0 {
				maxIndex := len(f.inputs) + 3
				if f.monoServer {
					maxIndex++
				}
				f.focusIndex = maxIndex
			}
			return f, f.updateFocus()
			
		case " ":
			// Toggle checkbox
			nameInputs := 1 // Just name
			monoCheckboxPos := nameInputs
			
			if f.focusIndex == monoCheckboxPos {
				// Toggle mono server
				f.monoServer = !f.monoServer
			} else if f.focusIndex > monoCheckboxPos {
				// Service checkboxes
				offset := monoCheckboxPos + 1
				if f.monoServer {
					offset++ // Skip IP input
				}
				
				checkboxIndex := f.focusIndex - offset
				keys := []string{"web", "database", "monitoring"}
				if checkboxIndex >= 0 && checkboxIndex < len(keys) {
					key := keys[checkboxIndex]
					f.checkboxes[key] = !f.checkboxes[key]
				}
			}
			
		case "enter":
			// If on an input field, don't submit - let the input handle it
			if f.focusIndex == 0 || (f.monoServer && f.focusIndex == 2) {
				// Let the input field handle enter (move to next field)
				f.focusIndex++
				maxIndex := len(f.inputs) + 3
				if f.monoServer {
					maxIndex++
				}
				if f.focusIndex > maxIndex {
					f.focusIndex = 0
				}
				return f, f.updateFocus()
			}
			
			// Otherwise submit form
			env, err := f.buildEnvironment()
			if err != nil {
				f.err = err
				return f, nil
			}
			
			// Save environment
			if err := f.storage.SaveEnvironment(*env); err != nil {
				f.err = fmt.Errorf("failed to save: %v", err)
				return f, nil
			}
			
			f.done = true
			f.environment = env
			// Return to menu
			return NewMainMenu(), nil
		}
	}
	
	// Update focused input
	if f.focusIndex == 0 {
		// Name input
		var cmd tea.Cmd
		f.inputs[0], cmd = f.inputs[0].Update(msg)
		return f, cmd
	} else if f.monoServer && f.focusIndex == 2 {
		// IP input (only if mono server is enabled)
		var cmd tea.Cmd
		f.inputs[1], cmd = f.inputs[1].Update(msg)
		return f, cmd
	}
	
	return f, nil
}

func (f *EnvironmentForm) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	
	// Name input at focus position 0
	if f.focusIndex == 0 {
		cmds = append(cmds, f.inputs[0].Focus())
	} else {
		f.inputs[0].Blur()
	}
	
	// IP input at focus position 2 (if mono server enabled)
	if f.monoServer {
		if f.focusIndex == 2 {
			cmds = append(cmds, f.inputs[1].Focus())
		} else {
			f.inputs[1].Blur()
		}
	}
	
	return tea.Batch(cmds...)
}

func (f *EnvironmentForm) buildEnvironment() (*inventory.Environment, error) {
	name := f.inputs[0].Value()
	if name == "" {
		name = "production"
	}
	
	// Validate environment name
	if err := f.validator.ValidateEnvironmentName(name); err != nil {
		return nil, err
	}
	
	// Check if exists
	if f.storage.EnvironmentExists(name) {
		return nil, fmt.Errorf("environment '%s' already exists", name)
	}
	
	monoIP := ""
	if f.monoServer {
		monoIP = f.inputs[1].Value()
		if monoIP == "" {
			return nil, fmt.Errorf("mono server IP is required")
		}
		// Validate IP
		if err := f.validator.ValidateIP(monoIP); err != nil {
			return nil, fmt.Errorf("invalid mono server IP: %v", err)
		}
	}
	
	env := &inventory.Environment{
		Name: name,
		Services: inventory.Services{
			Web:        f.checkboxes["web"],
			Database:   f.checkboxes["database"],
			Monitoring: f.checkboxes["monitoring"],
		},
		Config: inventory.Config{
			DeployUser: "root",
			Timezone:   "Europe/Paris",
		},
		Servers:    []inventory.Server{},
		MonoServer: f.monoServer,
		MonoIP:     monoIP,
	}
	
	return env, nil
}

func (f EnvironmentForm) View() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("📝 Create New Environment"))
	b.WriteString("\n\n")
	
	// Environment name input
	cursor := "  "
	if f.focusIndex == 0 {
		cursor = "▶ "
	}
	b.WriteString(fmt.Sprintf("%sEnvironment name:\n  %s\n\n", cursor, f.inputs[0].View()))
	
	// Mono server checkbox
	monoCheckboxPos := 1
	cursor = "  "
	if f.focusIndex == monoCheckboxPos {
		cursor = "▶ "
	}
	monoCheck := " "
	if f.monoServer {
		monoCheck = "✓"
	}
	b.WriteString(fmt.Sprintf("%s[%s] Mono server (all services on same IP)\n\n", cursor, monoCheck))
	
	// Mono IP input (only if mono server is enabled)
	if f.monoServer {
		cursor = "  "
		if f.focusIndex == 2 {
			cursor = "▶ "
		}
		b.WriteString(fmt.Sprintf("%sServer IP:\n  %s\n\n", cursor, f.inputs[1].View()))
	}
	
	// Service checkboxes
	b.WriteString("\nServices to enable:\n")
	checkboxLabels := []string{"Web servers", "Database servers", "Monitoring"}
	checkboxKeys := []string{"web", "database", "monitoring"}
	
	// Calculate offset for service checkboxes
	offset := 2 // name + mono checkbox
	if f.monoServer {
		offset++ // + IP input
	}
	
	for i, label := range checkboxLabels {
		cursor := "  "
		checkbox := "[ ]"
		if f.checkboxes[checkboxKeys[i]] {
			checkbox = "[✓]"
		}
		if f.focusIndex == offset+i {
			cursor = "▶ "
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, label))
	}
	
	b.WriteString("\n")
	
	// Error message
	if f.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", f.err)))
		b.WriteString("\n\n")
	}
	
	// Success message
	if f.done {
		b.WriteString(successStyle.Render("✅ Environment created successfully!"))
		b.WriteString("\n\n")
	}
	
	// Help
	b.WriteString(helpStyle.Render("[Tab/↑↓] Navigate  [Space] Toggle  [Enter] Create  [Esc] Cancel"))
	
	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}
