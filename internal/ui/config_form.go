package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigForm manages configuration options
type ConfigForm struct {
	envName      string
	config       *config.ConfigOptions
	configMgr    *config.Manager
	inputs       []textinput.Model
	focused      int
	tagSelection map[string][]bool // "provisioning" or "deployment" -> selected tags
	currentStep  int                // 0: general, 1: provisioning tags, 2: deployment tags
	err          error
	saved        bool
}

// NewConfigForm creates a new configuration form
func NewConfigForm(envName string, configMgr *config.Manager) (*ConfigForm, error) {
	cfg, err := configMgr.Load(envName)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Initialize text inputs
	inputs := make([]textinput.Model, 5)
	
	// Refresh interval
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "1"
	inputs[0].SetValue(fmt.Sprintf("%d", int(cfg.RefreshInterval.Seconds())))
	inputs[0].Focus()
	
	// Log retention
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "100"
	inputs[1].SetValue(fmt.Sprintf("%d", cfg.LogRetention))
	
	// Health check timeout
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "30"
	inputs[2].SetValue(fmt.Sprintf("%d", int(cfg.HealthCheckTimeout.Seconds())))
	
	// Health check retries
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "3"
	inputs[3].SetValue(fmt.Sprintf("%d", cfg.HealthCheckRetries))
	
	// Max retries
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "3"
	inputs[4].SetValue(fmt.Sprintf("%d", cfg.MaxRetries))

	// Initialize tag selections
	tagSelection := make(map[string][]bool)
	
	// Provisioning tags
	provTags := make([]bool, len(config.ProvisioningTags))
	for i, tag := range config.ProvisioningTags {
		provTags[i] = contains(cfg.ProvisioningTags, tag)
	}
	tagSelection["provisioning"] = provTags
	
	// Deployment tags
	deplTags := make([]bool, len(config.DeploymentTags))
	for i, tag := range config.DeploymentTags {
		deplTags[i] = contains(cfg.DeploymentTags, tag)
	}
	tagSelection["deployment"] = deplTags

	return &ConfigForm{
		envName:      envName,
		config:       cfg,
		configMgr:    configMgr,
		inputs:       inputs,
		focused:      0,
		tagSelection: tagSelection,
		currentStep:  0,
	}, nil
}

func (f *ConfigForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f *ConfigForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return f, tea.Quit
			
		case "esc":
			if f.currentStep > 0 {
				f.currentStep--
				return f, nil
			}
			return NewMainMenu(), nil
			
		case "tab", "down":
			if f.currentStep == 0 {
				f.focused = (f.focused + 1) % (len(f.inputs) + 5) // +5 for toggles
				f.updateFocus()
			} else if f.currentStep == 1 {
				// Cycle through provisioning tags
				for i := range f.tagSelection["provisioning"] {
					if f.tagSelection["provisioning"][i] {
						f.tagSelection["provisioning"][i] = false
						next := (i + 1) % len(f.tagSelection["provisioning"])
						f.tagSelection["provisioning"][next] = true
						break
					}
				}
			} else if f.currentStep == 2 {
				// Cycle through deployment tags
				for i := range f.tagSelection["deployment"] {
					if f.tagSelection["deployment"][i] {
						f.tagSelection["deployment"][i] = false
						next := (i + 1) % len(f.tagSelection["deployment"])
						f.tagSelection["deployment"][next] = true
						break
					}
				}
			}
			return f, nil
			
		case "shift+tab", "up":
			if f.currentStep == 0 {
				f.focused = (f.focused - 1 + len(f.inputs) + 5) % (len(f.inputs) + 5)
				f.updateFocus()
			}
			return f, nil
			
		case "enter":
			if f.currentStep == 0 && f.focused == len(f.inputs)+4 {
				// Save button pressed
				if err := f.saveConfig(); err != nil {
					f.err = err
					return f, nil
				}
				f.saved = true
				return f, tea.Quit
			} else if f.currentStep == 0 && (f.focused == len(f.inputs) || f.focused == len(f.inputs)+1) {
				// Strategy selection
				if f.focused == len(f.inputs) {
					// Toggle provisioning strategy
					if f.config.ProvisioningStrategy == "sequential" {
						f.config.ProvisioningStrategy = "parallel"
					} else {
						f.config.ProvisioningStrategy = "sequential"
					}
				} else {
					// Toggle deployment strategy
					switch f.config.DeploymentStrategy {
					case "rolling":
						f.config.DeploymentStrategy = "all_at_once"
					case "all_at_once":
						f.config.DeploymentStrategy = "blue_green"
					case "blue_green":
						f.config.DeploymentStrategy = "rolling"
					}
				}
			} else if f.currentStep == 0 && f.focused == len(f.inputs)+2 {
				// Health check toggle
				f.config.HealthCheckEnabled = !f.config.HealthCheckEnabled
			} else if f.currentStep == 0 && f.focused == len(f.inputs)+3 {
				// Auto retry toggle
				f.config.AutoRetryEnabled = !f.config.AutoRetryEnabled
			} else if f.currentStep == 0 {
				// Move to tag selection
				f.currentStep = 1
			}
			return f, nil
			
		case " ":
			// Space toggles current tag selection
			if f.currentStep == 1 {
				for i := range f.tagSelection["provisioning"] {
					if f.tagSelection["provisioning"][i] {
						// Toggle this tag
						f.tagSelection["provisioning"][i] = !f.tagSelection["provisioning"][i]
						break
					}
				}
			} else if f.currentStep == 2 {
				for i := range f.tagSelection["deployment"] {
					if f.tagSelection["deployment"][i] {
						f.tagSelection["deployment"][i] = !f.tagSelection["deployment"][i]
						break
					}
				}
			}
			return f, nil
			
		case "n":
			// Next step
			if f.currentStep < 2 {
				f.currentStep++
			}
			return f, nil
		}
	}

	// Update current input
	if f.currentStep == 0 && f.focused < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[f.focused], cmd = f.inputs[f.focused].Update(msg)
		return f, cmd
	}

	return f, nil
}

