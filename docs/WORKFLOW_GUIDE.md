# 📋 Ansible Workflow Guide

## Overview

The **"Work with your inventory"** feature provides a powerful interface to manage Ansible playbook execution, monitor server states, and view logs in real-time.

---

## 🚀 Quick Start

1. **Launch the application:**
   ```bash
   make run
   ```

2. **Select "Work with your inventory"** from the main menu

3. **Navigate environments** using `Tab` key

4. **Select servers** using `Space` and arrow keys

5. **Execute actions** using keyboard shortcuts

---

## 📊 Interface Layout

```
┌─────────────────────────────────────────────────────────┐
│  📋 Working with Inventory - production                 │
├─────────────────────────────────────────────────────────┤
│  [ production ] [ staging ] [ development ]              │
├─────────────────────────────────────────────────────────┤
│  Sel  Name          IP           Port  Type  Status      │
│  ──────────────────────────────────────────────────────  │
│  ▶ ✓  web-01        10.0.1.10    3000  web   ✓ Ready    │
│      web-02        10.0.1.11    3000  web   ✓ Ready    │
│    ✓  db-01         10.0.1.20    5432  db    ✓ Ready    │
├─────────────────────────────────────────────────────────┤
│  [Controls] [v] Validate [p] Provision [d] Deploy...    │
├─────────────────────────────────────────────────────────┤
│  Queue: 2 actions | Status: Running | Last: 19:30:15    │
└─────────────────────────────────────────────────────────┘
```

---

## ⌨️ Keyboard Controls

### Navigation
- `↑/k` - Move cursor up
- `↓/j` - Move cursor down
- `Tab` - Switch environment (cycles through all)
- `q/Esc` - Quit to main menu

### Selection
- `Space` - Toggle server selection
- `a` - Select/Deselect all servers

### Actions
- `v` - **Validate** selected servers (readiness check)
- `p` - **Provision** selected servers (queue provision.yml)
- `d` - **Deploy** selected servers (queue deploy.yml)
- `c` - **Check** selected servers (health check)

### Queue Management
- `s` - **Start/Stop** queue processing
- `x` - **Clear** entire queue
- `r` - **Refresh** status manually

### Logs
- `l` - View logs for server under cursor

---

## 📈 Server Status States

### Status Flow
```
Unknown → Not Ready → Ready → Provisioning → Provisioned
                                                ↓
                                            Deploying → Verifying → Deployed
                                                ↓
                                             Failed
```

### Status Indicators

| Icon | Status | Description |
|------|--------|-------------|
| `?` | Unknown | Server not validated yet |
| `✗` | Not Ready | Missing required information |
| `✓` | Ready | All checks passed, ready for provisioning |
| `⚡` | Provisioning | Running provision.yml playbook |
| `✓` | Provisioned | Server setup complete, ready for deploy |
| `⚡` | Deploying | Running deploy.yml playbook |
| `🔍` | Verifying | Running health check |
| `✓` | Deployed | Application deployed and healthy |
| `✗` | Failed | Action failed (see error message) |

---

## ✅ Readiness Validation

When you press `v`, each selected server is checked for:

1. **IP Valid**: Valid IP address format
2. **SSH Key Exists**: Private key file exists on disk
3. **Port Valid**: Port number between 1-65535
4. **All Fields Filled**: Name, IP, SSH key, Git repo, app port, Node version

**Result:** Server state changes to `Ready` or `Not Ready` with details.

---

## 🔄 Action Queue System

### How It Works

1. **FIFO (First In, First Out)**: Actions execute in order queued
2. **Multiple servers**: Can run simultaneously (async execution)
3. **Stop behavior**: Stops current action, continues with queue
4. **Priority**: Manual actions can jump queue (future feature)

### Queue Display

```
Queue: 3 actions | Status: Running | Last refresh: 19:30:15
```

- **Queue count**: Number of pending actions
- **Status**: Running/Stopped
- **Last refresh**: Time of last status update

---

## 📦 Provision vs Deploy

### Provision (`p` key)

**Playbook:** `playbooks/provision.yml`

**Actions:**
- Install system packages
- Configure UFW firewall
- Setup Fail2ban
- Create deploy user
- Configure SSH

**Requirements:**
- Server in `Ready` state

**Result:**
- Server moves to `Provisioned` state
- Server ready for application deployment

---

### Deploy (`d` key)

**Playbook:** `playbooks/deploy.yml`

**Actions:**
- Clone Git repository
- Install Node.js dependencies
- Configure PM2
- Start application
- Setup reverse proxy

**Requirements:**
- Server in `Provisioned` or `Deployed` state

**Result:**
- Server moves to `Deployed` state (after health check passes)
- Application accessible at `http://<ip>:<app_port>`

---

## 🏥 Health Checks

### Automatic Post-Deploy

After successful deployment, the system automatically:

1. Waits 5 seconds for app startup
2. Executes: `curl -sf -m 5 http://<server_ip>:<app_port>/`
3. Updates status:
   - **Success (200-299)**: State = `Deployed` ✓
   - **Failure**: State = `Failed` ✗ with error

### Manual Check (`c` key)

Use this to:
- Verify deployed application is still running
- Re-check after fixing issues
- Test connectivity

---

## 📄 Log System

### Log Files

**Location:** `logs/<environment>/<server>_<action>_<timestamp>.log`

**Example:**
```
logs/production/web-01_provision_20251111_193045.log
logs/production/web-01_deploy_20251111_194512.log
```

### Viewing Logs

1. **Navigate** to server with arrow keys
2. **Press `l`** to open log viewer
3. **View** last 100 lines with formatting:
   - `✓` - Success tasks
   - `❌` - Failed tasks
   - `⚡` - Changed tasks
4. **Press `q/Esc`** to return

### Log Format

