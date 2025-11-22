# 🚀 Operations Workflow Roadmap

## 📋 Current State (v1.0)

✅ **Completed Features:**
- Interactive TUI with Bubbletea
- Environment creation and management
- Server CRUD operations (Create, Read, Update, Delete)
- SSH connection testing
- Auto-generation of `hosts.yml` and `group_vars/`
- Mono-server deployment support (single IP for all servers)
- Shared SSH key option for all servers
- Input validation (IP, ports, SSH keys)
- Environment deletion

## 🎯 Next Phase: Operations Workflow (v2.0)

### Overview
Transform "Validate all inventories" into a comprehensive **"Working with Your Inventory"** section that allows users to:
1. Validate inventory configurations
2. Provision sites (server setup)
3. Deploy applications
4. Monitor deployment status in real-time

---

## 📐 Detailed Plan

### 1. **Rename and Restructure Main Menu Option**

**Change:**
```
"Validate all inventories" → "Working with Your Inventory"
```

**New Main Menu:**
```
🔧 Ansible Inventory Manager

📁 Existing environments:
   • production (3 servers)
   • staging (2 servers)

  ▶ Create new environment
    Manage existing environment
    Working with your inventory    ← NEW NAME
    Quit
```

---

### 2. **New Screen: Operations Dashboard**

When user selects "Working with your inventory":

```
╔══════════════════════════════════════════════════════════════════════╗
║                    🎯 Operations Dashboard                           ║
╚══════════════════════════════════════════════════════════════════════╝

Environment: [production ▼]    Filter: [All ▼]    Refresh: Auto

┌────────────────────────────────────────────────────────────────────┐
│ Site Name      │ Servers │ SSH │ Provision │ Deploy │ Status      │
├────────────────┼─────────┼─────┼───────────┼────────┼─────────────┤
│ production     │   3     │  ✓  │    ✓      │   ✓    │ ● Running   │
│ staging        │   2     │  ✓  │    ✓      │   ⧗    │ ⧗ Deploying │
│ development    │   1     │  ✓  │    ✗      │   -    │ ✗ Failed    │
└────────────────┴─────────┴─────┴───────────┴────────┴─────────────┘

[v] Validate Selected  [p] Provision  [d] Deploy  [l] View Logs
[a] All Environments   [q] Back
```

**Status Columns:**
- **SSH**: SSH connectivity (✓ connected, ✗ failed, - not tested)
- **Provision**: Server provisioning status (✓ done, ✗ failed, ⧗ in progress, - not started)
- **Deploy**: Application deployment (✓ deployed, ✗ failed, ⧗ deploying, - not deployed)
- **Status**: Overall site status (● Running, ⧗ Deploying, ✗ Failed, ○ Stopped)

---

### 3. **Status Persistence**

**File Structure:**
```
inventory/
└── production/
    ├── config.yml          # Environment config
    ├── hosts.yml           # Ansible inventory
    └── status.yml          # NEW: Deployment status
```

**status.yml Structure:**
```yaml
environment: production
last_updated: "2025-11-11T08:30:00Z"
servers:
  - name: production-web-01
    ssh_status: connected
    ssh_last_check: "2025-11-11T08:25:00Z"
    ssh_latency_ms: 45
    provision_status: completed
    provision_date: "2025-11-10T15:30:00Z"
    deploy_status: deployed
    deploy_date: "2025-11-11T08:15:00Z"
    deploy_version: "v1.2.3"
    last_error: null
  - name: production-web-02
    ssh_status: connected
    ssh_last_check: "2025-11-11T08:25:00Z"
    ssh_latency_ms: 52
    provision_status: in_progress
    provision_date: "2025-11-11T08:20:00Z"
    deploy_status: not_deployed
    deploy_date: null
    deploy_version: null
    last_error: "Connection timeout on port 3000"

playbook_history:
  - type: provision
    playbook: playbooks/provision.yml
    started: "2025-11-11T08:20:00Z"
    completed: "2025-11-11T08:28:00Z"
    status: success
    servers: [production-web-01, production-web-02]
  - type: deploy
    playbook: playbooks/deploy.yml
    started: "2025-11-11T08:15:00Z"
    completed: "2025-11-11T08:16:30Z"
    status: success
    servers: [production-web-01]
```

