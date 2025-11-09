# Troubleshooting Guide

## Common Issues and Solutions

### 1. SSH Connection Failures

#### Error: "Permission denied (publickey)"
```bash
# Solution 1: Ensure your key is in ssh-agent
ssh-add ~/.ssh/id_rsa

# Solution 2: Specify the key explicitly
ansible-playbook playbooks/provision.yml -i inventory/dev --private-key ~/.ssh/id_rsa

# Solution 3: Copy key to server
ssh-copy-id -i ~/.ssh/id_rsa.pub debian@YOUR_SERVER_IP
```

#### Error: "Host key verification failed"
```bash
# Remove old host key
ssh-keygen -R YOUR_SERVER_IP

# Or disable strict checking temporarily
export ANSIBLE_HOST_KEY_CHECKING=False
```

### 2. Application Deployment Issues

#### App won't start after deployment
```bash
# Check PM2 status
ssh deploy@YOUR_SERVER_IP
pm2 status
pm2 logs

# Common causes:
# - Wrong main file in ecosystem.config.js
# - Missing dependencies
# - Port already in use
# - Environment variables not set
```

#### Fix: Update ecosystem.config.js
Edit `roles/deploy-app/templates/ecosystem.config.js.j2`:
```javascript
script: './server.js',  // Change to your main file
```

#### Check application logs
```bash
ssh deploy@YOUR_SERVER_IP
tail -f /var/www/myapp/shared/logs/pm2-error.log
tail -f /var/www/myapp/shared/logs/pm2-out.log
```

### 3. Database Connection Issues

#### Can't connect to PostgreSQL from web server
```bash
# Test connection
ssh deploy@YOUR_WEB_SERVER_IP
psql -h DB_SERVER_IP -U myapp_user -d myapp_dev

# Check UFW rules on DB server
ssh debian@YOUR_DB_SERVER_IP
sudo ufw status

# Check PostgreSQL is listening
sudo netstat -tlnp | grep 5432
```

#### Fix: Allow web server IP
```bash
# On database server
sudo ufw allow from WEB_SERVER_IP to any port 5432
```

### 4. Nginx Issues

#### 502 Bad Gateway
```bash
# Check if app is running
ssh deploy@YOUR_SERVER_IP
pm2 status

# Check Nginx error logs
sudo tail -f /var/log/nginx/error.log

# Test Nginx config
sudo nginx -t

# Restart services
sudo systemctl restart nginx
pm2 restart all
```

#### SSL Certificate Issues
```bash
# Check if domain resolves
nslookup yourdomain.com

# Manually obtain certificate
sudo certbot --nginx -d yourdomain.com

# Check certificate status
sudo certbot certificates
```

### 5. Ansible Playbook Failures

#### "unreachable" errors
```bash
# Test basic connectivity
ansible all -i inventory/dev -m ping

# Run with verbose output
ansible-playbook playbooks/provision.yml -i inventory/dev -vvv
```

#### "Package not found" errors
```bash
# Update apt cache first
ansible all -i inventory/dev -m apt -a "update_cache=yes" --become
```

#### Permission denied on become
```bash
# Ensure user has sudo access
ssh debian@YOUR_SERVER_IP
sudo -l

# If needed, add to sudoers
echo "debian ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/debian
```

### 6. Monitoring Not Working

#### Prometheus not scraping targets
```bash
# Check Prometheus UI: http://YOUR_SERVER_IP:9090/targets

# Ensure Node Exporter is running
sudo systemctl status node_exporter

# Check firewall
sudo ufw status
```

#### Grafana not accessible
```bash
# Check service status
sudo systemctl status grafana-server

# View logs
sudo journalctl -u grafana-server -f

# Restart service
sudo systemctl restart grafana-server
```

### 7. Performance Issues

#### High memory usage
```bash
# Check PM2 memory limit
pm2 list

# Update in group_vars/all.yml:
pm2_max_memory: "1G"  # Increase if needed

# Deploy changes
ansible-playbook playbooks/deploy.yml -i inventory/production
```

#### Database slow
```bash
# Tune PostgreSQL settings in group_vars/dbservers.yml:
postgresql_shared_buffers: "512MB"     # Increase
postgresql_effective_cache_size: "2GB"  # Increase

# Re-run provision
ansible-playbook playbooks/provision.yml -i inventory/production --tags postgresql
```

### 8. Backup Issues

#### Backups not running
```bash
# Check cron job
ssh debian@YOUR_DB_SERVER_IP
sudo -u postgres crontab -l

# Manually run backup
sudo -u postgres /usr/local/bin/backup_postgres.sh

# Check backup logs
tail /var/log/postgresql_backup.log
```

### 9. Rollback Not Working

#### No previous release found
This happens when it's your first deployment. You need at least 2 deployments before rollback works.

```bash
# Check releases directory
ssh deploy@YOUR_SERVER_IP
ls -la /var/www/myapp/releases/
```

### 10. UFW Blocking Necessary Traffic

#### Temporarily disable UFW for testing
```bash
ssh debian@YOUR_SERVER_IP
sudo ufw disable

# Test your application
# Then re-enable
sudo ufw enable
```

#### Add custom rules
```bash
# Allow specific port
sudo ufw allow 8080/tcp

# Allow from specific IP
sudo ufw allow from 192.168.1.100
```

## Debug Mode

Run playbooks in check mode (dry-run):
```bash
ansible-playbook playbooks/deploy.yml -i inventory/dev --check
```

Run with maximum verbosity:
```bash
ansible-playbook playbooks/deploy.yml -i inventory/dev -vvvv
```

## Still Having Issues?

1. Check all IP addresses in inventory files
2. Verify all passwords are set correctly
3. Ensure your Node.js app works locally
4. Check server resources (RAM, disk space)
5. Review logs on the server

## Getting Logs

```bash
# Application logs
ssh deploy@YOUR_SERVER_IP
pm2 logs

# Nginx logs
sudo tail -f /var/log/nginx/error.log
sudo tail -f /var/log/nginx/access.log

# System logs
sudo journalctl -xe
sudo journalctl -u nginx -f
sudo journalctl -u postgresql -f
```
