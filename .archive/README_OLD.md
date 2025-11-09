# Ansible Deployment for Node.js Applications on Scaleway

Complete Ansible setup for deploying and managing Node.js applications with PostgreSQL on Scaleway VPS instances running Debian.

## 🚀 Features

- **Full server provisioning** on Scaleway VPS (Debian)
- **Security hardening**: UFW firewall, fail2ban, SSH hardening, auto-updates
- **PostgreSQL** database setup with backups
- **Node.js** with PM2 process manager
- **Nginx** reverse proxy with Let's Encrypt SSL
- **Monitoring**: Prometheus + Grafana + Node Exporter
- **Zero-downtime deployments** with rollback capability
- **Multi-environment** support (dev, production)

## 📋 Prerequisites

- Ansible 2.9+ installed on your local machine
- Scaleway VPS instances (Debian 11/12)
- SSH access to servers (password or key)
- Domain name (optional, for SSL certificates)
- GitHub repository with your Node.js application

## 🛠️ Installation

### 1. Install Ansible

```bash
# On Ubuntu/Debian
sudo apt update
sudo apt install ansible

# On macOS
brew install ansible

# Verify installation
ansible --version
```

### 2. Install required Ansible collections

```bash
ansible-galaxy collection install -r requirements.yml
```

### 3. Generate SSH key (if you don't have one)

```bash
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
```

## ⚙️ Configuration

### 1. Update Inventory Files

Edit `inventory/dev/hosts.yml` and `inventory/production/hosts.yml` with your Scaleway server IPs:

```yaml
webservers:
  hosts:
    prod-web-01:
      ansible_host: YOUR_SCALEWAY_IP_1
    prod-web-02:
      ansible_host: YOUR_SCALEWAY_IP_2

dbservers:
  hosts:
    prod-db-01:
      ansible_host: YOUR_SCALEWAY_IP_3
```

### 2. Configure Variables

#### `group_vars/all.yml`
- Update `app_repo` with your GitHub repository URL
- Set `app_name` to your application name
- Update `ssh_key_path` to your public SSH key location

#### `group_vars/webservers.yml`
- Set `ssl_certbot_email` to your email
- Update `ssl_domains` with your domain name

#### `group_vars/dbservers.yml`
- Change `postgresql_users[].password` to a secure password
- Adjust PostgreSQL performance settings based on server resources

### 3. First-time Setup on Scaleway

If your servers are brand new and require password authentication initially:

```bash
# Copy your SSH key to the servers
ssh-copy-id debian@YOUR_SERVER_IP

# Or use ansible to do this
ansible all -i inventory/dev -m authorized_key \
  -a "user=debian key='{{ lookup('file', '~/.ssh/id_rsa.pub') }}' state=present" \
  --ask-pass
```

## 🚢 Deployment

### Full Provisioning (First Time)

This sets up everything: users, security, PostgreSQL, Node.js, Nginx, monitoring.

```bash
# For development environment
ansible-playbook playbooks/provision.yml -i inventory/dev

# For production environment
ansible-playbook playbooks/provision.yml -i inventory/production
```

### Deploy Application

```bash
# Development
ansible-playbook playbooks/deploy.yml -i inventory/dev

# Production
ansible-playbook playbooks/deploy.yml -i inventory/production
```

### Update Application

Quick update without full deployment (pulls latest code, updates dependencies):

```bash
ansible-playbook playbooks/update.yml -i inventory/production
```

### Rollback

If something goes wrong, rollback to the previous release:

```bash
ansible-playbook playbooks/rollback.yml -i inventory/production
```

## 📊 Monitoring

After provisioning, access monitoring tools:

### Prometheus
```
http://YOUR_MONITORING_SERVER_IP:9090
```

### Grafana
```
http://YOUR_MONITORING_SERVER_IP:3001
Default credentials: admin / admin
```

### Node Exporter Metrics
```
http://ANY_SERVER_IP:9100/metrics
```

## 🔐 Security Features

- **UFW Firewall**: Configured to allow only necessary ports
- **fail2ban**: Protects against brute-force attacks
- **SSH Hardening**: Root login disabled, password auth disabled
- **Automatic Security Updates**: Enabled via unattended-upgrades
- **Deploy User**: Separate user for deployments with sudo access

## 📂 Project Structure

```
.
├── ansible.cfg                 # Ansible configuration
├── requirements.yml            # Required collections
├── inventory/
│   ├── dev/hosts.yml          # Development servers
│   └── production/hosts.yml   # Production servers
├── group_vars/
│   ├── all.yml                # Global variables
│   ├── webservers.yml         # Web server variables
│   └── dbservers.yml          # Database variables
├── roles/
│   ├── common/                # Base setup (users, packages)
│   ├── security/              # UFW, fail2ban, SSH hardening
│   ├── postgresql/            # PostgreSQL setup
│   ├── nodejs/                # Node.js and PM2
│   ├── nginx/                 # Nginx and SSL
│   ├── monitoring/            # Prometheus + Grafana
│   └── deploy-app/            # Application deployment
└── playbooks/
    ├── provision.yml          # Full server provisioning
    ├── deploy.yml             # Deploy application
    ├── update.yml             # Quick update
    └── rollback.yml           # Rollback to previous release
```

