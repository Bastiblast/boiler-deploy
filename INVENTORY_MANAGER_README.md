# 🔧 Ansible Inventory Manager

A modern, interactive Terminal User Interface (TUI) for managing Ansible inventories, built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea).

## 🎯 Features

- ✅ **Interactive TUI** - Navigate with keyboard, no mouse required
- ✅ **Environment Management** - Create and manage multiple environments (prod, dev, staging)
- ✅ **Server Configuration** - Add, edit, and remove servers with validation
- ✅ **Auto-generation** - Automatically generates Ansible-compatible `hosts.yml` and `group_vars/`
- ✅ **Validation** - Real-time IP, port, and configuration validation
- ✅ **Lightweight** - Single binary, no dependencies (~5MB)
- ✅ **Fast** - Instant startup, responsive UI

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/bastiblast/boiler-deploy.git
cd boiler-deploy

# Build
make build

# Or install globally
make install
```

### Download Binary

Download the latest release from the [Releases page](https://github.com/bastiblast/boiler-deploy/releases).

## 🚀 Usage

### Launch the TUI

```bash
./bin/inventory-manager

# Or if installed globally
inventory-manager
```

### Navigation

```
[↑↓] or [j/k]  - Navigate menu
[Enter]         - Select option
[Tab]           - Next field (in forms)
[Space]         - Toggle checkbox
[Esc]           - Go back / Cancel
[q]             - Quit
```

### Create an Environment

1. Launch the application
2. Select "Create new environment"
3. Fill in the form:
   - Environment name (e.g., `production`)
   - Git repository URL
   - Git branch
   - Node.js version
   - Application port
4. Toggle services (Web, Database, Monitoring)
5. Press Enter to create

### Generated Files

The tool automatically generates:

```
inventory/
└── production/
    ├── config.yml          # Environment configuration
    └── hosts.yml           # Ansible inventory

group_vars/
└── production.yml          # Ansible variables
```

## 📚 Documentation

- [Full Documentation](docs/INVENTORY_MANAGER_PLAN.md)
- [Bubbletea vs Streamlit Comparison](docs/BUBBLETEA_VS_STREAMLIT.md)
- [Copilot Agent Guide](docs/COPILOT_AGENT_GUIDE.md)

## 🏗️ Architecture

```
cmd/
└── inventory-manager/
    └── main.go                 # Entry point

internal/
├── ui/                         # Bubbletea UI
│   ├── menu.go                 # Main menu
│   ├── form_environment.go     # Environment form
│   ├── styles.go               # Lipgloss styles
│   └── components/             # Reusable components
├── inventory/                  # Business logic
│   ├── models.go               # Data models
│   ├── validator.go            # Validation
│   └── generator.go            # YAML generation
├── ssh/                        # SSH utilities
│   └── tester.go               # Connection testing
└── storage/                    # Persistence
    └── yaml.go                 # YAML I/O
```

## 🛠️ Development

### Prerequisites

- Go 1.21 or higher
- Make (optional)

### Build

```bash
make build          # Build the application
make run            # Build and run
make clean          # Clean build artifacts
make test           # Run tests
make fmt            # Format code
```

### Dependencies

```go
github.com/charmbracelet/bubbletea  // TUI framework
github.com/charmbracelet/lipgloss   // Styling
github.com/charmbracelet/bubbles    // Components
gopkg.in/yaml.v3                    // YAML parsing
```

## 📝 Examples

### Example: Generated hosts.yml

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
          app_port: 3000
```

### Example: Generated group_vars/production.yml

```yaml
app_name: production-app
app_repo: https://github.com/user/repo.git
app_branch: main
nodejs_version: "20"
app_port: "3000"
deploy_user: root
timezone: Europe/Paris
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📜 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Amazing TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Beautiful terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - Reusable TUI components

## 📞 Support

For issues, questions, or suggestions, please open an issue on GitHub.

---

**Made with ❤️ and Go**
