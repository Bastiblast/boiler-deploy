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
	monoSSHKey  bool
	err         error
	validator   *inventory.Validator
	storage     *storage.Storage
	done        bool
	environment *inventory.Environment
}

func NewEnvironmentForm() EnvironmentForm {
	// Initialize text inputs - name + optional mono IP + optional mono SSH key
	inputs := make([]textinput.Model, 3)
	
	// Environment name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "production"
	inputs[0].Focus()
	inputs[0].Width = 40
	
	// Mono server IP (only shown if mono_server is checked)
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "192.168.1.10"
	inputs[1].Width = 20
	
	// Mono SSH key path (only shown if mono_ssh_key is checked)
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "~/.ssh/id_rsa"
	inputs[2].Width = 40
	
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
		monoSSHKey: false,
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
			// Calculate max index: name + mono server checkbox + (IP if mono) + mono SSH checkbox + (SSH key if mono SSH) + 3 service checkboxes
			maxIndex := 1 + 1 + 1 + 3 // name + mono_server + mono_ssh + services
			if f.monoServer {
				maxIndex++ // Add IP input
			}
			if f.monoSSHKey {
				maxIndex++ // Add SSH key input
			}
			if f.focusIndex > maxIndex {
				f.focusIndex = 0
			}
			return f, f.updateFocus()
			
		case "shift+tab", "up":
			f.focusIndex--
			if f.focusIndex < 0 {
				maxIndex := 1 + 1 + 1 + 3
				if f.monoServer {
					maxIndex++
				}
				if f.monoSSHKey {
					maxIndex++
				}
				f.focusIndex = maxIndex
			}
			return f, f.updateFocus()
			
		case " ":
			// Toggle checkbox
			// Layout: [0]=name, [1]=mono_server, [2]=IP (if mono), [3]=mono_ssh, [4]=SSH key (if mono_ssh), [5+]=services
			nameInputs := 1
			monoServerPos := nameInputs
			ipPos := monoServerPos + 1
			
			monoSSHPos := ipPos
			if f.monoServer {
				monoSSHPos++
			}
			
			sshKeyPos := monoSSHPos + 1
			servicesStartPos := sshKeyPos
			if f.monoSSHKey {
				servicesStartPos++
			}
			
			if f.focusIndex == monoServerPos {
				// Toggle mono server
				f.monoServer = !f.monoServer
			} else if f.focusIndex == monoSSHPos {
				// Toggle mono SSH key
				f.monoSSHKey = !f.monoSSHKey
			} else if f.focusIndex >= servicesStartPos {
				// Service checkboxes
				checkboxIndex := f.focusIndex - servicesStartPos
				keys := []string{"web", "database", "monitoring"}
				if checkboxIndex >= 0 && checkboxIndex < len(keys) {
					key := keys[checkboxIndex]
					f.checkboxes[key] = !f.checkboxes[key]
				}
			}
			
		case "enter":
			// If on an input field, don't submit - let the input handle it
			monoSSHPos := 2
			if f.monoServer {
				monoSSHPos = 3
			}
			
			sshKeyInputPos := monoSSHPos + 1
			
			isOnInput := f.focusIndex == 0 || 
				(f.monoServer && f.focusIndex == 2) ||
				(f.monoSSHKey && f.focusIndex == sshKeyInputPos)
			
			if isOnInput {
				// Let the input field handle enter (move to next field)
				f.focusIndex++
				maxIndex := 1 + 1 + 1 + 3
				if f.monoServer {
					maxIndex++
				}
				if f.monoSSHKey {
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
	} else if f.monoSSHKey {
		// SSH key input (only if mono SSH is enabled)
		monoSSHPos := 2
		if f.monoServer {
			monoSSHPos = 3
		}
		sshKeyInputPos := monoSSHPos + 1
		
		if f.focusIndex == sshKeyInputPos {
			var cmd tea.Cmd
			f.inputs[2], cmd = f.inputs[2].Update(msg)
			return f, cmd
		}
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
	
	// SSH key input (if mono SSH key enabled)
	if f.monoSSHKey {
		monoSSHPos := 2
		if f.monoServer {
			monoSSHPos = 3
		}
		sshKeyInputPos := monoSSHPos + 1
		
		if f.focusIndex == sshKeyInputPos {
			cmds = append(cmds, f.inputs[2].Focus())
		} else {
			f.inputs[2].Blur()
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
	
	monoSSHKey := ""
	if f.monoSSHKey {
		monoSSHKey = f.inputs[2].Value()
		if monoSSHKey == "" {
			return nil, fmt.Errorf("mono SSH key path is required")
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
		MonoSSHKey: f.monoSSHKey,
		MonoSSHKeyPath: monoSSHKey,
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
	
	// Mono SSH key checkbox
	monoSSHPos := 2
	if f.monoServer {
		monoSSHPos = 3
	}
	cursor = "  "
	if f.focusIndex == monoSSHPos {
		cursor = "▶ "
	}
	monoSSHCheck := " "
	if f.monoSSHKey {
		monoSSHCheck = "✓"
	}
	b.WriteString(fmt.Sprintf("%s[%s] Use same SSH key for all servers\n\n", cursor, monoSSHCheck))
	
	// Mono SSH key path input (only if mono SSH key is enabled)
	if f.monoSSHKey {
		sshKeyInputPos := monoSSHPos + 1
		cursor = "  "
		if f.focusIndex == sshKeyInputPos {
			cursor = "▶ "
		}
		b.WriteString(fmt.Sprintf("%sSSH Key Path:\n  %s\n\n", cursor, f.inputs[2].View()))
	}
	
	// Service checkboxes
	b.WriteString("\nServices to enable:\n")
	checkboxLabels := []string{"Web servers", "Database servers", "Monitoring"}
	checkboxKeys := []string{"web", "database", "monitoring"}
	
	// Calculate offset for service checkboxes
	offset := 3 // name + mono server + mono SSH
	if f.monoServer {
		offset++ // + IP input
	}
	if f.monoSSHKey {
		offset++ // + SSH key input
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