func (f *ConfigForm) View() string {
	if f.saved {
		return successStyle.Render("✓ Configuration saved successfully!") + "\n\nPress any key to return to menu..."
	}

	var b strings.Builder
	
	b.WriteString(titleStyle.Render(fmt.Sprintf("⚙️  Configuration - %s", f.envName)) + "\n\n")
	
	if f.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", f.err)) + "\n\n")
	}
	
	switch f.currentStep {
	case 0:
		f.renderGeneralSettings(&b)
	case 1:
		f.renderProvisioningTags(&b)
	case 2:
		f.renderDeploymentTags(&b)
	}
	
	b.WriteString("\n" + helpStyle.Render("tab/↑↓: navigate • enter: select/next • n: next step • esc: back • q: quit"))
	
	return b.String()
}

func (f *ConfigForm) renderGeneralSettings(b *strings.Builder) {
	b.WriteString(infoStyle.Render("General Settings") + "\n\n")
	
	// Text inputs
	labels := []string{
		"Refresh Interval (seconds):",
		"Log Retention (lines):",
		"Health Check Timeout (seconds):",
		"Health Check Retries:",
		"Max Retries:",
	}
	
	for i, label := range labels {
		cursor := " "
		if f.focused == i {
			cursor = "▶"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, label))
		b.WriteString(fmt.Sprintf("  %s\n\n", f.inputs[i].View()))
	}
	
	// Strategy selections
	cursor := " "
	if f.focused == len(f.inputs) {
		cursor = "▶"
	}
	b.WriteString(fmt.Sprintf("%s Provisioning Strategy: ", cursor))
	if f.config.ProvisioningStrategy == "sequential" {
		b.WriteString(activeStyle.Render("[Sequential]") + " " + inactiveStyle.Render("Parallel"))
	} else {
		b.WriteString(inactiveStyle.Render("Sequential") + " " + activeStyle.Render("[Parallel]"))
	}
	b.WriteString("\n\n")
	
	cursor = " "
	if f.focused == len(f.inputs)+1 {
		cursor = "▶"
	}
	b.WriteString(fmt.Sprintf("%s Deployment Strategy: %s\n\n", cursor, activeStyle.Render(f.config.DeploymentStrategy)))
	
	// Toggles
	cursor = " "
	if f.focused == len(f.inputs)+2 {
		cursor = "▶"
	}
	check := "☐"
	if f.config.HealthCheckEnabled {
		check = "☑"
	}
	b.WriteString(fmt.Sprintf("%s %s Health Check Enabled\n\n", cursor, check))
	
	cursor = " "
	if f.focused == len(f.inputs)+3 {
		cursor = "▶"
	}
	check = "☐"
	if f.config.AutoRetryEnabled {
		check = "☑"
	}
	b.WriteString(fmt.Sprintf("%s %s Auto Retry Enabled\n\n", cursor, check))
	
	// Save button
	cursor = " "
	if f.focused == len(f.inputs)+4 {
		cursor = "▶"
	}
	b.WriteString(fmt.Sprintf("\n%s %s\n", cursor, activeStyle.Render("[Save Configuration]")))
}

