# Multi-Server Setup Wizard

The setup wizard (`setup.sh`) helps you configure your deployment environment with multiple VPS servers.

## 🚀 Quick Start

```bash
# Create a new environment
./setup.sh

# Or specify environment name directly
./setup.sh production

# Add servers to existing environment
./setup.sh production --add-servers

# Resume interrupted setup
./setup.sh production --resume
```

## 📋 What the Wizard Does

The setup wizard guides you through 7 phases:

1. **Prerequisites Check** - Validates SSH, Ansible, Python3, Git
2. **Environment Setup** - Creates or selects deployment environment
3. **SSH Key Configuration** - Sets up authentication keys
4. **Web Server Configuration** - Configures 1-20 web servers
5. **Database Configuration** - Sets up PostgreSQL server
6. **Monitoring Configuration** - Configures Prometheus + Grafana
7. **Application Configuration** - Sets up Git repo and Node.js version

## 🎯 Features

### Multi-Server Support

- Deploy up to **20 web servers** per environment
- Automatic **load balancing** configuration
- **ID gap management** - next available ID assignment
- **Single SSH user** for all web servers

### Intelligent Validation

- **IP format validation** with octet checking
- **Duplicate IP detection** with port conflict checking
- **SSH connection testing** before saving configuration
- **Git repository validation** with branch checking

### Partial Configuration & Resume

- **State saving** - Resume interrupted setups
- **Partial rollback** - Save successful servers, skip failed ones
- **Troubleshooting tips** - Actionable advice for SSH failures
- **Setup logs** - Every run saved with timestamp

### Service Flexibility

- **Quick Mode** - All services on one VPS
- **Distributed Mode** - Separate IPs for each service
- **Shared Monitoring** - Install on existing web/db server
- **Custom hostnames** - Override default server names

## 📖 Usage Examples

### Example 1: Single VPS (Quick Mode)

```bash
./setup.sh production
```

**Configuration:**
- Environment: production
- Services: Web + Database + Monitoring
- VPS: All on one IP (different ports)
- Result: Perfect for small projects

### Example 2: Multi-Server with Load Balancing

```bash
./setup.sh production
```

**Configuration:**
- Environment: production
- Web servers: 3 instances
  - production-web-01 → 192.168.1.10
  - production-web-02 → 192.168.1.11
  - production-web-03 → 192.168.1.12
- Database: Dedicated server → 192.168.1.20
- Monitoring: On production-web-01
- Result: Nginx auto-configures load balancing

### Example 3: Add Servers to Existing Environment

```bash
./setup.sh production --add-servers
```

**Scenario:**
- Existing: web-01, web-02, web-03
- Add: 2 more web servers
- Result: Creates web-04, web-05 automatically

## 🔧 Configuration Details

### Generated Files

```
inventory/production/
├── hosts.yml              # Ansible inventory
├── hosts.yml.example      # Template copy
├── .setup_state.yml       # Resume state
└── .ssh_test_results.log  # SSH test results

group_vars/
├── all.yml                # Global configuration
├── webservers.yml         # Web server specific
└── dbservers.yml          # Database specific

setup_YYYYMMDD_HHMMSS.log  # Setup wizard log
```

### hosts.yml Structure

```yaml
---
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: 192.168.1.10
          ansible_user: deploy
          app_port: 3000
        production-web-02:
          ansible_host: 192.168.1.11
          ansible_user: deploy
          app_port: 3000
    
    dbservers:
      hosts:
        production-db-01:
          ansible_host: 192.168.1.20
          ansible_user: root
    
    monitoring:
      hosts:
        production-web-01:
          ansible_host: 192.168.1.10
          ansible_user: deploy
      vars:
        prometheus_targets:
          - targets:
              - '192.168.1.10:9100'
              - '192.168.1.11:9100'
              - '192.168.1.20:9100'
```

### Load Balancer Configuration (all.yml)

When you configure multiple web servers, the wizard automatically generates:

```yaml
load_balancer:
  enabled: true
  algorithm: least_conn
  backend_servers:
    - server production-web-01 192.168.1.10:3000 weight=1
    - server production-web-02 192.168.1.11:3000 weight=1
    - server production-web-03 192.168.1.12:3000 weight=1
  health_check:
    uri: "/health"
    interval: 10s
    timeout: 5s
```

## 🛡️ Validation & Safety

### IP Validation

```bash
# Valid IPs
192.168.1.1    ✓
10.0.0.1       ✓

# Invalid IPs
256.1.1.1      ✗ (octet > 255)
192.168.1      ✗ (incomplete)
192.168.1.1.1  ✗ (too many octets)
```

### Conflict Detection

```bash
# Scenario: Same IP different ports
web-01 → 192.168.1.10:3000  ✓ OK
web-02 → 192.168.1.10:3001  ✓ OK (different port)
web-03 → 192.168.1.10:3000  ✗ CONFLICT (same IP:port)
```

### SSH Connection Testing

The wizard tests SSH connectivity for each server:

```bash
→ Testing production-web-01 (192.168.1.10)...
  ✓ Connection successful
  ✓ Python3 detected: /usr/bin/python3

→ Testing production-web-02 (192.168.1.11)...
  ✗ Connection failed (timeout)
```

**On failure:**
- Shows troubleshooting steps
- Saves partial configuration
- Offers retry or skip options

## 🔄 Partial Configuration & Recovery

### When SSH Tests Fail

