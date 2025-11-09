# Configuration Guide

Complete configuration reference for deploying Node.js applications to any VPS.

## Table of Contents

- [Inventory Setup](#inventory-setup)
- [Global Variables](#global-variables)
- [Web Server Variables](#web-server-variables)
- [Database Variables](#database-variables)
- [Monitoring Variables](#monitoring-variables)
- [SSL Configuration](#ssl-configuration)
- [Security Options](#security-options)
- [Auto-Detection Overrides](#auto-detection-overrides)

## Inventory Setup

The inventory defines your servers. You can have separate environments (production, staging, development).

### Directory Structure

```
inventory/
├── production/
│   ├── hosts.yml
│   └── group_vars/  (optional, environment-specific vars)
├── staging/
│   └── hosts.yml
└── dev/
    └── hosts.yml
```

### Single VPS Configuration

For a single VPS running all services:

```yaml
# inventory/production/hosts.yml
---
all:
  children:
    webservers:
      hosts:
        vps-01:
          ansible_host: XX.XX.XX.XX
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
    
    dbservers:
      hosts:
        vps-01:
          ansible_host: XX.XX.XX.XX
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
    
    monitoring:
      hosts:
        vps-01:
          ansible_host: XX.XX.XX.XX
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
```

### Multi-VPS Configuration

For separate web, database, and monitoring servers:

```yaml
# inventory/production/hosts.yml
---
all:
  children:
    webservers:
      hosts:
        vps-web-01:
          ansible_host: XX.XX.XX.XX
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
        vps-web-02:
          ansible_host: YY.YY.YY.YY
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
    
    dbservers:
      hosts:
        vps-db-01:
          ansible_host: ZZ.ZZ.ZZ.ZZ
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
    
    monitoring:
      hosts:
        vps-monitor-01:
          ansible_host: WW.WW.WW.WW
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
```

### SSH Key Authentication

Recommended approach (more secure):

```yaml
vps-web-01:
  ansible_host: XX.XX.XX.XX
  ansible_user: root
  ansible_ssh_private_key_file: ~/.ssh/id_rsa
```

### Password Authentication

Less secure, but works:

```yaml
vps-web-01:
  ansible_host: XX.XX.XX.XX
  ansible_user: root
  ansible_ssh_pass: your-password  # Store in Ansible Vault!
```

**Note**: Use `ansible-vault` to encrypt sensitive data:

```bash
ansible-vault encrypt_string 'your-password' --name 'ansible_ssh_pass'
```

## Global Variables

Edit `group_vars/all.yml` for settings that apply to all servers.

### Application Settings

```yaml
# Application Configuration
app_name: myapp                    # Your application name
app_port: 3000                     # Port your app listens on
app_repo: "https://github.com/user/repo.git"
app_branch: "main"                 # Branch to deploy
app_environment: "production"      # Environment name
```

### Deploy User

The system creates a dedicated deploy user (not root):

```yaml
# Deploy User Configuration
deploy_user: deploy                # Username for deployments
ssh_key_path: "~/.ssh/id_rsa.pub" # Your public SSH key
```

### Node.js Configuration

```yaml
# Node.js Configuration
nodejs_version: "20"               # LTS version (18, 20, etc.)
```

### PM2 Configuration

```yaml
# PM2 Configuration
pm2_app_name: "{{ app_name }}"    # PM2 app name
pm2_instances: 2                   # Cluster instances (ignored for Next.js/Nuxt)
pm2_max_memory: "512M"             # Max memory before restart
```

### Directory Structure

```yaml
# Application Directories
app_dir: "/var/www/{{ app_name }}"
app_releases_dir: "{{ app_dir }}/releases"
app_current_dir: "{{ app_dir }}/current"
app_shared_dir: "{{ app_dir }}/shared"
```

The deployment creates this structure:

```
/var/www/myapp/
├── releases/
│   ├── 20231109T120000_abc1234/  # Timestamped releases
│   ├── 20231109T130000_def5678/
│   └── 20231109T140000_ghi9012/
├── current -> releases/20231109T140000_ghi9012/  # Symlink
└── shared/
    ├── config/
    │   └── .env                  # Shared environment variables
    └── logs/
        ├── pm2-error.log
        └── pm2-out.log
```

### Timezone

```yaml
# System Timezone
timezone: "Europe/Paris"           # Or "America/New_York", "Asia/Tokyo", etc.
```

### Backup Configuration

```yaml
# Backup Configuration
backup_dir: "/var/backups"
backup_retention_days: 7           # Days to keep backups
```

## Web Server Variables

Edit `group_vars/webservers.yml` for web server-specific settings.

### Nginx Configuration

```yaml
# Nginx Configuration
nginx_worker_processes: auto       # Or specific number
nginx_worker_connections: 1024
nginx_keepalive_timeout: 65
nginx_client_max_body_size: "20M"  # Max upload size
```

### SSL Configuration

```yaml
# SSL Configuration (Let's Encrypt)
ssl_enabled: false                 # Set to true when ready
ssl_certbot_email: "admin@example.com"
ssl_domains:
  - "myapp.com"
  - "www.myapp.com"
```

**To enable SSL:**

1. Point your domain to your VPS IP
2. Set `ssl_enabled: true`
3. Update `ssl_certbot_email` and `ssl_domains`
4. Run `./deploy.sh provision`

### Node.js Environment

```yaml
# Node.js environment variables
node_env: "{{ app_environment }}"  # production, staging, etc.
node_options: "--max-old-space-size=2048"
```

## Database Variables

Edit `group_vars/dbservers.yml` for database configuration.

### PostgreSQL Configuration

```yaml
# PostgreSQL Version
postgresql_version: "15"

# Database Configuration
postgresql_databases:
  - name: myapp_production
    encoding: UTF-8
    lc_collate: en_US.UTF-8
    lc_ctype: en_US.UTF-8

# Database Users
postgresql_users:
  - name: myapp
    password: "CHANGE_THIS_PASSWORD"  # Use ansible-vault!
    databases:
      - myapp_production
    privileges: ALL
```

**Secure password storage:**

```bash
ansible-vault encrypt_string 'strong-password' --name 'password'
```

### Performance Tuning

```yaml
# PostgreSQL Performance
postgresql_shared_buffers: "256MB"
postgresql_effective_cache_size: "1GB"
postgresql_maintenance_work_mem: "64MB"
postgresql_work_mem: "4MB"
postgresql_max_connections: 100
```

### Backup Configuration

```yaml
# Database Backup
postgresql_backup_enabled: true
postgresql_backup_hour: 3
postgresql_backup_minute: 0
```

## Monitoring Variables

Configure Prometheus and Grafana settings.

### Prometheus Configuration

```yaml
# Prometheus
prometheus_version: "2.48.0"
prometheus_retention_time: "15d"   # How long to keep metrics
```

### Grafana Configuration

```yaml
# Grafana
grafana_admin_password: "admin"    # Change this!
grafana_port: 3001
```

### Node Exporter

```yaml
# Node Exporter
node_exporter_version: "1.7.0"
node_exporter_port: 9100
```

## SSL Configuration

### Automatic SSL with Let's Encrypt

```yaml
# group_vars/webservers.yml
ssl_enabled: true
ssl_certbot_email: "your-email@domain.com"
ssl_domains:
  - "yourdomain.com"
  - "www.yourdomain.com"
```

**Prerequisites:**
- Domain must point to your VPS IP
- Ports 80 and 443 must be accessible

**Process:**
1. Set configuration
2. Run `./deploy.sh provision`
3. Certbot automatically obtains and configures certificates
4. Auto-renewal is configured via cron

### Manual SSL Certificates

If using your own certificates:

```yaml
ssl_enabled: true
ssl_certificate_path: "/etc/ssl/certs/your-cert.crt"
ssl_certificate_key_path: "/etc/ssl/private/your-key.key"
```

Copy your certificates to the server before running provision.

## Security Options

### Firewall Rules (UFW)

Configured automatically, but you can customize:

```yaml
# Custom firewall rules
ufw_rules:
  - { port: 8080, proto: tcp, rule: allow, comment: "Custom service" }
  - { port: 5432, proto: tcp, rule: allow, from: "10.0.0.0/8" }
```

### fail2ban Configuration

```yaml
# fail2ban settings
fail2ban_bantime: "10m"
fail2ban_findtime: "10m"
fail2ban_maxretry: 5
```

### SSH Hardening

Configured automatically:
- Root login: disabled (after deploy user setup)
- Password authentication: disabled
- Key-only authentication: enabled
- SSH port: 22 (change if needed)

To change SSH port:

```yaml
# group_vars/all.yml
ssh_port: 2222  # Custom port
```

Don't forget to update your inventory:

```yaml
vps-01:
  ansible_host: XX.XX.XX.XX
  ansible_port: 2222
```

## Auto-Detection Overrides

The system auto-detects framework and package manager, but you can override:

```yaml
# group_vars/all.yml

# Force specific application type
# Options: nodejs, nextjs, nuxtjs, express, fastify, nestjs
app_type_override: "nextjs"

# Force specific package manager
# Options: npm, pnpm, yarn
package_manager_override: "pnpm"

# Force specific entry file
app_entry_file_override: "dist/main.js"
```

**When to override:**
- Auto-detection fails
- Custom build setup
- Monorepo structure
- Non-standard entry point

See [Auto-Detection Guide](AUTO_DETECTION.md) for more details.

## Environment Variables

Application-specific environment variables go in `.env`:

```bash
# .env (stored in shared/config/.env)
DATABASE_URL=postgresql://user:pass@localhost:5432/dbname
API_KEY=your-secret-key
NODE_ENV=production
```

Create the file:

```bash
ssh deploy@your-vps-ip
cat > /var/www/myapp/shared/config/.env << EOF
DATABASE_URL=postgresql://myapp:password@localhost:5432/myapp_production
NODE_ENV=production
EOF
```

The deployment automatically symlinks it to each release.

## VPS Provider Examples

### DigitalOcean

```yaml
vps-01:
  ansible_host: 142.93.XXX.XXX
  ansible_user: root
```

### Vultr

```yaml
vps-01:
  ansible_host: 45.76.XXX.XXX
  ansible_user: root
```

### Linode

```yaml
vps-01:
  ansible_host: 172.105.XXX.XXX
  ansible_user: root
```

### Hetzner

```yaml
vps-01:
  ansible_host: 88.198.XXX.XXX
  ansible_user: root
```

### OVH

```yaml
vps-01:
  ansible_host: 51.38.XXX.XXX
  ansible_user: debian  # OVH uses debian user
  ansible_become: yes
```

## Troubleshooting Configuration

### Validate Configuration

```bash
# Check inventory
ansible-inventory -i inventory/production --list

# Test connection
ansible all -i inventory/production -m ping

# Check variables
ansible all -i inventory/production -m debug -a "var=app_name"
```

### Common Issues

**Variables not loading:**
- Ensure `group_vars/all.yml` exists
- Check YAML syntax (indentation matters!)
- Verify inventory structure

**SSH connection fails:**
- Verify IP address is correct
- Test SSH manually: `ssh root@your-vps-ip`
- Check firewall allows port 22

**Auto-detection not working:**
- Ensure `package.json` exists in repository
- Check dependencies list frameworks correctly
- Use overrides if needed

## Next Steps

- [Auto-Detection Guide](AUTO_DETECTION.md) - Understand how frameworks are detected
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions
- [Examples](EXAMPLES.md) - Real-world configurations

---

Need help? Check the [Troubleshooting Guide](TROUBLESHOOTING.md) or open an issue.
