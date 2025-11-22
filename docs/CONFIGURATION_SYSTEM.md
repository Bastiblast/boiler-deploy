# Configuration System Documentation

## Overview

The Inventory Manager now includes a comprehensive configuration system that allows users to customize runtime behavior, tag selection, and deployment strategies for each environment independently.

## Architecture

### Components

1. **`internal/config/types.go`** - Configuration data structures and defaults
2. **`internal/config/manager.go`** - Configuration persistence (load/save to YAML)
3. **`internal/ui/config_selector.go`** - Environment selection for configuration
4. **`internal/ui/config_form.go`** - Interactive configuration form
5. **`inventory/[env]/config.yml`** - Per-environment configuration storage

### Configuration Storage

Each environment stores its configuration in:
```
inventory/
├── docker/
│   ├── hosts.yml
│   ├── group_vars/
│   └── config.yml          ← Configuration for 'docker' environment
├── production/
│   ├── hosts.yml
│   ├── group_vars/
│   └── config.yml          ← Configuration for 'production' environment
```

## Available Options

### 1. Provisioning Options

#### Provisioning Tags
Select which Ansible tasks to execute during provisioning:
- `all` - Execute all provisioning tasks (default)
- `base` - Base system setup
- `security` - Security hardening
- `firewall` - UFW firewall configuration
- `ssh` - SSH configuration
- `nodejs` - Node.js/NVM installation
- `nginx` - Nginx web server
- `postgresql` - PostgreSQL database
- `monitoring` - Monitoring stack (Prometheus/Grafana)

#### Provisioning Strategy
- **Sequential** (default): Provision servers one at a time
- **Parallel**: Provision multiple servers simultaneously

### 2. Deployment Options

#### Deployment Tags
Select which deployment phases to execute:
- `all` - Execute all deployment steps (default)
- `dependencies` - Install dependencies only
- `build` - Build application only
- `deploy` - Deploy application files
- `restart` - Restart services only
- `health_check` - Run health checks

#### Deployment Strategy
- **Rolling** (default): Deploy to one server at a time with health checks
- **All at once**: Deploy to all servers simultaneously
- **Blue/green**: Deploy to standby servers first (requires proper setup)

### 3. Health Check Options

#### Health Check Enabled
- **True** (default): Run health checks after deployment
- **False**: Skip health checks

#### Health Check Timeout
- Default: **30 seconds**
- Range: 5-300 seconds
- Time to wait for application to respond

#### Health Check Retries
- Default: **3 retries**
- Range: 0-10
- Number of retry attempts before marking as failed

### 4. Display & Refresh Options

#### Refresh Interval
- Default: **1 second**
- Range: 1-10 seconds
- How often to refresh status during operations

#### Log Retention Lines
- Default: **100 lines**
- Range: 10-1000 lines
- Number of log lines to keep in memory for display

### 5. Retry Options

#### Auto Retry Enabled
- **False** (default): Manual retry only
- **True**: Automatically retry failed operations

#### Max Retries
- Default: **3 retries**
- Range: 0-10
- Maximum number of automatic retry attempts

## User Interface

### Accessing Configuration

1. Launch Inventory Manager
2. Select **"Configuration options"** from main menu
3. Choose environment to configure
4. Follow the three-step configuration process:

#### Step 1: General Settings
Configure numeric values, strategies, and toggles:
- Use **Tab/↑↓** to navigate between fields
- Type values in text inputs
- Press **Enter** on strategy options to cycle through choices
- Press **Enter** on checkboxes to toggle
- Press **Enter** or **n** to proceed to next step

#### Step 2: Provisioning Tags
Select which provisioning tags to use by default:
- Use **Tab/↑↓** to navigate tags
- Press **Space** to toggle individual tags
- Press **a** to select all tags
- Press **n** to select none
- Press **n** to proceed to next step

#### Step 3: Deployment Tags
Select which deployment tags to use by default:
- Same navigation as Step 2
- Press **Enter** or complete form to save configuration

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab`, `↑`, `↓` | Navigate between options |
| `Enter` | Select/toggle current option or proceed to next step |
| `Space` | Toggle tag selection (in tag selection steps) |
| `n` | Next step |
| `a` | Select all tags (in tag selection steps) |
| `Esc` | Go back to previous step or menu |
| `q`, `Ctrl+C` | Quit application |

## How It Works

### Configuration Loading

When you open "Work with your inventory" for an environment:

1. System loads configuration from `inventory/[env]/config.yml`
2. If no configuration exists, default values are used
3. Configuration affects:
   - Auto-refresh rate in the workflow view
   - Number of log lines displayed
   - Default tag selection when pressing `p` (provision) or `d` (deploy)

### Tag Pre-selection

When you press `p` or `d` to provision or deploy:

1. Tag selector opens with your configured default tags **pre-selected**
2. You can modify the selection as needed
3. Your base configuration remains unchanged (only affects this execution)
4. To change defaults permanently, use "Configuration options" menu

### Applying to Operations

Configuration values are applied as follows:

- **Provisioning/Deployment Tags**: Passed to `ansible-playbook --tags <tags>`
- **Refresh Interval**: Used for auto-refresh ticker in workflow view
- **Log Retention**: Limits realtime log buffer size
- **Health Check Settings**: Passed to health check functions (timeout, retries)
- **Strategies**: Currently informational (sequential is default behavior)

## Default Configuration

If no configuration file exists, these defaults are used:

```yaml
provisioning_tags:
  - all
