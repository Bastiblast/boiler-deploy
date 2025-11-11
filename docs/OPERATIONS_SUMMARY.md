# 📊 Operations Workflow - Executive Summary

## 🎯 Goal

Add a complete operations workflow to the Ansible Inventory Manager that allows users to:
- **Validate** inventory configurations
- **Provision** servers (install software, configure)
- **Deploy** applications
- **Monitor** deployment status in real-time

All from within the TUI, without leaving the interface.

---

## 📋 What We're Building

### Current (v1.0)
```
Main Menu:
  ▶ Create new environment       ✅ Done
    Manage existing environment  ✅ Done
    Validate all inventories     ❌ Placeholder
    Quit
```

### Target (v2.0)
```
Main Menu:
  ▶ Create new environment              ✅ Done
    Manage existing environment         ✅ Done
    Working with your inventory         🔄 NEW FEATURE
    Quit

"Working with your inventory" opens:

╔═══════════════════════════════════════════════════════════╗
║            🎯 Operations Dashboard                         ║
╠═══════════════════════════════════════════════════════════╣
║ Environment: [production ▼]                               ║
║                                                            ║
║ Site: production                                           ║
║ ┌────────────────────────────────────────────────────┐   ║
║ │ Server         │ SSH │ Provision │ Deploy │ Status │   ║
║ ├────────────────┼─────┼───────────┼────────┼────────┤   ║
║ │ prod-web-01    │  ✓  │     ✓     │   ✓    │ Running│   ║
║ │ prod-web-02    │  ✓  │     ✓     │   ⧗    │Deploying│  ║
║ │ prod-db-01     │  ✓  │     ✗     │   -    │ Failed │   ║
║ └────────────────┴─────┴───────────┴────────┴────────┘   ║
║                                                            ║
║ [v] Validate  [p] Provision  [d] Deploy  [l] Logs        ║
╚═══════════════════════════════════════════════════════════╝
```

---

## ✨ Key Features

### 1. **Status Tracking**
- Each environment has a `status.yml` file
- Tracks SSH connectivity, provision status, deploy status
- Persists across app restarts
- Shows last execution time and errors

### 2. **Operations**
- **Validate**: Check if inventory is valid (`ansible-inventory --list`)
- **Provision**: Run `playbooks/provision.yml` (install Node.js, Nginx, etc.)
- **Deploy**: Run `playbooks/deploy.yml` (deploy your app)
- **View Logs**: See ansible output from previous runs

### 3. **Execution Modes**
- **Individual**: Run operation on one environment
- **Batch**: Queue operations across multiple environments
- **Real-time output**: Stream ansible output as it runs
- **Cancel/Pause**: Stop operations mid-execution

### 4. **Status Display**
```
SSH Status:
  ✓ connected (45ms)
  ✗ failed (connection timeout)
  - not tested

Provision Status:
  ✓ completed (2025-11-10 15:45)
  ✗ failed (see logs)
  ⧗ in progress (installing packages...)
  - not started

Deploy Status:
  ✓ deployed (v1.2.3 at 08:15)
  ✗ failed (port 3000 unreachable)
  ⧗ deploying (2/5 tasks)
  - not deployed
```

---

## 🗂️ Files & Structure

### New Files Created
```
inventory/production/
  ├── config.yml          # Existing
  ├── hosts.yml           # Existing
  └── status.yml          # NEW - deployment status

logs/production/
  ├── provision_20251111_083000.log
  ├── deploy_20251111_084500.log
  └── validate_20251111_090000.log

internal/ansible/         # NEW package
  ├── executor.go         # Run ansible-playbook
  ├── parser.go           # Parse ansible output
  ├── status.go           # Load/save status.yml
  ├── queue.go            # Queue system for batch ops
  └── validator.go        # Validate inventory

internal/ui/
  ├── operations_dashboard.go  # NEW - main operations screen
  ├── execution_view.go        # NEW - watch playbook run
  ├── queue_view.go            # NEW - batch operations
  └── log_viewer.go            # NEW - view logs
```

