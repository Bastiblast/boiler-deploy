# Debug Guide - Validation & Check Issues

## Problem Analysis

Two main issues were reported:
1. **Validation freeze**: Pressing 'v' to validate servers causes the app to freeze
2. **Check crash**: Pressing 'c' to check servers causes the app to crash

## Root Causes Identified

### 1. Validation Freeze
- The `validateSelectedCmd()` function executes synchronously in a tea.Cmd
- If validation takes time (file I/O, status updates), it blocks the UI thread
- No feedback is given to the user during validation

### 2. Check Crash
- The orchestrator might not be running when check is triggered
- Health check errors were not properly handled
- Missing nil checks in various places

## Fixes Applied

### Added Comprehensive Logging

All components now have detailed logging:
- **Workflow View** (`internal/ui/workflow_view.go`): User actions, selections, validation flow
- **Orchestrator** (`internal/ansible/orchestrator.go`): Queue processing, action execution
- **Queue** (`internal/ansible/queue.go`): Action queueing, completion
- **Status Manager** (`internal/status/manager.go`): Status updates, validation checks
- **Executor** (`internal/ansible/executor.go`): Health check execution with full output

### Enhanced Error Handling

1. **Check for empty selections**: Both 'v' and 'c' now check if servers are selected
2. **Auto-start orchestrator**: The 'c' command now ensures the orchestrator is running
3. **Better health check errors**: Captures curl output for debugging
4. **Null safety**: Added checks to prevent nil pointer dereferences

### Debug Logging to File

- Application now writes debug logs to `debug.log`
- Includes timestamps, microseconds, and file locations
- Logs are appended, so history is preserved

## Testing Instructions

### Manual Testing

1. **Start the application**:
   ```bash
   make run
   ```

2. **In another terminal, watch the logs**:
   ```bash
   tail -f debug.log
   ```

3. **Test validation**:
   - Navigate to "Working with Inventory"
   - Select a server with SPACE
   - Press 'v' to validate
   - Watch for log entries like:
     ```
     [WORKFLOW] Key 'v' pressed - starting validation
     [WORKFLOW] Validating 1 selected servers
     [STATUS] UpdateReadyChecks for bast-web-01: IP=true SSH=true Port=true Fields=true
     [WORKFLOW] Validation complete, sending message
     ```

4. **Test check**:
   - Select a server with SPACE
   - Press 'c' to check
   - Watch for log entries like:
     ```
     [WORKFLOW] Key 'c' pressed - starting check
     [ORCHESTRATOR] QueueCheck called with 1 servers
     [QUEUE] Adding action: check for server bast-web-01
     [ORCHESTRATOR] Processing action: check for server bast-web-01
     [EXECUTOR] Running health check: curl -sf -m 5 http://127.0.0.1:3000/
     ```

### Using Test Script

```bash
./test_workflow.sh
```

This script:
- Clears the debug log
- Shows instructions
- Runs the application
- Allows easy log review after testing

## Log Analysis

### Successful Validation Sequence

```
[WORKFLOW] Key 'v' pressed - starting validation
[WORKFLOW] Validating 1 selected servers
[WORKFLOW] Got 1 servers to validate
[WORKFLOW] Validating server 1/1: bast-web-01
[WORKFLOW] Validation checks for bast-web-01: IP=true SSH=true Port=true Fields=true
[STATUS] UpdateReadyChecks for bast-web-01: IP=true SSH=true Port=true Fields=true
[STATUS] Server bast-web-01 is ready, updating state to Ready
[STATUS] Successfully saved ready checks for bast-web-01
[WORKFLOW] Successfully updated status for bast-web-01
[WORKFLOW] Validation complete, sending message
[WORKFLOW] Received validationCompleteMsg, refreshing statuses
[WORKFLOW] Statuses refreshed
```

### Successful Check Sequence

