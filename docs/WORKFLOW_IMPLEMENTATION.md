# 🏗️ Workflow Implementation Documentation

## Architecture Overview

The Ansible workflow system is built with a modular architecture using Go and the Bubbletea TUI framework.

---

## 📦 Component Structure

```
internal/
├── status/           # Server status tracking & persistence
│   ├── models.go     # Status data structures
│   └── manager.go    # Status CRUD operations
├── ansible/          # Playbook execution & queue management
│   ├── executor.go   # Ansible playbook runner
│   ├── queue.go      # FIFO action queue
│   └── orchestrator.go # High-level workflow coordination
├── logging/          # Log file management
│   └── reader.go     # Log reading & formatting
└── ui/               # User interface
    └── workflow_view.go # Main workflow TUI view
```

---

## 🔄 Status System

### File: `internal/status/models.go`

#### ServerState Enum

```go
const (
    StateUnknown      ServerState = "unknown"
    StateNotReady     ServerState = "not_ready"
    StateReady        ServerState = "ready"
    StateProvisioning ServerState = "provisioning"
    StateProvisioned  ServerState = "provisioned"
    StateDeploying    ServerState = "deploying"
    StateDeployed     ServerState = "deployed"
    StateFailed       ServerState = "failed"
    StateVerifying    ServerState = "verifying"
)
```

#### ServerStatus Structure

```go
type ServerStatus struct {
    Name          string
    State         ServerState
    LastAction    ActionType
    LastUpdate    time.Time
    ErrorMessage  string
    ReadyChecks   ReadyChecks
}
```

**Persistence:** JSON file at `inventory/<env>/.status/servers.json`

### File: `internal/status/manager.go`

#### Key Methods

**`NewManager(environment string)`**
- Creates status manager for environment
- Loads existing status from disk
- Creates `.status/` directory if needed

**`ValidateServer(server *inventory.Server)`**
- Checks IP validity (regex pattern)
- Verifies SSH key file exists
- Validates port range (1-65535)
- Ensures all required fields filled

**`UpdateStatus(serverName, state, action, errorMsg)`**
- Updates server status atomically
- Saves to disk immediately
- Thread-safe with mutex locking

**`UpdateReadyChecks(serverName, checks)`**
- Stores validation results
- Auto-updates state to Ready/NotReady
- Persists checks for display

---

## 📋 Action Queue System

### File: `internal/ansible/queue.go`

#### QueuedAction Structure

```go
type QueuedAction struct {
    ID          string
    ServerName  string
    Action      ActionType
    Priority    int
    QueuedAt    time.Time
    StartedAt   *time.Time
}
```

**Persistence:** JSON file at `inventory/<env>/.queue/actions.json`

#### Key Methods

**`Add(serverName, action, priority)`**
- Generates unique UUID for action
- Appends to queue
- Sorts by priority (higher first)
- Returns action ID

**`Next()`**
- Returns first action in queue
- Marks as started (sets StartedAt)
- Does NOT remove from queue

**`Complete()`**
- Removes first action from queue
- Clears current action
- Saves updated queue to disk

**`Stop()` / `Resume()`**
- Controls queue processing
- Uses channel for stop signal
- Thread-safe

---

## ⚙️ Ansible Executor

### File: `internal/ansible/executor.go`

#### Execution Flow

```
1. Create log file: logs/<env>/<server>_<action>_<timestamp>.log
2. Build command: ansible-playbook -i <inventory> <playbook> --limit <server>
3. Set environment: ANSIBLE_STDOUT_CALLBACK=json
4. Stream stdout to: log file + progress channel
5. Wait for completion
6. Return result with success/error
```

#### Key Methods

**`RunPlaybook(playbook, serverName, progressChan)`**
- Executes Ansible with JSON callback
- Streams output to log file
- Parses JSON events for progress
- Returns ExecutionResult

**`parseProgress(line, progressChan)`**
- Parses Ansible JSON events:
  - `playbook_on_task_start`: Send task name
  - `runner_on_ok`: Send success message
  - `runner_on_failed`: Send error message
- Sends to progress channel for UI display

**`Provision(serverName, progressChan)`**
- Shortcut for `RunPlaybook("provision.yml", ...)`

**`Deploy(serverName, progressChan)`**
- Shortcut for `RunPlaybook("deploy.yml", ...)`

**`HealthCheck(ip, port)`**
- Executes: `curl -sf -m 5 http://<ip>:<port>/`
- Returns error if non-2xx status
- Used for post-deploy verification

---