Logs use Ansible JSON callback plugin for structured output:

```json
{"event": "playbook_on_task_start", "task": "Install Node.js"}
{"event": "runner_on_ok", "result": "success"}
```

Formatted display shows:
```
✓ Task completed: Install Node.js
⚡ Task changed: Start PM2
❌ Failed: Connection timeout
```

---

## 🌍 Multi-Environment Support

### Switching Environments

**Press `Tab`** to cycle through: `production → staging → development → ...`

### Environment Isolation

Each environment has:
- **Separate servers** and configurations
- **Independent status** tracking
- **Isolated queue** (actions don't mix)
- **Separate logs** directory

### Status Persistence

Status is saved in: `inventory/<env>/.status/servers.json`

**Persists:**
- Server state
- Last action
- Error messages
- Ready checks
- Timestamp

**Survives:** Application restarts

---

## 🔁 Auto-Refresh

### Refresh Rates

- **During execution:** 3 seconds
- **Idle state:** 5 seconds

### What Refreshes

- Server statuses
- Queue size
- Progress messages
- Last update timestamp

### Manual Refresh

Press `r` to force immediate status update.

---

## ⚠️ Error Handling

### Failed Actions

When an action fails:

1. **Status** changes to `Failed` ✗
2. **Error message** shows in status column (truncated to 20 chars)
3. **Full error** visible in logs (`l` key)
4. **Retry** available - just queue action again

### Common Issues

| Error | Cause | Solution |
|-------|-------|----------|
| Cannot parse SSH key | .pub file used | Use private key (remove .pub) |
| Connection refused | Wrong IP/port | Verify server network settings |
| Permission denied | SSH key issues | Check key permissions (600) |
| Server must be provisioned | Deploy before provision | Run provision first |
| Health check failed | App not responding | Check app logs, verify port |

---

## 💡 Best Practices

### 1. Validate First
Always run validation (`v`) before provisioning to catch configuration errors early.

### 2. Provision Before Deploy
Never try to deploy to an unprovisioned server - it will fail.

### 3. Check Logs on Failure
Press `l` immediately after failure to see detailed error information.

### 4. Use Batch Operations
Select multiple servers with `Space` + `a` to provision/deploy entire environment at once.

### 5. Monitor Queue
Keep an eye on queue size - too many actions may indicate issues.

### 6. Stop Queue If Needed
Press `s` to stop queue processing if you need to investigate issues.

---

## 🔧 Advanced Features

### Action Priority (Future)

Currently FIFO, but infrastructure supports priority levels:
- Priority `0`: Normal
- Priority `1-10`: Higher priority (executes first)

### Progress Messages

Real-time progress shown in "Progress" column:
- `Task: Install Node.js`
- `✓ Task completed`
- `✗ Failed: Connection timeout`

Updated every 3 seconds during execution.

---

## 📁 File Structure

```
boiler-deploy/
├── inventory/
│   ├── production/
│   │   ├── .status/
│   │   │   └── servers.json          # Status persistence
│   │   ├── .queue/
│   │   │   └── actions.json          # Queued actions
│   │   ├── hosts.yml                  # Ansible inventory
│   │   └── config.yml                 # Environment config
│   └── staging/
│       └── ...
├── logs/
│   ├── production/
│   │   ├── web-01_provision_*.log    # Raw Ansible logs
│   │   └── web-01_deploy_*.log
│   └── staging/
│       └── ...
└── playbooks/
    ├── provision.yml                  # Server setup
    ├── deploy.yml                     # App deployment
    ├── rollback.yml                   # Rollback deploy
    └── update.yml                     # Update app
```

---

## 🐛 Troubleshooting

### Application Crashes

**Symptom:** Panic or crash during operation

**Check:**
1. Log files for stack trace
2. Status files for corruption (`inventory/<env>/.status/`)
3. Queue files (`inventory/<env>/.queue/`)

**Fix:** Delete `.status/` and `.queue/` directories, restart app

### Queue Not Processing

**Symptom:** Actions queued but not executing

**Check:** Status line shows `Running` or `Stopped`

**Fix:** Press `s` to start queue

### Status Not Updating

**Symptom:** Status appears frozen

**Fix:**
1. Press `r` for manual refresh
2. Check auto-refresh is enabled
3. Verify `.status/servers.json` is writable

### Logs Not Appearing

**Symptom:** Press `l` but no logs shown

**Check:**
1. `logs/<env>/` directory exists
2. Server has been provisioned/deployed at least once
3. Permissions on log files

---

## 🎓 Tutorial: Complete Workflow

### Scenario: Deploy Application to New Server

**Step 1: Validate Server**
```
1. Navigate to server with ↑↓
2. Press Space to select
3. Press v to validate
4. Wait for "✓ Ready" status
```

**Step 2: Provision Server**
```
1. Press p to queue provision
2. Check "Queue: 1 actions"
3. Press s if not auto-started
4. Monitor progress in Progress column
5. Wait for "✓ Provisioned" status
```

**Step 3: Deploy Application**
```
1. Press d to queue deploy
2. Monitor deployment progress
3. Wait for "🔍 Verifying" status
4. Wait for "✓ Deployed" status
```

**Step 4: Verify Deployment**
```
1. Press c for manual health check
2. Or open browser: http://<server_ip>:<app_port>
```

**Step 5: View Logs (Optional)**
```
1. Press l to view deployment logs
2. Look for success/failure messages
3. Press q to return
```

**Total time:** ~5-10 minutes depending on server specs

---

## 📞 Support

For issues or questions:
1. Check logs with `l` key
2. Review this guide
3. Check playbook documentation in `docs/`
4. Review Ansible output in `logs/` directory

---

**Version:** 1.0  
**Last Updated:** 2025-11-11  
**Status:** Production Ready ✅
