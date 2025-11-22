# Ansible Invocation Analysis: Bash Script vs Go Code

## Problem Statement
The `./deploy.sh provision docker` and `./deploy.sh deploy docker` commands work perfectly, but when executed through the Go application (inventory-manager), both **provision** and **deploy** operations fail.

## Key Differences

### 1. **Command Structure**

#### Bash Script (`deploy.sh`)
```bash
# Provision
ansible-playbook playbooks/provision.yml -i "inventory/$ENVIRONMENT"

# Deploy  
ansible-playbook playbooks/deploy.yml -i "inventory/$ENVIRONMENT"
```

- **Inventory path**: Points to directory `inventory/docker/`
- **No host limiting**: Runs on ALL hosts in inventory
- **Simple invocation**: Just playbook + inventory directory

#### Go Code (Current Implementation)
```go
// In executor.go - RunPlaybook()
cmd := exec.Command("ansible-playbook",
    "-i", inventoryPath,              // inventory/docker/hosts.yml (FILE!)
    playbookPath,                     // playbooks/provision.yml
    "--limit", serverName,            // docker-web-01
)
```

- **Inventory path**: Points to FILE `inventory/docker/hosts.yml`
- **Host limiting**: Uses `--limit docker-web-01`
- **Environment vars**: Sets `ANSIBLE_STDOUT_CALLBACK=json`

### 2. **Inventory Path Resolution**

#### Why This Matters

When Ansible receives a **directory** as inventory:
```bash
ansible-playbook playbooks/provision.yml -i inventory/docker/
```

It automatically loads:
1. `inventory/docker/hosts.yml` - Host definitions
2. `inventory/docker/group_vars/all.yml` - Group-level variables
3. `inventory/docker/host_vars/{hostname}.yml` - Host-specific variables

When Ansible receives a **file** as inventory:
```bash
ansible-playbook playbooks/provision.yml -i inventory/docker/hosts.yml
```

It ONLY loads:
1. `inventory/docker/hosts.yml` - Host definitions
2. **IGNORES** `group_vars/` and `host_vars/`

### 3. **Critical Missing Variables**

Our playbooks depend on variables defined in `group_vars/all.yml`:

```yaml
# group_vars/all.yml (NOT loaded by Go code)
deploy_user: root
nodejs_version: "20"
app_name: docker
app_repo: https://github.com/Bastiblast/portefolio
app_branch: main
app_dir: /var/www/docker
pm2_app_name: "{{ app_name }}"
enable_firewall: false
```

And host-specific overrides in `host_vars/docker-web-01.yml`:
```yaml
app_port: 3000
app_repo: https://github.com/Bastiblast/portefolio  
app_branch: main
nodejs_version: "20"
```

**Without these variables, Ansible playbooks fail** because:
- Roles reference undefined variables (e.g., `{{ app_repo }}`, `{{ nodejs_version }}`)
- Conditional tasks fail
- Deployment paths are incorrect

### 4. **Host Limiting Behavior**

#### Bash Script
- No `--limit` flag
- Runs on ALL webservers defined in inventory
- Natural group targeting via playbook

#### Go Code
- Uses `--limit docker-web-01`
- Attempts to target specific server
- BUT combined with file-based inventory = missing variables

### 5. **Environment Variables**

#### Go Code Sets:
```go
cmd.Env = append(os.Environ(), 
    "ANSIBLE_STDOUT_CALLBACK=json",
    "ANSIBLE_FORCE_COLOR=false"
)
```

The `json` callback might not be compatible with all playbooks or could cause parsing issues.

## Root Cause Analysis

### Why Provision & Deploy Fail in Go

**Primary Issue**: Inventory path points to **file** instead of **directory**

```go
// executor.go:48 - THE PROBLEM
inventoryPath := filepath.Join("inventory", e.environment, "hosts.yml")
//                                                            ^^^^^^^^^
//                                                            FILE, not DIR!
```

