# 🗺️ Inventory Manager - Roadmap & Development Plan

## 📊 Project Overview

**Project Name:** Ansible Inventory Manager (TUI)  
**Tech Stack:** Go + Bubbletea  
**Git Branch:** `streamlit`  
**Status:** MVP Complete ✅  
**Created:** November 2025

### 🎯 Project Goal

Replace the confusing Bash setup scripts with a modern, user-friendly Terminal UI (TUI) to manage Ansible inventories for deploying Node.js applications across multiple environments.

---

## ✅ Implemented Features (MVP)

### 1. Core Functionality

- [x] **Environment Management**
  - Create multiple environments (prod, dev, staging, etc.)
  - Configure per environment:
    - Name
    - Git repository URL (default)
    - Git branch (default)
    - Node.js version (default)
    - Application port (default)
  - Mono-server deployment option (same IP for all servers)
  - Shared SSH key option (same key for all servers)
  - Delete environments
  - Switch between environments

- [x] **Server Management**
  - Add servers with types: Web, Database, Monitoring
  - Configure per server:
    - Name
    - IP address
    - SSH port
    - SSH key path (with .pub extension validation)
    - Git repository URL (overrides environment default)
    - Application port (overrides environment default)
  - Edit existing servers
  - Delete servers
  - Visual server list with status indicators

- [x] **SSH Testing**
  - Test individual server SSH connection (`t` key)
  - Test all servers at once (`T` key)
  - Real-time connection status
  - Detailed error messages:
    - Connection timeout
    - Invalid SSH key
    - Authentication failures
  - Connection latency display

- [x] **File Generation**
  - Auto-generate `inventory/{env}/hosts.yml` (Ansible inventory)
  - Auto-generate `group_vars/{env}.yml` (Ansible variables)
  - YAML preview before saving (`g` key)
  - Save to disk (`s` key)

- [x] **User Interface**
  - Clean, intuitive TUI with Bubbletea
  - Keyboard navigation (arrows, j/k, Tab, Enter, Esc)
  - Real-time validation (IP, ports, paths)
  - Error messages with visual feedback
  - Color-coded status indicators
  - Responsive forms

### 2. Technical Features

- [x] Single binary executable (~5MB)
- [x] No external dependencies
- [x] Fast startup time
- [x] YAML-based storage
- [x] Makefile for easy build/run
- [x] Structured Go architecture:
  - `cmd/` - Entry point
  - `internal/ui/` - Bubbletea components
  - `internal/inventory/` - Business logic
  - `internal/ssh/` - SSH utilities
  - `internal/storage/` - File I/O

---

## 🚀 Planned Features (Next Steps)

### Phase 1: Enhanced Validation & UX (Priority: HIGH)

- [ ] **Improved Input Validation**
  - [ ] SSH key existence check (warn if file not found)
  - [ ] Git repository URL validation (optional ping test)
  - [ ] Port conflict detection (warn if multiple services use same port)
  - [ ] Hostname/IP duplicate detection

- [ ] **Better Error Handling**
  - [ ] Graceful handling of corrupt YAML files
  - [ ] Auto-backup before overwriting files
  - [ ] Recovery mode if configuration is invalid

- [ ] **UX Improvements**
  - [ ] Confirmation dialogs for destructive actions (delete environment/server)
  - [ ] Undo last action (Ctrl+Z)
  - [ ] Search/filter servers by name or IP
  - [ ] Bulk operations (delete multiple servers, test multiple)

### Phase 2: Deployment Integration (Priority: HIGH)

- [ ] **Ansible Playbook Execution**
  - [ ] Run Ansible playbooks directly from TUI
  - [ ] Real-time deployment logs display
  - [ ] Progress indicators for long-running tasks
  - [ ] Deployment history/logs

- [ ] **Pre-deployment Checks**
  - [ ] Verify all SSH connections before deploy
  - [ ] Check disk space on target servers
  - [ ] Verify Git repository accessibility
  - [ ] Validate Node.js version availability

- [ ] **Post-deployment Verification**
  - [ ] Health check endpoints
  - [ ] Service status verification
  - [ ] Rollback on failure

### Phase 3: Advanced Features (Priority: MEDIUM)

- [ ] **Configuration Templates**
  - [ ] Save/load environment templates
  - [ ] Clone existing environments
  - [ ] Import/export configurations

- [ ] **Multi-Environment Comparison**
  - [ ] Side-by-side environment comparison
  - [ ] Diff viewer for configurations
  - [ ] Sync configurations between environments

- [ ] **Advanced SSH Management**
  - [ ] SSH agent integration
  - [ ] Generate SSH key pairs from UI
  - [ ] Test SSH with custom commands
  - [ ] Jump host/bastion support

