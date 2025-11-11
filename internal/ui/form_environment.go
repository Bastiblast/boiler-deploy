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
	err         error
	validator   *inventory.Validator
	storage     *storage.Storage
	done        bool
	environment *inventory.Environment
}

func NewEnvironmentForm() EnvironmentForm {
	// Initialize text inputs
	inputs := make([]textinput.Model, 5)
	
	// Environment name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "production"
	inputs[0].Focus()
	inputs[0].Width = 40
	
	// Git repository
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "https://github.com/user/repo.git"
	inputs[1].Width = 50
	
	// Git branch
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "main"
	inputs[2].Width = 30
	
	// Node.js version
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "20"
	inputs[3].Width = 10
	
	// App port
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "3000"
	inputs[4].Width = 10
	
	// Get current working directory for storage
	stor := storage.NewStorage(".")
	
	return EnvironmentForm{
		inputs: inputs,
		checkboxes: map[string]bool{
			"web":        true,
			"database":   false,
			"monitoring": false,
		},
		validator: inventory.NewValidator(),
		storage:   stor,
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
			if f.focusIndex > len(f.inputs)+2 { // inputs + 3 checkboxes
				f.focusIndex = 0
			}
			return f, f.updateFocus()
			
		case "shift+tab", "up":
			f.focusIndex--
			if f.focusIndex < 0 {
				f.focusIndex = len(f.inputs) + 2
			}
			return f, f.updateFocus()
			
		case " ":
			// Toggle checkbox
			if f.focusIndex >= len(f.inputs) {
				checkboxIndex := f.focusIndex - len(f.inputs)
				keys := []string{"web", "database", "monitoring"}
				if checkboxIndex < len(keys) {
					key := keys[checkboxIndex]
					f.checkboxes[key] = !f.checkboxes[key]
				}
			}
			
		case "enter":
			// Submit form
			if f.focusIndex == len(f.inputs)+2 { // On last element
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
	}
	
	// Update focused input
	if f.focusIndex < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)
		return f, cmd
	}
	
	return f, nil
}

func (f *EnvironmentForm) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(f.inputs))
	
	for i := 0; i < len(f.inputs); i++ {
		if i == f.focusIndex {
			cmds[i] = f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
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
	
	repo := f.inputs[1].Value()
	if repo == "" {
		repo = "https://github.com/user/repo.git"
	}
	
	branch := f.inputs[2].Value()
	if branch == "" {
		branch = "main"
	}
	
	nodeVersion := f.inputs[3].Value()
	if nodeVersion == "" {
		nodeVersion = "20"
	}
	
	port := f.inputs[4].Value()
	if port == "" {
		port = "3000"
	}
	
	env := &inventory.Environment{
		Name: name,
		Services: inventory.Services{
			Web:        f.checkboxes["web"],
			Database:   f.checkboxes["database"],
			Monitoring: f.checkboxes["monitoring"],
		},
		Config: inventory.Config{
			AppName:       name + "-app",
			AppRepo:       repo,
			AppBranch:     branch,
			NodeJSVersion: nodeVersion,
			AppPort:       port,
			DeployUser:    "root",
			Timezone:      "Europe/Paris",
		},
		Servers: []inventory.Server{},
	}
	
	return env, nil
}

func (f EnvironmentForm) View() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("📝 Create New Environment"))
	b.WriteString("\n\n")
	
	// Text inputs
	labels := []string{
		"Environment name:",
		"Git repository:",
		"Git branch:",
		"Node.js version:",
		"Application port:",
	}
	
	for i, input := range f.inputs {
		cursor := "  "
		if f.focusIndex == i {
			cursor = "▶ "
		}
		b.WriteString(fmt.Sprintf("%s%s\n  %s\n\n", cursor, labels[i], input.View()))
	}
	
	// Checkboxes
	b.WriteString("\nServices to enable:\n")
	checkboxLabels := []string{"Web servers", "Database servers", "Monitoring"}
	checkboxKeys := []string{"web", "database", "monitoring"}
	
	for i, label := range checkboxLabels {
		cursor := "  "
		checkbox := "[ ]"
		if f.checkboxes[checkboxKeys[i]] {
			checkbox = "[✓]"
		}
		if f.focusIndex == len(f.inputs)+i {
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