## 🔧 Useful Commands

### Check connectivity
```bash
ansible all -i inventory/production -m ping
```

### Run specific role
```bash
ansible-playbook playbooks/provision.yml -i inventory/dev --tags "security"
```

### Check playbook syntax
```bash
ansible-playbook playbooks/deploy.yml --syntax-check
```

### Dry run (check mode)
```bash
ansible-playbook playbooks/deploy.yml -i inventory/dev --check
```

### View PM2 logs on server
```bash
ssh deploy@YOUR_SERVER_IP
pm2 logs
pm2 status
```

### Database backup
Backups are automatically created daily at 2 AM. To manually trigger:
```bash
ssh deploy@YOUR_DB_SERVER_IP
sudo -u postgres /usr/local/bin/backup_postgres.sh
```

## 🐛 Troubleshooting

### Connection issues
```bash
# Test SSH connection
ssh -v debian@YOUR_SERVER_IP

# Check inventory
ansible-inventory -i inventory/dev --list
```

### Application not starting
```bash
# Check PM2 status
ssh deploy@YOUR_SERVER_IP
pm2 status
pm2 logs

# Check Nginx
sudo systemctl status nginx
sudo nginx -t
```

### Database connection issues
```bash
# Check PostgreSQL
sudo systemctl status postgresql
sudo -u postgres psql -c "\l"

# Test connection from web server
psql -h DB_SERVER_IP -U myapp_user -d myapp_production
```

## 📝 Application Requirements

Your Node.js application should:

1. **Have a main entry file** (index.js, app.js, or server.js)
2. **Listen on the port** specified in environment variable `PORT`
3. **Include a health endpoint** (optional but recommended): `GET /health`
4. **Use environment variables** from `.env` file

Example minimal app:
```javascript
const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

app.listen(port, () => {
  console.log(`Server running on port ${port}`);
});
```

## 🔄 Update Process

The deployment uses a releases directory structure:
```
/var/www/myapp/
├── current -> releases/20250123_143000_abc1234
├── releases/
│   ├── 20250123_143000_abc1234/
│   ├── 20250122_120000_def5678/
│   └── ...
└── shared/
    ├── logs/
    └── config/.env
```

This allows:
- Quick rollbacks
- Zero-downtime deployments
- Shared configuration between releases

## 📚 Additional Resources

- [Ansible Documentation](https://docs.ansible.com/)
- [PM2 Documentation](https://pm2.keymetrics.io/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)

## 🤝 Contributing

Feel free to submit issues or pull requests for improvements!

## 📄 License

MIT License - feel free to use this for your projects!

## 🔒 Version Control & Security

### Git Setup

This project includes sensitive configuration files that should NOT be committed to Git:
- `group_vars/all.yml` - Contains app repo URLs and settings
- `group_vars/dbservers.yml` - Contains database passwords
- `inventory/*/hosts.yml` - Contains server IPs

These files are in `.gitignore` for safety.

### Initial Git Setup

```bash
# Initialize git repository
git init

# The .gitignore is already configured
# Example files (.example.yml) can be safely committed

# Add files
git add .

# Commit
git commit -m "Initial Ansible deployment setup"

# Add remote
git remote add origin https://github.com/yourusername/your-ansible-repo.git

# Push
git push -u origin main
```

### Sharing Configuration Safely

Use the `.example` files for sharing project structure:
- `group_vars/all.yml.example`
- `group_vars/dbservers.yml.example`
- `inventory/dev/hosts.yml.example`

Team members can copy these:
```bash
cp group_vars/all.yml.example group_vars/all.yml
cp group_vars/dbservers.yml.example group_vars/dbservers.yml
cp inventory/dev/hosts.yml.example inventory/dev/hosts.yml
# Then update with real values
```

### Using Ansible Vault (Recommended)

For production, use Ansible Vault to encrypt sensitive data:

```bash
# Encrypt sensitive files
ansible-vault encrypt group_vars/dbservers.yml

# Edit encrypted file
ansible-vault edit group_vars/dbservers.yml

# Run playbook with vault
ansible-playbook playbooks/provision.yml -i inventory/production --ask-vault-pass

# Or use password file
echo "your-vault-password" > .vault-password
chmod 600 .vault-password
ansible-playbook playbooks/provision.yml -i inventory/production --vault-password-file .vault-password
```