---

## 🔄 User Workflows

### Workflow 1: First-Time Provisioning
```
1. User creates environment "production" with 3 web servers
2. Navigate to "Working with your inventory"
3. Select "production"
4. Press [v] to validate → ✓ Inventory valid
5. Press [p] to provision → Runs playbooks/provision.yml
   - Shows real-time ansible output
   - "Installing Nginx... ✓"
   - "Configuring firewall... ✓"
   - Takes 5-10 minutes
6. Status saved: provision_status = completed
7. Press [d] to deploy → Runs playbooks/deploy.yml
   - "Cloning repository... ✓"
   - "Installing npm packages... ✓"
   - "Starting PM2... ✓"
8. Status saved: deploy_status = deployed
9. All servers show ✓ ✓ ✓ Running
```

### Workflow 2: Update Application
```
1. Navigate to "Working with your inventory"
2. Select "production"
3. Press [d] to deploy → Redeploys latest code
4. Watch progress in real-time
5. Status updated with new version
```

### Workflow 3: Debug Failed Provision
```
1. Dashboard shows prod-web-02: ✗ ✗ - Failed
2. Select prod-web-02 and press [l] for logs
3. View: "ERROR: Port 22 connection refused"
4. Fix SSH issue on server
5. Press [t] to test SSH → ✓ Connected
6. Press [p] to re-provision → Success!
```

---

## 🔧 Technical Implementation

### Ansible Execution (Go)
```go
// Run ansible-playbook
cmd := exec.Command("ansible-playbook",
    "-i", "inventory/production/hosts.yml",
    "playbooks/provision.yml",
)

// Stream output line by line
scanner := bufio.NewScanner(stdout)
for scanner.Scan() {
    line := scanner.Text()
    // Parse: "ok: [server]", "changed: [server]", "failed: [server]"
    // Update UI in real-time
    updateUI(line)
}
```

### Status Persistence (YAML)
```yaml
# inventory/production/status.yml
environment: production
last_updated: "2025-11-11T08:30:00Z"
servers:
  - name: production-web-01
    ssh:
      status: connected
      latency_ms: 45
    provision:
      status: completed
      completed: "2025-11-10T15:45:00Z"
    deploy:
      status: deployed
      version: "v1.2.3"
history:
  - type: provision
    started: "2025-11-11T08:30:00Z"
    status: success
```

---

## ❓ Questions to Answer

Before starting implementation, please answer:

### 1. Queue Behavior
- If one site fails in batch mode, continue to next or stop?
- Can user cancel individual jobs in queue, or only entire queue?

### 2. Status Display
- Auto-refresh interval? (5s, 10s, 30s)
- Show estimated time remaining for operations?

### 3. Playbooks
- Fixed 4 playbooks (provision, deploy, rollback, update)?
- Or allow custom playbooks?

### 4. Logging
- Save all ansible logs to `logs/` folder?
- Keep last N logs (10, 50, 100)?

### 5. Error Handling
- On provision failure, allow retry on specific servers?
- Show "Retry" button in UI?

### 6. Multiple Environments
- When selecting "All Environments", show merged table or separate views?
- Can user select 2-3 environments (not all) for batch operations?

---

## 📅 Timeline

- **Phase 1-2** (1.5 weeks): Data models + Ansible integration
- **Phase 3-4** (1.5 weeks): UI components
- **Phase 5** (1 week): Operations logic + queue system
- **Phase 6** (1 week): Testing & polish

**Total: ~5 weeks**

---

## 📚 Documentation Created

1. **OPERATIONS_WORKFLOW_ROADMAP.md** - Complete feature spec with mockups
2. **OPERATIONS_TECHNICAL_SPEC.md** - Code structure, data models, implementation details
3. **OPERATIONS_SUMMARY.md** - This document (executive overview)

---

## ✅ Ready to Start?

Please review the documents and answer the 6 questions above. Once confirmed, we can begin Phase 1 implementation!

**Next Step:** Create branch `operations-workflow` and start building `internal/ansible/` package.
