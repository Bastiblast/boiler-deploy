# Setup 3 Docker Servers for Testing

**Script:** `tests/setup-3-servers.sh`

Automated setup script based on `test-docker-vps.sh` methodology. Creates 3 SSH-ready Docker containers for full deployment testing.

---

## 🎯 Features

### Automated Setup
- ✅ Cleans existing containers
- ✅ Creates 3 Ubuntu 22.04 containers
- ✅ Installs all necessary tools (SSH, Git, Nginx, Python, build-tools)
- ✅ Configures SSH with key + password auth
- ✅ Generates Ansible inventory
- ✅ Validates all connections
- ✅ Production-like environment

### Container Configuration

**test-web-01:**
- SSH: `localhost:2222`
- HTTP: `localhost:8080`
- HTTPS: `localhost:8443`
- APP: `localhost:3000`

**test-web-02:**
- SSH: `localhost:2223`
- HTTP: `localhost:8081`
- HTTPS: `localhost:8444`
- APP: `localhost:3001`

**test-web-03:**
- SSH: `localhost:2224`
- HTTP: `localhost:8082`
- HTTPS: `localhost:8445`
- APP: `localhost:3002`

---

## 🚀 Quick Start

### 1. Run Setup Script

```bash
cd /home/basthook/devIronMenth/boiler-deploy
./tests/setup-3-servers.sh
```

**Duration:** ~2-3 minutes

**Output:**
```
╔══════════════════════════════════════════════════════════════╗
║    Docker Test Environment Setup (3 Servers)                ║
╚══════════════════════════════════════════════════════════════╝

[INFO] Cleaning up existing containers...
[✓] Cleanup completed
[INFO] Setting up SSH keys...
[✓] SSH key generated: /home/basthook/.ssh/boiler_test_rsa
[INFO] Creating 3 containers (this may take 1-2 minutes)...
[✓] Container test-web-01 created successfully
[✓] Container test-web-02 created successfully
[✓] Container test-web-03 created successfully
[INFO] Waiting for containers to initialize (30 seconds)...
[✓] SSH key authentication working for test-web-01
[✓] SSH key authentication working for test-web-02
[✓] SSH key authentication working for test-web-03
[✓] Inventory generated: inventory/docker/hosts.yml

╔══════════════════════════════════════════════════════════════╗
║         3 Docker Servers Setup - COMPLETE                   ║
╚══════════════════════════════════════════════════════════════╝
```

---

## 📋 What Gets Installed

### Base Packages
- `openssh-server` - SSH daemon
- `sudo` - Privilege escalation
- `curl`, `wget` - HTTP tools
- `git` - Version control
- `python3`, `python3-pip` - Python runtime
- `build-essential` - Compilation tools
- `nginx` - Web server
- `net-tools`, `iputils-ping` - Network utilities
- `vim`, `htop` - Admin tools
- `ca-certificates` - SSL certificates

### Configuration
- SSH root login enabled
- SSH key authentication configured
- Password authentication enabled (root:root)
- Deploy user created (deploy:deploy)
- Proper permissions on ~/.ssh
- SSHD running on port 22 (mapped to host)

---

## 🔐 SSH Access

### Method 1: SSH Key (Recommended)

```bash
# test-web-01
ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost

# test-web-02
ssh -i ~/.ssh/boiler_test_rsa -p 2223 root@localhost

# test-web-03
ssh -i ~/.ssh/boiler_test_rsa -p 2224 root@localhost
```

### Method 2: SSH Password

```bash
# Password: root
ssh -p 2222 root@localhost  # test-web-01
ssh -p 2223 root@localhost  # test-web-02
ssh -p 2224 root@localhost  # test-web-03
```

### Method 3: Docker Exec (Direct)

```bash
docker exec -it test-web-01 bash
docker exec -it test-web-02 bash
docker exec -it test-web-03 bash
```

---

## 📦 Generated Files

### SSH Keys

