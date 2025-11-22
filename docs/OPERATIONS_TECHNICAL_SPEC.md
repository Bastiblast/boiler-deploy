# 🔧 Operations Workflow - Technical Specification

## 📋 Overview

This document provides the technical implementation details for the Operations Workflow feature (v2.0).

---

## 🗂️ File Structure

### New Files to Create

```
internal/
├── ansible/
│   ├── executor.go         # Ansible playbook execution
│   ├── parser.go           # Output parsing and streaming
│   ├── status.go           # Status management
│   ├── queue.go            # Queue system for batch operations
│   └── validator.go        # Inventory validation
├── ui/
│   ├── operations_dashboard.go  # Main operations screen
│   ├── execution_view.go        # Single operation view
│   ├── queue_view.go            # Batch operations view
│   └── log_viewer.go            # Log viewing component
└── models/
    └── status.go           # Status data structures

inventory/
└── {environment}/
    ├── config.yml          # Existing
    ├── hosts.yml           # Existing
    └── status.yml          # NEW: Deployment status

logs/
└── {environment}/
    ├── provision_20251111_083000.log
    ├── deploy_20251111_084500.log
    └── validate_20251111_090000.log
```

---

## 📊 Data Structures

### status.yml Format

```yaml
environment: production
last_updated: "2025-11-11T08:30:00Z"

# Per-server status
servers:
  - name: production-web-01
    # SSH connectivity
    ssh:
      status: connected          # connected, failed, unknown
      last_check: "2025-11-11T08:25:00Z"
      latency_ms: 45
      error: null
    
    # Provisioning status
    provision:
      status: completed          # not_started, in_progress, completed, failed
      started: "2025-11-10T15:30:00Z"
      completed: "2025-11-10T15:45:00Z"
      error: null
    
    # Deployment status
    deploy:
      status: deployed           # not_deployed, deploying, deployed, failed
      version: "v1.2.3"
      started: "2025-11-11T08:15:00Z"
      completed: "2025-11-11T08:16:30Z"
      error: null

# Playbook execution history
history:
  - id: "exec_20251111_083000"
    type: provision              # provision, deploy, update, rollback, validate
    playbook: playbooks/provision.yml
    started: "2025-11-11T08:30:00Z"
    completed: "2025-11-11T08:35:00Z"
    status: success              # success, failed, cancelled
    servers: [production-web-01, production-web-02]
    log_file: logs/production/provision_20251111_083000.log
    summary:
      total_tasks: 45
      ok: 40
      changed: 5
      failed: 0
      skipped: 0
```

### Go Status Models

```go
// internal/models/status.go

package models

import "time"

type EnvironmentStatus struct {
    Environment string         `yaml:"environment"`
    LastUpdated time.Time      `yaml:"last_updated"`
    Servers     []ServerStatus `yaml:"servers"`
    History     []ExecutionHistory `yaml:"history"`
}

type ServerStatus struct {
    Name      string          `yaml:"name"`
    SSH       ConnectionStatus `yaml:"ssh"`
    Provision OperationStatus  `yaml:"provision"`
    Deploy    OperationStatus  `yaml:"deploy"`
}

type ConnectionStatus struct {
    Status    string    `yaml:"status"` // connected, failed, unknown
    LastCheck time.Time `yaml:"last_check"`
    LatencyMs int       `yaml:"latency_ms"`
    Error     string    `yaml:"error,omitempty"`
}

type OperationStatus struct {
    Status    string    `yaml:"status"` // not_started, in_progress, completed, failed
    Started   *time.Time `yaml:"started,omitempty"`
    Completed *time.Time `yaml:"completed,omitempty"`
    Version   string    `yaml:"version,omitempty"` // For deploy only
    Error     string    `yaml:"error,omitempty"`
}

type ExecutionHistory struct {
    ID        string    `yaml:"id"`
    Type      string    `yaml:"type"` // provision, deploy, update, rollback, validate
    Playbook  string    `yaml:"playbook"`
    Started   time.Time `yaml:"started"`
    Completed *time.Time `yaml:"completed,omitempty"`
    Status    string    `yaml:"status"` // success, failed, cancelled, running
    Servers   []string  `yaml:"servers"`
    LogFile   string    `yaml:"log_file"`
    Summary   *ExecutionSummary `yaml:"summary,omitempty"`
}

type ExecutionSummary struct {
    TotalTasks int `yaml:"total_tasks"`
    OK         int `yaml:"ok"`
    Changed    int `yaml:"changed"`
    Failed     int `yaml:"failed"`
    Skipped    int `yaml:"skipped"`
}
```