- [ ] **Monitoring Integration**
  - [ ] Display server metrics (CPU, RAM, disk)
  - [ ] Application health status
  - [ ] Log aggregation view

### Phase 4: Scalability & Collaboration (Priority: LOW)

- [ ] **Team Collaboration**
  - [ ] Git-based configuration sync
  - [ ] Lock mechanism to prevent conflicts
  - [ ] Change history/audit log

- [ ] **Advanced Inventory Features**
  - [ ] Dynamic inventory (cloud providers integration)
  - [ ] Host groups and tags
  - [ ] Conditional variables
  - [ ] Vault encryption for secrets

- [ ] **Customization**
  - [ ] Custom themes/color schemes
  - [ ] Configurable keyboard shortcuts
  - [ ] Plugin system for extensions

---

## 🔧 Technical Improvements

### Code Quality

- [ ] **Testing**
  - [ ] Unit tests for business logic (inventory, validator)
  - [ ] Integration tests for SSH testing
  - [ ] TUI component tests
  - [ ] CI/CD pipeline (GitHub Actions)

- [ ] **Documentation**
  - [ ] Code documentation (godoc)
  - [ ] Architecture diagrams
  - [ ] Developer guide
  - [ ] API documentation

- [ ] **Refactoring**
  - [ ] Extract common UI patterns into reusable components
  - [ ] Improve error handling consistency
  - [ ] Add logging framework (structured logs)
  - [ ] Performance profiling and optimization

### Build & Distribution

- [ ] **Cross-platform Support**
  - [ ] Linux (amd64, arm64)
  - [ ] macOS (Intel, Apple Silicon)
  - [ ] Windows (optional)

- [ ] **Distribution**
  - [ ] GitHub Releases with pre-built binaries
  - [ ] Homebrew formula
  - [ ] APT/YUM repository
  - [ ] Docker image (optional)

---

## 🐛 Known Issues

1. **Minor UI Glitches**
   - ~~Tab navigation crash on empty fields~~ (FIXED ✅)
   - ~~SSH key .pub extension not validated~~ (FIXED ✅)

2. **Feature Gaps**
   - No confirmation dialog for destructive actions
   - Cannot reorder servers in list
   - No search/filter functionality

3. **Edge Cases**
   - Very long server lists may have scrolling issues
   - Large YAML files may cause slow rendering

---

## 📅 Development Timeline (Estimate)

| Phase | Duration | Effort |
|-------|----------|--------|
| Phase 1: Enhanced Validation & UX | 1-2 weeks | 15-20 hours |
| Phase 2: Deployment Integration | 2-3 weeks | 30-40 hours |
| Phase 3: Advanced Features | 3-4 weeks | 40-50 hours |
| Phase 4: Scalability & Collaboration | 4-6 weeks | 50-70 hours |

**Total Estimated Effort:** 135-180 hours (~4-6 months part-time)

---

## 🎓 Learning Resources

### Bubbletea & Go TUI Development

- [Bubbletea Official Docs](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Styling Guide](https://github.com/charmbracelet/lipgloss)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Charm CLI Examples](https://github.com/charmbracelet/charm)

### Go Best Practices

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)

---

## 📝 Decision Log

### Why Go + Bubbletea?

**Alternatives Considered:**
1. **Streamlit** (Python web UI)
   - ❌ Requires Python runtime + dependencies
   - ❌ Web-based (overkill for local-only tool)
   - ❌ Heavier resource usage

2. **Semaphore** (Ansible UI)
   - ❌ Too complex for our needs
   - ❌ Requires database + authentication
   - ❌ Not customizable enough

3. **Bash + Dialog/Whiptail**
   - ❌ Hard to maintain
   - ❌ Poor user experience
   - ❌ Limited UI capabilities

**Why Bubbletea Won:**
- ✅ Single binary, no dependencies
- ✅ Fast and lightweight
- ✅ Beautiful, modern TUI
- ✅ Easy to customize
- ✅ Great for terminal-first workflows
- ✅ Active community and good documentation

### Key Design Decisions

1. **Local-only, no authentication**
   - Simplified architecture
   - Faster development
   - Suitable for single-user use case

2. **YAML-based storage**
   - Human-readable
   - Git-friendly
   - Easy to edit manually if needed
   - Compatible with Ansible

3. **Server-specific overrides**
   - Flexibility to customize per server
   - Inherit defaults from environment
   - Repository and port can differ per server

4. **SSH key validation**
   - Validate .pub extension (prevent common mistake)
   - Test actual connectivity before deployment
   - Clear error messages

---

## 🚦 Getting Back to Development

### Quick Start Checklist

1. **Environment Setup**
   ```bash
   cd /home/basthook/devIronMenth/boiler-deploy
   git checkout streamlit
   make build
   ```

2. **Run the Application**
   ```bash
   make run
   ```