**Location:** `~/.ssh/boiler_test_rsa`

**Files:**
- `boiler_test_rsa` - Private key (600)
- `boiler_test_rsa.pub` - Public key (644)

**Used by:** Ansible inventory, SSH connections

### Ansible Inventory

**Location:** `inventory/docker/hosts.yml`

**Content:**
```yaml
all:
  children:
    webservers:
      hosts:
        test-web-01:
          ansible_host: 127.0.0.1
          ansible_user: root
          ansible_port: 2222
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: /home/basthook/.ssh/boiler_test_rsa
          ansible_become: true
          app_port: 3000
        test-web-02:
          # ... same pattern
        test-web-03:
          # ... same pattern
```

**Backup:** Original backed up to `hosts.yml.backup.YYYYMMDD_HHMMSS`

---

## 🧪 Testing Workflow

### 1. Verify Containers

```bash
docker ps --filter "name=test-web"
```

**Expected:**
```
NAMES         STATUS        PORTS
test-web-03   Up X minutes  0.0.0.0:2224->22/tcp, ...
test-web-02   Up X minutes  0.0.0.0:2223->22/tcp, ...
test-web-01   Up X minutes  0.0.0.0:2222->22/tcp, ...
```

### 2. Test SSH Connections

```bash
# Quick test all 3
for port in 2222 2223 2224; do
  echo "Testing port $port..."
  ssh -i ~/.ssh/boiler_test_rsa -o StrictHostKeyChecking=no -p $port root@localhost "hostname"
done
```

**Expected:**
```
Testing port 2222...
test-web-01
Testing port 2223...
test-web-02
Testing port 2224...
test-web-03
```

### 3. Run Inventory Manager

```bash
./bin/inventory-manager
```

**Expected:**
- Environment "docker" detected
- 3 servers listed (test-web-01/02/03)
- Status validation runs at startup
- All servers show "ready" or "provisioned"

### 4. Provision Servers

**Via Inventory Manager UI:**
1. Select "docker" environment
2. Select server(s)
3. Choose "Provision"
4. Monitor progress

**Via Ansible directly:**
```bash
ansible-playbook -i inventory/docker/hosts.yml playbooks/provision.yml
```

### 5. Deploy Application

**Via Inventory Manager:**
1. Select provisioned server
2. Choose "Deploy"
3. Monitor deployment

**Via Ansible:**
```bash
ansible-playbook -i inventory/docker/hosts.yml playbooks/deploy.yml
```

### 6. Verify Deployment

```bash
# Check PM2 status
ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost "pm2 list"

# Check app response
curl http://localhost:3000

# Check all 3 apps
for port in 3000 3001 3002; do
  echo "Testing app on port $port..."
  curl -s http://localhost:$port | head -1
done
```

---

## 🔧 Maintenance

### View Container Logs

```bash
docker logs test-web-01
docker logs test-web-02
docker logs test-web-03
```

### Restart Container

```bash
docker restart test-web-01
```

### Stop All Containers

```bash
docker stop test-web-01 test-web-02 test-web-03
```

### Remove All Containers

```bash
docker rm -f test-web-01 test-web-02 test-web-03
```

### Re-run Setup

```bash
# Clean and recreate
./tests/setup-3-servers.sh
```

---

## 🐛 Troubleshooting

### SSH Connection Refused

**Symptom:** `Connection refused` on SSH port

**Diagnosis:**
```bash
# Check if container running
docker ps --filter "name=test-web-01"

# Check SSHD status
docker exec test-web-01 ps aux | grep sshd

# Check SSH logs
docker logs test-web-01 | grep sshd
```

**Fix:**
```bash
# Restart SSHD
docker exec test-web-01 /usr/sbin/sshd

# Or restart container
docker restart test-web-01
```

### SSH Key Permission Denied

**Symptom:** `Permission denied (publickey)`