---

## 🔧 Implementation Details

### 1. Status Management (`internal/ansible/status.go`)

```go
package ansible

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
    
    "gopkg.in/yaml.v3"
    "github.com/bastiblast/boiler-deploy/internal/models"
)

type StatusManager struct {
    basePath string
}

func NewStatusManager(basePath string) *StatusManager {
    return &StatusManager{basePath: basePath}
}

// Load status for an environment
func (sm *StatusManager) Load(env string) (*models.EnvironmentStatus, error) {
    statusPath := filepath.Join(sm.basePath, "inventory", env, "status.yml")
    
    // If status file doesn't exist, return empty status
    if _, err := os.Stat(statusPath); os.IsNotExist(err) {
        return sm.initializeStatus(env), nil
    }
    
    data, err := os.ReadFile(statusPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read status: %v", err)
    }
    
    var status models.EnvironmentStatus
    if err := yaml.Unmarshal(data, &status); err != nil {
        return nil, fmt.Errorf("failed to parse status: %v", err)
    }
    
    return &status, nil
}

// Save status for an environment
func (sm *StatusManager) Save(status *models.EnvironmentStatus) error {
    status.LastUpdated = time.Now()
    
    statusPath := filepath.Join(sm.basePath, "inventory", status.Environment, "status.yml")
    
    data, err := yaml.Marshal(status)
    if err != nil {
        return fmt.Errorf("failed to marshal status: %v", err)
    }
    
    // Atomic write: write to temp file, then rename
    tempPath := statusPath + ".tmp"
    if err := os.WriteFile(tempPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write status: %v", err)
    }
    
    if err := os.Rename(tempPath, statusPath); err != nil {
        return fmt.Errorf("failed to save status: %v", err)
    }
    
    return nil
}

// Initialize empty status for new environment
func (sm *StatusManager) initializeStatus(env string) *models.EnvironmentStatus {
    return &models.EnvironmentStatus{
        Environment: env,
        LastUpdated: time.Now(),
        Servers:     []models.ServerStatus{},
        History:     []models.ExecutionHistory{},
    }
}

// Update SSH status for a server
func (sm *StatusManager) UpdateSSHStatus(env, serverName string, connected bool, latencyMs int, err error) error {
    status, loadErr := sm.Load(env)
    if loadErr != nil {
        return loadErr
    }
    
    // Find or create server status
    var serverStatus *models.ServerStatus
    for i := range status.Servers {
        if status.Servers[i].Name == serverName {
            serverStatus = &status.Servers[i]
            break
        }
    }
    
    if serverStatus == nil {
        // Create new server status
        status.Servers = append(status.Servers, models.ServerStatus{
            Name: serverName,
        })
        serverStatus = &status.Servers[len(status.Servers)-1]
    }
    
    // Update SSH status
    serverStatus.SSH.LastCheck = time.Now()
    serverStatus.SSH.LatencyMs = latencyMs
    
    if connected {
        serverStatus.SSH.Status = "connected"
        serverStatus.SSH.Error = ""
    } else {
        serverStatus.SSH.Status = "failed"
        if err != nil {
            serverStatus.SSH.Error = err.Error()
        }
    }
    
    return sm.Save(status)
}

// Add execution to history
func (sm *StatusManager) AddExecution(env string, exec models.ExecutionHistory) error {
    status, err := sm.Load(env)
    if err != nil {
        return err
    }
    
    status.History = append(status.History, exec)
    
    // Keep only last 50 executions
    if len(status.History) > 50 {
        status.History = status.History[len(status.History)-50:]
    }
    
    return sm.Save(status)
}

// Update operation status (provision/deploy)
func (sm *StatusManager) UpdateOperationStatus(env, serverName, operation, newStatus string, err error) error {
    status, loadErr := sm.Load(env)
    if loadErr != nil {
        return loadErr
    }
    
    // Find server
    for i := range status.Servers {
        if status.Servers[i].Name == serverName {
            now := time.Now()
            
            var opStatus *models.OperationStatus
            if operation == "provision" {
                opStatus = &status.Servers[i].Provision
            } else if operation == "deploy" {
                opStatus = &status.Servers[i].Deploy
            } else {
                return fmt.Errorf("unknown operation: %s", operation)
            }
            
            opStatus.Status = newStatus
            
            // Set timestamps
            if newStatus == "in_progress" && opStatus.Started == nil {
                opStatus.Started = &now
            } else if newStatus == "completed" || newStatus == "failed" {
                if opStatus.Started == nil {
                    opStatus.Started = &now
                }
                opStatus.Completed = &now
            }
            
            // Set error if failed
            if err != nil {
                opStatus.Error = err.Error()
            } else {
                opStatus.Error = ""
            }
            
            break
        }
    }
    
    return sm.Save(status)
}
```

