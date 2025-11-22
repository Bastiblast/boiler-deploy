# Test Environment with Docker

This document explains how to use the Docker-based test environment to simulate a VPS for testing provisioning and deployment.

## Overview

The `test-docker-vps.sh` script creates a **lightweight Ubuntu container** with only SSH enabled. This simulates a fresh VPS similar to what you'd get from Scaleway or other providers.

**Important**: The container is intentionally minimal. It only includes:
- SSH server (OpenSSH)
- Python3 (minimal, for Ansible)
- Basic CA certificates

Everything else (Node.js, Nginx, PostgreSQL, UFW, Fail2ban, etc.) will be installed by your Ansible provision playbooks, exactly as it would be on a real VPS.

## Quick Start

### 1. Setup the Test Environment

```bash
./test-docker-vps.sh setup
```

This will:
- Generate SSH keys (if needed) at `~/.ssh/boiler_test_rsa`
- Build a minimal Ubuntu Docker image with SSH
- Start a container accessible on:
  - SSH: `localhost:2222`
  - HTTP: `localhost:8080` (after deployment)
  - HTTPS: `localhost:8443` (after SSL configuration)

### 2. Configure Inventory Manager

Run the inventory manager:
```bash
make run
```

Create a new environment with these settings:
- **Environment Name**: `test-docker`
- **Mono Server**: Yes (for simplicity)
- **Server Configuration**:
  - Name: `test-web-01`
  - IP: `127.0.0.1`
  - SSH Port: `2222`
  - SSH Key: `~/.ssh/boiler_test_rsa`
  - Type: `web`
  - Repository: `https://github.com/Bastiblast/ansible-next-test.git`
  - App Port: `3000`
  - Node Version: `20.x` (or your preference)

### 3. Test the Workflow

1. **Validate Inventory** - Check all settings are correct
2. **Provision Server** - Install Node.js, Nginx, UFW, Fail2ban, etc.
3. **Deploy Application** - Clone repo, install dependencies, start app
4. **Verify** - Check the app is running

### 4. Access Your Deployed App

After successful deployment:
```bash
curl http://localhost:8080
```

Or open in browser: `http://localhost:8080`

## Script Commands

```bash
# Setup environment
./test-docker-vps.sh setup

# Check status
./test-docker-vps.sh status

# SSH into container
./test-docker-vps.sh ssh

# Execute bash in container
./test-docker-vps.sh exec

# View logs
./test-docker-vps.sh logs

# Restart container
./test-docker-vps.sh restart

# Clean up everything
./test-docker-vps.sh cleanup
```

## Manual SSH Connection

```bash
ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost
```

## Troubleshooting

### Container won't start
```bash
# Check Docker is running
docker info

# Clean up and retry
./test-docker-vps.sh cleanup
./test-docker-vps.sh setup
```

### SSH connection fails
```bash
# Check SSH key permissions
ls -la ~/.ssh/boiler_test_rsa
# Should be: -rw------- (600)

# Test SSH manually
ssh -i ~/.ssh/boiler_test_rsa -p 2222 -v root@localhost
```

### Provision fails
The provision playbook might fail if:
- SSH connection is not working (test first)
- Container doesn't have internet access
- Required Ansible roles are missing

Check logs and verify container networking:
```bash
docker exec boiler-test-vps ping -c 3 8.8.8.8
```

## What Gets Installed

### Pre-installed (in container)
- Ubuntu 22.04
- OpenSSH Server
- Python3 (minimal)
- CA certificates

### Installed by Ansible Provision
Based on `playbooks/provision.yml`:
- **Common role**: System updates, basic packages
- **Security role**: UFW firewall, Fail2ban
- **Node.js role**: Node.js runtime and npm
- **Nginx role**: Web server and reverse proxy
- **PostgreSQL role**: Database (on dbservers)
- **Monitoring role**: Prometheus and Grafana

### Installed by Ansible Deploy
Based on `playbooks/deploy.yml`:
- Your application code (from Git)
- Node.js dependencies (npm install)
- PM2 process manager
- Application service configuration

## Port Mapping

| Service | Container Port | Host Port |
|---------|----------------|-----------|
| SSH     | 22             | 2222      |
| HTTP    | 80             | 8080      |
| HTTPS   | 443            | 8443      |
| App     | 3000           | -         |

## Differences from Real VPS

This test environment is very close to a real VPS, but:

1. **Networking**: Uses Docker port mapping instead of public IPs
2. **Systemd**: Uses simple SSH daemon instead of full systemd
3. **Performance**: May be slower than real VPS depending on Docker setup
4. **Persistence**: Container storage is ephemeral (use volumes if needed)

## Testing Multiple Servers

To test multi-server setups:

```bash
# Create multiple containers
for i in 1 2 3; do
    docker run -d \
        --name "boiler-test-vps-$i" \
        -p "222$i:22" \
        -p "808$i:80" \
        boiler-test-ubuntu
done
```

Then configure each in the inventory manager with different ports.

## Cleanup

Remove test environment completely:
```bash
./test-docker-vps.sh cleanup
```

This removes:
- Docker container
- Docker image

**Note**: SSH keys in `~/.ssh/` are kept for reuse. Delete manually if needed.

## Best Practices

1. **Always test in Docker first** before deploying to real servers
2. **Start fresh** - Run cleanup between major changes
3. **Check logs** - Use `docker logs boiler-test-vps` for debugging
4. **Validate inventory** - Before provisioning, validate all settings
5. **Monitor resources** - Docker containers can use significant resources

## Next Steps

After successful testing in Docker:
1. Deploy to a real Scaleway VPS
2. Update inventory with real IP addresses
3. Configure SSL with real domain names
4. Set up monitoring and backups

## Support

If you encounter issues:
1. Check `./test-docker-vps.sh status`
2. Review container logs: `docker logs boiler-test-vps`
3. Test SSH manually: `ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost`
4. Check Ansible verbose output: Add `-vvv` to ansible-playbook commands