**Diagnosis:**
```bash
# Check key permissions
ls -la ~/.ssh/boiler_test_rsa*

# Check authorized_keys in container
docker exec test-web-01 ls -la /root/.ssh/
```

**Fix:**
```bash
# Fix local permissions
chmod 600 ~/.ssh/boiler_test_rsa
chmod 644 ~/.ssh/boiler_test_rsa.pub

# Reinstall key in container
cat ~/.ssh/boiler_test_rsa.pub | docker exec -i test-web-01 bash -c "cat > /root/.ssh/authorized_keys && chmod 600 /root/.ssh/authorized_keys"
```

### Port Already in Use

**Symptom:** `Bind for 0.0.0.0:2222 failed: port is already allocated`

**Diagnosis:**
```bash
# Check what's using the port
sudo lsof -i :2222
```

**Fix:**
```bash
# Stop conflicting service/container
docker rm -f <conflicting_container>

# Or modify ports in script
# Edit BASE_SSH_PORT in setup-3-servers.sh
```

### Container Exits Immediately

**Symptom:** Container created but not running

**Diagnosis:**
```bash
# Check exit status
docker ps -a --filter "name=test-web-01"

# View logs
docker logs test-web-01
```

**Fix:**
```bash
# Usually apt-get update failed (network issue)
# Remove and retry
docker rm test-web-01
./tests/setup-3-servers.sh
```

---

## 📊 Resource Usage

**Per Container:**
- CPU: ~0.5% idle, ~5% during provision
- RAM: ~50MB idle, ~200MB during build
- Disk: ~500MB (base + packages)

**Total (3 containers):**
- RAM: ~600MB
- Disk: ~1.5GB

---

## 🎯 Use Cases

### 1. Development Testing
- Test provisioning playbooks
- Validate deployment scripts
- Debug SSH issues
- Test multi-server orchestration

### 2. Integration Testing
- CI/CD pipeline validation
- Automated test runs
- Regression testing
- Performance benchmarks

### 3. Training/Demos
- Safe environment for learning
- Reproducible setup
- No cloud costs
- Fast iteration

### 4. Debugging
- Isolate issues
- Test fixes
- Verify rollbacks
- State inspection

---

## 🔄 Cleanup

### Remove Containers Only

```bash
docker rm -f test-web-01 test-web-02 test-web-03
```

### Full Cleanup (Including Keys)

```bash
# Remove containers
docker rm -f test-web-01 test-web-02 test-web-03

# Remove SSH keys
rm ~/.ssh/boiler_test_rsa*

# Remove inventory backup
rm inventory/docker/hosts.yml.backup.*
```

---

## 📝 Script Configuration

### Customization

Edit variables in `setup-3-servers.sh`:

```bash
# Change container prefix
CONTAINER_PREFIX="my-server"

# Change SSH key name
SSH_KEY_NAME="my_test_key"

# Change base ports
BASE_SSH_PORT=3000
BASE_HTTP_PORT=9000
BASE_HTTPS_PORT=9100
BASE_APP_PORT=4000

# Change Ubuntu version
UBUNTU_VERSION="20.04"
```

### Add More Containers

Modify the loops:

```bash
# Change from 1 2 3 to 1 2 3 4 5
for i in 1 2 3 4 5; do
  create_container $i
done
```

---

## ✅ Validation Checklist

After running script, verify:

- [ ] 3 containers running: `docker ps`
- [ ] SSH ports open: `nc -zv localhost 2222`
- [ ] SSH key auth works: `ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost`
- [ ] Inventory generated: `cat inventory/docker/hosts.yml`
- [ ] Containers have tools: `docker exec test-web-01 which curl git nginx`
- [ ] Python available: `docker exec test-web-01 python3 --version`
- [ ] Inventory manager sees servers: `./bin/inventory-manager`

---

**Created by:** Boiler Expert Agent v2  
**Date:** 2025-11-22  
**Based on:** test-docker-vps.sh methodology  
**Status:** ✅ Production Ready