---

### 4. **Operation Modes**

#### Mode A: All Sites Together (Queue System)

**Behavior:**
- User selects multiple sites
- Actions are queued and executed sequentially
- Shows progress for current operation
- Displays queue status

```
╔══════════════════════════════════════════════════════════════════════╗
║                      Deployment Queue                                ║
╚══════════════════════════════════════════════════════════════════════╝

Current: Provisioning 'production' (2/3 servers complete)
[████████████████░░░░░░░░] 67% - Installing Node.js on web-03...

Queue:
  1. ⧗ Provision 'production'    (in progress)
  2. ⏸ Deploy 'production'        (waiting)
  3. ⏸ Provision 'staging'        (waiting)

[x] Cancel Queue  [p] Pause  [l] View Logs
```

#### Mode B: Individual Site Operations

**Behavior:**
- User selects one site
- Can run validation, provision, or deploy independently
- Real-time ansible output display

```
╔══════════════════════════════════════════════════════════════════════╗
║            Provisioning: production                                  ║
╚══════════════════════════════════════════════════════════════════════╝

Server: production-web-01 (192.168.1.10)
Task: Installing Nginx...

Ansible Output:
─────────────────────────────────────────────────────────────────────
TASK [nginx : Install Nginx] ********************************************
ok: [production-web-01]

TASK [nginx : Configure Nginx] ******************************************
changed: [production-web-01]
─────────────────────────────────────────────────────────────────────

Progress: [████████████████████░░░░] 80% (4/5 tasks)

[x] Cancel  [↓] Scroll Down  [↑] Scroll Up
```

---

### 5. **Ansible Execution Integration**

**Go Package: `internal/ansible/`**

```
internal/ansible/
├── executor.go        # Ansible playbook executor
├── parser.go          # Parse ansible output
├── status.go          # Status tracking
└── queue.go           # Queue management
```

**Key Functions:**
```go
// executor.go
func RunPlaybook(env, playbook string, limit []string) (*Execution, error)
func GetPlaybookPath(name string) string
func ValidateInventory(env string) error

// status.go
func LoadStatus(env string) (*Status, error)
func SaveStatus(env string, status *Status) error
func UpdateServerStatus(env, server string, update StatusUpdate) error

// queue.go
func NewQueue() *Queue
func (q *Queue) Add(job Job) error
func (q *Queue) Start() error
func (q *Queue) Cancel() error
func (q *Queue) GetStatus() QueueStatus
```

---

### 6. **Implementation Steps**

#### **Phase 1: Data Models & Storage (Week 1)**
- [ ] Create `internal/ansible/` package
- [ ] Define status data structures
- [ ] Implement `status.yml` read/write
- [ ] Add status persistence to storage layer

#### **Phase 2: Ansible Integration (Week 1-2)**
- [ ] Implement `executor.go` - run ansible-playbook commands
- [ ] Implement `parser.go` - parse ansible stdout in real-time
- [ ] Test with existing playbooks (provision.yml, deploy.yml)
- [ ] Handle errors and timeouts

#### **Phase 3: UI - Operations Dashboard (Week 2)**
- [ ] Create `operations_dashboard.go`
- [ ] Display environments in table format
- [ ] Show status columns (SSH, Provision, Deploy)
- [ ] Implement environment selector dropdown
- [ ] Add refresh functionality (manual + auto)

#### **Phase 4: UI - Execution Views (Week 2-3)**
- [ ] Create `execution_view.go` for individual operations
- [ ] Create `queue_view.go` for batch operations
- [ ] Real-time ansible output display (scrollable)
- [ ] Progress bars and status indicators
- [ ] Cancel/pause functionality

#### **Phase 5: Operations Logic (Week 3)**
- [ ] Implement validate inventory action
- [ ] Implement provision action (calls playbooks/provision.yml)
- [ ] Implement deploy action (calls playbooks/deploy.yml)
- [ ] Queue system for batch operations
- [ ] Status updates during execution

#### **Phase 6: Testing & Polish (Week 4)**
- [ ] Test all operations with real servers
- [ ] Error handling and recovery
- [ ] Log viewer implementation
- [ ] Documentation updates
- [ ] Performance optimization

