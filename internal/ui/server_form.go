package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

type ServerForm struct {
	environment  *inventory.Environment
	editingServer *inventory.Server
	inputs       []textinput.Model
	focusIndex   int
	typeIndex    int // 0=web, 1=db, 2=monitoring
	err          error
	validator    *inventory.Validator
	storage      *storage.Storage
}

// NewServerForm creates a form for adding or editing a server
func NewServerForm(env *inventory.Environment, editServer ...*inventory.Server) ServerForm {
	inputs := make([]textinput.Model, 10)

	// Server name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = fmt.Sprintf("%s-web-01", env.Name)
	inputs[0].Width = 40
	inputs[0].Focus()

	// IP address - use mono IP if available
	inputs[1] = textinput.New()
	if env.MonoServer && env.MonoIP != "" {
		inputs[1].Placeholder = env.MonoIP
		inputs[1].SetValue(env.MonoIP)
	} else {
		inputs[1].Placeholder = "192.168.1.10"
	}
	inputs[1].Width = 20

	// SSH Port
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "22"
	inputs[2].Width = 10

	// SSH User
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "root"
	inputs[3].Width = 20

	// SSH Key Path
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "~/.ssh/id_rsa"
	inputs[4].Width = 50

	// App-specific fields (for web servers)
	// App Port
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "3000"
	inputs[5].Width = 10

	// Git Repository
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "https://github.com/user/repo.git"
	inputs[6].Width = 50

	// Git Branch
	inputs[7] = textinput.New()
	inputs[7].Placeholder = "main"
	inputs[7].Width = 30

	// Node.js Version
	inputs[8] = textinput.New()
	inputs[8].Placeholder = "20"
	inputs[8].Width = 10

	form := ServerForm{
		environment: env,
		inputs:      inputs,
		validator:   inventory.NewValidator(),
		storage:     storage.NewStorage("."),
		typeIndex:   0,
	}

	// If editing, populate fields
	if len(editServer) > 0 && editServer[0] != nil {
		form.editingServer = editServer[0]
		form.inputs[0].SetValue(editServer[0].Name)
		form.inputs[1].SetValue(editServer[0].IP)
		form.inputs[2].SetValue(strconv.Itoa(editServer[0].Port))
		form.inputs[3].SetValue(editServer[0].SSHUser)
		form.inputs[4].SetValue(editServer[0].SSHKeyPath)
		
		// App config
		if editServer[0].AppPort > 0 {
			form.inputs[5].SetValue(strconv.Itoa(editServer[0].AppPort))
		}
		form.inputs[6].SetValue(editServer[0].GitRepo)
		form.inputs[7].SetValue(editServer[0].GitBranch)
		form.inputs[8].SetValue(editServer[0].NodeVersion)
		
		// Set type index
		switch editServer[0].Type {
		case "web":
			form.typeIndex = 0
		case "db":
			form.typeIndex = 1
		case "monitoring":
			form.typeIndex = 2
		}
	}

	return form
}