## 🎭 Orchestrator

### File: `internal/ansible/orchestrator.go`

High-level coordinator between queue, executor, and status manager.

#### Key Methods

**`ValidateInventory(servers)`**
- Validates all servers
- Updates ready checks
- Called by UI on `v` key

**`QueueProvision/Deploy/Check(serverNames, priority)`**
- Adds actions to queue
- Can handle multiple servers
- Priority 0 = normal (FIFO)

**`Start(servers)`**
- Launches background goroutine
- Continuously processes queue
- Stops on Stop() call

**`processQueue(servers)`**
```go
for {
    select {
    case <-stopChan:
        return
    default:
        action := queue.Next()
        if action == nil {
            continue
        }
        executeAction(action, servers)
        queue.Complete()
    }
}
```

**`executeAction(action, servers)`**
1. Find server by name
2. Update status to "in progress" state
3. Create progress channel
4. Call executor method (Provision/Deploy/Check)
5. Update status based on result:
   - Success: Move to next state
   - Failure: Set Failed with error message
6. For Deploy: Run health check automatically
7. Close progress channel

---

## 📊 Log System

### File: `internal/logging/reader.go`

#### Key Methods

**`GetServerLogs(serverName)`**
- Globs: `logs/<env>/<serverName>_*.log`
- Returns sorted list of log files
- Used for log viewer navigation

**`ReadLog(logFile, maxLines)`**
- Reads entire log file
- Returns last N lines if maxLines > 0
- Handles large files efficiently

**`FormatLogLine(line)`**
- Adds emoji prefixes:
  - `❌` for failed/error
  - `✓` for ok/success
  - `⚡` for changed
- Preserves raw format otherwise

---

## 🖥️ Workflow UI View

### File: `internal/ui/workflow_view.go`

#### WorkflowView Structure

```go
type WorkflowView struct {
    environments    []string              // All available envs
    currentEnvIndex int                   // Active env index
    servers         []*inventory.Server   // Current env servers
    statuses        map[string]*ServerStatus
    selectedServers map[string]bool       // Checkboxes
    cursor          int                   // Table cursor position
    orchestrator    *ansible.Orchestrator
    statusMgr       *status.Manager
    logReader       *logging.Reader
    showLogs        bool                  // Log viewer active
    currentLogFile  string
    logLines        []string
    progress        map[string]string     // Real-time progress
    lastRefresh     time.Time
    autoRefresh     bool
}
```

#### Bubbletea Lifecycle

**`Init()`**
- Starts orchestrator
- Returns tick command for auto-refresh

**`Update(msg tea.Msg)`**
- Handles keyboard input
- Processes tick messages (auto-refresh)
- Routes to main view or log viewer

**`View()`**
- Renders current screen
- Switches between main view and log viewer

#### Key Methods

**`loadEnvironment()`**
1. Get environment name from index
2. Load servers from storage
3. Create status manager
4. Create orchestrator
5. Set up log reader
6. Refresh statuses

**`refreshStatuses()`**
- Gets all statuses from status manager
- Updates local status map
- Records refresh timestamp

**`handleMainKeys(msg)`**
- Navigation: up/down/cursor
- Selection: space/a
- Actions: v/p/d/c
- Logs: l
- Queue: s/x
- Environment: tab

**`renderServerTable()`**
```
Sel  Name          IP           Port  Type  Status        Progress
─────────────────────────────────────────────────────────────────
▶ ✓  web-01        10.0.1.10    3000  web   ✓ Ready      -
     web-02        10.0.1.11    3000  web   ⚡ Deploying  Task: PM2 setup
  ✓  db-01         10.0.1.20    5432  db    ✓ Deployed   -
```

**Components:**
- Cursor: `▶` on active row
- Checkbox: `✓` if selected
- Status: Formatted with icon
- Progress: Real-time from orchestrator callback

**`onProgress(serverName, message)`**
- Callback from orchestrator
- Updates progress map
- Displayed in table on next render

---

## 🔄 Data Flow

### Provision Action Flow

```
User presses 'p'
    ↓
handleMainKeys() → provisionSelected()
    ↓
orchestrator.QueueProvision(serverNames)
    ↓
queue.Add(serverName, ActionProvision, 0)
    ↓
processQueue() picks up action
    ↓
executeAction(action)
    ↓
statusMgr.UpdateStatus(server, StateProvisioning)
    ↓
executor.Provision(server, progressChan)
    ↓
ansible-playbook provision.yml --limit <server>
    ↓
parseProgress() sends updates to progressChan
    ↓
onProgress() updates progress map
    ↓
UI renders updated progress
    ↓
executor returns result
    ↓
statusMgr.UpdateStatus(server, StateProvisioned or StateFailed)
    ↓
queue.Complete()
    ↓
processQueue() picks next action
```

