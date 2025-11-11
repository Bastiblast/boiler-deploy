# Docker Test Container Guide

## Overview

This guide explains how to use the systemd-enabled Docker container to test provisioning and deployment without needing a real VPS.

## What's Special About This Container?

Unlike a basic Docker container, this test container:

1. **Runs systemd** - The full init system that manages services
2. **Supports service management** - Can start/stop/enable services like Nginx, Fail2ban, UFW
3. **SSH access** - Accessible via SSH just like a real VPS
4. **Minimal base** - Only includes SSH and Python; Ansible installs everything else
5. **Realistic environment** - Behaves like an actual Ubuntu 22.04 server

## Quick Start

### 1. Setup the Test Container

```bash
./test-docker-vps.sh setup
```

This will:
- Generate SSH keys (`~/.ssh/boiler_test_rsa`)
- Build a systemd-enabled Ubuntu 22.04 image
- Start the container with proper privileges
- Configure SSH access
- Create test inventory files

### 2. Access the Container

**Via SSH:**
```bash
ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost
```

**Via Docker:**
```bash
docker exec -it boiler-test-vps bash
```

### 3. Use with Inventory Manager

Run the inventory manager:
```bash
make run
```

Create a new environment called `docker-test` with these settings:

**Server Configuration:**
- Name: `docker-web-01`
- IP: `127.0.0.1`
- SSH Port: `2222`
- SSH Key: `~/.ssh/boiler_test_rsa`
- Type: `web`
- Repository: `https://github.com/Bastiblast/portefolio` (or any other)
- App Port: `3000`
- Node Version: `20`

**Then:**
1. ✅ Validate Inventory
2. 🔧 Provision Server (installs Node.js, Nginx, UFW, Fail2ban, PostgreSQL, etc.)
3. 🚀 Deploy Application

### 4. Access Deployed Application

After deployment completes:
```bash
curl http://localhost:8080
```

Or open in browser: http://localhost:8080

## Commands Reference

### Container Management

```bash
# Setup everything
./test-docker-vps.sh setup

# Check container status
./test-docker-vps.sh status

# Restart container
./test-docker-vps.sh restart

# View logs
./test-docker-vps.sh logs

# SSH into container
./test-docker-vps.sh ssh

# Execute bash in container
./test-docker-vps.sh exec

# Remove everything
./test-docker-vps.sh cleanup
```

### Docker Commands

```bash
# Check container status
docker ps | grep boiler-test-vps

# View container logs
docker logs boiler-test-vps

# Stop container
docker stop boiler-test-vps

# Start container
docker start boiler-test-vps

# Remove container
docker rm -f boiler-test-vps
```

### Systemd Commands (inside container)

```bash
# Check SSH service
systemctl status ssh

# Check Nginx (after provisioning)
systemctl status nginx

# Check Fail2ban (after provisioning)
systemctl status fail2ban

# List all services
systemctl list-units --type=service

# View service logs
journalctl -u nginx
journalctl -u fail2ban
```

## Technical Details

### Why Systemd in Docker?

Ansible playbooks expect a real server environment with systemd to:
- Install and enable services (Nginx, Fail2ban, UFW)
- Manage service states (start/stop/restart)
- Configure service autostart on boot
- Use `systemctl` commands

A regular Docker container without systemd would fail when Ansible tries to:
```bash
systemctl enable nginx
systemctl start nginx
systemctl enable fail2ban
```

### Container Configuration

The container uses the `jrei/systemd-ubuntu:22.04` base image with special flags:

```bash
docker run -d \
    --privileged \              # Required for systemd
    --tmpfs /tmp \              # Systemd needs these tmpfs mounts
    --tmpfs /run \
    --tmpfs /run/lock \
    -v /sys/fs/cgroup:/sys/fs/cgroup:rw \  # Cgroup support
    --cgroupns=host \           # Use host cgroup namespace
    --stop-signal SIGRTMIN+3 \  # Proper systemd shutdown
    -p 2222:22 \                # SSH port
    -p 8080:80 \                # HTTP port
    -p 8443:443 \               # HTTPS port
    boiler-test-ubuntu
```