```
[WORKFLOW] Key 'c' pressed - starting check
[WORKFLOW] Checking 1 selected servers: [bast-web-01]
[WORKFLOW] Orchestrator not running, starting it
[ORCHESTRATOR] processQueue started
[WORKFLOW] Queueing check actions for: [bast-web-01]
[ORCHESTRATOR] QueueCheck called with 1 servers: [bast-web-01]
[ORCHESTRATOR] Adding check action for server: bast-web-01
[QUEUE] Adding action: check for server bast-web-01
[QUEUE] Action added, queue size now: 1
[ORCHESTRATOR] Queue size after adding checks: 1
[WORKFLOW] Queue size after adding checks: 1
[QUEUE] Next action: check for server bast-web-01
[ORCHESTRATOR] Processing action: check for server bast-web-01
[ORCHESTRATOR] Executing check for bast-web-01 (IP: 127.0.0.1, Port: 3000)
[STATUS] Updating status for bast-web-01: state=verifying action=check error=""
[STATUS] Successfully saved status for bast-web-01
[ORCHESTRATOR] Running health check: curl http://127.0.0.1:3000/
[EXECUTOR] Running health check: curl -sf -m 5 http://127.0.0.1:3000/
[EXECUTOR] Health check failed: exit status 7, output: curl: (7) Failed to connect...
[ORCHESTRATOR] Health check FAILED for bast-web-01: curl failed...
[STATUS] Updating status for bast-web-01: state=failed action=check error="Health check failed..."
[QUEUE] Completing action: check for server bast-web-01
[QUEUE] Action completed, queue size now: 0
[ORCHESTRATOR] Completed action: check for server bast-web-01
```

### Common Error Patterns

1. **No servers selected**:
   ```
   [WORKFLOW] Key 'v' pressed - starting validation
   [WORKFLOW] Validating 0 selected servers
   [WORKFLOW] No servers selected for validation
   ```

2. **Health check connection failure** (expected when service not running):
   ```
   [EXECUTOR] Health check failed: exit status 7, output: curl: (7) Failed to connect to 127.0.0.1 port 3000
   ```

3. **SSH key not found**:
   ```
   [WORKFLOW] Validation checks for server: IP=true SSH=false Port=true Fields=true
   [STATUS] Server bast-web-01 is not ready, updating state to NotReady
   ```

## Next Steps

### If Validation Still Freezes

Check the logs for:
- Is `validateSelectedCmd()` being called?
- Is `validationCompleteMsg` being received?
- Are there errors in `UpdateReadyChecks`?
- Is file I/O hanging (check disk space, permissions)?

### If Check Still Crashes

Check the logs for:
- Where exactly does it crash? (last log entry before crash)
- Is the orchestrator started?
- Is the queue properly initialized?
- Are there nil pointer issues? (add more nil checks where crashes occur)

### Performance Issues

If the queue processing is slow:
- Consider adding a small sleep in `processQueue()` when queue is empty
- Monitor CPU usage with `top` or `htop`
- Check if file I/O (status saves, queue saves) is causing bottlenecks

## Code Improvements Needed

Based on testing results, consider:

1. **Async validation**: Run validation in background, update UI incrementally
2. **Progress indicators**: Show "Validating..." spinner during validation
3. **Better error messages**: Display errors in the UI, not just logs
4. **Orchestrator lifecycle**: Better handling of start/stop states
5. **Queue optimization**: Batch file I/O instead of saving after every action

## Debugging Commands

```bash
# Watch logs in real-time
tail -f debug.log

# Filter for specific component
grep "\[WORKFLOW\]" debug.log
grep "\[ORCHESTRATOR\]" debug.log
grep "\[QUEUE\]" debug.log
grep "\[STATUS\]" debug.log
grep "\[EXECUTOR\]" debug.log

# Find errors
grep -i "error\|failed\|panic" debug.log

# Check last 50 lines
tail -50 debug.log

# Clear log before testing
> debug.log
```

## Contact Points

If issues persist, focus debugging on:
1. `internal/ui/workflow_view.go:227-236` - validateSelectedCmd
2. `internal/ansible/orchestrator.go:88-103` - processQueue loop
3. `internal/ansible/queue.go:106-120` - Next() action retrieval
4. `internal/status/manager.go:106-132` - UpdateReadyChecks

All these now have extensive logging to trace execution flow.
