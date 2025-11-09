# Ansible Deployment for Node.js Applications

🚀 Universal deployment system for Node.js applications on any VPS provider with intelligent auto-detection, zero-downtime deployments, and built-in monitoring.

## ✨ Features

- **🤖 Auto-Detection**: Automatically detects Next.js, Nuxt.js, Express, Fastify, NestJS, and vanilla Node.js applications
- **📦 Multi Package Manager**: Supports npm, pnpm, and yarn with automatic detection
- **🔄 Zero-Downtime Deployments**: Rolling deployments with automatic rollback on failure
- **📊 Built-in Monitoring**: Prometheus + Grafana + Node Exporter pre-configured
- **🔒 Security Hardening**: UFW firewall, fail2ban, SSH key authentication, and automated security updates
- **⚡ One-Command Deployment**: Simple CLI for provisioning and deploying
- **🎯 PM2 Process Management**: Automatic PM2 configuration optimized per framework
- **🔧 Smart Build System**: Detects build requirements and runs them automatically

## 🎯 Supported Technologies

### Frameworks
- **Next.js** (auto-detected, optimized PM2 config)
- **Nuxt.js** (auto-detected, optimized PM2 config)
- **Express** (auto-detected)
- **Fastify** (auto-detected)
- **NestJS** (auto-detected)
- **Node.js** (standard applications)

### Package Managers
- **pnpm** (auto-detected via pnpm-lock.yaml)
- **yarn** (auto-detected via yarn.lock)
- **npm** (auto-detected via package-lock.json)

### VPS Providers
Works with any VPS provider:
- DigitalOcean
- Vultr
- Linode
- OVH
- Hetzner
- And any other VPS with SSH access

### Operating Systems
- Debian 12 (Bookworm)
- Debian 13 (Trixie) ✅ Tested
- Ubuntu 20.04 LTS
- Ubuntu 22.04 LTS

## 🚀 Quick Start

Get your application deployed in 10 minutes. See the [Quick Start Guide](QUICKSTART.md) for detailed instructions.

```bash
# 1. Install dependencies
ansible-galaxy collection install -r requirements.yml

# 2. Configure your VPS
cp inventory/production/hosts.yml.example inventory/production/hosts.yml
# Edit with your VPS IP and settings

# 3. Set your application details
cp group_vars/all.yml.example group_vars/all.yml
# Edit with your Git repository and configuration

# 4. Deploy
./deploy.sh provision production  # First time: install all services
./deploy.sh deploy production     # Deploy your application
```

## 📚 Documentation

- **[Quick Start Guide](QUICKSTART.md)** - Get up and running in 10 minutes
- **[Configuration Guide](docs/CONFIGURATION.md)** - Complete configuration reference
- **[SSL Setup Guide](docs/SSL_SETUP.md)** - Configure HTTPS with Let's Encrypt
- **[Auto-Detection System](docs/AUTO_DETECTION.md)** - How the auto-detection works
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Examples](docs/EXAMPLES.md)** - Real-world application examples
- **[Changelog](docs/CHANGELOG.md)** - Version history and changes

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Your VPS Server                   │
├─────────────────────────────────────────────────────┤
│  Nginx (Reverse Proxy)                              │
│    ↓                                                 │
│  PM2 (Process Manager)                              │
│    ↓                                                 │
│  Your Node.js App (Auto-detected & Configured)      │
├─────────────────────────────────────────────────────┤
│  PostgreSQL 15 (Database)                           │
│  Prometheus (Metrics Collection)                    │
│  Grafana (Visualization)                            │
│  Node Exporter (System Metrics)                     │
├─────────────────────────────────────────────────────┤
│  Security: UFW + fail2ban + SSH hardening          │
└─────────────────────────────────────────────────────┘
```

## 🔧 What Gets Installed

**Web Stack:**
- Nginx (reverse proxy)
- Node.js 20 LTS
- PM2 (process manager)

**Database:**
- PostgreSQL 15

**Monitoring:**
- Prometheus (metrics)
- Grafana (dashboards)
- Node Exporter (system metrics)

**Security:**
- UFW (firewall)
- fail2ban (intrusion prevention)
- Automated security updates
- SSH hardening (key-only auth, no root login)

## 🎯 How It Works

1. **Auto-Detection**: The system reads your `package.json` to detect:
   - Framework type (Next.js, Express, etc.)
   - Package manager (pnpm, yarn, npm)
   - Build requirements
   - Entry point

2. **Smart Installation**: Installs dependencies using the correct package manager:
   - Full dependencies for Next.js/Nuxt.js (build needed)
   - Production-only for other frameworks

3. **Optimized PM2 Config**: Generates PM2 configuration based on framework:
   - Fork mode for Next.js/Nuxt.js (framework handles scaling)
   - Cluster mode for Express/Fastify/NestJS

4. **Zero-Downtime Deploy**: 
   - Keeps last 5 releases
   - Symlink swap for instant rollback
   - Health checks before switching

## 📊 Monitoring

Access your monitoring dashboards:

- **Prometheus**: `http://your-vps-ip:9090`
- **Grafana**: `http://your-vps-ip:3001` (default: admin/admin)
- **Node Exporter**: `http://your-vps-ip:9100`

