# Testing the Ansible Inventory Fix

## Prerequisites
1. Docker container running: `./test-docker-vps.sh setup`
2. App built: `make build`
3. Docker environment created (already done)

## Quick Verification

### 1. Check No Ansible Warnings
```bash
ansible-inventory -i inventory/docker --list 2>&1 | grep -i warning
# Should output nothing (no warnings)
```

### 2. Test SSH Connectivity
```bash
ansible all -i inventory/docker -m ping
# Should return: docker-web-01 | SUCCESS => { "ping": "pong" }
```

### 3. Test Provision (Syntax)
```bash
ansible-playbook -i inventory/docker playbooks/provision.yml --syntax-check
# Should return: playbook: playbooks/provision.yml
```

### 4. Run Provision via App
```bash
make run
# Navigate to: Working with Inventory > Select docker > Check servers > Validate > Provision
```

### 5. Run Deploy via App
```bash
# After provision completes successfully:
# Navigate to: Working with Inventory > Select docker > Select server > Deploy
```

## Expected Results

✅ **No Ansible warnings** about skipping 'servers', 'mono_server', config keys, etc.
✅ **SSH connections work** without "too many authentication failures"
✅ **Provision runs successfully** and installs Node.js, Nginx, etc.
✅ **Deploy runs successfully** and application is accessible at http://localhost:8080

## Troubleshooting

### Container Not Responding
```bash
docker exec boiler-test-vps systemctl status ssh
# SSH should be active (running)
```

### SSH Key Issues
```bash
# Verify key exists and has correct permissions
ls -la ~/.ssh/boiler_test_rsa
chmod 600 ~/.ssh/boiler_test_rsa

# Test direct SSH
ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@127.0.0.1 'echo OK'
```

### Ansible Warnings Return
```bash
# Verify .env-config.yml exists (not config.yml)
ls -la inventory/docker/.env-config.yml

# Verify Ansible only sees proper inventory files
ls inventory/docker/*.yml
# Should show: hosts.yml (NOT config.yml)
```

## Clean Slate Test

To test everything from scratch:
```bash
# 1. Clean up
./test-docker-vps.sh cleanup
rm -rf inventory/docker

# 2. Rebuild
./test-docker-vps.sh setup
make build

# 3. Recreate environment manually or via app
# See ANSIBLE_INVENTORY_FIX.md for structure

# 4. Test
ansible all -i inventory/docker -m ping
make run
```

## Performance Notes

- **Provision**: ~5-10 minutes (installs Node.js, Nginx, UFW, Fail2ban, etc.)
- **Deploy**: ~2-3 minutes (clones repo, npm install, PM2 setup, Nginx config)
- **Check**: ~5 seconds (curl the app)
