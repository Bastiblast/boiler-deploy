# Ansible Inventory Fix - 12 Nov 2025

## Problems Fixed

### 1. Ansible Warnings on Inventory
**Problem**: When running provision/deploy, Ansible showed multiple warnings:
```
[WARNING]: Skipping 'servers' as this is not a valid group definition
[WARNING]: Skipping 'mono_server' as this is not a valid group definition
[WARNING]: Skipping key (timezone) in group (config) as it is not a mapping
```

**Cause**: The `config.yml` file in each environment directory contained application-specific metadata (mono_server, mono_ip, services, etc.) that Ansible tried to parse as inventory variables.

**Solution**: Renamed `config.yml` to `.env-config.yml` (files starting with `.` are ignored by Ansible). This file contains the full Environment struct for the Go application.

**Files modified**:
- `internal/storage/yaml.go`: Changed all references from `config.yml` to `.env-config.yml`
- Migrated existing inventories: `inventory/*/config.yml` → `inventory/*/.env-config.yml`

### 2. SSH "Too Many Authentication Failures"
**Problem**: SSH connections failed with "Too many authentication failures" because ssh-agent had 7+ keys loaded.

**Solution**: Added `-o IdentitiesOnly=yes` to `ansible.cfg` ssh_args to force SSH to use only the explicitly specified key file.

**Files modified**:
- `ansible.cfg`: Updated `ssh_args` line in `[ssh_connection]` section

### 3. SSH Permission Denied (Docker Container)
**Problem**: Initial Docker container had `PermitRootLogin no` in sshd_config.

**Solution**: `test-docker-vps.sh` already correctly sets `PermitRootLogin yes` in the Dockerfile. Issue was with a stale container - fixed by rebuilding.

## Current Inventory Structure

Each environment directory (e.g., `inventory/docker/`) contains:

```
inventory/docker/
├── .env-config.yml         # Application config (hidden from Ansible)
├── .queue/                 # Queue files for the app
├── .status/                # Status files for the app  
├── hosts.yml               # Ansible inventory (servers definition)
├── group_vars/
│   └── all.yml            # Common Ansible variables
└── host_vars/
    └── docker-web-01.yml  # Per-host Ansible variables
```

### Files Used By:
- **Ansible**: `hosts.yml`, `group_vars/`, `host_vars/`
- **Go App**: `.env-config.yml`, `.queue/`, `.status/`

## Testing

All tests pass:
```bash
# No Ansible warnings
ansible-inventory -i inventory/docker --list

# SSH connectivity works
ansible all -i inventory/docker -m ping

# Provisioning works
./deploy.sh provision docker

# Deployment works  
./deploy.sh deploy docker
```

## Migration Notes

If you have old environments with `config.yml`, they need to be renamed:
```bash
cd inventory/
for env in */config.yml; do
  mv "$env" "${env%config.yml}.env-config.yml"
done
```

The Go app has been updated to automatically use `.env-config.yml`.
