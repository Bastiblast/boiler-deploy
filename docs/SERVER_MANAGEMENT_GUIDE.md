# 🖥️ Server Management Guide

## Overview

The Inventory Manager now includes complete server management capabilities, allowing you to add, edit, and delete servers for each environment.

## 🎯 Features

### Environment Selection
- View all existing environments
- See server count for each environment
- Quick navigation with keyboard

### Server Management
- Add new servers
- Edit existing servers
- Delete servers
- View server details in a table
- Real-time validation
- Conflict detection

### Server Configuration
- **Name**: Auto-generated or custom
- **IP Address**: With format validation
- **SSH Port**: Default 22
- **App Port**: Application port (e.g., 3000, 3001)
- **SSH User**: Default root
- **SSH Key Path**: Path to SSH private key
- **Type**: Web, Database, or Monitoring

---

## 📋 Workflows

### 1. Create a New Environment

```
Main Menu → Create new environment
```

1. Enter environment name (e.g., `production`)
2. Configure Git repository
3. Set Node.js version and app port
4. Toggle services (Web/Database/Monitoring)
5. Press Enter to create

**Result**: Empty environment ready for servers

---

### 2. Add Servers to Environment

```
Main Menu → Manage existing environment → Select environment
```

**In Server Manager:**
- Press **`a`** to add a server

**In Server Form:**
1. **Name**: Leave empty for auto-generation (e.g., `production-web-01`)
2. **IP**: `192.168.1.10`
3. **SSH Port**: `22` (default)
4. **App Port**: `3000`
5. **SSH User**: `root`
6. **SSH Key**: `~/.ssh/id_rsa`
7. **Type**: Use ←→ arrows to select (Web/Database/Monitoring)
8. Press **Enter** to save

**Auto-naming Pattern:**
- `{env}-web-01`, `{env}-web-02`, etc. for Web
- `{env}-db-01` for Database
- `{env}-monitoring-01` for Monitoring

---

### 3. Edit a Server

```
Server Manager → Navigate to server → Press 'e'
```

1. Modify any field
2. Press Enter to save
3. Changes are validated and saved

**Validations:**
- IP format check
- Port range (1-65535)
- SSH key file existence
- No IP:Port conflicts

---

### 4. Delete a Server

```
Server Manager → Navigate to server → Press 'd'
```

- Server is immediately removed
- Environment is auto-saved
- Cursor adjusts to valid position

---

### 5. View Generated Configuration

```
Server Manager → Press 'g'
```

Shows a summary with:
- Environment name
- Enabled services
- Configuration details
- All servers and their details

---

## 🎨 UI Elements

### Server Manager Screen

```
╔════════════════════════════════════════════════════════════╗
║  🖥️  Manage Environment: production                        ║
╚════════════════════════════════════════════════════════════╝

Services: Web, Database

Servers (3):

  Name                IP              Port    Type    Status
  ──────────────────────────────────────────────────────────
▶ production-web-01   192.168.1.10    3000    web     ⚠ Not tested
  production-web-02   192.168.1.11    3001    web     ⚠ Not tested
  production-db-01    192.168.1.20    5432    db      ⚠ Not tested


[a] Add  [e] Edit  [d] Delete  [s] Save  [g] Summary  [Esc] Back
```

### Server Form

```
╔════════════════════════════════════════════════════════════╗
║  ➕ Add New Server                                         ║
╚════════════════════════════════════════════════════════════╝

▶ Server name:
  production-web-03

  IP address:
  192.168.1.12

  SSH port:
  22

  Application port:
  3002

  SSH user:
  root

  SSH key path:
  ~/.ssh/id_rsa

  Server type:
  [Web]  Database   Monitoring

[Tab/↑↓] Navigate  [←→] Change type  [Enter] Save  [Esc] Cancel
```

---

## ⌨️ Keyboard Shortcuts

### Environment Selector
| Key | Action |
|-----|--------|
| `↑↓` or `j/k` | Navigate environments |
| `Enter` | Select environment |
| `Esc` or `q` | Back to main menu |

### Server Manager
| Key | Action |
|-----|--------|
| `↑↓` or `j/k` | Navigate servers |
| `a` | Add new server |
| `e` | Edit selected server |
| `d` | Delete selected server |
| `s` | Save environment |
| `g` | Generate summary |
| `Esc` | Back to main menu |

### Server Form
| Key | Action |
|-----|--------|
| `Tab` or `↓` | Next field |
| `Shift+Tab` or `↑` | Previous field |
| `←→` | Change server type |
| `Enter` | Save server |
| `Esc` | Cancel |

