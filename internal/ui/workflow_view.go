package ui

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/ansible"
	"github.com/bastiblast/boiler-deploy/internal/config"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/logging"
	"github.com/bastiblast/boiler-deploy/internal/status"
	"github.com/bastiblast/boiler-deploy/internal/storage"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorkflowView struct {
	mu                 sync.Mutex // Protects concurrent access to shared state
	environment        string
	servers            []*inventory.Server
	statuses           map[string]*status.ServerStatus
	selectedServers    map[string]bool
	cursor             int
	orchestrator       *ansible.Orchestrator
	statusMgr          *status.Manager
	logReader          *logging.Reader
	showLogs           bool
	currentLogFile     string
	logLines           []string
	progress           map[string]string
	lastRefresh        time.Time
	autoRefresh        bool
	realtimeLogs       []string // Live streaming output from ansible
	maxRealtimeLogs    int
	logsViewport       viewport.Model
	logsReady          bool
	userScrolling      bool // Track if user is manually scrolling logs
	tagSelector        *TagSelector
	showTagSelector    bool
	pendingAction      string // "provision" or "deploy"
	width              int
	height             int
	configMgr          *config.Manager
	configOpts         *config.ConfigOptions
	deploySuccessChan  chan deploySuccessMsg // Channel for deploy success events
}

type tickMsg time.Time
type statusUpdateMsg struct{}
type validationCompleteMsg struct{}
type deploySuccessMsg struct {
	serverName string
	serverIP   string
}

func NewWorkflowView() (*WorkflowView, error) {
	stor := storage.NewStorage(".")
	envs, err := stor.ListEnvironments()
	if err != nil || len(envs) == 0 {
		return nil, fmt.Errorf("no environments found")
	}

	return NewWorkflowViewWithEnv(envs[0])
}

