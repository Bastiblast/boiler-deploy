# SSL Setup Guide

Complete guide for configuring HTTPS with Let's Encrypt on your deployed Node.js application.

## Quick Start

```bash
./configure-ssl.sh
```

The interactive script will guide you through the entire SSL setup process.

## Prerequisites

Before running the SSL configuration:

1. **Domain Configuration**
   - You must own a domain name
   - Domain DNS must be configurable

2. **DNS Setup**
   - Add A record pointing to your VPS IP
   - Wait for DNS propagation (usually 5-30 minutes)

3. **Server Requirements**
   - Application already deployed
   - Port 80 accessible (required for Let's Encrypt validation)
   - Port 443 will be opened automatically

4. **Email Address**
   - Valid email for Let's Encrypt notifications
   - Used for certificate expiry reminders

## DNS Configuration Examples

### For Primary Domain

```
Type: A
Name: @
Value: XX.XX.XX.XX (your VPS IP)
TTL: 3600
```

### For www Subdomain

```
Type: A
Name: www
Value: XX.XX.XX.XX (your VPS IP)
TTL: 3600
```

Or use CNAME:

```
Type: CNAME
Name: www
Value: yourdomain.com
TTL: 3600
```

### For API Subdomain

```
Type: A
Name: api
Value: XX.XX.XX.XX (your VPS IP)
TTL: 3600
```

## Script Workflow

### 1. Detection Phase

The script automatically detects:
- Deployed applications
- Server IP addresses
- Current environments (production, dev)
- Existing SSL configuration

```
→ Detecting deployed applications...
✓ Found application: myapp
✓ Found server: 72.61.146.126 (production)
ℹ SSL Status: Disabled
```

### 2. Environment Selection

If multiple environments exist, you'll select which to configure:

```
? Select environment [1-2]: 1
  1) production - 72.61.146.126
  2) dev - 142.93.1.1
```

### 3. Domain Configuration

**Primary Domain:**
```
? Primary domain (e.g., myapp.com): myportfolio.com
```

**www Subdomain:**
```
? Add www subdomain? (www.myportfolio.com) (Y/n): y
```

**Additional Domains:**
```
? Add another domain? (y/N): y
? Additional domain: api.myportfolio.com
? Add another domain? (y/N): n
```

Result:
```
ℹ Domains to configure: myportfolio.com www.myportfolio.com api.myportfolio.com
```

### 4. Email Configuration

```
? Email for Let's Encrypt notifications: admin@myportfolio.com
```

This email receives:
- Certificate expiry notifications
- Important updates from Let's Encrypt
- Renewal failure alerts

### 5. Redirect Configuration

```
? Enable HTTP → HTTPS redirect? (Y/n): y
```

If enabled:
- `http://yourdomain.com` → `https://yourdomain.com`
- Automatic 301 permanent redirect
- Recommended for production

### 6. DNS Validation

The script checks DNS configuration:

**If DNS is configured correctly:**
```
→ Checking DNS for myportfolio.com...
✓ myportfolio.com → 72.61.146.126
→ Checking DNS for www.myportfolio.com...
✓ www.myportfolio.com → 72.61.146.126
```

**If DNS issues detected:**
```
→ Checking DNS for myportfolio.com...
! No A record found for myportfolio.com
ℹ Please add: myportfolio.com → 72.61.146.126

? Continue anyway? (Certificate may fail) (y/N): n
→ Waiting for DNS propagation...
Press Enter when DNS is configured...
```

### 7. Configuration Preview

Review all settings before applying:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CONFIGURATION SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Environment:     production
Server:          72.61.146.126
Application:     myapp
Domains:         myportfolio.com www.myportfolio.com
Email:           admin@myportfolio.com
Redirect HTTP:   Yes

Changes to apply:
  • Update group_vars/webservers.yml
  • Install SSL certificate via Let's Encrypt
  • Configure Nginx for HTTPS
  • Enable certificate auto-renewal
  • Open port 443 in firewall

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

? Proceed with SSL configuration? (Y/n): y
```

### 8. Execution Phase

The script performs:

```
→ Backing up current configuration...
✓ Backup saved to .ssl_backup/20251109_120000

→ Updating Ansible variables...
✓ Variables updated

→ Applying SSL configuration...
ℹ Running Ansible playbook (this may take a few minutes)...
[Ansible output...]
✓ SSL configuration applied

→ Verifying SSL configuration...
ℹ Testing HTTPS access to myportfolio.com...
✓ HTTPS is working!
```

### 9. Success Report

```
╔═══════════════════════════════════════════════════════════════════════╗
║  ✓ SSL Configuration Complete!
╚═══════════════════════════════════════════════════════════════════════╝

Certificate Details:
  • Domains: myportfolio.com www.myportfolio.com
  • Issuer: Let's Encrypt
  • Email: admin@myportfolio.com
  • Auto-renewal: Enabled (via cron)

Access your application:
  • https://myportfolio.com
  • https://www.myportfolio.com

Next Steps:
  • Test your application over HTTPS
  • Update any hardcoded HTTP URLs to HTTPS
  • Certificate will auto-renew 30 days before expiry

ℹ Backup saved in: .ssl_backup
ℹ To rollback: restore files from backup and run provision again
```

## Updating Existing SSL

If SSL is already configured, the script detects it and offers to update:

```
ℹ SSL is already configured
ℹ Current domains: oldomain.com www.olddomain.com
ℹ Current email: old@email.com

? Update SSL configuration? (Y/n): y
```

You can then:
- Change domains
- Add new domains
- Update email address
- Reconfigure settings

## Certificate Management

### Auto-Renewal

Let's Encrypt certificates expire after 90 days. Auto-renewal is configured via cron:

```bash
# Runs daily at 2:00 AM
0 2 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'
```

Check renewal cron:
```bash
ssh deploy@your-vps-ip 'crontab -l | grep certbot'
```

### Manual Renewal

Force certificate renewal:

```bash
ssh deploy@your-vps-ip 'sudo certbot renew --force-renewal'
```

### Check Certificate Status

View certificate details:

```bash
ssh deploy@your-vps-ip 'sudo certbot certificates'
```

Example output:
```
Certificate Name: myportfolio.com
  Domains: myportfolio.com www.myportfolio.com
  Expiry Date: 2026-02-07 12:00:00+00:00 (VALID: 89 days)
  Certificate Path: /etc/letsencrypt/live/myportfolio.com/fullchain.pem
  Private Key Path: /etc/letsencrypt/live/myportfolio.com/privkey.pem
```

## Troubleshooting

### DNS Not Propagated

**Issue**: DNS check fails

**Solution**:
1. Wait 5-30 minutes for DNS propagation
2. Check DNS: `dig yourdomain.com`
3. Verify A record points to correct IP
4. Use [DNS Checker](https://dnschecker.org) to verify globally

### Port 80 Blocked

**Issue**: Let's Encrypt validation fails

**Solution**:
```bash
# On VPS, check firewall
ssh deploy@your-vps-ip 'sudo ufw status'

# If port 80 not allowed
ssh deploy@your-vps-ip 'sudo ufw allow 80/tcp'
```

### Certificate Acquisition Failed

**Issue**: Certbot fails to obtain certificate

**Possible causes**:
1. **DNS not configured**: Domain doesn't point to server
2. **Port 80 blocked**: Firewall blocking HTTP
3. **Rate limit**: Let's Encrypt has rate limits (5 certificates per week)

**Solution for rate limit**:
Wait 1 week, or use staging certificates for testing:

Edit the script temporarily:
```bash
# In apply_ssl_config(), add --staging flag
certbot --nginx --staging -d yourdomain.com ...
```

### Certificate Not Trusted

**Issue**: Browser shows "Not Secure"

**Causes**:
1. Using staging certificate (for testing)
2. Certificate not yet propagated (wait a few minutes)
3. Mixed content (HTTP resources on HTTPS page)

**Solution**:
- For staging: Remove `--staging` flag and rerun
- For mixed content: Update all resources to HTTPS

### HTTP Still Works (No Redirect)

**Issue**: Site accessible via HTTP and HTTPS

**Check Nginx config**:
```bash
ssh deploy@your-vps-ip 'cat /etc/nginx/sites-enabled/myapp'
```

Should contain:
```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$host$request_uri;
}
```

**Fix**: Rerun configure-ssl.sh with redirect enabled

### Auto-Renewal Not Working

**Check cron**:
```bash
ssh deploy@your-vps-ip 'crontab -l'
```

**Test renewal**:
```bash
ssh deploy@your-vps-ip 'sudo certbot renew --dry-run'
```

**Check logs**:
```bash
ssh deploy@your-vps-ip 'sudo cat /var/log/letsencrypt/letsencrypt.log'
```

## Rollback

If you need to disable SSL:

### 1. Restore Backup

```bash
# List backups
ls -la .ssl_backup/