---

## 🔍 Validation Rules

### IP Address
- Must be valid IPv4 format
- Examples: `192.168.1.10`, `10.0.0.1`

### Ports
- SSH Port: 1-65535 (typically 22)
- App Port: 1-65535 (typically 3000-9999)

### Server Name
- Auto-generated if empty
- Pattern: `{env}-{type}-{number}`
- Must be unique within environment

### SSH Key Path
- File must exist
- Expands `~` to home directory
- Validates file accessibility

### Conflicts
- No two servers can have same IP:Port combination
- Warns if potential conflict detected

---

## 📁 Generated Files

After adding servers, the following files are auto-generated:

### `inventory/{env}/hosts.yml`

```yaml
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: 192.168.1.10
          ansible_user: root
          ansible_port: 22
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          ansible_become: true
          app_port: 3000
        production-web-02:
          ansible_host: 192.168.1.11
          ansible_user: root
          ansible_port: 22
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          ansible_become: true
          app_port: 3001
    dbservers:
      hosts:
        production-db-01:
          ansible_host: 192.168.1.20
          ansible_user: root
          ansible_port: 22
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          ansible_become: true
```

### `inventory/{env}/config.yml`

```yaml
name: production
services:
  web: true
  database: true
  monitoring: false
config:
  app_name: production-app
  app_repo: https://github.com/user/repo.git
  app_branch: main
  nodejs_version: "20"
  app_port: "3000"
  deploy_user: root
  timezone: Europe/Paris
servers:
  - name: production-web-01
    ip: 192.168.1.10
    port: 22
    ssh_user: root
    ssh_key_path: ~/.ssh/id_rsa
    type: web
    app_port: 3000
    ansible_become: true
  - name: production-web-02
    ip: 192.168.1.11
    port: 22
    ssh_user: root
    ssh_key_path: ~/.ssh/id_rsa
    type: web
    app_port: 3001
    ansible_become: true
```

---

## 💡 Tips & Best Practices

### Naming Conventions

**Good:**
```
production-web-01
production-web-02
staging-db-01
dev-monitoring-01
```

**Avoid:**
```
server1          # Not descriptive
my_web_server   # Mixed conventions
prod-server     # Ambiguous type
```

### IP Organization

Organize IPs by service type:
- Web servers: `192.168.1.10-19`
- Database servers: `192.168.1.20-29`
- Monitoring: `192.168.1.30-39`

### Port Allocation

Use sequential ports for multiple instances:
- Web-01: `3000`
- Web-02: `3001`
- Web-03: `3002`

### SSH Keys

Use different keys for different environments:
- Production: `~/.ssh/id_rsa_prod`
- Staging: `~/.ssh/id_rsa_staging`
- Development: `~/.ssh/id_rsa_dev`

---

## 🐛 Troubleshooting

### "SSH key file not found"

**Solution**: Verify the key exists
```bash
ls -l ~/.ssh/id_rsa
```

If missing, generate one:
```bash
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
```

### "IP:Port conflict detected"

**Solution**: Either:
1. Change the IP address
2. Change the port number
3. Delete/edit the conflicting server

### "Failed to save environment"

**Solution**: Check permissions
```bash
chmod 755 inventory/
```

### Server not showing in list

**Solution**: Ensure it was saved (press `s` in Server Manager)

---

## 🚀 Quick Start Example

### Create a 3-tier environment

```bash
# Launch
./bin/inventory-manager

# 1. Create environment
Create new environment
  Name: production
  Repo: https://github.com/myuser/myapp.git
  Branch: main
  Services: [✓] Web [✓] Database [✓] Monitoring

# 2. Add web servers
Manage → production → Add
  IP: 192.168.1.10, Port: 3000, Type: Web
Manage → production → Add
  IP: 192.168.1.11, Port: 3001, Type: Web

# 3. Add database
Manage → production → Add
  IP: 192.168.1.20, Port: 5432, Type: Database

# 4. Add monitoring
Manage → production → Add
  IP: 192.168.1.30, Port: 9090, Type: Monitoring

# 5. Save
Press 's' to save

# 6. View summary
Press 'g' to see full configuration
```

---

## 📚 Related Documentation

- [Main README](../INVENTORY_MANAGER_README.md)
- [Architecture Plan](INVENTORY_MANAGER_PLAN.md)
- [Bubbletea vs Streamlit](BUBBLETEA_VS_STREAMLIT.md)

---

**Happy Server Managing! 🎉**