func (f *ConfigForm) renderProvisioningTags(b *strings.Builder) {
	b.WriteString(infoStyle.Render("Select Provisioning Tags") + "\n\n")
	b.WriteString(helpStyle.Render("Use space to toggle, tab to navigate, n for next step") + "\n\n")
	
	for i, tag := range config.ProvisioningTags {
		check := "☐"
		if f.tagSelection["provisioning"][i] {
			check = "☑"
		}
		
		style := inactiveStyle
		if f.tagSelection["provisioning"][i] {
			style = activeStyle
		}
		
		b.WriteString(fmt.Sprintf("  %s %s\n", check, style.Render(tag)))
	}
}

func (f *ConfigForm) renderDeploymentTags(b *strings.Builder) {
	b.WriteString(infoStyle.Render("Select Deployment Tags") + "\n\n")
	b.WriteString(helpStyle.Render("Use space to toggle, tab to navigate, enter to save") + "\n\n")
	
	for i, tag := range config.DeploymentTags {
		check := "☐"
		if f.tagSelection["deployment"][i] {
			check = "☑"
		}
		
		style := inactiveStyle
		if f.tagSelection["deployment"][i] {
			style = activeStyle
		}
		
		b.WriteString(fmt.Sprintf("  %s %s\n", check, style.Render(tag)))
	}
	
	b.WriteString("\n" + activeStyle.Render("[Save Configuration]"))
}

func (f *ConfigForm) updateFocus() {
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
	if f.focused < len(f.inputs) {
		f.inputs[f.focused].Focus()
	}
}

func (f *ConfigForm) saveConfig() error {
	// Parse text inputs
	refreshSec, err := strconv.Atoi(f.inputs[0].Value())
	if err != nil || refreshSec < 1 {
		return fmt.Errorf("invalid refresh interval")
	}
	f.config.RefreshInterval = time.Duration(refreshSec) * time.Second
	
	logRetention, err := strconv.Atoi(f.inputs[1].Value())
	if err != nil || logRetention < 10 {
		return fmt.Errorf("invalid log retention (minimum 10)")
	}
	f.config.LogRetention = logRetention
	
	healthTimeout, err := strconv.Atoi(f.inputs[2].Value())
	if err != nil || healthTimeout < 5 {
		return fmt.Errorf("invalid health check timeout (minimum 5)")
	}
	f.config.HealthCheckTimeout = time.Duration(healthTimeout) * time.Second
	
	healthRetries, err := strconv.Atoi(f.inputs[3].Value())
	if err != nil || healthRetries < 0 {
		return fmt.Errorf("invalid health check retries")
	}
	f.config.HealthCheckRetries = healthRetries
	
	maxRetries, err := strconv.Atoi(f.inputs[4].Value())
	if err != nil || maxRetries < 0 {
		return fmt.Errorf("invalid max retries")
	}
	f.config.MaxRetries = maxRetries
	
	// Build selected tags
	f.config.ProvisioningTags = []string{}
	for i, selected := range f.tagSelection["provisioning"] {
		if selected {
			f.config.ProvisioningTags = append(f.config.ProvisioningTags, config.ProvisioningTags[i])
		}
	}
	if len(f.config.ProvisioningTags) == 0 {
		f.config.ProvisioningTags = []string{"all"}
	}
	
	f.config.DeploymentTags = []string{}
	for i, selected := range f.tagSelection["deployment"] {
		if selected {
			f.config.DeploymentTags = append(f.config.DeploymentTags, config.DeploymentTags[i])
		}
	}
	if len(f.config.DeploymentTags) == 0 {
		f.config.DeploymentTags = []string{"all"}
	}
	
	// Save to disk
	return f.configMgr.Save(f.envName, f.config)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var (
	activeStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10"))
	
	inactiveStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)
