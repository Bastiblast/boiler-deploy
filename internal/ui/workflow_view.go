package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/ansible"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/logging"
	"github.com/bastiblast/boiler-deploy/internal/status"
	"github.com/bastiblast/boiler-deploy/internal/storage"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkflowView struct {
	environment     string
	servers         []*inventory.Server
	statuses        map[string]*status.ServerStatus
	selectedServers map[string]bool
	cursor          int
	orchestrator    *ansible.Orchestrator
	statusMgr       *status.Manager
	logReader       *logging.Reader
	showLogs        bool
	currentLogFile  string
	logLines        []string
	progress        map[string]string
	lastRefresh     time.Time
	autoRefresh     bool
	realtimeLogs    []string // Live streaming output from ansible
	maxRealtimeLogs int
	logsViewport    viewport.Model
	logsReady       bool
}

type tickMsg time.Time
type statusUpdateMsg struct{}
type validationCompleteMsg struct{}

func NewWorkflowView() (*WorkflowView, error) {
	stor := storage.NewStorage(".")
	envs, err := stor.ListEnvironments()
	if err != nil || len(envs) == 0 {
		return nil, fmt.Errorf("no environments found")
	}

	return NewWorkflowViewWithEnv(envs[0])
}

func NewWorkflowViewWithEnv(envName string) (*WorkflowView, error) {
	wv := &WorkflowView{
		environment:     envName,
		selectedServers: make(map[string]bool),
		progress:        make(map[string]string),
		autoRefresh:     true,
		realtimeLogs:    make([]string, 0),
		maxRealtimeLogs: 100, // Keep more logs for scrolling
		logsViewport:    viewport.New(120, 10), // Initial size, will be updated
		logsReady:       true, // Set to true immediately
	}
	
	// Initialize viewport content immediately
	wv.updateLogsViewport()

	if err := wv.loadEnvironment(); err != nil {
		return nil, err
	}

	return wv, nil
}

func (wv *WorkflowView) loadEnvironment() error {
	stor := storage.NewStorage(".")
	env, err := stor.LoadEnvironment(wv.environment)
	if err != nil {
		return err
	}

	wv.servers = make([]*inventory.Server, len(env.Servers))
	for i := range env.Servers {
		wv.servers[i] = &env.Servers[i]
	}

	statusMgr, err := status.NewManager(wv.environment)
	if err != nil {
		return err
	}
	wv.statusMgr = statusMgr

	orchestrator, err := ansible.NewOrchestrator(wv.environment, statusMgr)
	if err != nil {
		return err
	}
	wv.orchestrator = orchestrator
	wv.orchestrator.SetProgressCallback(wv.onProgress)

	wv.logReader = logging.NewReader(wv.environment)

	wv.refreshStatuses()
	
	// Auto-validate all servers on startup
	go wv.validateAllServers()

	return nil
}

func (wv *WorkflowView) validateAllServers() {
	for _, server := range wv.servers {
		checks := wv.statusMgr.ValidateServer(server)
		wv.statusMgr.UpdateReadyChecks(server.Name, checks)
	}
}

func (wv *WorkflowView) onProgress(serverName, message string) {
	wv.progress[serverName] = message
	
	// Add to realtime logs
	logLine := fmt.Sprintf("[%s] %s", serverName, message)
	wv.realtimeLogs = append(wv.realtimeLogs, logLine)
	
	// Keep only last N lines
	if len(wv.realtimeLogs) > wv.maxRealtimeLogs {
		wv.realtimeLogs = wv.realtimeLogs[len(wv.realtimeLogs)-wv.maxRealtimeLogs:]
	}
	
	// Update viewport content
	wv.updateLogsViewport()
}

func (wv *WorkflowView) updateLogsViewport() {
	if !wv.logsReady {
		return
	}
	
	var b strings.Builder
	
	if len(wv.realtimeLogs) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		b.WriteString(dimStyle.Render("Waiting for actions...") + "\n")
	} else {
		for _, line := range wv.realtimeLogs {
			// Truncate very long lines
			displayLine := line
			maxWidth := wv.logsViewport.Width - 4
			if maxWidth > 0 && len(displayLine) > maxWidth {
				displayLine = displayLine[:maxWidth-3] + "..."
			}
			
			// Apply different styles based on content
			styledLine := wv.styleLogLine(displayLine)
			b.WriteString(styledLine + "\n")
		}
	}
	
	wv.logsViewport.SetContent(b.String())
	// Auto-scroll to bottom on new content
	wv.logsViewport.GotoBottom()
}