func (f ServerForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f ServerForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// Return to server manager
			return NewServerManager(f.environment), nil

		case "tab", "down":
			f.focusIndex++
			// Calculate max index based on server type
			maxIndex := 5 // Common fields + type selector
			serverType := []string{"web", "db", "monitoring"}[f.typeIndex]
			if serverType == "web" {
				maxIndex = 9 // Common + app fields + type selector
			}
			if f.focusIndex > maxIndex {
				f.focusIndex = 0
			}
			return f, f.updateFocus()

		case "shift+tab", "up":
			f.focusIndex--
			if f.focusIndex < 0 {
				// Calculate max index based on server type
				maxIndex := 5
				serverType := []string{"web", "db", "monitoring"}[f.typeIndex]
				if serverType == "web" {
					maxIndex = 9
				}
				f.focusIndex = maxIndex
			}
			return f, f.updateFocus()

		case "left":
			// Type selector position depends on whether we're showing app fields
			serverType := []string{"web", "db", "monitoring"}[f.typeIndex]
			typeSelectorPos := 5
			if serverType == "web" {
				typeSelectorPos = 9
			}
			if f.focusIndex == typeSelectorPos && f.typeIndex > 0 {
				f.typeIndex--
				// Reset focus to account for different field counts
				f.focusIndex = 0
				return f, f.updateFocus()
			}

		case "right":
			serverType := []string{"web", "db", "monitoring"}[f.typeIndex]
			typeSelectorPos := 5
			if serverType == "web" {
				typeSelectorPos = 9
			}
			if f.focusIndex == typeSelectorPos && f.typeIndex < 2 {
				f.typeIndex++
				// Reset focus to account for different field counts
				f.focusIndex = 0
				return f, f.updateFocus()
			}

		case "enter":
			// Submit form
			server, err := f.buildServer()
			if err != nil {
				f.err = err
				return f, nil
			}

			// Add or update server
			if f.editingServer != nil {
				// Update existing server
				for i, s := range f.environment.Servers {
					if s.Name == f.editingServer.Name {
						f.environment.Servers[i] = *server
						break
					}
				}
			} else {
				// Add new server
				f.environment.Servers = append(f.environment.Servers, *server)
			}

			// Save environment
			if err := f.storage.SaveEnvironment(*f.environment); err != nil {
				f.err = fmt.Errorf("failed to save: %v", err)
				return f, nil
			}

			// Return to server manager
			return NewServerManager(f.environment), nil
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

func (f *ServerForm) updateFocus() tea.Cmd {
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

func (f *ServerForm) buildServer() (*inventory.Server, error) {
	// Get values with defaults
	name := f.inputs[0].Value()
	if name == "" {
		typeStr := []string{"web", "db", "monitoring"}[f.typeIndex]
		serverCount := 0
		for _, s := range f.environment.Servers {
			if s.Type == typeStr {
				serverCount++
			}
		}
		name = fmt.Sprintf("%s-%s-%02d", f.environment.Name, typeStr, serverCount+1)
	}

	ip := f.inputs[1].Value()
	if ip == "" {
		return nil, fmt.Errorf("IP address is required")
	}

	portStr := f.inputs[2].Value()
	if portStr == "" {
		portStr = "22"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SSH port: %v", err)
	}

	sshUser := f.inputs[3].Value()
	if sshUser == "" {
		sshUser = "root"
	}

	sshKeyPath := f.inputs[4].Value()
	if sshKeyPath == "" {
		sshKeyPath = "~/.ssh/id_rsa"
	}

	serverType := []string{"web", "db", "monitoring"}[f.typeIndex]

	server := &inventory.Server{
		Name:          name,
		IP:            ip,
		Port:          port,
		SSHUser:       sshUser,
		SSHKeyPath:    sshKeyPath,
		Type:          serverType,
		AnsibleBecome: true,
	}

	// For web servers, get app-specific configuration
	if serverType == "web" {
		appPortStr := f.inputs[5].Value()
		if appPortStr == "" {
			return nil, fmt.Errorf("application port is required for web servers")
		}
		appPort, err := strconv.Atoi(appPortStr)
		if err != nil {
			return nil, fmt.Errorf("invalid app port: %v", err)
		}
		server.AppPort = appPort

		gitRepo := f.inputs[6].Value()
		if gitRepo == "" {
			return nil, fmt.Errorf("git repository is required for web servers")
		}
		server.GitRepo = gitRepo

		gitBranch := f.inputs[7].Value()
		if gitBranch == "" {
			gitBranch = "main"
		}
		server.GitBranch = gitBranch

		nodeVersion := f.inputs[8].Value()
		if nodeVersion == "" {
			nodeVersion = "20"
		}
		server.NodeVersion = nodeVersion
	}

	// Validate server
	errors := f.validator.ValidateServer(*server)
	if len(errors) > 0 {
		return nil, errors[0]
	}

	// Check for conflicts (skip if editing same server)
	excludeName := ""
	if f.editingServer != nil {
		excludeName = f.editingServer.Name
	}
	if err := f.validator.CheckIPPortConflict(f.environment.Servers, ip, port, excludeName); err != nil {
		return nil, err
	}

	return server, nil
}

func (f ServerForm) View() string {
	var b strings.Builder

	if f.editingServer != nil {
		b.WriteString(titleStyle.Render("✏️  Edit Server"))
	} else {
		b.WriteString(titleStyle.Render("➕ Add New Server"))
	}
	b.WriteString("\n\n")

	// Common fields (all server types)
	commonLabels := []string{
		"Server name:",
		"IP address:",
		"SSH port:",
		"SSH user:",
		"SSH key path:",
	}

	for i := 0; i < 5; i++ {
		cursor := "  "
		if f.focusIndex == i {
			cursor = "▶ "
		}
		b.WriteString(fmt.Sprintf("%s%s\n  %s\n\n", cursor, commonLabels[i], f.inputs[i].View()))
	}

	// App-specific fields (only for web servers)
	serverType := []string{"web", "db", "monitoring"}[f.typeIndex]
	if serverType == "web" {
		b.WriteString("─── Application Configuration ───\n\n")
		
		appLabels := []string{
			"Application port:",
			"Git repository:",
			"Git branch:",
			"Node.js version:",
		}

		for i := 0; i < 4; i++ {
			cursor := "  "
			if f.focusIndex == i+5 {
				cursor = "▶ "
			}
			b.WriteString(fmt.Sprintf("%s%s\n  %s\n\n", cursor, appLabels[i], f.inputs[i+5].View()))
		}
	}

	// Server type selector
	typeSelectorPos := 5
	if serverType == "web" {
		typeSelectorPos = 9
	}
	
	typeCursor := "  "
	if f.focusIndex == typeSelectorPos {
		typeCursor = "▶ "
	}
	b.WriteString(typeCursor + "Server type:\n  ")

	types := []string{"Web", "Database", "Monitoring"}
	for i, t := range types {
		if i == f.typeIndex {
			b.WriteString(selectedItemStyle.Render("[" + t + "]"))
		} else {
			b.WriteString(normalItemStyle.Render(" " + t + " "))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n\n")

	// Error message
	if f.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", f.err)))
		b.WriteString("\n\n")
	}

	// Help
	helpText := "[Tab/↑↓] Navigate  [←→] Change type  [Enter] Save  [Esc] Cancel"
	b.WriteString(helpStyle.Render(helpText))

	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}
