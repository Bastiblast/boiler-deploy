# Docker VPS Test Environment Guide

## Quick Start

```bash
# Setup complete test environment
./tests/test-docker-vps.sh setup

# The script will:
# 1. Generate SSH keys (~/.ssh/boiler_test_rsa)
# 2. Build Docker image with systemd and SSH
# 3. Start container (ports: 2222→22, 8080→80, 8443→443)
# 4. Configure SSH access
# 5. Test SSH connectivity
```

## Common Issues & Solutions

### Issue: "ROOT LOGIN REFUSED"

**Symptom:** SSH logs show `ROOT LOGIN REFUSED FROM 172.17.0.1`

**Cause:** Container has `PermitRootLogin no` in SSH config

**Solution:**
```bash
# Enable root login in existing container
docker exec boiler-test-vps sed -i 's/^PermitRootLogin no/PermitRootLogin yes/' /etc/ssh/sshd_config
docker exec boiler-test-vps systemctl restart ssh
```

**Prevention:** Always rebuild container from scratch with the current script:
```bash
./tests/test-docker-vps.sh cleanup
./tests/test-docker-vps.sh setup
```

### Issue: "Too many authentication failures"

**Symptom:** SSH error `Received disconnect: Too many authentication failures`

**Cause:** SSH agent has too many keys loaded (>5), and SSH tries them all before your specific key

**Solution:**
```bash
# Test with IdentitiesOnly option
ssh -i ~/.ssh/boiler_test_rsa -p 2222 -o IdentitiesOnly=yes root@127.0.0.1

# Or temporarily remove keys from agent
ssh-add -D  # Remove all keys from agent
ssh-add ~/.ssh/boiler_test_rsa  # Add only the test key
```

**Note:** The Go SSH library doesn't have this issue as it doesn't use ssh-agent.

### Issue: "Host key verification failed"

**Symptom:** Warning about changed host identification

**Solution:**
```bash
# Remove old host key
ssh-keygen -f ~/.ssh/known_hosts -R '[127.0.0.1]:2222'
```

## Container Management

```bash
# Check status
./tests/test-docker-vps.sh status

# View logs
./tests/test-docker-vps.sh logs

# SSH into container
./tests/test-docker-vps.sh ssh

# Execute bash in container
./tests/test-docker-vps.sh exec

# Restart container
./tests/test-docker-vps.sh restart

# Complete cleanup (remove container and image)
./tests/test-docker-vps.sh cleanup
```

## Testing with Inventory Manager

### 1. Create Environment in App

```bash
make run
```

In the app:
- Select "Create new environment"
- Name: `docker`
- Services: Enable `web`
- Mono server: `Yes`
- IP: `127.0.0.1`
- SSH Key: `~/.ssh/boiler_test_rsa`

### 2. Add Server

- Name: `docker-web-01`
- IP: `127.0.0.1` (or press Enter for mono-server IP)
- Port: `2222`
- Type: `web`
- SSH Key: `~/.ssh/boiler_test_rsa` (or press Enter for mono-server key)
- Git Repo: `https://github.com/Bastiblast/portefolio`
- App Port: `3000`
- Node Version: `20`

### 3. Test and Provision

- Select "Work with your inventory"
- Choose `docker` environment
- Press `Space` to select `docker-web-01`
- Press `c` for SSH Check (should show ✓ Connected)
- Press `v` for Validate (checks all fields)
- Press `p` for Provision (installs all dependencies)
- Press `d` for Deploy (deploys your app)

### 4. Access Deployed App

After successful deployment:
```bash
# Access via browser
http://localhost:8080
```

## Debugging

### View Container SSH Logs
```bash
docker exec boiler-test-vps journalctl -u ssh -n 50 --no-pager
```

### View Ansible Logs
```bash
# Logs are in logs/docker/ after running provision/deploy
cat logs/docker/provision_docker-web-01_*.log
cat logs/docker/deploy_docker-web-01_*.log
```

### Manual SSH Test
```bash
ssh -i ~/.ssh/boiler_test_rsa -p 2222 -o StrictHostKeyChecking=no root@127.0.0.1 "hostname && whoami"
```

## Architecture

The test container simulates a real VPS with:
- **Systemd** - Full init system (required for services)
- **SSH Server** - OpenSSH with key-based auth
- **Python 3** - Required by Ansible
- **Minimal packages** - Everything else installed by Ansible

### What Ansible Provisions

The `provision` playbook installs:
- Node.js (version specified in inventory)
- Nginx (reverse proxy)
- UFW (firewall)
- Fail2ban (intrusion prevention)
- PostgreSQL (if database server)
- PM2 (Node.js process manager)
- Certbot (SSL certificates)

### What Ansible Deploys

The `deploy` playbook:
- Clones your Git repository
- Installs npm dependencies
- Configures environment variables
- Sets up PM2 process
- Configures Nginx reverse proxy
- Starts the application

## Troubleshooting Provision/Deploy

### Provision Fails

Check logs:
```bash
cat logs/docker/provision_docker-web-01_*.log
```

Common issues:
- Network errors: Check internet connectivity in container
- Package conflicts: May need to rebuild container
- SSH issues: Verify key permissions and PermitRootLogin

### Deploy Fails

Check logs:
```bash
cat logs/docker/deploy_docker-web-01_*.log
```

Common issues:
- Git clone fails: Check repository URL and access
- npm install fails: Check Node.js version compatibility
- Build fails: Check application build requirements
- Port conflicts: Check if port 3000 is already in use

## Cleanup

```bash
# Stop and remove container (keeps image)
docker stop boiler-test-vps
docker rm boiler-test-vps

# Complete cleanup (removes everything)
./tests/test-docker-vps.sh cleanup

# Remove SSH keys (optional, manual)
rm ~/.ssh/boiler_test_rsa*

# Remove inventory (optional, manual)
rm -rf inventory/docker/
```