**Result**:
1. ✗ `group_vars/all.yml` is NOT loaded
2. ✗ `host_vars/{hostname}.yml` is NOT loaded  
3. ✗ Playbooks fail with undefined variable errors
4. ✗ Roles cannot find required configuration

**Secondary Issue**: Possible conflicts with JSON callback

The JSON callback is designed for machine parsing, but:
- May not be supported by all Ansible versions
- Could cause issues with certain task outputs
- Might interfere with progress parsing

## Solution Options

### Option 1: Change Inventory Path to Directory (RECOMMENDED)

**Modify `executor.go`:**
```go
// Change line 48 from:
inventoryPath := filepath.Join("inventory", e.environment, "hosts.yml")

// To:
inventoryPath := filepath.Join("inventory", e.environment)
```

**Pros**:
- ✓ Matches bash script behavior exactly
- ✓ Loads all variable files automatically
- ✓ Minimal code change
- ✓ Maintains consistency

**Cons**:
- None

### Option 2: Keep Using deploy.sh Script (CURRENT FALLBACK)

**Status**: Already implemented via `ScriptExecutor` + `useScript = true`

**Pros**:
- ✓ Works perfectly
- ✓ Proven solution
- ✓ Streams output line-by-line
- ✓ Handles all edge cases

**Cons**:
- Depends on external script
- Harder to add Go-native features later

### Option 3: Manually Pass Extra Variables (NOT RECOMMENDED)

Load and pass variables explicitly:
```go
cmd := exec.Command("ansible-playbook",
    "-i", inventoryPath,
    "-e", "@inventory/docker/group_vars/all.yml",
    "-e", "@inventory/docker/host_vars/docker-web-01.yml",
    playbookPath,
)
```

**Pros**:
- Keeps file-based inventory

**Cons**:
- ✗ Complex and error-prone
- ✗ Doesn't scale with multiple hosts
- ✗ Variable precedence issues
- ✗ Still doesn't match bash behavior

## Recommended Fix

### Implement Option 1 + Keep Fallback

1. **Fix `executor.go`** to use directory-based inventory:
   ```go
   inventoryPath := filepath.Join("inventory", e.environment)
   ```

2. **Remove `--limit` flag** for provision (use playbook's natural targeting):
   ```go
   // For provision.yml, let playbook handle targeting
   cmd := exec.Command("ansible-playbook",
       "-i", inventoryPath,
       playbookPath,
       // No --limit needed
   )
   ```

3. **Keep `ScriptExecutor` as primary** until Go executor is validated:
   ```go
   useScript: true  // Keep this default
   ```

4. **Add toggle for testing** Go executor:
   ```go
   // Add environment variable or config flag
   if os.Getenv("USE_NATIVE_ANSIBLE") == "true" {
       o.useScript = false
   }
   ```

## Testing Plan

1. Start test container:
   ```bash
   ./tests/test-docker-vps.sh setup
   docker start boiler-test-vps
   ```

2. Test with script executor (should work):
   ```bash
   make run
   # Select docker environment
   # Check server
   # Press 'p' for provision
   ```

3. Apply fix to `executor.go`

4. Test with native Go executor:
   ```bash
   USE_NATIVE_ANSIBLE=true make run
   # Repeat provision test
   ```

5. Compare outputs and logs

## Verification Commands

### Check Inventory Loading:
```bash
# What bash script sees
ansible-inventory -i inventory/docker --list

# What Go code currently sees
ansible-inventory -i inventory/docker/hosts.yml --list
```

### Check Variable Resolution:
```bash
# With directory (CORRECT)
ansible-playbook playbooks/provision.yml -i inventory/docker --list-tasks

# With file (WRONG)
ansible-playbook playbooks/provision.yml -i inventory/docker/hosts.yml --list-tasks
```

## Conclusion

The root cause is clear: **File-based inventory path prevents Ansible from loading `group_vars/` and `host_vars/` directories**, which contain critical configuration variables required by our playbooks.

The fix is simple: **Use directory path instead of file path**, exactly as the working bash script does.

Current fallback (ScriptExecutor) works perfectly and should remain the default until the native Go executor is validated with the fix.