### Auto-Refresh Flow

```
Init() returns tickCmd()
    ↓
tea.Tick(3*time.Second) waits
    ↓
Sends tickMsg
    ↓
Update(tickMsg) → refreshStatuses()
    ↓
statusMgr.GetAllStatuses()
    ↓
Update local statuses map
    ↓
Return tickCmd() for next tick
    ↓
Repeat every 3 seconds
```

---

## 🔐 Thread Safety

### Status Manager
- Uses `sync.RWMutex`
- Read lock for GetStatus/GetAllStatuses
- Write lock for UpdateStatus/UpdateReadyChecks
- Saves to disk on every update

### Queue
- Uses `sync.RWMutex`
- Read lock for GetAll/GetCurrent/Size
- Write lock for Add/Complete/Clear
- Channel-based stop signal

### Orchestrator
- Uses `sync.RWMutex` for running flag
- Goroutine-safe queue processing
- Progress callback thread-safe (map updates in UI goroutine)

---

## 📁 File Persistence

### Status Files

**Path:** `inventory/<env>/.status/servers.json`

**Format:**
```json
{
  "web-01": {
    "name": "web-01",
    "state": "deployed",
    "last_action": "deploy",
    "last_update": "2025-11-11T19:30:15Z",
    "error_message": "",
    "ready_checks": {
      "ip_valid": true,
      "ssh_key_exists": true,
      "port_valid": true,
      "all_fields_filled": true
    }
  }
}
```

### Queue Files

**Path:** `inventory/<env>/.queue/actions.json`

