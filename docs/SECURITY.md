# 🔒 Security Best Practices

## ⚠️ Deploy User Configuration

### Current State (NOT RECOMMENDED)

The default configuration uses **root** as deploy user for simplicity:

```yaml
deploy_user: root
allow_root_login: true
```

**⚠️ Security Risks:**
- Maximum attack surface
- No privilege separation
- Non-compliant with security standards (PCI-DSS, ISO27001, CIS)
- Single point of failure

### Recommended Configuration

**Create dedicated deploy user with limited sudo:**

```yaml
deploy_user: deploy
deploy_user_groups:
  - sudo
  - www-data
allow_root_login: false  # Disable after initial provision
```

---

## 🔄 Migration Guide: Root → Deploy User

### Phase 1: New Deployments (Recommended)

For **new servers**, use secure configuration from start:

1. **Edit `group_vars/all.yml` or `inventory/<env>/group_vars/all.yml`:**

```yaml
# Deploy user configuration
deploy_user: deploy
deploy_user_groups:
  - sudo
  - www-data

# SSH Configuration
allow_root_login: false  # Will be enforced after provision
```

2. **Provision with root initially (SSH key must be configured for root):**

```bash
# First provision creates deploy user and transfers SSH key
./deploy.sh provision production
```

3. **Update inventory to use deploy user:**

Edit `inventory/<env>/hosts.yml`:
```yaml
all:
  children:
    webservers:
      hosts:
        server1:
          ansible_host: 192.168.1.10
          ansible_user: deploy  # Changed from root
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          ansible_become: yes    # Enables sudo
```

4. **Deploy application:**

```bash
./deploy.sh deploy production
```

### Phase 2: Existing Servers (Migration)

For **already provisioned** servers currently using root:

#### Option A: Clean Re-Provision (Safest)

1. Backup data if needed
2. Follow Phase 1 steps
3. Redeploy from scratch

#### Option B: In-Place Migration (Advanced)

1. **Create deploy user manually:**

```bash
# SSH to server as root
ssh root@your-server

# Create user
useradd -m -s /bin/bash deploy
usermod -aG sudo,www-data deploy

# Copy SSH keys
mkdir -p /home/deploy/.ssh
cp /root/.ssh/authorized_keys /home/deploy/.ssh/
chown -R deploy:deploy /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys

# Test sudo access
su - deploy
sudo ls /root  # Should work without password
```

2. **Configure passwordless sudo:**

```bash
# As root
echo "deploy ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/deploy
chmod 440 /etc/sudoers.d/deploy
```

3. **Update inventory and configs (see Phase 1 step 3)**

4. **Test deployment:**

```bash
# Test with deploy user
ansible -i inventory/production webservers -m ping --become

# Deploy
./deploy.sh deploy production
```

5. **Disable root login (when confident):**

```bash
# Edit group_vars
allow_root_login: false

# Re-run security role
ansible-playbook -i inventory/production playbooks/provision.yml \
  --tags security --limit your-server
```

---

## 🛡️ Security Hardening Checklist

### SSH Configuration

- [x] **Key-based authentication only** (PasswordAuthentication no)
- [ ] **Disable root login** (`allow_root_login: false`)
- [x] **Non-standard SSH port** (optional: `ssh_port: 2222`)
- [x] **fail2ban enabled** (auto-ban after failed attempts)

### User Configuration

- [ ] **Dedicated deploy user** (not root)
- [ ] **Limited sudo access** (sudoers.d with specific commands)
- [x] **Strong SSH keys** (RSA 4096 or ED25519)

### Firewall

- [x] **UFW enabled** (`enable_firewall: true`)
- [x] **Minimal open ports** (22/SSH, 80/HTTP, 443/HTTPS, app_port)
- [x] **Rate limiting** (UFW limits on SSH)

### Application Security

