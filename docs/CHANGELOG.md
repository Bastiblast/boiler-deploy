# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-09

### Added - Auto-Detection System ⭐

- **Automatic Framework Detection**: Detects Next.js, Nuxt.js, Express, Fastify, NestJS, and standard Node.js applications
- **Package Manager Detection**: Automatically identifies npm, pnpm, or yarn based on lockfiles
- **Smart Build Detection**: Determines if build step is needed from package.json scripts
- **Entry Point Detection**: Finds application entry point automatically
- **Adaptive PM2 Configuration**: Generates optimal PM2 config per framework (fork mode for Next.js/Nuxt, cluster for others)

### Added - Debian 13 Support

- Full compatibility with Debian 13 (Trixie)
- Modern GPG key management (replaced deprecated apt-key)
- Added cron to default packages
- Fixed PostgreSQL repository setup for Debian 13
- Fixed Grafana repository setup for Debian 13
- Added wait for apt lock mechanism

### Added - Monitoring Stack

- Prometheus 2.48.0 for metrics collection
- Grafana with pre-configured data sources
- Node Exporter for system metrics
- Automatic firewall configuration for monitoring ports
- Default dashboards support

### Added - Security Hardening

- UFW firewall with sensible defaults
- fail2ban for SSH protection
- SSH hardening (key-only auth, root login disabled after setup)
- Deploy user creation with sudo access
- Automated security updates
- Systemd journal retention limits

### Added - Documentation

- Comprehensive README with quick start
- Complete configuration guide
- Auto-detection system documentation
- Troubleshooting guide
- Real-world examples for all frameworks
- VPS-agnostic approach (works with any provider)

### Changed

- **Breaking**: Replaced `environment` variable with `app_environment` (Ansible reserved word conflict)
- **Breaking**: SSL disabled by default (enable when ready with real domain)
- Unified deployment script with clear commands
- Improved error messages and deployment feedback
- Better handling of Next.js devDependencies (full install for build)

### Fixed

- Next.js deployments now work correctly (fork mode, npm start)
- pnpm global installation added to nodejs role
- PM2 configuration for framework apps (Next.js, Nuxt.js)
- Monitoring ports accessible from outside (UFW rules)
- Build process for apps requiring devDependencies
- Entry point detection for various project structures
- PostgreSQL backup cron job (cron package added)

### Deprecated

- Provider-specific naming (Scaleway, Hostinger) in favor of generic VPS terminology

## [0.9.0] - 2025-11-08

### Added

- Initial Ansible playbooks structure
- Basic provisioning for web servers
- PostgreSQL 15 setup
- Node.js 20 LTS installation
- Nginx reverse proxy configuration
- PM2 process manager
- Basic security setup

### Known Issues

- Manual PM2 configuration required
- No framework auto-detection
- Limited to specific providers
- SSL configuration required manual setup

## Roadmap

### [1.1.0] - Planned

#### Features Under Consideration

- [ ] Docker support for containerized deployments
- [ ] Redis role for caching
- [ ] Automated database migrations
- [ ] Blue-green deployment strategy
- [ ] Canary deployments
- [ ] Health check endpoints configuration
- [ ] Custom Grafana dashboard provisioning
- [ ] Prometheus alerting rules
- [ ] Slack/Discord notifications
- [ ] Backup automation for application data

#### Improvements

- [ ] Support for Bun.js (new JavaScript runtime)
- [ ] Support for Deno applications
- [ ] TypeScript project detection
- [ ] Monorepo support (Nx, Turborepo)
- [ ] Environment-specific builds
- [ ] Cached dependencies for faster deployments
- [ ] Parallel deployments to multiple servers
- [ ] Deployment rollback with one command

#### Developer Experience

- [ ] Interactive CLI wizard for initial setup
- [ ] Deployment preview URLs
- [ ] Local development environment provisioning
- [ ] VS Code extension
- [ ] GitHub Actions integration examples
- [ ] GitLab CI integration examples

### [2.0.0] - Future