---

### 2. Ansible Executor (`internal/ansible/executor.go`)

```go
package ansible

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/bastiblast/boiler-deploy/internal/models"
)

type Executor struct {
    basePath      string
    statusManager *StatusManager
}

func NewExecutor(basePath string) *Executor {
    return &Executor{
        basePath:      basePath,
        statusManager: NewStatusManager(basePath),
    }
}

// ExecutionOptions configures playbook execution
type ExecutionOptions struct {
    Environment string
    Playbook    string
    Limit       []string // Limit to specific servers
    ExtraVars   map[string]string
    Tags        []string
    Verbose     bool
}

// ExecutionResult contains execution results
type ExecutionResult struct {
    Success   bool
    StartTime time.Time
    EndTime   time.Time
    Summary   *models.ExecutionSummary
    LogFile   string
    Error     error
}

// Execute runs an Ansible playbook
func (e *Executor) Execute(ctx context.Context, opts ExecutionOptions, outputChan chan<- string) (*ExecutionResult, error) {
    result := &ExecutionResult{
        StartTime: time.Now(),
    }
    
    // Prepare paths
    inventoryPath := filepath.Join(e.basePath, "inventory", opts.Environment, "hosts.yml")
    playbookPath := filepath.Join(e.basePath, opts.Playbook)
    
    // Check if files exist
    if _, err := os.Stat(inventoryPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("inventory not found: %s", inventoryPath)
    }
    if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("playbook not found: %s", playbookPath)
    }
    
    // Prepare log file
    logDir := filepath.Join(e.basePath, "logs", opts.Environment)
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create log directory: %v", err)
    }
    
    timestamp := time.Now().Format("20060102_150405")
    playbookName := strings.TrimSuffix(filepath.Base(opts.Playbook), ".yml")
    logFile := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", playbookName, timestamp))
    result.LogFile = logFile
    
    logWriter, err := os.Create(logFile)
    if err != nil {
        return nil, fmt.Errorf("failed to create log file: %v", err)
    }
    defer logWriter.Close()
    
    // Build ansible-playbook command
    args := []string{
        "-i", inventoryPath,
        playbookPath,
    }
    
    if len(opts.Limit) > 0 {
        args = append(args, "--limit", strings.Join(opts.Limit, ","))
    }
    
    if len(opts.Tags) > 0 {
        args = append(args, "--tags", strings.Join(opts.Tags, ","))
    }
    
    if opts.Verbose {
        args = append(args, "-vvv")
    }
    
    for key, value := range opts.ExtraVars {
        args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
    }
    
    // Create command
    cmd := exec.CommandContext(ctx, "ansible-playbook", args...)
    cmd.Dir = e.basePath
    
    // Capture stdout and stderr
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
    }
    
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to get stderr pipe: %v", err)
    }
    
    // Start command
    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("failed to start ansible-playbook: %v", err)
    }
    
    // Stream output
    parser := NewParser()
    
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            line := scanner.Text()
            
            // Write to log
            fmt.Fprintln(logWriter, line)
            
            // Send to output channel
            if outputChan != nil {
                outputChan <- line
            }
            
            // Parse line for statistics
            parser.ParseLine(line)
        }
    }()
    
    go func() {
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() {
            line := scanner.Text()
            fmt.Fprintln(logWriter, "[STDERR] "+line)
            if outputChan != nil {
                outputChan <- "[ERROR] " + line
            }
        }
    }()
    
    // Wait for completion
    err = cmd.Wait()
    result.EndTime = time.Now()
    
    if err != nil {
        result.Success = false
        result.Error = err
    } else {
        result.Success = true
    }
    
    // Get summary from parser
    result.Summary = parser.GetSummary()
    
    // Close output channel
    if outputChan != nil {
        close(outputChan)
    }
    
    return result, nil
}

// ValidateInventory validates an Ansible inventory
func (e *Executor) ValidateInventory(env string) error {
    inventoryPath := filepath.Join(e.basePath, "inventory", env, "hosts.yml")
    
    cmd := exec.Command("ansible-inventory", "-i", inventoryPath, "--list")
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        return fmt.Errorf("inventory validation failed: %v\n%s", err, string(output))
    }
    
    return nil
}

// GetPlaybookPath returns the full path to a playbook
func (e *Executor) GetPlaybookPath(name string) string {
    return filepath.Join("playbooks", name+".yml")
}
```

