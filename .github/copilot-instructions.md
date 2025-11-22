# Copilot Instructions: boiler-deploy

## Project Overview

**boiler-deploy** is a hybrid Go/Ansible deployment system for Node.js applications on any VPS. It combines:
- **Go TUI** (Bubbletea) for interactive inventory management (`cmd/inventory-manager/`)
- **Ansible playbooks** for provisioning and deployment (`playbooks/`, `roles/`)
- **Smart auto-detection** of Node.js frameworks (Next.js, Nuxt, Express, etc.)
- **Parallel execution** engine for multi-server deployments
- **Tag-based** granular control over provisioning/deployment steps

Architecture: The Go TUI generates Ansible inventories and orchestrates playbook execution. State persists in `inventory/<env>/` with YAML storage. Real-time SSH validation and status tracking via `internal/status/` and `internal/ssh/`.

## Critical Workflows

### Build & Run
```bash
make build          # Compiles to bin/inventory-manager
make run            # Build + run TUI
make test           # Run Go tests
./deploy.sh provision production  # Ansible provisioning (bash wrapper)
./deploy.sh deploy production     # Ansible deployment
```

### Testing Infrastructure
- **Docker-based testing**: `tests/multi-container-setup.sh` spins up 3 SSH-enabled Ubuntu containers
- **Autonomous test suite**: `tests/autonomous-agent-test.sh` validates parallel execution
- Use `inventory/test-docker/` or `inventory/test-multi/` for test environments
- Test SSH keys: `~/.ssh/boiler_test_rsa` (auto-generated)

### Key Entry Points
- `cmd/inventory-manager/main.go` - TUI entry, orchestrates startup validation
- `internal/ui/menu.go` - Main menu navigation
- `internal/ansible/executor.go` - Ansible playbook execution with context/cancellation
- `deploy.sh` - Bash wrapper for Ansible operations (supports `--yes` for automation)

## Project-Specific Conventions

### State Management Pattern
All state lives in `inventory/<env>/` per-environment:
```
inventory/production/
├── hosts.yml              # Ansible inventory (generated)
├── host_vars/*.yml        # Per-server vars (generated)
├── group_vars/all.yml     # Common vars (generated)
├── config.yml             # ConfigOptions (max_parallel_workers, tags, etc.)
├── status.yml             # Status tracking (managed by status.Manager)
└── .env-config.yml        # Internal: full Environment struct (hidden from Ansible)
```
**Rule**: Ansible files are generated—never edit manually. Source of truth: `.env-config.yml` loaded via `storage.LoadEnvironment()`.

### Parallel Execution Architecture
Feature branch: `feature/parallel-action-execution`

- **Orchestrator** (`internal/ansible/orchestrator.go`): Worker pool pattern
  - `SetMaxWorkers(n)` enables parallel mode (0=sequential, 3-5=optimal)
  - `processQueueParallel()` vs `processQueueSequential()`
  - Thread-safe queue (`internal/status/queue.go`) with `NextBatch()`, `CompleteByID()`
- **Configuration**: `inventory/<env>/config.yml` → `max_parallel_workers: 3`
- **Logging**: Workers prefix logs with `[ORCHESTRATOR] Worker N processing...`

### Tag System
Tags control granular Ansible task execution (see `docs/ANSIBLE_TAGS.md`):
- **Provision tags**: `common`, `security`, `firewall`, `ssh`, `nodejs`, `nginx`, `postgresql`, `monitoring`
- **Deploy tags**: `dependencies`, `build`, `deploy`, `restart`, `health_check`
- **Implementation**: `internal/ui/tag_selector.go` provides TUI selection → passed to `ansible/executor.go`
- **Usage**: Tags append to playbook invocation via `--tags` flag

### Auto-Detection System
See `docs/AUTO_DETECTION.md` for framework detection logic:
- Reads `package.json` to detect Next.js/Nuxt/Express/Fastify/NestJS
- Detects package manager: pnpm (lock) > yarn (lock) > npm (default)
- Generates PM2 config: fork mode (Next/Nuxt) vs cluster mode (Express)
- Ansible role: `roles/deploy-app/tasks/main.yml` runs detection playbook tasks

### SSH State Detection
`internal/ssh/state_detector.go` validates real SSH connectivity:
- Used at startup (`cmd/inventory-manager/main.go:validateAllServers()`)
- Updates `status.yml` with actual server reachability
- Pattern: Always validate before operations to prevent stale state

## Critical Files & Their Roles

