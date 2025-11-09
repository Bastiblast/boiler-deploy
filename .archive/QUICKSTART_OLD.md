# Quick Start Guide

## Before You Begin

1. **Get your Scaleway VPS IPs** from your Scaleway console
2. **Have your GitHub repository** ready (must contain a Node.js app)
3. **Ensure SSH access** to your servers

## Step-by-Step Setup

### 1. Install Ansible Collections
```bash
ansible-galaxy collection install -r requirements.yml
```

### 2. Configure Your Setup

#### A. Update Server IPs
Edit `inventory/dev/hosts.yml`:
```yaml
ansible_host: YOUR_SCALEWAY_IP_HERE
```

#### B. Update Application Settings
Edit `group_vars/all.yml`:
```yaml
app_name: myapp                           # Your app name
app_repo: "https://github.com/user/repo"  # Your GitHub repo
ssh_key_path: "~/.ssh/id_rsa.pub"        # Your SSH public key
```

#### C. Update Domain (for SSL)
Edit `group_vars/webservers.yml`:
```yaml
ssl_certbot_email: "you@example.com"
ssl_domains:
  - "yourapp.example.com"
```

#### D. Set Database Password
Edit `group_vars/dbservers.yml`:
```yaml
postgresql_users:
  - name: "myapp_user"
    password: "CHANGE_ME_TO_SECURE_PASSWORD"  # Important!
```

### 3. First-Time Setup

If servers are brand new, copy your SSH key:
```bash
ssh-copy-id debian@YOUR_SERVER_IP
```

### 4. Provision Servers

This installs everything (takes 5-10 minutes):
```bash
./deploy.sh dev provision
```

### 5. Deploy Your App

```bash
./deploy.sh dev deploy
```

### 6. Access Your Services

- **Your App**: http://YOUR_SERVER_IP (or https://yourdomain.com if SSL configured)
- **Prometheus**: http://YOUR_SERVER_IP:9090
- **Grafana**: http://YOUR_SERVER_IP:3001 (admin/admin)

## Common Commands

```bash
# Deploy to development
./deploy.sh dev deploy

# Deploy to production
./deploy.sh production deploy

# Update existing deployment
./deploy.sh production update

# Rollback if needed
./deploy.sh production rollback

# Check server connectivity
ansible all -i inventory/dev -m ping
```

## Checklist Before Going Live

- [ ] Updated all IP addresses in inventory files
- [ ] Changed database passwords
- [ ] Updated app_repo URL
- [ ] Configured domain name
- [ ] Updated SSL email
- [ ] Tested deployment on dev environment
- [ ] Verified monitoring is working
- [ ] Tested rollback procedure

## Need Help?

Check the full README.md for detailed information and troubleshooting!
