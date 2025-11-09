# Quick Start Guide

Deploy your Node.js application to any VPS in 10 minutes with automatic framework detection and zero configuration.

## 📋 Prerequisites

Before you begin, ensure you have:

- **Ansible 2.9+** installed on your local machine
- **SSH access** to your VPS (root or sudo user)
- **Python 3.8+** on your VPS (usually pre-installed)
- **Git repository** with your Node.js application
- **SSH key** configured (recommended) or password access

### Install Ansible

**On macOS:**
```bash
brew install ansible
```

**On Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install ansible
```

**On other systems:**
```bash
pip3 install ansible
```

**Verify installation:**
```bash
ansible --version
```

## 🚀 Step-by-Step Setup

### 1. Install Ansible Collections

Install required Ansible collections:

```bash
cd boiler-deploy
ansible-galaxy collection install -r requirements.yml
```

### 2. Configure Your VPS Connection

Create your inventory file from the example:

```bash
cp inventory/production/hosts.yml.example inventory/production/hosts.yml
```

Edit `inventory/production/hosts.yml`:

```yaml
---
all:
  children:
    webservers:
      hosts:
        vps-web-01:
          ansible_host: XX.XX.XX.XX        # Your VPS IP address
          ansible_user: root               # Or your sudo user
          ansible_python_interpreter: /usr/bin/python3
    
    dbservers:
      hosts:
        vps-db-01:
          ansible_host: XX.XX.XX.XX        # Same or different VPS
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
    
    monitoring:
      hosts:
        vps-monitor-01:
          ansible_host: XX.XX.XX.XX        # Same or different VPS
          ansible_user: root
          ansible_python_interpreter: /usr/bin/python3
```

**Note**: For a single VPS setup, use the same IP for all three hosts.

### 3. Configure Your Application

Copy the variables template:

```bash
cp group_vars/all.yml.example group_vars/all.yml
```

Edit `group_vars/all.yml` with your application details:

```yaml
# Application Configuration
app_name: myapp
app_port: 3000
app_repo: "https://github.com/yourusername/your-app.git"
app_branch: "main"

# Deploy User (will be created automatically)
deploy_user: deploy
ssh_key_path: "~/.ssh/id_rsa.pub"

# Node.js Configuration (auto-detected)
nodejs_version: "20"  # LTS version
```

**That's it!** The system will auto-detect:
- Framework type (Next.js, Express, etc.)
- Package manager (npm, pnpm, yarn)
- Build requirements
- Entry point

### 4. Test Connection

Verify Ansible can connect to your VPS:

```bash
ansible all -i inventory/production -m ping
```

Expected output:
```
vps-web-01 | SUCCESS => {
    "changed": false,
    "ping": "pong"
}
```

### 5. Provision Your VPS

First-time setup installs all services (takes 5-10 minutes):

```bash
./deploy.sh provision
```

This installs:
- ✅ Node.js 20 LTS
- ✅ PostgreSQL 15
- ✅ Nginx
- ✅ PM2
- ✅ Prometheus
- ✅ Grafana
- ✅ Security (UFW + fail2ban)

### 6. Deploy Your Application

Deploy your application:

```bash
./deploy.sh deploy
```

The system will:
1. Auto-detect your framework
2. Clone your repository
3. Install dependencies (with correct package manager)
4. Build if needed (Next.js, Nuxt.js, etc.)
5. Configure PM2 optimally
6. Start your application

## ✅ Verification

### Check Application Status

```bash
./deploy.sh status
```

Or SSH to your VPS:

```bash
ssh deploy@your-vps-ip 'pm2 status'
```

### Access Your Application

- **Application**: `http://your-vps-ip`
- **Prometheus**: `http://your-vps-ip:9090`
- **Grafana**: `http://your-vps-ip:3001` (login: admin/admin)

### View Logs

```bash
ssh deploy@your-vps-ip 'pm2 logs'
```

## 🔄 Common Operations

### Update Application

Deploy latest changes:

```bash
./deploy.sh deploy
```

### Rollback to Previous Version

If something goes wrong:

```bash
./deploy.sh rollback
```

### Restart Application

```bash
ssh deploy@your-vps-ip 'pm2 restart all'
```

### View Application Logs

```bash
ssh deploy@your-vps-ip 'pm2 logs myapp --lines 100'
```

## 🎯 What Just Happened?

The deployment system automatically:

1. **Detected** your framework (Next.js, Express, etc.)
2. **Identified** your package manager (npm, pnpm, yarn)
3. **Installed** the right dependencies
4. **Built** your application if needed
5. **Configured** PM2 optimally for your framework
6. **Started** your application with monitoring

## 🔧 Framework-Specific Notes

### Next.js / Nuxt.js

- **PM2 Mode**: Fork (1 instance)
- **Dependencies**: Full install (devDependencies included for build)
- **Build**: Automatically runs `build` script
- **Start**: Uses `npm start`

### Express / Fastify / NestJS

- **PM2 Mode**: Cluster (multiple instances)
- **Dependencies**: Production only
- **Build**: Runs if `build` script exists
- **Start**: Uses detected entry point

## 🐛 Troubleshooting

### Can't Connect to VPS

```bash
# Test SSH connection
ssh root@your-vps-ip

# Check SSH key
ssh-add -l
```

### Application Not Starting

```bash
# View detailed logs
ssh deploy@your-vps-ip 'pm2 logs --err'

# Check PM2 status
ssh deploy@your-vps-ip 'pm2 status'
```

### Port Already in Use

Check if another service is using port 3000:

```bash
ssh deploy@your-vps-ip 'sudo netstat -tlnp | grep 3000'
```

For more troubleshooting, see [Troubleshooting Guide](docs/TROUBLESHOOTING.md).

## 📚 Next Steps

Now that your application is deployed:

1. **Configure SSL**: See [Configuration Guide](docs/CONFIGURATION.md#ssl-configuration)
2. **Set up Monitoring**: Configure Grafana dashboards
3. **Custom Domain**: Point your domain to your VPS IP
4. **Database**: Configure PostgreSQL connection
5. **Environment Variables**: Add secrets to `.env`

## 🆘 Need Help?

- **Configuration**: [docs/CONFIGURATION.md](docs/CONFIGURATION.md)
- **Auto-Detection**: [docs/AUTO_DETECTION.md](docs/AUTO_DETECTION.md)
- **Troubleshooting**: [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)
- **Examples**: [docs/EXAMPLES.md](docs/EXAMPLES.md)

---

**Congratulations!** Your application is now deployed with professional-grade infrastructure. 🎉