### Go Modules
- `internal/inventory/` - Core Environment/Server data models
- `internal/storage/yaml.go` - YAML persistence (SaveEnvironment, LoadEnvironment)
- `internal/status/models.go` - ActionType (validate/provision/deploy/check), QueuedAction, ServerStatus
- `internal/config/types.go` - ConfigOptions with defaults (DefaultConfig())
- `internal/ui/workflow_view.go` - Main operations dashboard (keyboard shortcuts: v/p/d/c/l)

### Ansible Structure
- `playbooks/provision.yml` - Multi-role orchestration (common, security, nodejs, nginx, postgresql, monitoring)
- `playbooks/deploy.yml` - Application deployment (auto-detection, PM2, nginx config)
- `roles/*/tasks/main.yml` - Role logic with tags applied
- `group_vars/all.yml.example` - Template for app configuration (app_repo, app_branch, app_port)

### Documentation
- `docs/WORKFLOW_GUIDE.md` - TUI keyboard controls and status flow
- `docs/PARALLEL_EXECUTION.md` - Worker pool architecture
- `docs/COPILOT_AGENT_GUIDE.md` - French guide for using AI agents (context patterns)
- `docs/OPERATIONS_SUMMARY.md` - Operations dashboard feature spec

## Common Pitfalls

### Ansible Execution
- **Always use absolute paths** for playbooks in `executor.go`
- **Context cancellation**: Use `RunPlaybookWithContext()` for timeout handling (default 30min)
- **Progress channels**: Non-blocking sends to `progressChan` or risk deadlock
- **Log files**: Created per action in `logs/<env>/<server>_<action>_<timestamp>.log`

### Bubbletea State Management
- **Model transitions**: Return new model + command from `Update()` for screen changes
  ```go
  return NewWorkflowView(env), nil  // NOT: m.view = "workflow"
  ```
- **Keyboard conflicts**: Check existing shortcuts in `workflow_view.go` before adding
- **Thread safety**: Status updates via `status.Manager` are mutex-protected

### Configuration Loading
- **Config file missing**: `config.Manager.LoadConfig()` returns defaults if `config.yml` absent
- **Parallel workers = 0**: Sequential mode (default), not disabled
- **Duration fields**: Use `time.Duration` in Go, serialize as `"30s"` in YAML

### Testing Anti-Patterns
- **Don't hardcode IPs**: Use `inventory/test-docker/hosts.yml` generated by `multi-container-setup.sh`
- **Clean state**: Remove `logs/test-*` and `inventory/test-*/status.yml` between test runs
- **SSH key reuse**: Test containers expect `~/.ssh/boiler_test_rsa` key

## Integration Points

### Bash ↔ Go TUI
- `deploy.sh --yes` flag: Auto-confirms for TUI invocation (skips terminal prompts)
- TUI calls `deploy.sh` via `run_in_terminal` tool for Ansible operations
- Environment variable: `AUTO_CONFIRM=true` for non-interactive mode

### Ansible ↔ Go Executor
- Executor streams output to log file + progress channel
- Exit code handling: Non-zero = failure, updates `status.yml` error message
- Tags passed via `--tags "tag1,tag2"` flag to ansible-playbook

### Storage ↔ UI
- All UI forms write to storage via `storage.SaveEnvironment()`
- UI reads via `storage.LoadEnvironment()` → updates display
- Validation happens in forms (IP format, SSH connectivity) before save

## Code Style & Patterns

- **Error handling**: Wrap with context: `fmt.Errorf("failed to X: %w", err)`
- **Logging**: Use `logger.Get("component")` from `internal/logger/`, not `log.Printf`
- **UI styling**: Reuse `lipgloss.Style` instances in `internal/ui/` (titleStyle, focusedStyle)
- **Mutex naming**: `fooMu sync.RWMutex` for field `foo`, always embed `sync.Mutex` first
- **YAML tags**: Always include for config structs: `yaml:"field_name"`

## AI Agent Guidance

When implementing features:
1. **Check existing patterns** in `internal/ui/` for similar TUI screens before creating new ones
2. **Validate in parallel**: Use goroutines for SSH checks but respect rate limits
3. **Test with Docker**: Run `tests/multi-container-setup.sh` for integration tests
4. **Document in `docs/`**: Add markdown guide for user-facing features
5. **Update Makefile**: Add new binaries to targets if creating new `cmd/` tools

For debugging Ansible issues:
- Check `logs/<env>/<server>_<action>_<timestamp>.log` for playbook output
- Verify inventory generation with `ansible-inventory -i inventory/<env>/hosts.yml --list`
- Test playbook syntax: `ansible-playbook playbooks/provision.yml --syntax-check`