- Kubernetes support
- Multi-cloud deployment
- Infrastructure as Code (Terraform integration)
- Cost optimization recommendations
- Auto-scaling configuration
- CDN integration

## Migration Guides

### Migrating from 0.9.0 to 1.0.0

If you're upgrading from the previous version:

#### 1. Update Variable Names

```yaml
# Old
environment: production

# New
app_environment: production
```

#### 2. SSL Configuration

SSL is now disabled by default. To enable:

```yaml
# group_vars/webservers.yml
ssl_enabled: true  # Was true by default before
ssl_certbot_email: "your@email.com"
ssl_domains:
  - "yourdomain.com"
```

#### 3. Inventory Structure

If you used provider-specific inventory names, they still work but consider renaming:

```yaml
# Old
scaleway-web-01
hostinger-db-01

# Recommended
vps-web-01
vps-db-01
```

#### 4. Re-provision (Optional but Recommended)

To get all new features:

```bash
./deploy.sh provision
```

This will:
- Install pnpm globally
- Update monitoring stack
- Configure firewall for monitoring
- Add cron package
- Update PostgreSQL configuration

#### 5. No Action Required

Auto-detection works automatically on next deployment. No configuration changes needed.

## Version Support

| Version | Supported          | Notes |
|---------|--------------------|-------|
| 1.0.x   | ✅ Current         | Full support |
| 0.9.x   | ⚠️ Limited support | Security fixes only |
| < 0.9   | ❌ Unsupported     | Please upgrade |

## Compatibility Matrix

### Operating Systems

| OS | Version | Status |
|----|---------|--------|
| Debian | 13 (Trixie) | ✅ Tested |
| Debian | 12 (Bookworm) | ✅ Supported |
| Ubuntu | 22.04 LTS | ✅ Supported |
| Ubuntu | 20.04 LTS | ✅ Supported |
| Ubuntu | 24.04 LTS | ⚠️ Not tested |

### Node.js Versions

| Version | Status |
|---------|--------|
| 20.x LTS | ✅ Recommended |
| 18.x LTS | ✅ Supported |
| 16.x LTS | ⚠️ EOL soon |
| 22.x | ⚠️ Not tested |

### Frameworks

| Framework | Version | Status |
|-----------|---------|--------|
| Next.js | 13.x, 14.x, 15.x | ✅ Auto-detected |
| Nuxt.js | 3.x | ✅ Auto-detected |
| Express | 4.x | ✅ Auto-detected |
| Fastify | 4.x | ✅ Auto-detected |
| NestJS | 10.x | ✅ Auto-detected |

### Package Managers

| Package Manager | Version | Status |
|----------------|---------|--------|
| pnpm | 8.x, 9.x, 10.x | ✅ Auto-detected |
| yarn | 1.x, 3.x, 4.x | ✅ Auto-detected |
| npm | 9.x, 10.x | ✅ Auto-detected |

## Contributing

We welcome contributions! Please see:

- [Configuration Guide](CONFIGURATION.md) for adding new features
- [Auto-Detection Guide](AUTO_DETECTION.md) for adding framework support
- [Examples](EXAMPLES.md) for contributing examples

### How to Add Framework Support

1. Edit `roles/deploy-app/tasks/detect-app-type.yml`
2. Create `roles/deploy-app/templates/ecosystem.config.FRAMEWORK.js.j2`
3. Update `roles/deploy-app/tasks/main.yml`
4. Add example in `docs/EXAMPLES.md`
5. Test on real VPS
6. Submit PR

## Security

### Reporting Security Issues

Please report security vulnerabilities to: security@example.com

Do not open public issues for security vulnerabilities.

### Security Updates

Security patches are released as needed. Subscribe to releases to stay informed.

## Credits

Built with:
- Ansible
- PM2
- Nginx
- PostgreSQL
- Prometheus
- Grafana

Tested on real VPS deployments.

Special thanks to all contributors and early testers.

---

For detailed documentation, see [README.md](../README.md).
