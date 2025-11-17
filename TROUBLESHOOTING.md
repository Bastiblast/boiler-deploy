# Troubleshooting Guide - Ansible Inventory Manager

## SSH Connection Issues

### "ROOT LOGIN REFUSED" Error

**Symptom:**
- SSH test fails in Inventory Manager
- Container logs show: `ROOT LOGIN REFUSED FROM 172.17.0.1`
- Error: `Connection failed: ssh: handshake failed: ssh: unable to authenticate`

**Cause:**
The SSH server has `PermitRootLogin no` in its configuration file.

**Solution for Docker Test Container:**
```bash
# Enable root login in the running container
docker exec boiler-test-vps sed -i 's/^PermitRootLogin no/PermitRootLogin yes/' /etc/ssh/sshd_config
docker exec boiler-test-vps systemctl restart ssh

# Test the connection
ssh -i ~/.ssh/boiler_test_rsa -p 2222 -o IdentitiesOnly=yes root@127.0.0.1 "echo OK"
```

**Prevention:**
Always rebuild the container from scratch:
```bash
./tests/test-docker-vps.sh cleanup
./tests/test-docker-vps.sh setup
```

**For Production Servers:**
Edit `/etc/ssh/sshd_config` and set:
```
PermitRootLogin yes
# or for key-only:
PermitRootLogin prohibit-password
```
Then: `systemctl restart sshd`

### "Too many authentication failures"

**Symptom:**
```
Received disconnect: Too many authentication failures
```

**Cause:**
SSH agent has too many keys (>5). SSH tries each one before your specified key.

**Solution:**
```bash
# Option 1: Use IdentitiesOnly
ssh -i ~/.ssh/your_key -o IdentitiesOnly=yes -p 2222 root@server

# Option 2: Temporarily clear agent
ssh-add -D  # Remove all
ssh-add ~/.ssh/your_key  # Add specific key
```

**Note:** Inventory Manager doesn't use ssh-agent, so it avoids this issue.

### "Host key verification failed"

**Symptom:**
```
WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!
```

**Solution:**
```bash
# For localhost with custom port
ssh-keygen -f ~/.ssh/known_hosts -R '[127.0.0.1]:2222'

# For regular host
ssh-keygen -f ~/.ssh/known_hosts -R '192.168.1.100'
```

## Quick Start Testing

### Step 1: Build the Application
```bash
make build
```

### Step 2: Start Logging in Another Terminal
```bash
# Terminal 1 - Watch logs
tail -f debug.log
```

### Step 3: Run the Application
```bash
# Terminal 2 - Run app
make run
```

### Step 4: Test Actions

1. Navigate to "Working with Inventory"
2. Select environment "bast" (or your environment)
3. Use arrow keys to navigate to a server
4. Press SPACE to select the server (you should see ✓)
5. Press 'v' to validate - watch Terminal 1 for logs
6. Press 'c' to check - watch Terminal 1 for logs

## Common Issues & Solutions

### Issue 1: Application Freezes on 'v' (Validation)

**Symptoms:**
- Press 'v' and nothing happens
- UI doesn't update
- Can't press any other keys

**Debug Steps:**

1. Check if logs show the key press:
   ```bash
   grep "Key 'v' pressed" debug.log
   ```

2. Check if validation is completing:
   ```bash
   grep "Validation complete" debug.log
   ```

3. Check for errors in status updates:
   ```bash
   grep -A 5 "\[STATUS\].*error" debug.log
   ```

**Possible Causes:**

- **No servers selected**: Logs will show "No servers selected for validation"
  - Solution: Make sure to press SPACE on a server first

- **File I/O hanging**: Check disk space and permissions
  ```bash
  df -h .
  ls -la inventory/*/. status/
  ```

- **Status manager stuck**: Look for the last status update in logs
  ```bash
  grep "\[STATUS\]" debug.log | tail -20
  ```

### Issue 2: Application Crashes on 'c' (Check)

**Symptoms:**
- Press 'c' and application exits with panic
- See "runtime error" or "nil pointer dereference"

**Debug Steps:**

1. Check if orchestrator is starting:
   ```bash
   grep "Orchestrator.*running\|processQueue started" debug.log
   ```

2. Check if check action is queued:
   ```bash
   grep "QueueCheck\|Adding action: check" debug.log
   ```

3. Look for panic stack trace:
   ```bash
   grep -A 20 "panic\|runtime error" debug.log
   ```

**Possible Causes:**

- **Orchestrator not running**: Should auto-start now, check logs
  - If not starting, there may be an issue with queue initialization

- **Server not found**: Check that server exists in inventory
  ```bash
  cat inventory/bast/hosts.yml
  ```

- **Health check failure**: Expected if service isn't running
  ```bash
  # This is normal if you don't have a service on that port
  curl -sf -m 5 http://127.0.0.1:3000/
  ```

### Issue 3: Validation Shows "Not Ready" but Everything Looks Good

**Check each validation criterion:**

1. **IP Valid**: Must be a valid IPv4 address
   ```bash
   grep "IPValid" debug.log | tail -5
   ```

2. **SSH Key Exists**: File must exist at the path
   ```bash
   # Check what path is being used
   grep "SSHKeyExists" debug.log | tail -5
   
   # Verify the file exists
   cat inventory/bast/hosts.yml | grep ssh_key
   ls -la ~/.ssh/id_rsa  # or whatever path is configured
   ```

3. **Port Valid**: Must be between 1-65535
   ```bash
   grep "PortValid" debug.log | tail -5
   ```