3. **Check Current Status**
   - Review this roadmap
   - Check `git log` for recent changes
   - Run the app to see current state

4. **Choose Next Feature**
   - Pick from "Planned Features" above
   - Start with Phase 1 for quick wins
   - Check "Known Issues" for bugs to fix

### Development Workflow

```bash
# Edit code
vim internal/ui/menu.go

# Test changes
make run

# Format code
make fmt

# Commit changes
git add .
git commit -m "feat: add confirmation dialog for delete"

# Push to remote
git push origin streamlit
```

---

## 📞 Questions & Clarifications

### Architecture Questions

**Q: Should we support remote configuration storage (cloud)?**  
A: Not in MVP. Keep it local-only for simplicity. Can add in Phase 4 if needed.

**Q: Should we add a web UI in addition to TUI?**  
A: No, TUI-only keeps it simple and lightweight. Web UI would require significant effort.

**Q: How to handle secrets (passwords, API keys)?**  
A: For now, store SSH key paths only. In Phase 3, consider Ansible Vault integration.

### Feature Prioritization

**Q: What's the most important next feature?**  
A: **Ansible playbook execution** (Phase 2). This completes the deployment workflow.

**Q: Should we support non-Node.js deployments?**  
A: Not initially. Keep scope focused. Can generalize later if needed.

**Q: Do we need multi-user support?**  
A: No. Local-only, single-user is sufficient for current use case.

---

## 🎯 Success Metrics

### MVP Success (Completed ✅)

- [x] Can create and manage multiple environments
- [x] Can configure servers with all required fields
- [x] Can test SSH connections before deployment
- [x] Generates valid Ansible inventory and variables
- [x] Replaces confusing Bash scripts completely

### Phase 1 Success

- [ ] Zero configuration errors due to typos
- [ ] Confirmation on all destructive actions
- [ ] Can find servers quickly (search/filter)

### Phase 2 Success

- [ ] Can deploy from TUI without command line
- [ ] Real-time deployment feedback
- [ ] Automatic rollback on failure

### Phase 3 Success

- [ ] Can clone environments in <30 seconds
- [ ] Can compare configurations visually
- [ ] SSH management fully integrated

### Phase 4 Success

- [ ] Team can collaborate without conflicts
- [ ] Configuration changes are auditable
- [ ] Scales to 100+ servers per environment

---

## 📚 Related Documentation

- [README.md](../README.md) - Project overview and Ansible playbooks
- [INVENTORY_MANAGER_README.md](../INVENTORY_MANAGER_README.md) - User guide
- [docs/INVENTORY_MANAGER_PLAN.md](./INVENTORY_MANAGER_PLAN.md) - Original design document
- [docs/BUBBLETEA_VS_STREAMLIT.md](./BUBBLETEA_VS_STREAMLIT.md) - Framework comparison
- [docs/COPILOT_AGENT_GUIDE.md](./COPILOT_AGENT_GUIDE.md) - GitHub Copilot usage

---

## 🤝 Contributing

### How to Contribute

1. Pick a feature from "Planned Features"
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Implement the feature
4. Test thoroughly with `make run`
5. Submit a pull request to `streamlit` branch

### Code Style

- Follow Go conventions (`gofmt`, `golint`)
- Write descriptive commit messages
- Add comments for complex logic
- Keep functions small and focused

---

## 📊 Project Stats

**Lines of Code:**
```bash
# Run this to get current stats
find cmd internal -name "*.go" | xargs wc -l
```

**Files Structure:**
```
cmd/
└── inventory-manager/
    └── main.go                 # Entry point (17 lines)

internal/
├── ui/
│   ├── menu.go                 # Main menu
│   ├── form_environment.go     # Environment creation form
│   ├── server_form.go          # Server add/edit form
│   ├── environment_selector.go # Environment selection
│   ├── server_manager.go       # Server management view
│   └── styles.go               # Shared styles
├── inventory/
│   ├── models.go               # Data structures
│   ├── validator.go            # Input validation
│   └── generator.go            # YAML generation
├── ssh/
│   └── tester.go               # SSH connection testing
└── storage/
    └── yaml.go                 # File I/O operations
```

---

## 🔄 Version History

### v0.1.0 - MVP (Current)
- Initial TUI implementation
- Environment and server management
- SSH testing
- YAML generation

### v0.2.0 - Planned (Phase 1)
- Enhanced validation
- Confirmation dialogs
- Search/filter
- Better error handling

### v0.3.0 - Planned (Phase 2)
- Ansible playbook execution
- Deployment logs
- Pre/post-deployment checks

### v1.0.0 - Planned (Phase 3)
- Configuration templates
- Environment comparison
- Advanced SSH features

---

**Last Updated:** November 11, 2025  
**Maintained by:** BastiBlast  
**Project Status:** 🟢 Active Development
