# Troubleshooting Guide

Common issues and solutions for deploying Node.js applications.

## Table of Contents

- [SSH and Connection Issues](#ssh-and-connection-issues)
- [Ansible Errors](#ansible-errors)
- [Deployment Failures](#deployment-failures)
- [Application Not Starting](#application-not-starting)
- [Monitoring Issues](#monitoring-issues)
- [Debian 13 Specific Issues](#debian-13-specific-issues)
- [Performance Issues](#performance-issues)
- [Database Issues](#database-issues)

## SSH and Connection Issues

### Cannot Connect to VPS

**Symptoms:**
```
fatal: [vps-01]: UNREACHABLE! => {"changed": false, "msg": "Failed to connect to the host via ssh"}
```

**Solutions:**

1. **Test SSH manually:**
```bash
ssh root@your-vps-ip
```

2. **Check if SSH key is loaded:**
```bash
ssh-add -l
```

If empty, add your key:
```bash
ssh-add ~/.ssh/id_rsa
```

3. **Use specific key:**
```yaml
# inventory/production/hosts.yml
vps-01:
  ansible_ssh_private_key_file: ~/.ssh/your_key
```

4. **Check firewall:**
```bash
# On VPS
sudo ufw status
sudo ufw allow 22/tcp
```

### Permission Denied (publickey)

**Symptoms:**
```
Permission denied (publickey,password)
```

**Solutions:**

1. **Copy your SSH key to VPS:**
```bash
ssh-copy-id root@your-vps-ip
```

2. **Or manually:**
```bash
cat ~/.ssh/id_rsa.pub | ssh root@your-vps-ip "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

3. **Check SSH config on VPS:**
```bash
ssh root@your-vps-ip
cat /etc/ssh/sshd_config | grep PubkeyAuthentication
# Should be: PubkeyAuthentication yes
```

### Host Key Verification Failed

**Symptoms:**
```
Host key verification failed
```

**Solutions:**

1. **Remove old host key:**
```bash
ssh-keygen -R your-vps-ip
```

2. **Or disable strict checking (not recommended for production):**
```yaml
# inventory/production/hosts.yml
vps-01:
  ansible_ssh_common_args: '-o StrictHostKeyChecking=no'
```

## Ansible Errors

### Module Not Found

**Symptoms:**
```
ERROR! couldn't resolve module/action 'community.postgresql.postgresql_db'
```

**Solution:**

Install required Ansible collections:
```bash
ansible-galaxy collection install -r requirements.yml
```

### Variables Undefined

**Symptoms:**
```
The task includes an option with an undefined variable. The error was: 'app_name' is undefined
```

**Solutions:**

1. **Check `group_vars/all.yml` exists:**
```bash
ls -la group_vars/all.yml
```

2. **Verify variables are set:**
```bash
ansible all -i inventory/production -m debug -a "var=app_name"
```

3. **Check inventory structure:**
```bash
ansible-inventory -i inventory/production --list
```

### YAML Syntax Error

**Symptoms:**
```
ERROR! We were unable to read either as JSON nor YAML
```

**Solution:**

Validate YAML syntax:
```bash
# Check inventory
ansible-inventory -i inventory/production --list

# Check playbook
ansible-playbook playbooks/provision.yml --syntax-check
```

Common YAML mistakes:
- Wrong indentation (use 2 spaces, not tabs)
- Missing colon after key
- Unquoted strings with special characters

## Deployment Failures

### Git Clone Fails

**Symptoms:**
```
fatal: [vps-01]: FAILED! => {"changed": false, "cmd": "/usr/bin/git checkout --force main", "msg": "Failed to checkout main"}
```

**Solutions:**

1. **Check branch exists:**
```bash
git ls-remote --heads https://github.com/user/repo.git
```

2. **Update branch in config:**
```yaml
# group_vars/all.yml
app_branch: "main"  # or "master", "develop", etc.
```

3. **Check repository is accessible:**
```bash
git ls-remote https://github.com/user/repo.git
```

### Build Fails

**Symptoms:**
```
npm ERR! missing script: build
```

**Solutions:**

1. **Check if build script exists:**
```json
// package.json
{
  "scripts": {
    "build": "next build"  // Must exist for Next.js
  }
}
```

2. **For apps without build:**
The auto-detection will skip build if no script exists.

3. **Check build logs:**
```bash
ssh deploy@your-vps-ip 'cat /var/www/myapp/current/.npm/_logs/*.log'
```

### Dependencies Installation Fails

**Symptoms:**
```
npm ERR! Cannot find module 'some-package'
```

**Solutions:**

1. **Check lockfile exists:**
```bash
# Your project should have one of:
package-lock.json  # npm
yarn.lock         # yarn
pnpm-lock.yaml    # pnpm
```

2. **For Next.js/Nuxt.js, ensure devDependencies are included:**
The system automatically installs devDependencies for framework apps.

3. **Check package.json is valid:**
```bash
cat package.json | jq .
```

### PM2 Not Starting

**Symptoms:**
```
[PM2][ERROR] Script not found: ./index.js
```

**Solutions:**

1. **Check entry point:**
```json
// package.json
{
  "main": "server.js"  // Must point to existing file
}
```

2. **For Next.js/Nuxt.js:** Entry point is not needed (uses `npm start`)

3. **Check PM2 logs:**
```bash
ssh deploy@your-vps-ip 'pm2 logs --err --lines 50'
```

4. **Restart PM2:**
```bash
ssh deploy@your-vps-ip 'pm2 restart all'
```

## Application Not Starting

### Port Already in Use

**Symptoms:**
```
Error: listen EADDRINUSE: address already in use :::3000
```

**Solutions:**

1. **Check what's using the port:**
```bash
ssh deploy@your-vps-ip 'sudo netstat -tlnp | grep 3000'
```

2. **Kill the process:**
```bash
ssh deploy@your-vps-ip 'sudo kill -9 <PID>'
```

3. **Or change your app port:**
```yaml
# group_vars/all.yml
app_port: 3001
```

### Application Crashes Immediately

**Symptoms:**
```
PM2 status shows "errored" or constant restarts
```

**Solutions:**

1. **View error logs:**
```bash
ssh deploy@your-vps-ip 'pm2 logs myapp --err --lines 100'
```

2. **Check environment variables:**
```bash
ssh deploy@your-vps-ip 'cat /var/www/myapp/shared/config/.env'
```

3. **Test manually:**
```bash
ssh deploy@your-vps-ip
cd /var/www/myapp/current
NODE_ENV=production npm start
```

4. **Check PM2 config:**
```bash
ssh deploy@your-vps-ip 'cat /var/www/myapp/current/ecosystem.config.js'
```

### Next.js 404 on All Routes

**Symptoms:**
- Application starts
- All routes return 404

**Solution:**

This often happens with i18n. Next.js might redirect `/` to `/en` or `/fr`.

Try accessing: `http://your-vps-ip/en` or check middleware.ts

### Cannot Access Application

**Symptoms:**
- PM2 shows "online"
- Cannot access `http://your-vps-ip`

**Solutions:**

1. **Check Nginx is running:**
```bash
ssh deploy@your-vps-ip 'sudo systemctl status nginx'
```

2. **Check Nginx config:**
```bash
ssh deploy@your-vps-ip 'sudo nginx -t'
```

3. **Check firewall:**
```bash
ssh deploy@your-vps-ip 'sudo ufw status'
# Port 80 should be allowed
```

4. **Test locally on VPS:**
```bash
ssh deploy@your-vps-ip 'curl http://localhost:3000'
```

## Monitoring Issues

### Prometheus Not Accessible

**Symptoms:**
- `http://your-vps-ip:9090` times out

**Solutions:**

1. **Check Prometheus is running:**
```bash
ssh deploy@your-vps-ip 'sudo systemctl status prometheus'
```

2. **Check firewall:**
```bash
ssh deploy@your-vps-ip 'sudo ufw status | grep 9090'
```

If not allowed:
```bash
ssh deploy@your-vps-ip 'sudo ufw allow 9090/tcp'
```

3. **Check Prometheus logs:**
```bash
ssh deploy@your-vps-ip 'sudo journalctl -u prometheus -n 50'
```

### Grafana Not Accessible

**Symptoms:**
- `http://your-vps-ip:3001` times out

**Solutions:**

1. **Check Grafana is running:**
```bash
ssh deploy@your-vps-ip 'sudo systemctl status grafana-server'
```

2. **Check firewall:**
```bash
ssh deploy@your-vps-ip 'sudo ufw status | grep 3001'
```

If not allowed:
```bash
ssh deploy@your-vps-ip 'sudo ufw allow 3001/tcp'
```

3. **Reset Grafana admin password:**
```bash
ssh deploy@your-vps-ip 'sudo grafana-cli admin reset-admin-password newpassword'
```

### Node Exporter Not Working

**Symptoms:**
- No metrics in Prometheus

**Solutions:**

1. **Check Node Exporter:**
```bash
ssh deploy@your-vps-ip 'sudo systemctl status node_exporter'
```

2. **Test metrics endpoint:**
```bash
ssh deploy@your-vps-ip 'curl http://localhost:9100/metrics'
```

## Debian 13 Specific Issues

### apt-key Deprecated

**Symptoms:**
```
Warning: apt-key is deprecated
```

**Solution:**

This is already fixed in the roles. The system uses modern GPG key management.

### software-properties-common Not Found

**Symptoms:**
```
No package matching 'software-properties-common' is available
```

**Solution:**

This package doesn't exist in Debian 13. Already removed from our roles.

### Cron Not Installed

**Symptoms:**
```
Failed to find required executable "crontab"
```

**Solution:**

Already fixed. Cron is now in the common packages list.

## Performance Issues

### High Memory Usage

**Symptoms:**
- PM2 constantly restarting
- OOM (Out of Memory) errors

**Solutions:**

1. **Increase PM2 memory limit:**
```yaml
# group_vars/all.yml
pm2_max_memory: "1G"  # or higher
```

2. **Reduce PM2 instances:**
```yaml
pm2_instances: 1  # For low-memory VPS
```

3. **Check actual memory usage:**
```bash
ssh deploy@your-vps-ip 'free -h'
ssh deploy@your-vps-ip 'pm2 monit'
```

### Slow Build Times

**Symptoms:**
- Deployment takes very long

**Solutions:**

1. **For Next.js with Turbopack:**
```json
{
  "scripts": {
    "build": "next build"  // Remove --turbopack if issues
  }
}
```

2. **Use smaller VPS for build:**
Build locally and deploy pre-built artifacts (advanced).

3. **Check available disk space:**
```bash
ssh deploy@your-vps-ip 'df -h'
```

## Database Issues

### Cannot Connect to PostgreSQL

**Symptoms:**
```
Error: connect ECONNREFUSED 127.0.0.1:5432
```

**Solutions:**

1. **Check PostgreSQL is running:**
```bash
ssh deploy@your-vps-ip 'sudo systemctl status postgresql'
```

2. **Check connection string:**
```bash
# In your .env
DATABASE_URL=postgresql://user:password@localhost:5432/dbname
```

3. **Test connection:**
```bash
ssh deploy@your-vps-ip 'sudo -u postgres psql -c "\l"'
```

### Permission Denied for Database

**Symptoms:**
```
permission denied for database myapp_production
```

**Solution:**

Grant permissions:
```bash
ssh deploy@your-vps-ip 'sudo -u postgres psql'
```

```sql
GRANT ALL PRIVILEGES ON DATABASE myapp_production TO myapp;
```

## Diagnostic Commands

### Check All Services

```bash
ssh deploy@your-vps-ip 'sudo systemctl status nginx postgresql prometheus grafana-server node_exporter'
```

### Check PM2 Status

```bash
ssh deploy@your-vps-ip 'pm2 status && pm2 logs --lines 20'
```

### Check Disk Space

```bash
ssh deploy@your-vps-ip 'df -h && du -sh /var/www/*'
```

### Check Memory

```bash
ssh deploy@your-vps-ip 'free -h && ps aux --sort=-%mem | head -10'
```

### Check Logs

```bash
# Application logs
ssh deploy@your-vps-ip 'pm2 logs myapp --lines 100'

# Nginx logs
ssh deploy@your-vps-ip 'sudo tail -f /var/log/nginx/error.log'

# System logs
ssh deploy@your-vps-ip 'sudo journalctl -xe'
```

### Check Firewall

```bash
ssh deploy@your-vps-ip 'sudo ufw status verbose'
```

### Check Listening Ports

```bash
ssh deploy@your-vps-ip 'sudo netstat -tlnp'
```

## Getting Help

If you're still stuck:

1. **Enable verbose mode:**
```bash
./deploy.sh deploy -vvv
```

2. **Check documentation:**
- [Configuration Guide](CONFIGURATION.md)
- [Auto-Detection](AUTO_DETECTION.md)
- [Examples](EXAMPLES.md)

3. **Open an issue** with:
- Error message
- Deployment logs
- PM2 status output
- System info (OS, Node version)

---

For configuration options, see [Configuration Guide](CONFIGURATION.md).  
For examples, see [Examples Guide](EXAMPLES.md).
