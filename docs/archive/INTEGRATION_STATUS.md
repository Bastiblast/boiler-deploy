# Integration Status: deploy.sh Script

## ✅ Completed

### 1. Script Executor Implementation
- Created `internal/ansible/script_executor.go`
- Implements `RunAction()` to execute `./deploy.sh ACTION ENVIRONMENT`
- Streams output line-by-line to UI via channels
- Strips ANSI color codes for clean display
- Logs output to files in `logs/{environment}/`

### 2. Orchestrator Integration
- Added `ScriptExecutor` to `Orchestrator` struct
- Added `useScript` flag (enabled by default)
- Modified `executeAction()` to use ScriptExecutor for:
  - Provision (`./deploy.sh provision ENV`)
  - Deploy (`./deploy.sh deploy ENV`)
  - Check (still uses direct health check for now)

### 3. Real-time Output Display
- Added `realtimeLogs` array to `WorkflowView`
- Created `renderRealtimeLogs()` function
- Shows last 10 lines of live output at bottom of screen
- Updates via progress callback from orchestrator

### 4. Build Status
- ✅ Application compiles successfully
- ✅ No syntax errors

## ⚠️ Known Issues

### 1. Interactive Prompts
The `deploy.sh` script contains interactive prompts that block execution:
- SSH config warning (line 123-137)
- Confirmation prompts for actions (line 170-183)

**Impact**: Script hangs waiting for user input when run from app

**Solutions**:
1. **Option A**: Add `--yes` flag to deploy.sh to skip confirmations
2. **Option B**: Use direct ansible-playbook commands instead
3. **Option C**: Provide stdin input via pipe

### 2. Environment Structure Mismatch
- Old inventory structure (docker): Uses `hosts.yml` + `config.yml`
- New inventory structure (bast, test-docker): Uses `environment.json`

**Impact**: deploy.sh expects old structure, new app uses new structure

### 3. Validation Still Broken
The validation feature in workflow view doesn't work properly:
- Status stays on "Validating..." forever
- No actual validation happens

## 🔄 Next Steps

### Priority 1: Fix Script Interactivity
```bash
# Modify deploy.sh to add non-interactive mode
# Option 1: Add --yes flag
if [ "$3" = "--yes" ]; then
    AUTO_YES=true
fi

# Option 2: Detect if running in non-TTY mode
if [ ! -t 0 ]; then
    AUTO_YES=true
fi
```

### Priority 2: Test Full Workflow
1. Start app: `make run`
2. Select `docker` environment  
3. Check server (Space + C)
4. Provision server (Space + P)
5. Deploy server (Space + D)
6. Verify logs display in realtime

### Priority 3: Fallback Strategy
Keep JSON callback parser as fallback:
- If ScriptExecutor fails → use Executor
- Add toggle command to switch modes
- Log which method is being used

## 📝 Testing Commands

```bash
# Test container is running
docker ps | grep boiler-test-vps

# Test SSH access
ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost

# Test ansible directly
ansible all -i inventory/docker -m ping

# Test provision (manual)
./deploy.sh provision docker

# Test app
make run
```

## 🎯 Goal

Create seamless integration where:
1. User clicks provision/deploy in app
2. Script executes without blocking
3. Output streams to UI in real-time
4. Status updates automatically
5. Errors are captured and displayed

## 📊 Current State

```
┌─────────────────┬──────────┐
│ Feature         │ Status   │
├─────────────────┼──────────┤
│ Script Executor │ ✅ Done  │
│ Orchestrator    │ ✅ Done  │
│ UI Streaming    │ ✅ Done  │
│ Build/Compile   │ ✅ Done  │
│ Non-interactive │ ❌ TODO  │
│ Full Testing    │ ❌ TODO  │
│ Validation Fix  │ ❌ TODO  │
└─────────────────┴──────────┘
```

## 💡 Alternative Approach

If deploy.sh proves too difficult to adapt, we can:
1. Extract the ansible-playbook commands from deploy.sh
2. Run them directly via ScriptExecutor
3. Keep deploy.sh for manual CLI usage
4. App uses direct ansible commands with proper callbacks

This gives us:
- ✅ No interactive prompts
- ✅ Better control over output parsing
- ✅ Same functionality
- ✅ Keep both interfaces (CLI + TUI)
