# Pre-Deployment Checklist

## ☐ Initial Setup

### Ansible Environment
- [ ] Ansible installed (2.9+)
- [ ] Python 3 installed
- [ ] SSH client configured
- [ ] Ansible collections installed: `ansible-galaxy collection install -r requirements.yml`

### SSH Configuration
- [ ] SSH key pair generated (`ssh-keygen -t rsa -b 4096`)
- [ ] Public key available at `~/.ssh/id_rsa.pub` (or update path in config)
- [ ] SSH key copied to servers: `ssh-copy-id debian@SERVER_IP`
- [ ] Can SSH into servers without password

### Scaleway Servers
- [ ] VPS instances created on Scaleway
- [ ] Debian 11 or 12 installed
- [ ] Root or debian user access available
- [ ] Server IPs documented

## ☐ Configuration Files

### inventory/dev/hosts.yml
- [ ] Updated web server IPs
- [ ] Updated database server IPs
- [ ] Correct ansible_user set (debian)

### inventory/production/hosts.yml
- [ ] Updated web server IPs
- [ ] Updated database server IPs
- [ ] All server IPs are reachable

### group_vars/all.yml
- [ ] `app_name` set to your application name
- [ ] `app_repo` set to your GitHub repository URL
- [ ] `app_branch` configured (main/develop)
- [ ] `ssh_key_path` points to your public key
- [ ] `timezone` set to your timezone

### group_vars/webservers.yml
- [ ] `ssl_certbot_email` set to your email
- [ ] `ssl_domains` configured with your domain(s)
- [ ] Domain DNS A records point to server IPs
- [ ] Nginx settings adjusted if needed

### group_vars/dbservers.yml
- [ ] PostgreSQL user password changed from default
- [ ] Database name matches your app requirements
- [ ] PostgreSQL version is correct
- [ ] Performance settings adjusted for server size

## ☐ Application Requirements

### Your Node.js Application
- [ ] Repository is accessible (public or SSH key added)
- [ ] Has valid package.json
- [ ] Main file exists (index.js, app.js, or server.js)
- [ ] Application listens on PORT environment variable
- [ ] Health check endpoint exists (`/health`)
- [ ] Uses environment variables for configuration
- [ ] Database connection configured via env vars
- [ ] Handles SIGTERM for graceful shutdown

### Application Testing
- [ ] Application runs locally
- [ ] Database connection works
- [ ] All dependencies in package.json
- [ ] Build script (if needed) works
- [ ] No hardcoded configuration

## ☐ Pre-Deployment Tests

### Connectivity Tests
```bash
# Test SSH access
ssh debian@WEB_SERVER_IP
ssh debian@DB_SERVER_IP

# Test Ansible connectivity
ansible all -i inventory/dev -m ping

# Check inventory
ansible-inventory -i inventory/dev --list
```

- [ ] All servers reachable via SSH
- [ ] Ansible ping succeeds for all hosts
- [ ] Inventory file syntax is valid

### Dry Run
```bash
# Test playbook syntax
ansible-playbook playbooks/provision.yml --syntax-check

# Dry run (check mode)
ansible-playbook playbooks/provision.yml -i inventory/dev --check
```

- [ ] No syntax errors
- [ ] Dry run completes without errors

## ☐ Security Checklist

- [ ] Changed all default passwords
- [ ] SSH key authentication only (no passwords)
- [ ] Strong database password set
- [ ] UFW firewall will be enabled
- [ ] fail2ban will be configured
- [ ] Root login will be disabled
- [ ] Automatic updates will be enabled

## ☐ Backup & Recovery

- [ ] Understand backup schedule (daily at 2 AM)
- [ ] Know where backups are stored (`/var/backups/postgresql`)
- [ ] Tested manual backup procedure
- [ ] Have rollback plan ready

## ☐ Monitoring Setup

- [ ] Know Prometheus URL (http://SERVER_IP:9090)
- [ ] Know Grafana URL (http://SERVER_IP:3001)
- [ ] Grafana default credentials documented (admin/admin)
- [ ] Plan to change Grafana password after first login

## ☐ First Deployment

### Development Environment
```bash
# 1. Provision servers
./deploy.sh dev provision

# 2. Deploy application
./deploy.sh dev deploy

# 3. Verify deployment
curl http://DEV_SERVER_IP
curl http://DEV_SERVER_IP/health
```

- [ ] Provision completed successfully
- [ ] No errors during deployment
- [ ] Application accessible
- [ ] Health check returns 200 OK
- [ ] Database connection works
- [ ] Monitoring accessible

### Post-Deployment Verification
```bash
# Check PM2 status
ssh deploy@SERVER_IP
pm2 status
pm2 logs

# Check Nginx
sudo systemctl status nginx
sudo nginx -t

# Check PostgreSQL
sudo systemctl status postgresql
```

- [ ] PM2 shows app running
- [ ] No errors in PM2 logs
- [ ] Nginx is running
- [ ] PostgreSQL is running
- [ ] Firewall rules applied
- [ ] fail2ban is active

## ☐ Production Deployment

### Pre-Production
- [ ] Development deployment tested thoroughly
- [ ] All issues resolved
- [ ] Application stable in dev environment
- [ ] Database migrations prepared
- [ ] Backup of any existing data

### Production Deploy
```bash
# 1. Final configuration check
cat inventory/production/hosts.yml
cat group_vars/dbservers.yml

# 2. Provision production servers
./deploy.sh production provision

# 3. Deploy to production
./deploy.sh production deploy

# 4. Smoke tests
curl https://PRODUCTION_DOMAIN/health
```

- [ ] All production IPs correct
- [ ] SSL certificate obtained
- [ ] Application accessible via domain
- [ ] HTTPS working correctly
- [ ] Database connection working
- [ ] No errors in logs

## ☐ Post-Deployment

### Verification
- [ ] Application loads correctly
- [ ] Database queries working
- [ ] SSL/HTTPS working
- [ ] Monitoring data appearing in Prometheus
- [ ] Grafana showing metrics
- [ ] Logs are being written
- [ ] Backups scheduled

### Documentation
- [ ] Document server IPs
- [ ] Document login credentials
- [ ] Save Grafana admin password
- [ ] Document database credentials
- [ ] Update team on deployment

### Monitoring Setup
- [ ] Access Grafana (http://SERVER_IP:3001)
- [ ] Change admin password
- [ ] Add Prometheus data source
- [ ] Import Node.js dashboard
- [ ] Set up alerts (optional)

## ☐ Ongoing Maintenance

### Weekly
- [ ] Check application logs
- [ ] Review monitoring dashboards
- [ ] Verify backups are running

### Monthly
- [ ] Review security updates
- [ ] Check disk space
- [ ] Review performance metrics
- [ ] Test rollback procedure

## Emergency Contacts & Resources

- Scaleway Support: [Add URL]
- GitHub Repository: [Add URL]
- Team Contacts: [Add names/contacts]
- Documentation: See README.md, QUICKSTART.md, TROUBLESHOOTING.md

## Rollback Plan

If deployment fails:
```bash
# Rollback to previous release
./deploy.sh production rollback

# Check status
ssh deploy@SERVER_IP
pm2 status
pm2 logs
```

## Support Resources

- README.md - Full documentation
- QUICKSTART.md - Quick setup guide
- TROUBLESHOOTING.md - Common issues
- EXAMPLE_APP.md - Application structure guide

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Remember: Always test in dev environment before production!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