4. **All Fields Filled**: Name, IP, SSH key path, Git repo, App port, Node version
   ```bash
   grep "AllFieldsFilled" debug.log | tail -5
   cat inventory/bast/hosts.yml  # Check all fields are present
   ```

### Issue 4: Health Check Always Fails

**This is expected if no service is running!**

The health check does:
```bash
curl -sf -m 5 http://SERVER_IP:APP_PORT/
```

**To test with a real service:**

1. Start a simple HTTP server:
   ```bash
   # In the directory where you want to serve
   python3 -m http.server 3000
   # or
   npx http-server -p 3000
   ```

2. Update your server configuration to use port 3000

3. Run check again - it should now succeed!

**Check the actual curl command in logs:**
```bash
grep "Running health check: curl" debug.log | tail -5
```

Then test it manually:
```bash
# Copy the exact command from logs and run it
curl -sf -m 5 http://127.0.0.1:3000/
echo $?  # Should be 0 for success
```

### Issue 5: Queue Not Processing

**Symptoms:**
- Actions are added to queue but never execute
- Queue size stays the same

**Debug Steps:**

1. Check if orchestrator is running:
   ```bash
   grep "processQueue started" debug.log
   ```

2. Check queue additions:
   ```bash
   grep "\[QUEUE\] Adding action" debug.log | tail -10
   ```

3. Check queue processing:
   ```bash
   grep "\[QUEUE\] Next action\|\[QUEUE\].*completed" debug.log | tail -10
   ```

4. Check for stop signals:
   ```bash
   grep "stop\|Stop" debug.log | tail -10
   ```

**Solutions:**

- Press 's' to toggle start/stop of orchestrator
- Check logs to see if it stopped unexpectedly
- Clear queue with 'x' and try again

## Log Analysis Patterns

### Healthy Validation Flow
```
[WORKFLOW] Key 'v' pressed - starting validation
[WORKFLOW] Validating 1 selected servers
[WORKFLOW] Got 1 servers to validate
[WORKFLOW] Validating server 1/1: bast-web-01
[STATUS] UpdateReadyChecks for bast-web-01: IP=true SSH=true Port=true Fields=true
[STATUS] Server bast-web-01 is ready, updating state to Ready
[WORKFLOW] Validation complete, sending message
[WORKFLOW] Received validationCompleteMsg, refreshing statuses
```

### Healthy Check Flow
```
[WORKFLOW] Key 'c' pressed - starting check
[WORKFLOW] Checking 1 selected servers: [bast-web-01]
[ORCHESTRATOR] QueueCheck called with 1 servers
[QUEUE] Adding action: check for server bast-web-01
[QUEUE] Next action: check for server bast-web-01
[ORCHESTRATOR] Processing action: check for server bast-web-01
[EXECUTOR] Running health check: curl -sf -m 5 http://127.0.0.1:3000/
[EXECUTOR] Health check failed: exit status 7, output: curl: (7) Failed...
[ORCHESTRATOR] Health check FAILED
[STATUS] Updating status for bast-web-01: state=failed
[QUEUE] Completing action
```

### Problem Patterns

**Pattern 1: Silent Failure**
```
[WORKFLOW] Key 'v' pressed
[WORKFLOW] Validating 1 selected servers
# Then nothing... = validation stuck
```

**Pattern 2: Repeated Errors**
```
[STATUS] Error saving status for bast-web-01: permission denied
[STATUS] Error saving status for bast-web-01: permission denied
# = File permission issue
```

**Pattern 3: Nil Pointer**
```
[QUEUE] Next action: check for server bast-web-01
runtime error: invalid memory address or nil pointer dereference
# = Something is nil that shouldn't be
```

## Emergency Debugging

### If Nothing in Logs

Check if log file is being created:
```bash
ls -la debug.log
```

If not, the application isn't writing logs. Check:
```bash
# Run with explicit error output
make run 2>&1 | tee run.log
```

### If Application Won't Start

```bash
# Check for build errors
make build 2>&1

# Check for missing dependencies
go mod tidy
go mod download

# Try running directly
./bin/inventory-manager
```

### If Logs are Too Verbose

Edit `cmd/inventory-manager/main.go` and comment out:
```go
// log.SetOutput(logFile)
```

Or filter logs:
```bash
# Only show errors
tail -f debug.log | grep -i error

# Only show specific component
tail -f debug.log | grep "\[WORKFLOW\]"
```

## Performance Monitoring

### Check if CPU is pegged
```bash
# While app is running
top -p $(pgrep inventory-manager)
```

### Check if file I/O is slow
```bash
# Monitor file operations
strace -e trace=file -p $(pgrep inventory-manager) 2>&1 | grep -v ENOENT
```

### Check queue responsiveness
```bash
# Queue should process actions within seconds
grep "Adding action\|Completing action" debug.log | tail -20
# Look for time gaps between add and complete
```

## Reporting Issues

If problems persist, provide:

1. **Full debug log** (at least last 100 lines):
   ```bash
   tail -100 debug.log > issue.log
   ```

2. **Your inventory configuration**:
   ```bash
   cat inventory/YOUR_ENV/hosts.yml > config.log
   ```

3. **System information**:
   ```bash
   uname -a > system.log
   go version >> system.log
   ```

4. **Exact steps to reproduce**:
   - What you clicked
   - What you expected
   - What actually happened

5. **Screenshots** of the UI state when the issue occurs

## Getting Help

1. Check `DEBUG_GUIDE.md` for detailed analysis
2. Check `BUGFIX_SUMMARY.md` for recent fixes
3. Search logs for error patterns mentioned above
4. Create a GitHub issue with logs and steps to reproduce