provisioning_strategy: sequential
deployment_strategy: rolling
deployment_tags:
  - all
health_check_enabled: true
health_check_timeout: 30s
health_check_retries: 3
refresh_interval: 1s
log_retention_lines: 100
auto_retry_enabled: false
max_retries: 3
```

## Examples

### Example 1: Fast Development Workflow

For rapid testing, configure:
- **Provisioning Tags**: Only `base`, `nodejs`, `nginx`
- **Deployment Tags**: Only `dependencies`, `deploy`, `restart`
- **Refresh Interval**: 1 second
- **Health Check**: Disabled or short timeout

This skips security hardening, monitoring setup, and lengthy health checks.

### Example 2: Production Deployment

For careful production rollout:
- **Provisioning Tags**: `all`
- **Deployment Tags**: `all`
- **Deployment Strategy**: `rolling`
- **Refresh Interval**: 3 seconds
- **Health Check**: Enabled, 60s timeout, 5 retries
- **Auto Retry**: Disabled (manual control)

This ensures all steps run, with careful verification.

### Example 3: Security-Only Update

To update only security configurations:
- **Provisioning Tags**: Only `security`, `firewall`, `ssh`
- **Deployment**: Not needed
- Run provision action on servers

### Example 4: Quick Restart

To restart services without full deploy:
- **Deployment Tags**: Only `restart`, `health_check`
- Run deploy action

## Best Practices

1. **Environment-Specific Configs**: Create different configurations for dev, staging, and production environments

2. **Tag Usage**:
   - Use `all` for initial setups
   - Use specific tags for targeted updates
   - Combine related tags (e.g., `security,firewall,ssh` for security updates)

3. **Health Checks**:
   - Always enabled for production
   - Can be disabled for development
   - Adjust timeout based on application startup time

4. **Refresh Interval**:
   - 1 second for active monitoring during operations
   - 3-5 seconds for background operations

5. **Auto Retry**:
   - Generally keep disabled
   - Only enable if you're confident in idempotency
   - Always monitor the first retry manually

## Integration with Ansible

The configuration system integrates seamlessly with Ansible:

### Tags

Tags are passed directly to Ansible:
```bash
ansible-playbook -i inventory/[env]/hosts.yml playbooks/provision.yml \
  --limit server-01 \
  --tags "base,security,nodejs"
```

### Check Mode Support

When combined with Ansible's check mode (planned feature):
```bash
ansible-playbook --check --diff --tags <selected-tags>
```

### Idempotency

Tags allow you to safely re-run specific tasks:
- Ansible tasks are idempotent
- Re-running with same tags makes no changes if already applied
- Perfect for validation or partial updates

## Future Enhancements

Potential additions to the configuration system:

1. **Strategy Implementation**:
   - Parallel provisioning (currently sequential only)
   - Blue/green deployment automation
   
2. **Advanced Health Checks**:
   - Custom health check endpoints
   - Multi-step health validation
   
3. **Notification Settings**:
   - Email/Slack notifications on completion
   - Alert thresholds
   
4. **Rollback Configuration**:
   - Automatic rollback triggers
   - Snapshot management
   
5. **Performance Tuning**:
   - SSH multiplexing settings
   - Ansible parallelism configuration

## Troubleshooting

### Configuration Not Loading

**Problem**: Changes to configuration don't appear

**Solutions**:
- Ensure you saved the configuration (press Enter at final step)
- Check file exists: `ls inventory/[env]/config.yml`
- Verify YAML syntax: `cat inventory/[env]/config.yml`
- Restart application to reload configuration

### Tags Not Working

**Problem**: Selected tags don't affect Ansible execution

**Solutions**:
- Verify tags exist in playbook tasks
- Check Ansible output for "skipping" messages
- Tags must match exactly (case-sensitive)
- Use `all` tag to run everything

### Refresh Too Slow/Fast

**Problem**: Screen updates feel wrong

**Solutions**:
- Adjust "Refresh Interval" in configuration
- Balance between responsiveness and CPU usage
- Recommended: 1s for active ops, 3-5s for background

## Related Documentation

- [Ansible Best Practices Review](./ANSIBLE_BEST_PRACTICES_REVIEW.md) - Tag implementation details
- [Inventory Manager README](../INVENTORY_MANAGER_README.md) - General usage guide
- [Troubleshooting Guide](../TROUBLESHOOTING.md) - Common issues

---

*Last Updated: 2025-11-19*  
*Version: 1.0.0*