## 🔐 Security

Security is enabled by default:

- **Firewall (UFW)**: Only necessary ports open
- **fail2ban**: Automatic IP ban after failed login attempts
- **SSH Hardening**: 
  - Key-only authentication
  - Root login disabled
  - Deploy user with sudo access
- **Automated Updates**: Security patches applied automatically

## 🛠️ Commands

### Deployment Script

Unified deployment script with syntax: `./deploy.sh ACTION [ENVIRONMENT]`

```bash
# Provisioning (first time setup)
./deploy.sh provision production

# Deploy application
./deploy.sh deploy production
# or simply (production is default)
./deploy.sh deploy

# Quick update (pull latest + restart)
./deploy.sh update production

# Rollback to previous version
./deploy.sh rollback production

# Dry-run check (no changes)
./deploy.sh check production

# Check application status
./deploy.sh status production

# View logs
ssh deploy@your-vps-ip 'pm2 logs'
```

### SSL Configuration

Interactive SSL setup with Let's Encrypt:

```bash
./configure-ssl.sh
```

Automatically configures HTTPS with:
- Domain validation
- Certificate installation
- Auto-renewal setup
- Nginx HTTPS configuration

See [SSL Setup Guide](docs/SSL_SETUP.md) for details.

### Health Check

```bash
./health_check.sh production
```

## 📦 Project Structure

```
boiler-deploy/
├── playbooks/          # Ansible playbooks
│   ├── provision.yml   # Full server setup
│   ├── deploy.yml      # Application deployment
│   ├── update.yml      # Quick updates
│   └── rollback.yml    # Version rollback
├── roles/              # Ansible roles
│   ├── common/         # Base system setup
│   ├── postgresql/     # Database
│   ├── nodejs/         # Node.js + PM2
│   ├── nginx/          # Reverse proxy
│   ├── monitoring/     # Prometheus + Grafana
│   ├── security/       # Firewall + fail2ban
│   └── deploy-app/     # Application deployment (with auto-detection)
├── inventory/          # Server configurations
│   ├── production/     # Production environment (default)
│   └── dev/            # Development environment
├── group_vars/         # Configuration variables
├── docs/               # Documentation
│   ├── CONFIGURATION.md
│   ├── SSL_SETUP.md
│   ├── AUTO_DETECTION.md
│   ├── TROUBLESHOOTING.md
│   ├── EXAMPLES.md
│   └── CHANGELOG.md
├── deploy.sh           # Unified deployment script
├── configure-ssl.sh    # SSL configuration helper
└── health_check.sh     # Health check script
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/boiler-deploy/issues)
- **Documentation**: [docs/](docs/)
- **Examples**: [docs/EXAMPLES.md](docs/EXAMPLES.md)

## ⭐ Acknowledgments

Built with Ansible, tested on real VPS deployments, designed for simplicity and reliability.

---

**Ready to deploy?** Start with the [Quick Start Guide](QUICKSTART.md) →