---

### 7. **User Workflows**

#### **Workflow A: Provision New Site**
1. Create environment + add servers
2. Navigate to "Working with your inventory"
3. Select environment (e.g., production)
4. Press `v` to validate inventory → Shows validation results
5. Press `p` to provision → Runs `playbooks/provision.yml`
6. Monitor progress in real-time
7. Status saved to `inventory/production/status.yml`

#### **Workflow B: Deploy Application**
1. Navigate to "Working with your inventory"
2. Select environment
3. Verify SSH status is ✓ and Provision status is ✓
4. Press `d` to deploy → Runs `playbooks/deploy.yml`
5. Monitor deployment progress
6. Status updates automatically

#### **Workflow C: Batch Provisioning**
1. Navigate to "Working with your inventory"
2. Press `a` to select all environments
3. Press `p` to provision all
4. Operations queued: production → staging → development
5. Monitor queue progress
6. Review results for each environment

---

### 8. **Technical Considerations**

#### **Ansible Execution**
```go
// Using Go's exec package
cmd := exec.Command("ansible-playbook",
    "-i", fmt.Sprintf("inventory/%s/hosts.yml", env),
    "--limit", strings.Join(servers, ","),
    playbookPath,
)
cmd.Stdout = parser  // Parse in real-time
cmd.Stderr = errorParser
```

#### **Real-time Output Parsing**
```go
// Stream ansible output line by line
scanner := bufio.NewScanner(stdout)
for scanner.Scan() {
    line := scanner.Text()
    // Parse TASK, PLAY, ok, changed, failed
    update := parseAnsibleLine(line)
    statusChan <- update
}
```

#### **Status Persistence**
- Save status after each task completion
- Atomic writes to prevent corruption
- Lock file mechanism for concurrent access

---

### 9. **Questions for Clarification**

1. **Queue Behavior:**
   - Should queued operations be cancellable individually, or only the entire queue?
   - Should failed operations stop the queue, or continue to next?

2. **Status Display:**
   - Auto-refresh interval (every 5s, 10s, 30s)?
   - Should we show estimated time remaining?

3. **Playbook Selection:**
   - Are the 4 playbooks (provision, deploy, rollback, update) fixed, or should users be able to add custom playbooks?
   - Should we show playbook descriptions in the UI?

4. **Logging:**
   - Should ansible logs be saved to files (e.g., `logs/production_provision_20251111.log`)?
   - Log retention policy (keep last N logs)?

5. **Error Handling:**
   - On provision failure, should we allow retry on specific servers?
   - Should we support rollback from the UI?

6. **Multiple Environments:**
   - When showing "All Environments", should it be a separate view or merged table?
   - Can user select multiple (but not all) environments for batch operations?

---

### 10. **Success Metrics**

**v2.0 will be successful when:**
- ✅ User can provision a new site without leaving the TUI
- ✅ User can deploy applications to all servers in one action
- ✅ User can see real-time progress of ansible operations
- ✅ Status is persisted and survives application restart
- ✅ User can manage multiple environments efficiently
- ✅ Logs are accessible for debugging

---

## 📅 Estimated Timeline

- **Phase 1-2:** 1.5 weeks (Data models + Ansible integration)
- **Phase 3-4:** 1.5 weeks (UI components)
- **Phase 5:** 1 week (Operations logic)
- **Phase 6:** 1 week (Testing & polish)

**Total: ~5 weeks for v2.0**

---

## 🔄 Future Enhancements (v3.0+)

- Real-time server metrics (CPU, RAM, disk)
- Deployment history visualization (timeline)
- One-click rollback to previous version
- Multi-user support (lock mechanism)
- Webhook notifications (Slack, Discord)
- SSH terminal multiplexer (execute commands on multiple servers)
- Configuration drift detection
- Backup/restore functionality

---

## 📝 Notes

- All UI interactions remain keyboard-driven (no mouse)
- Maintain the clean, minimalist aesthetic of v1.0
- Performance: operations should feel instant (< 100ms for UI updates)
- Keep binary size small (target < 10MB)

---

**Last Updated:** 2025-11-11
**Status:** Planning Phase
**Next Step:** Answer clarification questions