func (wv *WorkflowView) refreshStatuses() {
	wv.statuses = wv.statusMgr.GetAllStatuses()
	wv.lastRefresh = time.Now()
}

func (wv *WorkflowView) Init() tea.Cmd {
	wv.orchestrator.Start(wv.servers)
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (wv *WorkflowView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update viewport size when terminal is resized
		if !wv.logsReady {
			wv.logsViewport = viewport.New(msg.Width, 10)
			wv.logsReady = true
		} else {
			wv.logsViewport.Width = msg.Width
		}
		wv.updateLogsViewport()
		return wv, nil
		
	case tea.KeyMsg:
		if wv.showLogs {
			return wv.handleLogsKeys(msg)
		}
		return wv.handleMainKeys(msg)

	case tickMsg:
		if wv.autoRefresh {
			wv.refreshStatuses()
			wv.updateLogsViewport() // Update viewport content on refresh
		}
		return wv, tickCmd()

	case statusUpdateMsg:
		wv.refreshStatuses()
		return wv, nil

	case validationCompleteMsg:
		wv.refreshStatuses()
		return wv, nil
	}

	// Update viewport for scrolling
	wv.logsViewport, cmd = wv.logsViewport.Update(msg)
	return wv, cmd
}

func (wv *WorkflowView) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.String() {
	case "q", "esc":
		// Return to workflow selector
		return NewWorkflowSelector(), nil
		
	// Logs viewport scrolling
	case "pgup":
		wv.logsViewport, cmd = wv.logsViewport.Update(msg)
		return wv, cmd
		
	case "pgdown":
		wv.logsViewport, cmd = wv.logsViewport.Update(msg)
		return wv, cmd

	case "up", "k":
		if wv.cursor > 0 {
			wv.cursor--
		}

	case "down", "j":
		if wv.cursor < len(wv.servers)-1 {
			wv.cursor++
		}

	case " ":
		if wv.cursor < len(wv.servers) {
			serverName := wv.servers[wv.cursor].Name
			wv.selectedServers[serverName] = !wv.selectedServers[serverName]
		}

	case "a":
		allSelected := len(wv.selectedServers) == len(wv.servers)
		wv.selectedServers = make(map[string]bool)
		if !allSelected {
			for _, s := range wv.servers {
				wv.selectedServers[s.Name] = true
			}
		}

	case "v":
		selected := wv.getSelectedServers()
		names := wv.getSelectedServerNames()
		if len(selected) == 0 {
			return wv, nil
		}
		
		// Step 1: Run local validation (fields check)
		for _, server := range selected {
			checks := wv.statusMgr.ValidateServer(server)
			wv.statusMgr.UpdateReadyChecks(server.Name, checks)
		}
		
		// Step 2: Queue network validation (SSH + connectivity check)
		if !wv.orchestrator.IsRunning() {
			wv.orchestrator.Start(wv.servers)
		}
		
		for _, name := range names {
			wv.statusMgr.UpdateStatus(name, status.StateVerifying, status.ActionCheck, "Validating...")
		}
		
		wv.orchestrator.QueueCheck(names, 0)
		
		// Immediate refresh for instant feedback
		wv.refreshStatuses()
		wv.updateLogsViewport()
		
		return wv, nil

	case "p":
		wv.provisionSelected()
		// Immediate refresh for instant feedback
		wv.refreshStatuses()
		wv.updateLogsViewport()

	case "d":
		wv.deploySelected()
		// Immediate refresh for instant feedback
		wv.refreshStatuses()
		wv.updateLogsViewport()

	case "l":
		if wv.cursor < len(wv.servers) {
			serverName := wv.servers[wv.cursor].Name
			logFile, err := wv.logReader.GetLatestLog(serverName)
			if err == nil {
				wv.showLogs = true
				wv.currentLogFile = logFile
				wv.loadLogs()
			}
		}

	case "r":
		wv.refreshStatuses()

	case "s":
		if wv.orchestrator.IsRunning() {
			wv.orchestrator.Stop()
		} else {
			wv.orchestrator.Start(wv.servers)
		}

	case "x":
		wv.orchestrator.ClearQueue()
	}

	return wv, nil
}

func (wv *WorkflowView) handleLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		wv.showLogs = false
		return wv, nil
	}
	return wv, nil
}