### What Gets Installed?

**By Docker Image (minimal):**
- OpenSSH Server
- Python 3 (for Ansible)
- Python3-apt (for Ansible package management)
- Sudo
- CA certificates

**By Ansible Provision (everything else):**
- Node.js (version specified)
- Nginx (web server)
- UFW (firewall)
- Fail2ban (intrusion prevention)
- PostgreSQL (database)
- PM2 (process manager)
- Git
- Certbot (SSL certificates)
- And all dependencies

## Ports

| Service | Container Port | Host Port | Purpose |
|---------|---------------|-----------|---------|
| SSH     | 22            | 2222      | Remote access |
| HTTP    | 80            | 8080      | Web application |
| HTTPS   | 443           | 8443      | Secure web (if SSL enabled) |
| App     | 3000          | -         | Internal Node.js app |

## Troubleshooting

### SSH Connection Fails

```bash
# Check if container is running
docker ps | grep boiler-test-vps

# Check SSH service inside container
docker exec boiler-test-vps systemctl status ssh

# Test with verbose SSH
ssh -vvv -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost
```

### Provision Fails

```bash
# Check Ansible can connect
ansible -i inventory/docker-test/hosts all -m ping

# View full provision logs in inventory manager
# Logs are saved to logs/docker-test/provision/

# Manually SSH and check
./test-docker-vps.sh ssh
apt-get update  # Test if apt works
python3 --version  # Check Python
```

### Container Won't Start

```bash
# Remove and rebuild
./test-docker-vps.sh cleanup
./test-docker-vps.sh setup

# Check Docker logs
docker logs boiler-test-vps

# Verify systemd is working
docker exec boiler-test-vps systemctl status
```

### App Not Accessible After Deploy

```bash
# Check Nginx status
docker exec boiler-test-vps systemctl status nginx

# Check app is running
docker exec boiler-test-vps pm2 list

# Check Nginx config
docker exec boiler-test-vps nginx -t

# View Nginx logs
docker exec boiler-test-vps tail -f /var/log/nginx/error.log
```

## Best Practices

1. **Clean Start**: Always run `./test-docker-vps.sh cleanup` before testing major changes
2. **Save Logs**: Inventory manager saves logs to `logs/` directory - review them after failures
3. **Test Incrementally**: Test validation → provision → deploy separately
4. **Monitor Resources**: Container uses privileged mode - monitor Docker resource usage
5. **Persistent Data**: Container data is ephemeral - use volumes if you need persistence

## Differences from Real VPS

| Feature | Test Container | Real VPS |
|---------|---------------|----------|
| Systemd | ✅ Full support | ✅ Native |
| Service Management | ✅ Works | ✅ Works |
| Firewall (UFW) | ⚠️ Limited | ✅ Full |
| Kernel Access | ⚠️ Shared host | ✅ Isolated |
| Performance | 🚀 Fast | 🐢 Slower |
| Cost | 💰 Free | 💰 Paid |
| Network Isolation | ⚠️ Limited | ✅ Full |
| Persistence | ❌ Ephemeral | ✅ Persistent |

## Next Steps

After testing with the Docker container:

1. ✅ Verify all playbooks work correctly
2. ✅ Test validation, provision, and deploy flows
3. ✅ Review logs and fix any issues
4. 🎯 Deploy to real VPS with confidence

## Notes

- **Security**: This container is for testing only. It uses key-based SSH but runs as root.
- **Performance**: Container shares host kernel, so it's much faster than a real VPS.
- **Cleanup**: Always cleanup when done to free resources: `./test-docker-vps.sh cleanup`
- **Updates**: Pull latest Ubuntu image periodically: `docker pull jrei/systemd-ubuntu:22.04`
