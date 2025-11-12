# Quick Start - Test with Docker

Test your deployment workflow locally with Docker before deploying to real servers.

## 1. Setup Test VPS Container (1 minute)

```bash
./test-docker-vps.sh setup
```

This creates a **minimal Ubuntu container** with only SSH. Everything else (Node, Nginx, etc.) will be installed by Ansible.

## 2. Configure in Inventory Manager (2 minutes)

```bash
make run
```

**Create Environment:**
- Name: `test-docker`
- Mono Server: `Yes`
- IP: `127.0.0.1`

**Add Server:**
- Name: `test-web-01`
- IP: `127.0.0.1`
- SSH Port: `2222`
- SSH Key: `~/.ssh/boiler_test_rsa`
- Type: `web`
- Repository: `https://github.com/Bastiblast/ansible-next-test.git`
- App Port: `3000`
- Node Version: `20.x`

## 3. Deploy (5-10 minutes)

In the "Work with Inventory" menu:

1. **Validate** - Check all settings ✓
2. **Provision** - Install Node.js, Nginx, UFW, Fail2ban, etc.
3. **Deploy** - Deploy your application
4. **Verify** - Check status

## 4. Access Your App

```bash
curl http://localhost:8080
```

## Cleanup

```bash
./test-docker-vps.sh cleanup
```

## Troubleshooting

```bash
# Check status
./test-docker-vps.sh status

# SSH into container
./test-docker-vps.sh ssh

# View logs
docker logs boiler-test-vps
```

---

**What's Installed:**
- ✗ Container: Only SSH + Python3 (minimal)
- ✓ Provision: Node.js, Nginx, UFW, Fail2ban, PM2
- ✓ Deploy: Your app + dependencies

See [TEST_ENVIRONMENT.md](TEST_ENVIRONMENT.md) for detailed documentation.
