# Docker Testing Environment

Complete Docker-based testing environment for Ansible deployments.

## 🐳 Overview

This Docker setup simulates your Scaleway infrastructure locally with:
- 2 web servers (Debian 12)
- 1 database server (Debian 12)
- SSH access configured
- Network isolated with custom subnet

## 🚀 Quick Start

### 1. Start the Environment

```bash
./docker-start.sh
```

This will:
- Build Docker images
- Start 3 containers
- Configure networking

### 2. Setup SSH Keys

```bash
./docker/setup-ssh-keys.sh
```

This creates SSH keys and distributes them to all containers.

### 3. Test Connection

```bash
ansible all -i inventory/docker -m ping
```

### 4. Provision Test Servers

```bash
./deploy.sh docker provision
```

### 5. Deploy Application

```bash
./deploy.sh docker deploy
```

## 📦 Container Details

| Container | IP | SSH Port | HTTP Port | Purpose |
|-----------|-------|----------|-----------|---------|
| ansible-web-01 | 172.28.0.11 | 2201 | 8001 | Web Server 1 |
| ansible-web-02 | 172.28.0.12 | 2202 | 8002 | Web Server 2 |
| ansible-db-01 | 172.28.0.21 | 2203 | 5432 | Database Server |

## 🔐 Access Credentials

- **User:** debian
- **Password:** debian (for initial access)
- **SSH Key:** ~/.ssh/ansible_docker_rsa (after setup)

## 📋 Common Commands

### Access a Container

```bash
# Interactive shell
docker exec -it ansible-web-01 bash
docker exec -it ansible-db-01 bash

# As debian user
docker exec -it -u debian ansible-web-01 bash
```

### Check Container Status

```bash
docker-compose ps
docker-compose logs web-01
docker-compose logs -f db-01
```

### Run Ansible Commands

```bash
# Ping all servers
ansible all -i inventory/docker -m ping

# Check disk space
ansible all -i inventory/docker -a "df -h"

# Check services
ansible webservers -i inventory/docker -a "systemctl status nginx" --become
```

### Testing Playbooks

```bash
# Full provision
./deploy.sh docker provision

# Deploy app
./deploy.sh docker deploy

# Update app
./deploy.sh docker update

# Rollback
./deploy.sh docker rollback

# Health check
./health_check.sh inventory/docker
```

## 🛑 Stop Environment

```bash
./docker-stop.sh
```

### Remove All Data

```bash
docker-compose down -v
```

## 🔧 Troubleshooting

### Containers not starting

```bash
# Check logs
docker-compose logs

# Rebuild images
docker-compose build --no-cache
docker-compose up -d
```

### SSH connection issues

```bash
# Regenerate and copy keys
rm ~/.ssh/ansible_docker_rsa*
./docker/setup-ssh-keys.sh

# Test direct SSH
ssh -i ~/.ssh/ansible_docker_rsa debian@172.28.0.11
```

### Network issues

```bash
# Check network
docker network inspect deploy-me_ansible_net

# Restart containers
docker-compose restart
```

### Port conflicts

If ports 2201-2203 are in use, modify docker-compose.yml:
```yaml
ports:
  - "3301:22"  # Change to available port
```

## 📊 Monitoring in Docker

After provisioning, access:
- **Prometheus:** http://localhost:9090 (if monitoring role applied)
- **Grafana:** http://localhost:3001 (if monitoring role applied)
- **Application:** http://localhost:8001, http://localhost:8002

## 💡 Tips

### Speed up testing

Skip certain roles during testing:
```bash
ansible-playbook playbooks/provision.yml -i inventory/docker --skip-tags "monitoring"
```

### Test specific roles

```bash
ansible-playbook playbooks/provision.yml -i inventory/docker --tags "security,nodejs"
```

### Debug mode

```bash
ansible-playbook playbooks/deploy.yml -i inventory/docker -vvv
```

### Test changes without committing

Make changes, test in Docker, then commit if successful.

## 🎯 Differences from Production

- Uses password authentication initially (removed after provisioning)
- Containers run with `privileged: true` for systemd
- Network is isolated (172.28.0.0/16)
- SSL certificates won't validate (no real domain)
- Firewall rules apply but less critical

## ✅ Testing Checklist

Before deploying to production:

- [ ] All playbooks run successfully
- [ ] Services start and stay running
- [ ] Database connections work
- [ ] Application deploys correctly
- [ ] Rollback works
- [ ] Health checks pass
- [ ] Monitoring shows data
- [ ] Backups are created

## 🔄 Reset Environment

To start fresh:

```bash
./docker-stop.sh
docker-compose down -v
docker system prune -f
./docker-start.sh
./docker/setup-ssh-keys.sh
```

## 📚 Next Steps

After successful Docker testing:
1. Document any issues found
2. Update playbooks if needed
3. Test on dev environment (real VPS)
4. Deploy to production

## 🆘 Need Help?

Check the main documentation:
- README.md - Complete guide
- TROUBLESHOOTING.md - Common issues
- QUICKSTART.md - Setup guide
