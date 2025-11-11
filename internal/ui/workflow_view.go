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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkflowView struct {
	environments    []string
	currentEnvIndex int
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

	wv := &WorkflowView{
		environments:    envs,
		currentEnvIndex: 0,
		selectedServers: make(map[string]bool),
		progress:        make(map[string]string),
		autoRefresh:     true,
	}

	if err := wv.loadEnvironment(); err != nil {
		return nil, err
	}

	return wv, nil
}

func (wv *WorkflowView) loadEnvironment() error {
	envName := wv.environments[wv.currentEnvIndex]

	stor := storage.NewStorage(".")
	env, err := stor.LoadEnvironment(envName)
	if err != nil {
		return err
	}

	wv.servers = make([]*inventory.Server, len(env.Servers))
	for i := range env.Servers {
		wv.servers[i] = &env.Servers[i]
	}

	statusMgr, err := status.NewManager(envName)
	if err != nil {
		return err
	}
	wv.statusMgr = statusMgr

	orchestrator, err := ansible.NewOrchestrator(envName, statusMgr)
	if err != nil {
		return err
	}
	wv.orchestrator = orchestrator
	wv.orchestrator.SetProgressCallback(wv.onProgress)

	wv.logReader = logging.NewReader(envName)

	wv.refreshStatuses()

	return nil
}

func (wv *WorkflowView) onProgress(serverName, message string) {
	wv.progress[serverName] = message
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
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (wv *WorkflowView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if wv.showLogs {
			return wv.handleLogsKeys(msg)
		}
		return wv.handleMainKeys(msg)

	case tickMsg:
		if wv.autoRefresh {
			wv.refreshStatuses()
		}
		return wv, tickCmd()

	case statusUpdateMsg:
		wv.refreshStatuses()
		return wv, nil

	case validationCompleteMsg:
		wv.refreshStatuses()
		return wv, nil
	}

	return wv, nil
}

func (wv *WorkflowView) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return wv, tea.Quit

	case "tab":
		wv.currentEnvIndex = (wv.currentEnvIndex + 1) % len(wv.environments)
		wv.selectedServers = make(map[string]bool)
		wv.cursor = 0
		wv.loadEnvironment()
		return wv, nil

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
		if len(selected) == 0 {
			return wv, nil
		}
		
		// Run validation synchronously (it's fast, no need for goroutine or intermediate state)
		for _, server := range selected {
			checks := wv.statusMgr.ValidateServer(server)
			wv.statusMgr.UpdateReadyChecks(server.Name, checks)
		}
		
		// Refresh to show updated status
		wv.refreshStatuses()
		
		return wv, nil

	case "p":
		wv.provisionSelected()

	case "d":
		wv.deploySelected()

	case "c":
		selected := wv.getSelectedServerNames()
		if len(selected) == 0 {
			return wv, nil
		}
		
		// Ensure orchestrator is running
		if !wv.orchestrator.IsRunning() {
			wv.orchestrator.Start(wv.servers)
		}
		
		// Update status to Verifying immediately for visual feedback
		for _, name := range selected {
			wv.statusMgr.UpdateStatus(name, status.StateVerifying, status.ActionCheck, "Queued...")
		}
		
		// Queue checks
		wv.orchestrator.QueueCheck(selected, 0)
		wv.refreshStatuses()

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
	if len(names) > 0 {
		wv.orchestrator.QueueProvision(names, 0)
	}
}

func (wv *WorkflowView) deploySelected() {
	names := wv.getSelectedServerNames()
	if len(names) > 0 {
		wv.orchestrator.QueueDeploy(names, 0)
	}
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

	title := titleStyle.Render(fmt.Sprintf("📋 Working with Inventory - %s", wv.environments[wv.currentEnvIndex]))
	b.WriteString(title + "\n\n")

	envTabs := wv.renderEnvironmentTabs()
	b.WriteString(envTabs + "\n\n")

	table := wv.renderServerTable()
	b.WriteString(table + "\n\n")

	controls := wv.renderControls()
	b.WriteString(controls + "\n\n")

	queue := wv.renderQueue()
	b.WriteString(queue + "\n")

	return b.String()
}

func (wv *WorkflowView) renderEnvironmentTabs() string {
	tabs := make([]string, len(wv.environments))
	for i, env := range wv.environments {
		if i == wv.currentEnvIndex {
			tabs[i] = selectedItemStyle.Render(" " + env + " ")
		} else {
			tabs[i] = helpStyle.Render(" " + env + " ")
		}
	}
	return strings.Join(tabs, " ")
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
		"[Space] Select",
		"[a] Select All",
		"[v] Validate",
		"[p] Provision",
		"[d] Deploy",
		"[c] Check",
		"[l] Logs",
		"[r] Refresh",
		"[s] Start/Stop",
		"[x] Clear Queue",
		"[Tab] Switch Env",
		"[q] Quit",
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