- [x] **PM2 user isolation** (runs as deploy user, not root)
- [x] **Nginx reverse proxy** (app not directly exposed)
- [ ] **Environment variables** (use .env files, not hardcoded)
- [ ] **SSL/TLS enabled** (Let's Encrypt via `configure-ssl.sh`)

### Monitoring & Updates

- [x] **Automatic security updates** (unattended-upgrades)
- [x] **Log rotation** (logrotate configured)
- [ ] **Monitoring alerts** (Prometheus + Grafana)
- [ ] **Regular audits** (monthly security scans)

---

## 🔐 Sudo Configuration Options

### Option 1: Full Sudo (Default)

```bash
# /etc/sudoers.d/deploy
deploy ALL=(ALL) NOPASSWD:ALL
```

**Pros:** Simple, works with any Ansible task  
**Cons:** Deploy user has root privileges

### Option 2: Limited Sudo (Recommended)

```bash
# /etc/sudoers.d/deploy
deploy ALL=(ALL) NOPASSWD: /usr/sbin/service, /bin/systemctl, /usr/bin/apt-get, \
    /usr/bin/npm, /usr/bin/pm2, /usr/sbin/nginx, /bin/chown, /bin/chmod, \
    /usr/bin/rsync, /bin/ln, /bin/rm
```

**Pros:** Minimal privileges  
**Cons:** May need adjustments for specific tasks

### Option 3: Targeted Sudo (Most Secure)

```bash
# /etc/sudoers.d/deploy
deploy ALL=(ALL) NOPASSWD: /bin/systemctl restart myapp, \
    /bin/systemctl reload nginx, \
    /usr/bin/pm2 * --uid deploy
Cmnd_Alias APT_CMDS = /usr/bin/apt-get update, /usr/bin/apt-get install
deploy ALL=(ALL) NOPASSWD: APT_CMDS
```

**Pros:** Maximum security  
**Cons:** Requires careful planning, may break automation

---

## 🚨 Common Pitfalls

### 1. NVM Path Issues

**Problem:** NVM installed under `/root/.nvm` when using root

**Solution:** Our roles detect NVM path dynamically:
```bash
# Checks both /home/<user>/.nvm and /root/.nvm
if [ -d "/home/{{ deploy_user }}/.nvm" ]; then
    export NVM_DIR="/home/{{ deploy_user }}/.nvm"
elif [ -d "/root/.nvm" ]; then
    export NVM_DIR="/root/.nvm"
fi
```

### 2. File Permissions

**Problem:** Files owned by root, app can't read

**Solution:** Roles use `become_user: {{ deploy_user }}` for app-related tasks

### 3. PM2 Startup

**Problem:** PM2 configured for root user

**Solution:** After migration, regenerate PM2 startup:
```bash
ssh deploy@server
pm2 startup
pm2 save
```

---

## 📋 Audit Commands

### Check Current User Configuration

```bash
# Ansible user
ansible -i inventory/production webservers -m shell -a "whoami"

# Deploy user
grep deploy_user group_vars/all.yml inventory/*/group_vars/all.yml

# SSH root login status
ansible -i inventory/production webservers -m shell -a "grep PermitRootLogin /etc/ssh/sshd_config" --become
```

### Verify Sudo Access

```bash
# Test deploy user sudo
ansible -i inventory/production webservers -m shell -a "sudo -n true" -u deploy

# List sudo permissions
ansible -i inventory/production webservers -m shell -a "sudo -l" -u deploy
```

### Check File Ownership

```bash
# App directory
ansible -i inventory/production webservers -m shell -a "ls -la /var/www/" --become

# PM2 processes
ansible -i inventory/production webservers -m shell -a "ps aux | grep PM2" --become
```

---

## 🎯 Compliance Matrix

| Standard | Requirement | Status with Root | Status with Deploy |
|----------|-------------|------------------|-------------------|
| **CIS Benchmark** | Disable root login | ❌ Fail | ✅ Pass |
| **PCI-DSS** | Principle of least privilege | ❌ Fail | ✅ Pass |
| **ISO 27001** | Access control | ⚠️ Partial | ✅ Pass |
| **NIST** | User account separation | ❌ Fail | ✅ Pass |
| **SOC 2** | Privileged access management | ❌ Fail | ✅ Pass |

---

## 📚 References

- [CIS Ubuntu Benchmark](https://www.cisecurity.org/benchmark/ubuntu_linux)
- [Ansible Privilege Escalation](https://docs.ansible.com/ansible/latest/user_guide/become.html)
- [SSH Hardening Guide](https://www.ssh.com/academy/ssh/sshd_config)
- [Linux Sudo Best Practices](https://www.sudo.ws/docs/man/sudoers.man/)

---

## 🆘 Troubleshooting

### Deploy user can't sudo

```bash
# Check sudoers file
sudo visudo -f /etc/sudoers.d/deploy

# Verify group membership
groups deploy
```

### SSH key not working

```bash
# Check permissions
ls -la /home/deploy/.ssh/
# Should be: drwx------ (700) for .ssh, -rw------- (600) for authorized_keys

# Fix if needed
sudo chown -R deploy:deploy /home/deploy/.ssh
sudo chmod 700 /home/deploy/.ssh
sudo chmod 600 /home/deploy/.ssh/authorized_keys
```

### PM2 process not found

```bash
# PM2 installed for which user?
which pm2

# Reinstall for deploy user
su - deploy
npm install -g pm2
pm2 startup
```

---

**Last Updated:** 2025-11-21  
**Maintainer:** Boiler Deploy Team  
**Review Cycle:** Quarterly