func NewWorkflowViewWithEnv(envName string) (*WorkflowView, error) {
	// Load configuration for this environment
	configMgr := config.NewManager("inventory")
	configOpts, err := configMgr.Load(envName)
	if err != nil {
		configOpts = config.DefaultConfig()
	}
	
	wv := &WorkflowView{
		environment:       envName,
		selectedServers:   make(map[string]bool),
		progress:          make(map[string]string),
		autoRefresh:       true,
		realtimeLogs:      make([]string, 0),
		maxRealtimeLogs:   configOpts.LogRetention,
		logsViewport:      viewport.New(120, 20), // Initial size, will be updated dynamically
		logsReady:         true, // Set to true immediately
		configMgr:         configMgr,
		configOpts:        configOpts,
		deploySuccessChan: make(chan deploySuccessMsg, 10), // Buffered channel
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
	wv.orchestrator.SetDeploySuccessCallback(wv.onDeploySuccess)
	wv.orchestrator.SetHealthCheckEnabled(wv.configOpts.HealthCheckEnabled)
	wv.orchestrator.SetMaxWorkers(wv.configOpts.MaxParallelWorkers)

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
	wv.mu.Lock()
	
	wv.progress[serverName] = message
	
	// Add to realtime logs
	logLine := fmt.Sprintf("[%s] %s", serverName, message)
	wv.realtimeLogs = append(wv.realtimeLogs, logLine)
	
	// Keep only last N lines
	if len(wv.realtimeLogs) > wv.maxRealtimeLogs {
		wv.realtimeLogs = wv.realtimeLogs[len(wv.realtimeLogs)-wv.maxRealtimeLogs:]
	}
	
	wv.mu.Unlock()
	
	// Update viewport content (after unlocking to avoid holding lock during render)
	wv.updateLogsViewport()
}

func (wv *WorkflowView) onDeploySuccess(serverName, serverIP string) {
	log.Printf("[WORKFLOW] onDeploySuccess callback called: serverName=%s, serverIP=%s", serverName, serverIP)
	// Send to channel (non-blocking)
	select {
	case wv.deploySuccessChan <- deploySuccessMsg{serverName: serverName, serverIP: serverIP}:
		log.Printf("[WORKFLOW] Deploy success message sent to channel")
	default:
		log.Printf("[WORKFLOW] Warning: deploy success channel full, message dropped")
	}
}

// detectServerPort finds the correct HTTP port for a server
// Priority: http_port (external/browser) → app_port (internal) → 80 (default)
func (wv *WorkflowView) detectServerPort(serverIP string) int {
	// Find server by IP in loaded inventory
	for _, server := range wv.servers {
		if server.IP == serverIP {
			// Priority 1: http_port (external port for browser access)
			// Used for: Docker containers (8080/8081/8082), custom nginx ports
			if server.HTTPPort > 0 {
				log.Printf("[WORKFLOW] Using http_port=%d for server %s (external/browser access)", server.HTTPPort, server.Name)
				return server.HTTPPort
			}
			
			// Priority 2: Default port 80
			// Used for: Classic servers with nginx on standard HTTP port
			// Note: app_port (3000) is internal only, never used for browser
			log.Printf("[WORKFLOW] Using default port 80 for server %s (standard nginx)", server.Name)
			return 80
		}
	}
	
	// Fallback: Port 80 (if server not found in inventory)
	log.Printf("[WORKFLOW] Server with IP %s not found in inventory, using default port 80", serverIP)
	return 80
}

func (wv *WorkflowView) updateLogsViewport() {
	if !wv.logsReady {
		return
	}
	
	// Copy logs slice under lock for safe iteration
	wv.mu.Lock()
	logsCopy := make([]string, len(wv.realtimeLogs))
	copy(logsCopy, wv.realtimeLogs)
	wv.mu.Unlock()
	
	// Get current server name from cursor position
	var currentServerName string
	if wv.cursor >= 0 && wv.cursor < len(wv.servers) {
		currentServerName = wv.servers[wv.cursor].Name
	}
	
	var b strings.Builder
	
	if len(logsCopy) == 0 {
		dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		b.WriteString(dimStyle.Render("Waiting for actions...") + "\n")
	} else {
		// Filter logs to show only current server's logs
		filteredLogs := []string{}
		for _, line := range logsCopy {
			// Log format: [serverName] message
			if currentServerName != "" && strings.HasPrefix(line, "["+currentServerName+"]") {
				filteredLogs = append(filteredLogs, line)
			}
		}
		
		// If no logs for current server, show a message
		if len(filteredLogs) == 0 {
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			if currentServerName != "" {
				b.WriteString(dimStyle.Render(fmt.Sprintf("No logs yet for %s...", currentServerName)) + "\n")
			} else {
				b.WriteString(dimStyle.Render("No server selected") + "\n")
			}
		} else {
			for _, line := range filteredLogs {
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
	}
	
	wv.logsViewport.SetContent(b.String())
	// Auto-scroll to bottom only if user is not manually scrolling
	if !wv.userScrolling {
		wv.logsViewport.GotoBottom()
	}
}

func (wv *WorkflowView) refreshStatuses() {
	wv.statuses = wv.statusMgr.GetAllStatuses()
	wv.lastRefresh = time.Now()
}

func (wv *WorkflowView) Init() tea.Cmd {
	wv.orchestrator.Start(wv.servers)
	return tea.Batch(
		wv.tickCmd(),
		wv.waitForDeploySuccess(),
	)
}

func (wv *WorkflowView) waitForDeploySuccess() tea.Cmd {
	return func() tea.Msg {
		msg := <-wv.deploySuccessChan
		log.Printf("[WORKFLOW] Received deploy success from channel: %s -> %s", msg.serverName, msg.serverIP)
		return msg
	}
}

func (wv *WorkflowView) tickCmd() tea.Cmd {
	return tea.Tick(wv.configOpts.RefreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (wv *WorkflowView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	// Always handle WindowSizeMsg first to ensure tag selector gets proper dimensions
	if wsMsg, ok := msg.(tea.WindowSizeMsg); ok {
		// Store terminal dimensions
		wv.width = wsMsg.Width
		wv.height = wsMsg.Height
		
		// Update viewport size when terminal is resized
		// Calculate available height: total height - header - table - controls - queue - margins
		// Approximate: title(3) + table(servers*1 + 2) + controls(3) + queue(2) + margins(5) = ~15 + servers
		logsHeight := wsMsg.Height - 18 - len(wv.servers)
		if logsHeight < 5 {
			logsHeight = 5 // Minimum height
		}
		if logsHeight > 30 {
			logsHeight = 30 // Maximum height for readability
		}
		
		if !wv.logsReady {
			wv.logsViewport = viewport.New(wsMsg.Width, logsHeight)
			wv.logsReady = true
		} else {
			wv.logsViewport.Width = wsMsg.Width
			wv.logsViewport.Height = logsHeight
		}
		wv.updateLogsViewport()
		
		// If tag selector is visible, pass the message to it
		if wv.showTagSelector && wv.tagSelector != nil {
			updatedSelector, _ := wv.tagSelector.Update(msg)
			if selector, ok := updatedSelector.(TagSelector); ok {
				*wv.tagSelector = selector
			}
		}
		return wv, nil
	}
	
	// Handle tag selector
	if wv.showTagSelector && wv.tagSelector != nil {
		updatedSelector, cmd := wv.tagSelector.Update(msg)
		if selector, ok := updatedSelector.(TagSelector); ok {
			*wv.tagSelector = selector
			
			if selector.IsConfirmed() {
				// Get selected tags and execute action
				tags := selector.GetTagString()
				wv.showTagSelector = false
				wv.tagSelector = nil
				action := wv.pendingAction
				wv.pendingAction = ""
				// Execute action and force immediate refresh
				wv.executeActionWithTags(action, tags)
				wv.refreshStatuses()
				return wv, wv.tickCmd()
			} else if selector.IsCancelled() {
				wv.showTagSelector = false
				wv.tagSelector = nil
				wv.pendingAction = ""
				wv.refreshStatuses()
				return wv, wv.tickCmd()
			}
		}
		return wv, cmd
	}
	
	switch msg := msg.(type) {
		
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
		return wv, wv.tickCmd()

	case statusUpdateMsg:
		wv.refreshStatuses()
		return wv, nil

	case validationCompleteMsg:
		wv.refreshStatuses()
		return wv, nil
		
	case deploySuccessMsg:
		log.Printf("[WORKFLOW] Processing deploySuccessMsg: %s -> %s", msg.serverName, msg.serverIP)
		
		logLine := fmt.Sprintf("[%s] ✓ Deployment successful! Site ready at http://%s", msg.serverName, msg.serverIP)
		
		wv.mu.Lock()
		wv.realtimeLogs = append(wv.realtimeLogs, logLine)
		// Note: Browser option shown in Progress column for deployed servers
		
		if len(wv.realtimeLogs) > wv.maxRealtimeLogs {
			wv.realtimeLogs = wv.realtimeLogs[len(wv.realtimeLogs)-wv.maxRealtimeLogs:]
		}
		wv.mu.Unlock()
		
		wv.updateLogsViewport()
		
		// Re-subscribe to channel for next deploy success
		return wv, wv.waitForDeploySuccess()
	}

	// Update viewport for scrolling
	wv.logsViewport, cmd = wv.logsViewport.Update(msg)
	return wv, cmd
}

func (wv *WorkflowView) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg.String() {
	case "o":
		// Open browser for selected server (if deployed)
		if wv.cursor >= 0 && wv.cursor < len(wv.servers) {
			server := wv.servers[wv.cursor]
			st := wv.statuses[server.Name]
			
			log.Printf("[WORKFLOW] 'o' key pressed for server: %s, status: %s", server.Name, st.State)
			
			// Only allow browser open if server is deployed
			if st != nil && st.State == status.StateDeployed {
				// Detect correct port from server configuration
				port := wv.detectServerPort(server.IP)
				url := fmt.Sprintf("http://%s:%d", server.IP, port)
				log.Printf("[WORKFLOW] Opening browser for URL: %s (detected port: %d)", url, port)
				
				var logLine string
				if err := OpenBrowser(url); err != nil {
					logLine = fmt.Sprintf("[%s] Failed to open browser: %v", server.Name, err)
					log.Printf("[WORKFLOW] Browser open failed: %v", err)
				} else {
					logLine = fmt.Sprintf("[%s] Opening %s in browser...", server.Name, url)
					log.Printf("[WORKFLOW] Browser opened successfully")
				}
				
				wv.mu.Lock()
				wv.realtimeLogs = append(wv.realtimeLogs, logLine)
				wv.mu.Unlock()
				
				wv.updateLogsViewport()
			} else {
				log.Printf("[WORKFLOW] Server %s not deployed (status: %s), cannot open browser", 
					server.Name, st.State)
			}
		}
	
	case "q", "esc":
		// Return to workflow selector
		return NewWorkflowSelector(), nil
		
	// Logs viewport scrolling
	case "pgup":
		wv.userScrolling = true // User is manually scrolling
		wv.logsViewport, cmd = wv.logsViewport.Update(msg)
		return wv, cmd
		
	case "pgdown":
		wv.userScrolling = true // User is manually scrolling
		wv.logsViewport, cmd = wv.logsViewport.Update(msg)
		// If user scrolled to bottom, re-enable auto-scroll
		if wv.logsViewport.AtBottom() {
			wv.userScrolling = false
		}
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
		// Use checked servers, or server at cursor if none checked
		selected := wv.getServersForAction()
		names := wv.getServerNamesForAction()
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
		// Open tag selector for provision
		// Use checked servers, or server at cursor if none checked
		if len(wv.getServersForAction()) > 0 {
			selector := NewTagSelectorWithDefaults("provision", wv.configOpts.ProvisioningTags)
			// Initialize with current terminal dimensions
			selector.width = wv.width
			selector.height = wv.height
			wv.tagSelector = &selector
			wv.showTagSelector = true
			wv.pendingAction = "provision"
		}

	case "d":
		// Open tag selector for deploy
		// Use checked servers, or server at cursor if none checked
		if len(wv.getServersForAction()) > 0 {
			selector := NewTagSelectorWithDefaults("deploy", wv.configOpts.DeploymentTags)
			// Initialize with current terminal dimensions
			selector.width = wv.width
			selector.height = wv.height
			wv.tagSelector = &selector
			wv.showTagSelector = true
			wv.pendingAction = "deploy"
		}

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



func (wv *WorkflowView) executeActionWithTags(action, tags string) {
	// Use checked servers, or server at cursor if none checked
	names := wv.getServerNamesForAction()
	
	if len(names) == 0 {
		return
	}
	
	// Ensure orchestrator is running
	if !wv.orchestrator.IsRunning() {
		wv.orchestrator.Start(wv.servers)
	}
	
	switch action {
	case "provision":
		wv.orchestrator.QueueProvisionWithTags(names, 0, tags)
	case "deploy":
		wv.orchestrator.QueueDeployWithTags(names, 0, tags)
	}
	
	// Immediate refresh for instant feedback
	wv.refreshStatuses()
	wv.updateLogsViewport()
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

// getServersForAction returns servers to use for action (provision/deploy)
// Logic: If servers are checked, use checked servers
//        If none checked, use server at cursor position
func (wv *WorkflowView) getServersForAction() []*inventory.Server {
	selected := wv.getSelectedServers()
	
	// If servers are checked, use them
	if len(selected) > 0 {
		return selected
	}
	
	// If no servers checked, use server at cursor (if valid)
	if wv.cursor >= 0 && wv.cursor < len(wv.servers) {
		return []*inventory.Server{wv.servers[wv.cursor]}
	}
	
	// Fallback: empty list
	return []*inventory.Server{}
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

// getServerNamesForAction returns server names for action (provision/deploy)
// Logic: If servers are checked, use checked servers
//        If none checked, use server at cursor position
func (wv *WorkflowView) getServerNamesForAction() []string {
	servers := wv.getServersForAction()
	result := make([]string, 0, len(servers))
	for _, s := range servers {
		result = append(result, s.Name)
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
	if wv.showTagSelector && wv.tagSelector != nil {
		return wv.tagSelector.View()
	}
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
		fmt.Sprintf("  %-2s %-20s %-15s %-7s %-7s %-22s %-43s",
			"✓", "Name", "IP", "Port", "Type", "Status", "Progress"))
	b.WriteString(header + "\n")
	b.WriteString(strings.Repeat("─", 125) + "\n")

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

		statusStr, progressDetails := wv.formatStatus(st)
		progressStr := wv.progress[server.Name]
		if progressStr == "" {
			progressStr = progressDetails
		} else if progressDetails != "" {
			progressStr = progressDetails + " | " + progressStr
		}
		if progressStr == "" {
			progressStr = "-"
		}

		line := fmt.Sprintf("%s %-2s %-20s %-15s %-7d %-7s %-22s %-43s",
			cursor, sel, server.Name, server.IP, server.Port, server.Type, statusStr, progressStr)

		if i == wv.cursor {
			line = selectedItemStyle.Render(line)
		}

		b.WriteString(line + "\n")
	}

	return b.String()
}

func (wv *WorkflowView) formatStatus(st *status.ServerStatus) (string, string) {
	var icon string
	var progressDetails string
	
	// Color styles
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	
	switch st.State {
	case status.StateReady:
		icon = yellowStyle.Render("✓ Ready")
		// When validated successfully, show "All checks passed" in progress
		progressDetails = "All checks passed"
	case status.StateNotReady:
		icon = redStyle.Render("✗ Not Ready")
		// Show what's missing in progress column
		details := []string{}
		if !st.ReadyChecks.IPValid {
			details = append(details, "Invalid IP")
		}
		if !st.ReadyChecks.SSHKeyExists {
			details = append(details, "SSH key not found")
		}
		if !st.ReadyChecks.PortValid {
			details = append(details, "Invalid port")
		}
		if !st.ReadyChecks.AllFieldsFilled {
			details = append(details, "Missing fields")
		}
		if len(details) > 0 {
			progressDetails = strings.Join(details, ", ")
		}
	case status.StateProvisioning:
		icon = yellowStyle.Render("⚡ Provisioning")
	case status.StateProvisioned:
		icon = blueStyle.Render("✓ Provisioned")
	case status.StateDeploying:
		icon = yellowStyle.Render("⚡ Deploying")
	case status.StateDeployed:
		icon = greenStyle.Render("✓ Deployed")
		// Show browser hint for deployed servers
		progressDetails = "Press 'o' to open in browser"
	case status.StateVerifying:
		icon = blueStyle.Render("🔍 Verifying")
	case status.StateFailed:
		icon = redStyle.Render("✗ Failed")
		if st.ErrorMessage != "" {
			progressDetails = st.ErrorMessage
		}
	case "validating":
		icon = yellowStyle.Render("⏳ Validating")
	default:
		icon = grayStyle.Render("? Unknown")
	}

	return icon, progressDetails
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
	
	// Header with current server name
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("cyan")).
		Padding(0, 1)
	
	var serverName string
	if wv.cursor >= 0 && wv.cursor < len(wv.servers) {
		serverName = wv.servers[wv.cursor].Name
	}
	
	headerText := "📡 Live Output (PgUp/PgDown to scroll)"
	if serverName != "" {
		headerText = fmt.Sprintf("📡 Live Output - %s (PgUp/PgDown to scroll)", serverName)
	}
	
	b.WriteString(headerStyle.Render(headerText) + "\n")
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