```bash
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ⚠ Partial Configuration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  ✓ production-web-01  (192.168.1.10)  Added
  ✓ production-web-02  (192.168.1.11)  Added
  ✗ production-web-03  (192.168.1.12)  SSH failed

? What do you want to do?
  1) Save partial configuration (only web-01, web-02)
  2) Show troubleshooting tips
  3) Cancel setup

Choice [1]: 2
```

### Troubleshooting Tips

The wizard provides actionable advice:

```bash
Failed server: production-web-03 (192.168.1.12)

Possible solutions:
  1. Check if SSH is running:
     ssh deploy@192.168.1.12
  
  2. Verify firewall allows SSH (port 22):
     sudo ufw status
  
  3. Check SSH key is added:
     ssh-add -l
  
  4. Try with password:
     ssh -o PreferredAuthentications=password deploy@192.168.1.12

Configuration saved at:
  inventory/production/.setup_state.yml

To resume setup:
  ./setup.sh production --resume
```

### Resume Interrupted Setup

```bash
# After fixing SSH issues
./setup.sh production --resume

# Wizard loads saved state and continues
```

## 🧪 Testing

### Validate Configuration

```bash
# Test setup wizard logic
./test_setup.sh

# Verify generated inventory
ansible-inventory -i inventory/production --list

# Test connectivity to all servers
ansible all -i inventory/production -m ping
```

### Dry Run Provision

```bash
# Check what would be installed
ansible-playbook playbooks/provision.yml -i inventory/production --check

# Run with verbose output
ansible-playbook playbooks/provision.yml -i inventory/production -vvv
```

## 📊 Next Steps After Setup

Once the wizard completes, follow these steps:

### 1. Verify Configuration

```bash
# Check generated files
cat inventory/production/hosts.yml
cat group_vars/all.yml

# Test Ansible connectivity
ansible all -i inventory/production -m ping
```

### 2. Provision Servers

```bash
# Install all required software
./deploy.sh provision production

# Or provision specific servers
ansible-playbook playbooks/provision.yml \
  -i inventory/production \
  --limit production-web-01,production-web-02
```

### 3. Deploy Application

```bash
# Deploy to all servers
./deploy.sh deploy production

# Monitor deployment
./deploy.sh status production
```

### 4. Configure SSL

```bash
# After DNS points to your servers
./configure-ssl.sh production
```

### 5. Access Services

```bash
# Application (load balanced)
http://your-domain.com

# Monitoring
http://monitoring-server-ip:9090  # Prometheus
http://monitoring-server-ip:3001  # Grafana
```

## 🎯 Best Practices

### SSH Key Management

1. **Generate dedicated keys** per environment
2. **Add to ssh-agent** for convenience
3. **Backup private keys** securely
4. **Use different keys** for production/staging

### Server Naming

1. **Use descriptive names** for custom hostnames
2. **Keep default naming** for consistency
3. **Document custom names** in your team wiki

### Multi-Server Deployment

1. **Start with 1 server** for testing
2. **Add servers gradually** as load increases
3. **Monitor performance** before scaling
4. **Use load balancer health checks**

### Environment Separation

```bash
# Recommended environments
./setup.sh production   # Live traffic
./setup.sh staging      # Pre-production testing
./setup.sh dev          # Development experiments
```

## 🐛 Troubleshooting

### Issue: SSH Key Not Working

```bash
# Check key permissions
chmod 600 ~/.ssh/your_key
chmod 644 ~/.ssh/your_key.pub

# Verify key is loaded
ssh-add -l

# Add key if missing
ssh-add ~/.ssh/your_key
```

### Issue: Git Repository Not Accessible

```bash
# Test repository access
git ls-remote YOUR_REPO_URL

# For GitHub, test SSH
ssh -T git@github.com

# For GitLab, test SSH
ssh -T git@gitlab.com

# Use HTTPS if SSH fails
https://github.com/user/repo.git
```

### Issue: Ansible Not Found

```bash
# Install Ansible
## Ubuntu/Debian
sudo apt update
sudo apt install ansible

## macOS
brew install ansible

## Python pip
pip3 install ansible
```

### Issue: Python Not Found on VPS

```bash
# SSH to VPS and install
ssh user@your-vps
sudo apt update
sudo apt install python3 python3-pip

# Verify installation
python3 --version
```

## 📚 Advanced Usage

### Custom Inventory Structure

You can manually edit generated files:

```bash
# Add custom variables per server
vim inventory/production/hosts.yml

# Add environment-specific settings
vim group_vars/all.yml
```

### Integration with CI/CD

```yaml
# .github/workflows/deploy.yml
- name: Setup infrastructure
  run: |
    # Use non-interactive mode with pre-saved config
    cp .ci/production-hosts.yml inventory/production/hosts.yml
    cp .ci/production-vars.yml group_vars/all.yml
    
- name: Deploy
  run: ./deploy.sh deploy production
```

### Custom Service Combinations

```bash
# Web only (no database)
./setup.sh web-cluster
# Select: Web servers only

# Database only (separate cluster)
./setup.sh db-cluster
# Select: Database only

# Monitoring for existing infrastructure
./setup.sh monitoring
# Select: Monitoring only
```

## 📞 Support

- **Documentation**: See `/docs` directory
- **Examples**: Check `QUICKSTART.md`
- **Logs**: Review `setup_*.log` files
- **State**: Check `.setup_state.yml` for resume info

## 🔄 Updates

The setup wizard automatically:
- Validates against current best practices
- Checks for required tools and versions
- Generates configuration compatible with latest Ansible
- Creates idempotent playbook configurations

---

**Ready to start?** Run `./setup.sh` and follow the wizard! 🚀