func (wv *WorkflowView) provisionSelected() {
	names := wv.getSelectedServerNames()
	if len(names) == 0 {
		return
	}
	
	// Ensure orchestrator is running
	if !wv.orchestrator.IsRunning() {
		wv.orchestrator.Start(wv.servers)
	}
	
	wv.orchestrator.QueueProvision(names, 0)
}

func (wv *WorkflowView) deploySelected() {
	names := wv.getSelectedServerNames()
	if len(names) == 0 {
		return
	}
	
	// Ensure orchestrator is running
	if !wv.orchestrator.IsRunning() {
		wv.orchestrator.Start(wv.servers)
	}
	
	wv.orchestrator.QueueDeploy(names, 0)
}



func (wv *WorkflowView) getSelectedServers() []*inventory.Server {
	result := make([]*inventory.Server, 0)
	for _, s := range wv.servers {
		if wv.selectedServers[s.Name] {
			result = append(result, s)
		}
	}
	return result
}

func (wv *WorkflowView) getSelectedServerNames() []string {
	result := make([]string, 0)
	for _, s := range wv.servers {
		if wv.selectedServers[s.Name] {
			result = append(result, s.Name)
		}
	}
	return result
}

func (wv *WorkflowView) loadLogs() {
	lines, err := wv.logReader.ReadLog(wv.currentLogFile, 100)
	if err != nil {
		wv.logLines = []string{"Error loading logs: " + err.Error()}
		return
	}

	formatted := make([]string, len(lines))
	for i, line := range lines {
		formatted[i] = wv.logReader.FormatLogLine(line)
	}
	wv.logLines = formatted
}

func (wv *WorkflowView) View() string {
	if wv.showLogs {
		return wv.renderLogs()
	}
	return wv.renderMain()
}

func (wv *WorkflowView) renderMain() string {
	var b strings.Builder

	title := titleStyle.Render(fmt.Sprintf("📋 Working with Inventory - %s", wv.environment))
	b.WriteString(title + "\n\n")

	table := wv.renderServerTable()
	b.WriteString(table + "\n\n")

	controls := wv.renderControls()
	b.WriteString(controls + "\n\n")

	queue := wv.renderQueue()
	b.WriteString(queue + "\n")
	
	// Always show realtime logs section
	b.WriteString("\n")
	logsSection := wv.renderRealtimeLogs()
	b.WriteString(logsSection)

	return b.String()
}

func (wv *WorkflowView) renderServerTable() string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("%-4s %-20s %-15s %-8s %-10s %-20s %-30s",
			"Sel", "Name", "IP", "Port", "Type", "Status", "Progress"))
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", 120) + "\n")

	for i, server := range wv.servers {
		sel := " "
		if wv.selectedServers[server.Name] {
			sel = "✓"
		}

		cursor := " "
		if i == wv.cursor {
			cursor = "▶"
		}

		st := wv.statuses[server.Name]
		if st == nil {
			st = &status.ServerStatus{State: status.StateUnknown}
		}

		statusStr := wv.formatStatus(st)
		progressStr := wv.progress[server.Name]
		if progressStr == "" {
			progressStr = "-"
		}

		line := fmt.Sprintf("%s %-2s %-20s %-15s %-8d %-10s %-20s %-30s",
			cursor, sel, server.Name, server.IP, server.Port, server.Type, statusStr, progressStr)

		if i == wv.cursor {
			line = selectedItemStyle.Render(line)
		}

		b.WriteString(line + "\n")
	}

	return b.String()
}