**Format:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "server_name": "web-02",
    "action": "deploy",
    "priority": 0,
    "queued_at": "2025-11-11T19:28:00Z",
    "started_at": null
  }
]
```

### Log Files

**Path:** `logs/<env>/<server>_<action>_<timestamp>.log`

**Format:** Raw Ansible JSON callback output

**Example:**
```json
{"event":"playbook_on_start","uuid":"abc123"}
{"event":"playbook_on_task_start","task":"Install Node.js"}
{"event":"runner_on_ok","task":"Install Node.js","result":{"changed":true}}
{"event":"playbook_on_stats","stats":{"web-01":{"ok":15,"changed":3,"failures":0}}}
```

---

## 🎨 UI Styling

Uses `github.com/charmbracelet/lipgloss` for styling.

### Defined Styles (from `ui/styles.go`)

```go
titleStyle          // Bold, centered, bordered
selectedItemStyle   // Bright foreground for selected items
normalItemStyle     // Regular item style
helpStyle          // Dim text for help/controls
infoBoxStyle       // Bordered box for info
```

### Color Scheme

- **Selected:** Bright/bold
- **Normal:** Standard terminal colors
- **Help/Subtle:** Dimmed
- **Error:** Red (via emoji/text)
- **Success:** Green (via emoji/text)
- **Warning:** Yellow (via emoji/text)

---

## 🧪 Testing Considerations

### Unit Testing Targets

1. **Status Manager**
   - ValidateServer() with various invalid inputs
   - UpdateStatus() persistence
   - Thread safety under concurrent access

2. **Queue**
   - Priority sorting
   - FIFO ordering for same priority
   - Stop/Resume behavior
   - Persistence across restarts

3. **Executor**
   - Mock ansible-playbook execution
   - JSON parsing edge cases
   - Progress message extraction
   - Error handling

4. **Orchestrator**
   - Action queuing
   - State transitions
   - Provision prerequisite checking
   - Health check triggering

### Integration Testing

1. **End-to-End Workflow**
   - Create environment → Validate → Provision → Deploy
   - Verify status persistence
   - Check log files created
   - Confirm queue processing

2. **Multi-Environment**
   - Switch between environments
   - Verify isolated queues
   - Confirm separate status tracking

3. **Error Scenarios**
   - Failed provision (continue queue)
   - Deploy without provision (should fail)
   - Network errors during execution
   - Ansible not installed

---

## 🚀 Performance Optimizations

### Current Optimizations

1. **Concurrent Execution**
   - Multiple servers can provision/deploy simultaneously
   - Each action runs in own goroutine

2. **Efficient File I/O**
   - Status/queue saved only on changes
   - Logs written with buffered writer

3. **Smart Refresh**
   - 3-second tick during activity
   - 5-second tick when idle (configurable)
   - Manual refresh available

### Future Optimizations

1. **Incremental Log Reading**
   - Currently reads full file each time
   - Could implement tail-like streaming

2. **Status Caching**
   - Cache in-memory, sync periodically
   - Reduce disk I/O

3. **Queue Batching**
   - Group actions for same server
   - Reduce Ansible invocations

---

## 🔮 Future Enhancements

### Planned Features

1. **Action Priority**
   - UI to set priority levels
   - Manual priority boost for urgent actions

2. **Parallel Execution Limits**
   - Max concurrent actions setting
   - Resource-based throttling

3. **Rollback Support**
   - One-click rollback to previous deployment
   - Automatic rollback on failed health check

4. **Email/Webhook Notifications**
   - Alert on deployment success/failure
   - Slack/Discord integration

5. **Deployment History**
   - Track all deployments with timestamps
   - Visualize deployment timeline

6. **Live Log Streaming**
   - Real-time log updates in viewer
   - No need to close/reopen

7. **Custom Playbooks**
   - User-defined playbook actions
   - Dynamic action buttons

8. **Server Groups**
   - Group servers by role/tier
   - Batch actions on groups

---

## 📚 Dependencies

### Core Libraries

```go
require (
    github.com/charmbracelet/bubbletea v1.3.10  // TUI framework
    github.com/charmbracelet/lipgloss v0.x      // Styling
    github.com/charmbracelet/bubbles v0.21.0    // UI components
    github.com/google/uuid v1.6.0               // Action IDs
    gopkg.in/yaml.v3 v3.x                       // YAML parsing
)
```

### External Tools

- **Ansible** 2.9+ (required for playbook execution)
- **curl** (for health checks)

---

## 🐛 Known Issues & Limitations

### Current Limitations

1. **No Rollback UI**
   - Rollback playbook exists but not integrated
   - Must run manually with ansible-playbook

2. **Log Viewer Read-Only**
   - Cannot scroll through large logs
   - Limited to last 100 lines

3. **No Action Cancellation**
   - Once action starts, cannot cancel
   - Stop only prevents queue from continuing

4. **Single Queue Processor**
   - One action at a time per environment
   - Could support multiple parallel executors

### Edge Cases

1. **Rapid Environment Switching**
   - May lose progress updates
   - Status might be stale momentarily

2. **Large Log Files**
   - Reading last 100 lines still reads full file
   - Could be slow for multi-GB logs

3. **Ansible Not Installed**
   - Will fail silently
   - Should check for ansible binary

---

## 📝 Code Style Guide

### Naming Conventions

- **Types:** PascalCase (`ServerStatus`, `WorkflowView`)
- **Methods:** camelCase (`validateServer`, `loadEnvironment`)
- **Constants:** PascalCase with prefix (`StateProvisioning`)
- **Private fields:** camelCase with receiver prefix (`wv.currentEnvIndex`)

### Error Handling

```go
if err != nil {
    return nil, fmt.Errorf("descriptive context: %w", err)
}
```

### Mutex Locking

```go
m.mu.Lock()
defer m.mu.Unlock()
// Critical section
```

### Channel Usage

```go
progressChan := make(chan string, 100)  // Buffered
go func() {
    for msg := range progressChan {
        // Process
    }
}()
defer close(progressChan)
```

---

## 🔧 Configuration

### Environment Variables

Currently none required. Future:

```bash
ANSIBLE_CONFIG=/path/to/ansible.cfg
ANSIBLE_INVENTORY=/path/to/inventory
WORKFLOW_REFRESH_INTERVAL=3s
WORKFLOW_LOG_MAX_LINES=100
```

### Configuration Files

**`ansible.cfg`** (standard Ansible config)
```ini
[defaults]
host_key_checking = False
stdout_callback = json
```

---

## 📞 Troubleshooting

### Debug Mode

Add debug logging:

```go
import "log"

log.Printf("DEBUG: Status updated: %s -> %s", serverName, state)
```

### Common Issues

1. **Queue not processing**
   - Check `orchestrator.IsRunning()`
   - Verify no panic in processQueue goroutine

2. **Status not persisting**
   - Check file permissions on `inventory/<env>/.status/`
   - Verify disk space

3. **Logs not appearing**
   - Check `logs/<env>/` directory exists
   - Verify Ansible executed (check process list)

---

**Version:** 1.0  
**Last Updated:** 2025-11-11  
**Author:** Development Team