---

### 3. Output Parser (`internal/ansible/parser.go`)

```go
package ansible

import (
    "regexp"
    "sync"
    
    "github.com/bastiblast/boiler-deploy/internal/models"
)

type Parser struct {
    mu      sync.Mutex
    summary models.ExecutionSummary
}

func NewParser() *Parser {
    return &Parser{}
}

var (
    // Regex patterns for Ansible output
    playRegex    = regexp.MustCompile(`^PLAY \[(.*)\]`)
    taskRegex    = regexp.MustCompile(`^TASK \[(.*)\]`)
    okRegex      = regexp.MustCompile(`^ok: \[(.*)\]`)
    changedRegex = regexp.MustCompile(`^changed: \[(.*)\]`)
    failedRegex  = regexp.MustCompile(`^failed: \[(.*)\]`)
    skippedRegex = regexp.MustCompile(`^skipping: \[(.*)\]`)
    recapRegex   = regexp.MustCompile(`^(.*)\s+:\s+ok=(\d+)\s+changed=(\d+)\s+unreachable=(\d+)\s+failed=(\d+)\s+skipped=(\d+)`)
)

// ParseLine parses a single line of Ansible output
func (p *Parser) ParseLine(line string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if taskRegex.MatchString(line) {
        p.summary.TotalTasks++
    } else if okRegex.MatchString(line) {
        p.summary.OK++
    } else if changedRegex.MatchString(line) {
        p.summary.Changed++
    } else if failedRegex.MatchString(line) {
        p.summary.Failed++
    } else if skippedRegex.MatchString(line) {
        p.summary.Skipped++
    }
}

// GetSummary returns the execution summary
func (p *Parser) GetSummary() *models.ExecutionSummary {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    return &models.ExecutionSummary{
        TotalTasks: p.summary.TotalTasks,
        OK:         p.summary.OK,
        Changed:    p.summary.Changed,
        Failed:     p.summary.Failed,
        Skipped:    p.summary.Skipped,
    }
}
```

---

## 🖥️ UI Components

### Operations Dashboard (`internal/ui/operations_dashboard.go`)

```go
package ui

import (
    "fmt"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/bastiblast/boiler-deploy/internal/ansible"
    "github.com/bastiblast/boiler-deploy/internal/storage"
    "github.com/bastiblast/boiler-deploy/internal/models"
)

type OperationsDashboard struct {
    environments []string
    statuses     map[string]*models.EnvironmentStatus
    cursor       int
    selectedEnv  string
    storage      *storage.Storage
    statusMgr    *ansible.StatusManager
}

func NewOperationsDashboard() OperationsDashboard {
    stor := storage.NewStorage(".")
    statusMgr := ansible.NewStatusManager(".")
    
    envs, _ := stor.ListEnvironments()
    
    // Load status for each environment
    statuses := make(map[string]*models.EnvironmentStatus)
    for _, env := range envs {
        status, err := statusMgr.Load(env)
        if err == nil {
            statuses[env] = status
        }
    }
    
    return OperationsDashboard{
        environments: envs,
        statuses:     statuses,
        cursor:       0,
        storage:      stor,
        statusMgr:    statusMgr,
    }
}

// ... Implementation continues ...
```

---

## 📋 Next Steps

1. **Review this technical spec** - Confirm approach is correct
2. **Answer clarification questions** from roadmap document
3. **Begin Phase 1 implementation** - Data models & storage
4. **Iterate with feedback**

---

**Last Updated:** 2025-11-11
**Status:** Technical Design