func (wv *WorkflowView) formatStatus(st *status.ServerStatus) string {
	var icon string
	switch st.State {
	case status.StateReady:
		icon = "✓ Ready"
		// Show ready checks details
		if st.ReadyChecks.IPValid && st.ReadyChecks.SSHKeyExists && st.ReadyChecks.PortValid && st.ReadyChecks.AllFieldsFilled {
			icon = "✓ Ready"
		} else {
			details := ""
			if !st.ReadyChecks.IPValid {
				details += "IP!"
			}
			if !st.ReadyChecks.SSHKeyExists {
				if details != "" { details += " " }
				details += "SSH!"
			}
			if !st.ReadyChecks.PortValid {
				if details != "" { details += " " }
				details += "Port!"
			}
			if !st.ReadyChecks.AllFieldsFilled {
				if details != "" { details += " " }
				details += "Fields!"
			}
			if details != "" {
				icon = "✓ Ready (" + details + ")"
			}
		}
	case status.StateNotReady:
		icon = "✗ Not Ready"
		// Show what's missing
		details := ""
		if !st.ReadyChecks.IPValid {
			details += "IP!"
		}
		if !st.ReadyChecks.SSHKeyExists {
			if details != "" { details += " " }
			details += "SSH!"
		}
		if !st.ReadyChecks.PortValid {
			if details != "" { details += " " }
			details += "Port!"
		}
		if !st.ReadyChecks.AllFieldsFilled {
			if details != "" { details += " " }
			details += "Fields!"
		}
		if details != "" {
			icon = "✗ Not Ready (" + details + ")"
		}
	case status.StateProvisioning:
		icon = "⚡ Provisioning"
	case status.StateProvisioned:
		icon = "✓ Provisioned"
	case status.StateDeploying:
		icon = "⚡ Deploying"
	case status.StateDeployed:
		icon = "✓ Deployed"
	case status.StateVerifying:
		icon = "🔍 Verifying"
	case status.StateFailed:
		icon = "✗ Failed"
	case "validating":
		icon = "⏳ Validating"
	default:
		icon = "? Unknown"
	}

	if st.ErrorMessage != "" && st.State != status.StateReady && st.State != status.StateNotReady {
		icon += " (" + st.ErrorMessage[:min(30, len(st.ErrorMessage))] + ")"
	}

	return icon
}

func (wv *WorkflowView) renderControls() string {
	controls := []string{
		"[↑↓] Navigate",
		"[Space] Select",
		"[a] Select All",
		"[v] Validate & Check",
		"[p] Provision",
		"[d] Deploy",
		"[PgUp/PgDn] Scroll Logs",
		"[l] Logs",
		"[r] Refresh",
		"[s] Start/Stop",
		"[x] Clear Queue",
		"[Esc] Back",
	}
	return helpStyle.Render(strings.Join(controls, " | "))
}

func (wv *WorkflowView) renderQueue() string {
	queueSize := wv.orchestrator.GetQueueSize()
	running := "Stopped"
	if wv.orchestrator.IsRunning() {
		running = "Running"
	}

	return infoBoxStyle.Render(fmt.Sprintf("Queue: %d actions | Status: %s | Last refresh: %s",
		queueSize, running, wv.lastRefresh.Format("15:04:05")))
}

func (wv *WorkflowView) renderRealtimeLogs() string {
	var b strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("cyan")).
		Padding(0, 1)
	
	b.WriteString(headerStyle.Render("📡 Live Output (PgUp/PgDown to scroll)") + "\n")
	b.WriteString(strings.Repeat("─", 120) + "\n")
	
	// Render viewport with scrollable logs
	b.WriteString(wv.logsViewport.View())
	
	// Show scroll position indicator if there's content to scroll
	if wv.logsViewport.TotalLineCount() > wv.logsViewport.Height {
		scrollInfo := fmt.Sprintf(" [%d/%d lines] ", 
			wv.logsViewport.YOffset+wv.logsViewport.Height,
			wv.logsViewport.TotalLineCount())
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		b.WriteString("\n" + infoStyle.Render(scrollInfo))
	}
	
	return b.String()
}

func (wv *WorkflowView) styleLogLine(line string) string {
	// Color-code log lines based on content
	switch {
	case strings.Contains(line, "✅") || strings.Contains(line, "✓"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Render(line)
	case strings.Contains(line, "❌") || strings.Contains(line, "✗"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Render(line)
	case strings.Contains(line, "⚙️") || strings.Contains(line, "Task:"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Render(line)
	case strings.Contains(line, "🚀") || strings.Contains(line, "Starting"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("yellow")).Render(line)
	case strings.Contains(line, "⚠️") || strings.Contains(line, "WARNING"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("orange")).Render(line)
	case strings.Contains(line, "📖") || strings.Contains(line, "▶️") || strings.Contains(line, "📊"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("magenta")).Render(line)
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("gray")).Render(line)
	}
}

func (wv *WorkflowView) renderLogs() string {
	var b strings.Builder

	title := titleStyle.Render("📄 Log Viewer: " + wv.currentLogFile)
	b.WriteString(title + "\n\n")

	for _, line := range wv.logLines {
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + helpStyle.Render("[q/esc] Back") + "\n")

	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