# Restore from backup
cp .ssl_backup/20251109_120000/webservers.yml group_vars/webservers.yml
```

### 2. Update Variables

Or manually edit `group_vars/webservers.yml`:

```yaml
ssl_enabled: false
ssl_certbot_email: "admin@example.com"
ssl_domains:
  - "example.com"
```

### 3. Reapply Configuration

```bash
./deploy.sh provision
```

## Advanced Usage

### Multiple Domains

Configure multiple separate domains (different apps):

```bash
# First app
./configure-ssl.sh
# Configure domain1.com

# Second app (update app_name first)
vim group_vars/all.yml  # Change app_name
./configure-ssl.sh
# Configure domain2.com
```

### Custom Certificate

If you have your own SSL certificate:

1. Copy certificate files to server:
```bash
scp mycert.crt deploy@vps:/etc/ssl/certs/
scp mykey.key deploy@vps:/etc/ssl/private/
```

2. Update Nginx config manually or create custom template

3. Skip `configure-ssl.sh` script

## Security Best Practices

1. **Use Strong Domains**: Avoid easily guessable domains
2. **Keep Email Active**: Monitor Let's Encrypt notifications
3. **Enable HSTS**: Add HTTP Strict Transport Security header
4. **Monitor Expiry**: Check certificate status regularly
5. **Test Renewals**: Run `--dry-run` periodically

## Let's Encrypt Limits

Be aware of rate limits:

- **50 certificates** per registered domain per week
- **5 duplicate certificates** per week
- **300 new orders** per account per 3 hours

For development/testing, use staging environment (no limits).

## Next Steps

After SSL configuration:

1. ✅ Test application over HTTPS
2. ✅ Update any HTTP URLs to HTTPS
3. ✅ Update OAuth redirect URIs (if applicable)
4. ✅ Update API endpoints in frontend
5. ✅ Set up monitoring for certificate expiry
6. ✅ Configure HSTS header (optional)
7. ✅ Test HTTP → HTTPS redirect

## Related Documentation

- [Configuration Guide](CONFIGURATION.md) - General configuration
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues
- [Examples](EXAMPLES.md) - Real-world setups

---

For issues or questions, see [Troubleshooting Guide](TROUBLESHOOTING.md).
