#!/bin/bash

##############################################################################
# Multi-Server Setup Wizard
# Creates inventory and configuration for VPS deployment
##############################################################################

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

# Symbols
CHECKMARK="✓"
CROSSMARK="✗"
ARROW="→"
INFO="ℹ"
WARNING="⚠"

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="${SCRIPT_DIR}/setup_$(date +%Y%m%d_%H%M%S).log"
STATE_FILE=""

# Configuration
MAX_WEB_SERVERS=20
DEFAULT_BRANCH="main"

# Global variables
ENVIRONMENT=""
MODE="create" # create or add
RESUME_MODE=false

# Service flags
ENABLE_WEB=true
ENABLE_DB=true
ENABLE_MONITORING=true

# Configuration storage
declare -A CONFIG
declare -a WEB_SERVERS
declare -A DB_SERVER
declare -A MONITORING_SERVER
declare -a FAILED_CONNECTIONS

##############################################################################
# Logging
##############################################################################

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_FILE" >&2
}

##############################################################################
# UI Functions
##############################################################################

print_header() {
    echo -e "\n${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
}

print_success() {
    echo -e "${GREEN}${CHECKMARK}${NC} $1"
}

print_error() {
    echo -e "${RED}${CROSSMARK}${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}${WARNING}${NC} $1"
}

print_info() {
    echo -e "${BLUE}${INFO}${NC} $1"
}

print_arrow() {
    echo -e "${CYAN}${ARROW}${NC} $1"
}

ask() {
    local prompt="$1"
    local default="${2:-}"
    local result
    
    if [[ -n "$default" ]]; then
        echo -ne "${BOLD}?${NC} ${prompt} ${DIM}[${default}]${NC}: "
    else
        echo -ne "${BOLD}?${NC} ${prompt}: "
    fi
    
    read -r result
    echo "${result:-$default}"
}

ask_yes_no() {
    local prompt="$1"
    local default="${2:-n}"
    local result
    
    if [[ "$default" == "y" ]]; then
        echo -ne "${BOLD}?${NC} ${prompt} ${DIM}[Y/n]${NC}: "
    else
        echo -ne "${BOLD}?${NC} ${prompt} ${DIM}[y/N]${NC}: "
    fi
    
    read -r result
    result="${result:-$default}"
    [[ "$result" =~ ^[Yy] ]]
}

##############################################################################
# State Management
##############################################################################

save_state() {
    local state_file="$1"
    
    cat > "$state_file" << EOF
---
# Setup state - DO NOT EDIT MANUALLY
mode: "$MODE"
environment: "$ENVIRONMENT"
services:
  web: $ENABLE_WEB
  database: $ENABLE_DB
  monitoring: $ENABLE_MONITORING

# Configuration
$(declare -p CONFIG 2>/dev/null | sed 's/declare -A//' || echo "config: {}")

# Web servers
web_servers:
EOF
    
    for server in "${WEB_SERVERS[@]}"; do
        echo "  - $server" >> "$state_file"
    done
    
    if [[ ${#DB_SERVER[@]} -gt 0 ]]; then
        echo -e "\n# Database server" >> "$state_file"
        echo "database_server:" >> "$state_file"
        for key in "${!DB_SERVER[@]}"; do
            echo "  $key: ${DB_SERVER[$key]}" >> "$state_file"
        done
    fi
    
    if [[ ${#MONITORING_SERVER[@]} -gt 0 ]]; then
        echo -e "\n# Monitoring server" >> "$state_file"
        echo "monitoring_server:" >> "$state_file"
        for key in "${!MONITORING_SERVER[@]}"; do
            echo "  $key: ${MONITORING_SERVER[$key]}" >> "$state_file"
        done
    fi
    
    log "State saved to $state_file"
}

load_state() {
    local state_file="$1"
    
    if [[ ! -f "$state_file" ]]; then
        return 1
    fi
    
    log "Loading state from $state_file"
    # State loading would require YAML parser - simplified for now
    print_info "Resuming from saved state..."
    return 0
}

##############################################################################
# Validation Functions
##############################################################################

validate_ip() {
    local ip="$1"
    local regex='^([0-9]{1,3}\.){3}[0-9]{1,3}$'
    
    if [[ ! $ip =~ $regex ]]; then
        return 1
    fi
    
    # Check each octet
    IFS='.' read -ra ADDR <<< "$ip"
    for i in "${ADDR[@]}"; do
        if ((i > 255)); then
            return 1
        fi
    done
    
    return 0
}

check_ip_conflict() {
    local ip="$1"
    local port="${2:-3000}"
    
    # Check web servers
    for server_data in "${WEB_SERVERS[@]}"; do
        IFS='|' read -r _ server_ip server_port _ <<< "$server_data"
        if [[ "$server_ip" == "$ip" && "$server_port" == "$port" ]]; then
            return 1
        fi
    done
    
    # Check database
    if [[ -n "${DB_SERVER[ip]:-}" && "${DB_SERVER[ip]}" == "$ip" ]]; then
        if [[ "${DB_SERVER[port]:-5432}" == "$port" ]]; then
            return 1
        fi
    fi
    
    return 0
}

test_ssh_connection() {
    local user="$1"
    local host="$2"
    local ssh_key="${3:-}"
    
    local ssh_opts="-o ConnectTimeout=10 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"
    
    if [[ -n "$ssh_key" ]]; then
        ssh_opts="$ssh_opts -i $ssh_key"
    fi
    
    if ssh $ssh_opts "${user}@${host}" "command -v python3 > /dev/null 2>&1" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

check_git_repo() {
    local repo="$1"
    local branch="$2"
    
    print_arrow "Testing Git repository access..."
    
    if timeout 15 git ls-remote --heads "$repo" "$branch" > /dev/null 2>&1; then
        print_success "Git repository accessible"
        return 0
    else
        print_error "Git repository not accessible!"
        echo
        print_info "Tried: $repo (branch: $branch)"
        echo
        echo "Troubleshooting:"
        echo "  1. Check repository URL:"
        echo "     git ls-remote $repo"
        echo
        echo "  2. Verify SSH key for Git (if using git@):"
        echo "     ssh -T git@github.com  # or gitlab.com"
        echo
        echo "  3. Add SSH key to ssh-agent:"
        echo "     ssh-add ~/.ssh/your_key"
        echo
        echo "  4. Use HTTPS instead of SSH:"
        echo "     https://github.com/user/repo.git"
        echo
        
        save_state "$STATE_FILE"
        
        print_warning "Configuration saved. Fix the issue and run:"
        echo "  ./setup.sh $ENVIRONMENT --resume"
        echo
        
        exit 1
    fi
}

##############################################################################
# Phase 1: Prerequisites Check
##############################################################################

check_prerequisites() {
    print_header "Phase 1/7: Prerequisites Check"
    
    local missing_deps=()
    
    # Check SSH client
    print_arrow "Checking SSH client..."
    if command -v ssh > /dev/null 2>&1; then
        print_success "SSH client installed"
    else
        print_error "SSH client not found!"
        echo
        echo "Please install SSH client:"
        echo "  • Ubuntu/Debian: sudo apt-get install openssh-client"
        echo "  • macOS: brew install openssh"
        echo "  • Windows: Install OpenSSH via Settings or Git for Windows"
        echo
        exit 1
    fi
    
    # Check Ansible
    print_arrow "Checking Ansible..."
    if command -v ansible > /dev/null 2>&1; then
        local ansible_version=$(ansible --version | head -n1 | awk '{print $2}')
        print_success "Ansible installed (v${ansible_version})"
    else
        missing_deps+=("ansible")
        print_warning "Ansible not found"
    fi
    
    # Check Python3
    print_arrow "Checking Python3..."
    if command -v python3 > /dev/null 2>&1; then
        local python_version=$(python3 --version | awk '{print $2}')
        print_success "Python3 installed (v${python_version})"
    else
        missing_deps+=("python3")
        print_warning "Python3 not found"
    fi
    
    # Check Git
    print_arrow "Checking Git..."
    if command -v git > /dev/null 2>&1; then
        local git_version=$(git --version | awk '{print $3}')
        print_success "Git installed (v${git_version})"
    else
        missing_deps+=("git")
        print_warning "Git not found"
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        echo
        print_error "Missing dependencies: ${missing_deps[*]}"
        echo
        echo "Install with:"
        echo "  • Ubuntu/Debian: sudo apt-get install ${missing_deps[*]}"
        echo "  • macOS: brew install ${missing_deps[*]}"
        echo
        exit 1
    fi
    
    print_success "All prerequisites met!"
}

##############################################################################
# Phase 2: Environment Setup
##############################################################################

setup_environment() {
    print_header "Phase 2/7: Environment Setup"
    
    # Check existing environments
    local existing_envs=()
    if [[ -d "${SCRIPT_DIR}/inventory" ]]; then
        for env_dir in "${SCRIPT_DIR}/inventory"/*; do
            if [[ -d "$env_dir" && -f "$env_dir/hosts.yml" ]]; then
                local env_name=$(basename "$env_dir")
                existing_envs+=("$env_name")
            fi
        done
    fi
    
    if [[ ${#existing_envs[@]} -gt 0 ]]; then
        echo "Found existing environments:"
        for env in "${existing_envs[@]}"; do
            echo -e "  ${GREEN}•${NC} $env"
        done
        echo
        
        echo "What do you want to do?"
        echo "  1) Create a new environment"
        echo "  2) Add servers to existing environment"
        echo "  3) Exit"
        echo
        
        echo -ne "${BOLD}?${NC} Choice ${DIM}[1]${NC}: "
        read -r choice
        choice="${choice:-1}"
        
        case "$choice" in
            1)
                MODE="create"
                ENVIRONMENT=$(ask "Environment name (e.g., production, staging)" "production")
                
                if [[ " ${existing_envs[*]} " =~ " ${ENVIRONMENT} " ]]; then
                    print_error "Environment '$ENVIRONMENT' already exists!"
                    exit 1
                fi
                ;;
            2)
                MODE="add"
                echo
                echo "Select environment:"
                for i in "${!existing_envs[@]}"; do
                    echo "  $((i+1))) ${existing_envs[$i]}"
                done
                echo
                
                echo -ne "${BOLD}?${NC} Choice ${DIM}[1]${NC}: "
                read -r env_choice
                env_choice="${env_choice:-1}"
                ENVIRONMENT="${existing_envs[$((env_choice-1))]}"
                
                print_info "Adding servers to: $ENVIRONMENT"
                load_existing_environment "$ENVIRONMENT"
                ;;
            3|exit|quit)
                echo "Exiting..."
                exit 0
                ;;
            *)
                print_error "Invalid choice"
                exit 1
                ;;
        esac
    else
        ENVIRONMENT=$(ask "Environment name (e.g., production, staging)" "production")
    fi
    
    # Create inventory directory
    CONFIG[inventory_dir]="${SCRIPT_DIR}/inventory/${ENVIRONMENT}"
    mkdir -p "${CONFIG[inventory_dir]}"
    
    STATE_FILE="${CONFIG[inventory_dir]}/.setup_state.yml"
    
    print_success "Environment: ${ENVIRONMENT}"
}

load_existing_environment() {
    local env="$1"
    local hosts_file="${SCRIPT_DIR}/inventory/${env}/hosts.yml"
    
    print_header "Loading Environment: $env"
    
    # Parse existing hosts (simplified - just count)
    local web_count=$(grep -c "ansible_host:" "$hosts_file" 2>/dev/null || echo "0")
    
    print_info "Current configuration:"
    echo "  • Web servers: $web_count"
    echo
    
    echo "Which services do you want to add?"
    echo "  [X] Web servers"
    echo "  [ ] Database server (already configured)"
    echo "  [ ] Monitoring (already configured)"
    echo
    
    ENABLE_WEB=true
    ENABLE_DB=false
    ENABLE_MONITORING=false
}

select_services() {
    print_header "Service Selection"
    
    echo "Which services would you like to create?"
    echo
    
    if ask_yes_no "Create web server(s)?" "y"; then
        ENABLE_WEB=true
    else
        ENABLE_WEB=false
    fi
    
    if ask_yes_no "Create database server?" "y"; then
        ENABLE_DB=true
    else
        ENABLE_DB=false
    fi
    
    if ask_yes_no "Create monitoring server?" "y"; then
        ENABLE_MONITORING=true
    else
        ENABLE_MONITORING=false
    fi
    
    echo
    print_info "Selected services:"
    [[ "$ENABLE_WEB" == "true" ]] && echo "  • Web servers"
    [[ "$ENABLE_DB" == "true" ]] && echo "  • Database server"
    [[ "$ENABLE_MONITORING" == "true" ]] && echo "  • Monitoring server"
    echo
}

##############################################################################
# Phase 3: SSH Key Configuration
##############################################################################

configure_ssh_keys() {
    print_header "Phase 3/7: SSH Key Configuration"
    
    # If adding to existing environment, check if we should use existing key
    if [[ "$MODE" == "add" ]]; then
        local existing_key="${CONFIG[inventory_dir]}/group_vars/all.yml"
        if [[ -f "$existing_key" ]]; then
            local saved_key=$(grep "ansible_ssh_private_key_file:" "$existing_key" 2>/dev/null | awk '{print $2}' | tr -d '"')
            if [[ -n "$saved_key" && -f "$saved_key" ]]; then
                print_info "Found existing SSH key: $saved_key"
                if ask_yes_no "Use this existing SSH key for new servers?" "y"; then
                    CONFIG[ssh_private_key]="$saved_key"
                    CONFIG[ssh_public_key]="${saved_key}.pub"
                    print_success "Using existing SSH key"
                    
                    print_info "Public key to add to your new VPS:"
                    echo
                    cat "${CONFIG[ssh_public_key]}"
                    echo
                    
                    if ! ask_yes_no "Have you added this key to your new VPS?" "n"; then
                        print_warning "Add the key to your VPS and run again"
                        exit 1
                    fi
                    return
                fi
            fi
        fi
    fi
    
    if ask_yes_no "Do you already have an SSH key to use?" "y"; then
        local key_path=$(ask "Path to SSH private key" "~/.ssh/id_rsa")
        key_path="${key_path/#\~/$HOME}"
        
        if [[ ! -f "$key_path" ]]; then
            print_error "SSH key not found: $key_path"
            exit 1
        fi
        
        CONFIG[ssh_private_key]="$key_path"
        CONFIG[ssh_public_key]="${key_path}.pub"
        
        if [[ ! -f "${CONFIG[ssh_public_key]}" ]]; then
            print_error "Public key not found: ${CONFIG[ssh_public_key]}"
            exit 1
        fi
        
        print_success "Using existing SSH key: $key_path"
    else
        local key_name=$(ask "SSH key name" "${ENVIRONMENT}_deploy")
        local key_path="${HOME}/.ssh/${key_name}"
        
        if [[ -f "$key_path" ]]; then
            print_warning "Key already exists: $key_path"
            if ! ask_yes_no "Use existing key?" "y"; then
                exit 1
            fi
        else
            print_arrow "Generating SSH key..."
            ssh-keygen -t ed25519 -f "$key_path" -N "" -C "${ENVIRONMENT}_deploy_key"
            print_success "SSH key created: $key_path"
        fi
        
        CONFIG[ssh_private_key]="$key_path"
        CONFIG[ssh_public_key]="${key_path}.pub"
    fi
    
    # Add to ssh-agent
    if ask_yes_no "Add key to ssh-agent?" "y"; then
        eval "$(ssh-agent -s)" > /dev/null 2>&1
        ssh-add "${CONFIG[ssh_private_key]}" 2>/dev/null
        print_success "Key added to ssh-agent"
    fi
    
    print_info "Public key to add to your VPS:"
    echo
    cat "${CONFIG[ssh_public_key]}"
    echo
    
    if ! ask_yes_no "Have you added this key to your VPS?" "n"; then
        print_warning "Add the key to your VPS and run again"
        exit 1
    fi
}

##############################################################################
# Phase 4: VPS Configuration - Web Servers
##############################################################################

configure_web_servers() {
    print_header "Phase 4/7: Web Server Configuration"
    
    local existing_count=0
    if [[ "$MODE" == "add" ]]; then
        existing_count=$(grep -c "ansible_host:" "${CONFIG[inventory_dir]}/hosts.yml" 2>/dev/null || echo "0")
        print_info "Current web servers: $existing_count"
    fi
    
    local max_additional=$((MAX_WEB_SERVERS - existing_count))
    local count=$(ask "How many web servers do you want to deploy?" "1")
    
    if ((count > max_additional)); then
        print_error "Maximum $max_additional additional servers allowed (limit: $MAX_WEB_SERVERS)"
        exit 1
    fi
    
    if ((count == 0)); then
        print_warning "No web servers to configure"
        return
    fi
    
    CONFIG[web_count]=$count
    
    print_info "You will need $count IP address(es)"
    
    local next_id=$((existing_count + 1))
    
    local ssh_user=$(ask "SSH user for web servers" "deploy")
    CONFIG[web_ssh_user]="$ssh_user"
    
    print_info "All web servers will use user: $ssh_user"
    echo
    
    for ((i=1; i<=count; i++)); do
        local server_id=$((next_id + i - 1))
        local server_name="${ENVIRONMENT}-web-$(printf '%02d' $server_id)"
        
        print_header "Web Server #${i}: $server_name"
        
        local ip
        local port=3000
        local valid_ip=false
        
        while ! $valid_ip; do
            ip=$(ask "IP address")
            
            if ! validate_ip "$ip"; then
                print_error "Invalid IP format"
                continue
            fi
            
            print_success "Valid IP format"
            
            if ! check_ip_conflict "$ip" "$port"; then
                print_error "IP $ip:$port already in use"
                
                if ask_yes_no "Use same IP with different port?" "n"; then
                    port=$(ask "Port for this instance" "3001")
                    if check_ip_conflict "$ip" "$port"; then
                        valid_ip=true
                    fi
                fi
            else
                valid_ip=true
            fi
        done
        
        local custom_hostname=$(ask "Custom hostname (optional, press Enter for default)")
        local hostname="${custom_hostname:-$server_name}"
        
        # Store server data: name|ip|port|hostname
        WEB_SERVERS+=("${server_name}|${ip}|${port}|${hostname}")
        
        print_success "Configured: $hostname → $ip:$port"
        echo
    done
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Web Servers Summary"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    for ((i=0; i<${#WEB_SERVERS[@]}; i++)); do
        IFS='|' read -r name ip port hostname <<< "${WEB_SERVERS[$i]}"
        echo -e "  $((i+1)). $hostname → $ip:$port"
    done
    echo
    
    if ! ask_yes_no "Continue?" "y"; then
        exit 1
    fi
}

##############################################################################
# Phase 5: VPS Configuration - Database
##############################################################################

configure_database() {
    if [[ "$ENABLE_DB" != "true" ]]; then
        return
    fi
    
    print_header "Phase 5/7: Database Server Configuration"
    
    if ask_yes_no "Use one of the web servers for database?" "y"; then
        echo
        echo "Select web server:"
        for ((i=0; i<${#WEB_SERVERS[@]}; i++)); do
            IFS='|' read -r name ip port hostname <<< "${WEB_SERVERS[$i]}"
            echo "  $((i+1))) $hostname ($ip)"
        done
        echo
        
        echo -ne "${BOLD}?${NC} Choice ${DIM}[1]${NC}: " >&2
        read -r choice </dev/tty
        choice="${choice:-1}"
        IFS='|' read -r name ip port hostname <<< "${WEB_SERVERS[$((choice-1))]}"
        
        DB_SERVER[name]="$name"
        DB_SERVER[ip]="$ip"
        DB_SERVER[user]="${CONFIG[web_ssh_user]}"
        DB_SERVER[hostname]="$hostname"
    else
        print_arrow "Configuring dedicated database server"
        echo
        
        local db_ip
        local valid_ip=false
        
        while ! $valid_ip; do
            db_ip=$(ask "Database server IP")
            
            if ! validate_ip "$db_ip"; then
                print_error "Invalid IP format"
                continue
            fi
            
            print_success "Valid IP format"
            
            if ! check_ip_conflict "$db_ip" "5432"; then
                print_error "IP conflict detected"
                continue
            fi
            
            print_success "No IP conflict"
            valid_ip=true
        done
        
        local db_user=$(ask "SSH user" "root")
        local custom_hostname=$(ask "Custom hostname (optional, press Enter for default)")
        local db_name="${ENVIRONMENT}-db-01"
        local hostname="${custom_hostname:-$db_name}"
        
        DB_SERVER[name]="$db_name"
        DB_SERVER[ip]="$db_ip"
        DB_SERVER[user]="$db_user"
        DB_SERVER[hostname]="$hostname"
    fi
    
    print_success "Database configured: ${DB_SERVER[hostname]} → ${DB_SERVER[ip]}"
}

##############################################################################
# Phase 6: VPS Configuration - Monitoring
##############################################################################

configure_monitoring() {
    if [[ "$ENABLE_MONITORING" != "true" ]]; then
        return
    fi
    
    print_header "Phase 6/7: Monitoring Configuration"
    
    local all_servers=("${WEB_SERVERS[@]}")
    if [[ -n "${DB_SERVER[name]:-}" ]]; then
        all_servers+=("${DB_SERVER[name]}|${DB_SERVER[ip]}|5432|${DB_SERVER[hostname]}")
    fi
    
    echo "Monitoring will track all configured servers:"
    for server_data in "${all_servers[@]}"; do
        IFS='|' read -r _ ip _ hostname <<< "$server_data"
        echo -e "  ${GREEN}•${NC} $hostname ($ip)"
    done
    echo
    
    echo "Where do you want to install Prometheus + Grafana?"
    for ((i=0; i<${#all_servers[@]}; i++)); do
        IFS='|' read -r _ ip _ hostname <<< "${all_servers[$i]}"
        echo "  $((i+1))) $hostname ($ip)"
    done
    echo "  $((${#all_servers[@]}+1))) Use a new dedicated server"
    echo
    
    echo -ne "${BOLD}?${NC} Choice ${DIM}[1]${NC}: " >&2
    read -r choice </dev/tty
    choice="${choice:-1}"
    
    if ((choice > ${#all_servers[@]})); then
        # New dedicated server
        local mon_ip=$(ask "Monitoring server IP")
        local mon_user=$(ask "SSH user" "root")
        local mon_name="${ENVIRONMENT}-monitoring-01"
        
        MONITORING_SERVER[name]="$mon_name"
        MONITORING_SERVER[ip]="$mon_ip"
        MONITORING_SERVER[user]="$mon_user"
        MONITORING_SERVER[hostname]="$mon_name"
    else
        IFS='|' read -r name ip _ hostname <<< "${all_servers[$((choice-1))]}"
        
        MONITORING_SERVER[name]="$name"
        MONITORING_SERVER[ip]="$ip"
        MONITORING_SERVER[user]="${CONFIG[web_ssh_user]}"
        MONITORING_SERVER[hostname]="$hostname"
    fi
    
    print_success "Monitoring configured on: ${MONITORING_SERVER[hostname]} (${MONITORING_SERVER[ip]})"
    echo
    print_info "Services:"
    echo "  • Prometheus → http://${MONITORING_SERVER[ip]}:9090"
    echo "  • Grafana    → http://${MONITORING_SERVER[ip]}:3001"
}

##############################################################################
# Phase 7: Application Configuration
##############################################################################

configure_application() {
    print_header "Phase 7/7: Application Configuration"
    
    CONFIG[app_name]=$(ask "Application name" "myapp")
    CONFIG[app_repo]=$(ask "Git repository URL" "git@github.com:user/repo.git")
    CONFIG[app_branch]=$(ask "Branch to deploy" "$DEFAULT_BRANCH")
    
    # Test Git repository
    check_git_repo "${CONFIG[app_repo]}" "${CONFIG[app_branch]}"
    
    CONFIG[nodejs_version]=$(ask "Node.js version" "20")
    CONFIG[app_port]=$(ask "Application port" "3000")
    
    print_success "Application configured"
}

##############################################################################
# SSH Connection Tests
##############################################################################

test_all_connections() {
    print_header "Testing SSH Connections"
    
    local test_results=()
    FAILED_CONNECTIONS=()
    
    # Test web servers
    for server_data in "${WEB_SERVERS[@]}"; do
        IFS='|' read -r name ip _ hostname <<< "$server_data"
        
        print_arrow "Testing $hostname ($ip)..."
        
        if test_ssh_connection "${CONFIG[web_ssh_user]}" "$ip" "${CONFIG[ssh_private_key]}"; then
            print_success "Connection successful"
            test_results+=("$hostname|$ip|success")
        else
            print_error "Connection failed"
            test_results+=("$hostname|$ip|failed")
            FAILED_CONNECTIONS+=("$hostname|$ip")
        fi
    done
    
    # Test database server
    if [[ -n "${DB_SERVER[name]:-}" ]]; then
        print_arrow "Testing ${DB_SERVER[hostname]} (${DB_SERVER[ip]})..."
        
        if test_ssh_connection "${DB_SERVER[user]}" "${DB_SERVER[ip]}" "${CONFIG[ssh_private_key]}"; then
            print_success "Connection successful"
            test_results+=("${DB_SERVER[hostname]}|${DB_SERVER[ip]}|success")
        else
            print_error "Connection failed"
            test_results+=("${DB_SERVER[hostname]}|${DB_SERVER[ip]}|failed")
            FAILED_CONNECTIONS+=("${DB_SERVER[hostname]}|${DB_SERVER[ip]}")
        fi
    fi
    
    # Test monitoring server
    if [[ -n "${MONITORING_SERVER[name]:-}" && "${MONITORING_SERVER[name]}" != "${DB_SERVER[name]:-}" ]]; then
        IFS='|' read -r _ ip _ _ <<< "${WEB_SERVERS[0]}"
        if [[ "${MONITORING_SERVER[ip]}" != "$ip" ]]; then
            print_arrow "Testing ${MONITORING_SERVER[hostname]} (${MONITORING_SERVER[ip]})..."
            
            if test_ssh_connection "${MONITORING_SERVER[user]}" "${MONITORING_SERVER[ip]}" "${CONFIG[ssh_private_key]}"; then
                print_success "Connection successful"
                test_results+=("${MONITORING_SERVER[hostname]}|${MONITORING_SERVER[ip]}|success")
            else
                print_error "Connection failed"
                test_results+=("${MONITORING_SERVER[hostname]}|${MONITORING_SERVER[ip]}|failed")
                FAILED_CONNECTIONS+=("${MONITORING_SERVER[hostname]}|${MONITORING_SERVER[ip]}")
            fi
        fi
    fi
    
    # Summary
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "SSH Test Results"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    for result in "${test_results[@]}"; do
        IFS='|' read -r hostname ip status <<< "$result"
        if [[ "$status" == "success" ]]; then
            echo -e "  ${GREEN}${CHECKMARK}${NC} $hostname ($ip)"
        else
            echo -e "  ${RED}${CROSSMARK}${NC} $hostname ($ip)"
        fi
    done
    echo
    
    if [[ ${#FAILED_CONNECTIONS[@]} -gt 0 ]]; then
        print_warning "Some connections failed"
        echo
        echo "What do you want to do?"
        echo "  1) Save partial configuration (skip failed servers)"
        echo "  2) Show troubleshooting tips"
        echo "  3) Cancel setup"
        echo
        
        echo -ne "${BOLD}?${NC} Choice ${DIM}[1]${NC}: " >&2
        read -r choice </dev/tty
        choice="${choice:-1}"
        
        case "$choice" in
            1)
                print_info "Continuing with successful connections only"
                remove_failed_servers
                ;;
            2)
                show_ssh_troubleshooting
                exit 1
                ;;
            3)
                exit 1
                ;;
        esac
    fi
}

remove_failed_servers() {
    local new_web_servers=()
    
    for server_data in "${WEB_SERVERS[@]}"; do
        IFS='|' read -r name ip _ hostname <<< "$server_data"
        local failed=false
        
        for failed_conn in "${FAILED_CONNECTIONS[@]}"; do
            IFS='|' read -r failed_host failed_ip <<< "$failed_conn"
            if [[ "$hostname" == "$failed_host" || "$ip" == "$failed_ip" ]]; then
                failed=true
                break
            fi
        done
        
        if ! $failed; then
            new_web_servers+=("$server_data")
        fi
    done
    
    WEB_SERVERS=("${new_web_servers[@]}")
}

show_ssh_troubleshooting() {
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Troubleshooting SSH Connection Failures"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    for failed in "${FAILED_CONNECTIONS[@]}"; do
        IFS='|' read -r hostname ip <<< "$failed"
        
        echo "Failed server: $hostname ($ip)"
        echo
        echo "Possible solutions:"
        echo "  1. Check if SSH is running:"
        echo "     ssh ${CONFIG[web_ssh_user]}@$ip"
        echo
        echo "  2. Verify firewall allows SSH (port 22):"
        echo "     # On the VPS:"
        echo "     sudo ufw status"
        echo
        echo "  3. Check SSH key is added:"
        echo "     ssh-add -l"
        echo
        echo "  4. Try with password authentication:"
        echo "     ssh -o PreferredAuthentications=password ${CONFIG[web_ssh_user]}@$ip"
        echo
        echo "  5. Verify public key is in authorized_keys:"
        echo "     # On the VPS:"
        echo "     cat ~/.ssh/authorized_keys"
        echo
    done
    
    save_state "$STATE_FILE"
    
    echo "Configuration saved at:"
    echo "  $STATE_FILE"
    echo
    echo "To resume setup after fixing issues:"
    echo "  ./setup.sh $ENVIRONMENT --resume"
    echo
}

##############################################################################
# Generate Configuration Files
##############################################################################

generate_hosts_yml() {
    local hosts_file="${CONFIG[inventory_dir]}/hosts.yml"
    
    cat > "$hosts_file" << 'EOF'
---
all:
  children:
    webservers:
      hosts:
EOF
    
    # Add web servers
    for server_data in "${WEB_SERVERS[@]}"; do
        IFS='|' read -r name ip port hostname <<< "$server_data"
        
        cat >> "$hosts_file" << EOF
        ${name}:
          ansible_host: ${ip}
          ansible_user: ${CONFIG[web_ssh_user]}
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ${CONFIG[ssh_private_key]}
          ansible_become: yes
          app_port: ${port}
EOF
    done
    
    # Add database
    if [[ -n "${DB_SERVER[name]:-}" ]]; then
        cat >> "$hosts_file" << 'EOF'
    
    dbservers:
      hosts:
EOF
        
        cat >> "$hosts_file" << EOF
        ${DB_SERVER[name]}:
          ansible_host: ${DB_SERVER[ip]}
          ansible_user: ${DB_SERVER[user]}
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ${CONFIG[ssh_private_key]}
          ansible_become: yes
EOF
    fi
    
    # Add monitoring
    if [[ -n "${MONITORING_SERVER[name]:-}" ]]; then
        cat >> "$hosts_file" << 'EOF'
    
    monitoring:
      hosts:
EOF
        
        cat >> "$hosts_file" << EOF
        ${MONITORING_SERVER[name]}:
          ansible_host: ${MONITORING_SERVER[ip]}
          ansible_user: ${MONITORING_SERVER[user]}
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ${CONFIG[ssh_private_key]}
          ansible_become: yes
EOF
        
        # Add monitoring targets
        cat >> "$hosts_file" << 'EOF'
      vars:
        prometheus_targets:
          - targets:
EOF
        
        for server_data in "${WEB_SERVERS[@]}"; do
            IFS='|' read -r _ ip _ _ <<< "$server_data"
            echo "              - '${ip}:9100'" >> "$hosts_file"
        done
        
        if [[ -n "${DB_SERVER[ip]:-}" ]]; then
            echo "              - '${DB_SERVER[ip]}:9100'" >> "$hosts_file"
        fi
        
        cat >> "$hosts_file" << 'EOF'
            labels:
              job: 'node_exporter'
          - targets:
EOF
        
        for server_data in "${WEB_SERVERS[@]}"; do
            IFS='|' read -r _ ip port _ <<< "$server_data"
            echo "              - '${ip}:${port}'" >> "$hosts_file"
        done
        
        cat >> "$hosts_file" << 'EOF'
            labels:
              job: 'nodejs_app'
EOF
    fi
    
    log "Generated hosts.yml: $hosts_file"
    print_success "Generated: inventory/${ENVIRONMENT}/hosts.yml"
}

generate_group_vars() {
    local all_vars="${SCRIPT_DIR}/group_vars/all.yml"
    
    cat > "$all_vars" << EOF
---
# Global variables for all hosts

# Deploy user configuration
deploy_user: ${CONFIG[web_ssh_user]}
deploy_user_groups:
  - sudo
  - www-data

# SSH Configuration
ssh_port: 22
ssh_key_path: "${CONFIG[ssh_public_key]}"

# Node.js Configuration
nodejs_version: "${CONFIG[nodejs_version]}"

# Application Configuration
app_name: ${CONFIG[app_name]}
app_port: ${CONFIG[app_port]}
app_repo: "${CONFIG[app_repo]}"
app_branch: "${CONFIG[app_branch]}"
app_dir: "/var/www/{{ app_name }}"
app_releases_dir: "{{ app_dir }}/releases"
app_current_dir: "{{ app_dir }}/current"
app_shared_dir: "{{ app_dir }}/shared"

# PM2 Configuration
pm2_app_name: "{{ app_name }}"
pm2_instances: 2
pm2_max_memory: "512M"

# Environment
app_environment: "${ENVIRONMENT}"

# Timezone
timezone: "Europe/Paris"

# Backup Configuration
backup_dir: "/var/backups"
backup_retention_days: 7
EOF
    
    # Add load balancer config if multiple web servers
    if [[ ${#WEB_SERVERS[@]} -gt 1 ]]; then
        cat >> "$all_vars" << 'EOF'

# Load Balancer Configuration
load_balancer:
  enabled: true
  algorithm: least_conn
  backend_servers:
EOF
        
        for server_data in "${WEB_SERVERS[@]}"; do
            IFS='|' read -r name ip port _ <<< "$server_data"
            echo "    - server ${name} ${ip}:${port} weight=1 max_fails=3 fail_timeout=30s" >> "$all_vars"
        done
        
        cat >> "$all_vars" << 'EOF'
  health_check:
    uri: "/health"
    interval: 10s
    timeout: 5s
EOF
    fi
    
    # Add database config
    if [[ -n "${DB_SERVER[ip]:-}" ]]; then
        cat >> "$all_vars" << EOF

# Database Configuration
database:
  host: ${DB_SERVER[ip]}
  port: 5432
  name: "{{ app_name }}_{{ app_environment }}"
EOF
    fi
    
    # Add monitoring config
    if [[ -n "${MONITORING_SERVER[ip]:-}" ]]; then
        cat >> "$all_vars" << EOF

# Monitoring Configuration
monitoring:
  prometheus:
    host: ${MONITORING_SERVER[ip]}
    port: 9090
  grafana:
    host: ${MONITORING_SERVER[ip]}
    port: 3001
EOF
    fi
    
    log "Generated group_vars/all.yml"
    print_success "Generated: group_vars/all.yml"
}

##############################################################################
# Final Summary
##############################################################################

show_final_summary() {
    print_header "Setup Complete!"
    
    echo "Configuration saved successfully!"
    echo
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Environment: ${ENVIRONMENT}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    if [[ ${#WEB_SERVERS[@]} -gt 0 ]]; then
        echo "Web Servers (${#WEB_SERVERS[@]}):"
        for server_data in "${WEB_SERVERS[@]}"; do
            IFS='|' read -r name ip port hostname <<< "$server_data"
            echo "  • $hostname → $ip:$port"
        done
        echo
    fi
    
    if [[ -n "${DB_SERVER[ip]:-}" ]]; then
        echo "Database Server:"
        echo "  • ${DB_SERVER[hostname]} → ${DB_SERVER[ip]}"
        echo
    fi
    
    if [[ -n "${MONITORING_SERVER[ip]:-}" ]]; then
        echo "Monitoring Server:"
        echo "  • ${MONITORING_SERVER[hostname]} → ${MONITORING_SERVER[ip]}"
        echo "  • Prometheus: http://${MONITORING_SERVER[ip]}:9090"
        echo "  • Grafana: http://${MONITORING_SERVER[ip]}:3001"
        echo
    fi
    
    echo "Application:"
    echo "  • Name: ${CONFIG[app_name]}"
    echo "  • Repository: ${CONFIG[app_repo]}"
    echo "  • Branch: ${CONFIG[app_branch]}"
    echo "  • Node.js: v${CONFIG[nodejs_version]}"
    echo
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Next Steps"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    echo "1. Provision your servers:"
    echo "   ./deploy.sh provision ${ENVIRONMENT}"
    echo
    echo "2. Deploy your application:"
    echo "   ./deploy.sh deploy ${ENVIRONMENT}"
    echo
    echo "3. Configure SSL (after DNS is pointing to your server):"
    echo "   ./configure-ssl.sh ${ENVIRONMENT}"
    echo
    echo "4. Check application status:"
    echo "   ./deploy.sh status ${ENVIRONMENT}"
    echo
    echo "5. View logs:"
    echo "   ssh ${CONFIG[web_ssh_user]}@${WEB_SERVERS[0]#*|} 'pm2 logs'"
    echo
    
    if [[ ${#FAILED_CONNECTIONS[@]} -gt 0 ]]; then
        print_warning "Note: Some servers had SSH connection issues"
        echo "Review the troubleshooting tips and add them later:"
        echo "  ./setup.sh ${ENVIRONMENT} --add-servers"
        echo
    fi
    
    echo "Setup log saved to:"
    echo "  $LOG_FILE"
    echo
}

##############################################################################
# Main Flow
##############################################################################

main() {
    # Parse arguments
    if [[ $# -ge 1 ]]; then
        ENVIRONMENT="$1"
        shift
        
        while [[ $# -gt 0 ]]; do
            case "$1" in
                --resume)
                    RESUME_MODE=true
                    ;;
                --add-servers)
                    MODE="add"
                    ;;
                *)
                    echo "Unknown option: $1"
                    exit 1
                    ;;
            esac
            shift
        done
    fi
    
    clear
    
    print_header "Multi-Server Setup Wizard"
    echo "Creates inventory and configuration for VPS deployment"
    echo
    
    log "=== Setup started ==="
    
    # Phase 1: Prerequisites
    check_prerequisites
    
    # Phase 2: Environment
    if [[ -z "$ENVIRONMENT" ]]; then
        setup_environment
    else
        CONFIG[inventory_dir]="${SCRIPT_DIR}/inventory/${ENVIRONMENT}"
        mkdir -p "${CONFIG[inventory_dir]}"
        STATE_FILE="${CONFIG[inventory_dir]}/.setup_state.yml"
        
        if [[ "$RESUME_MODE" == true && -f "$STATE_FILE" ]]; then
            load_state "$STATE_FILE"
        fi
    fi
    
    # Select services (only in create mode)
    if [[ "$MODE" == "create" ]]; then
        select_services
    fi
    
    # Phase 3: SSH Keys
    configure_ssh_keys
    
    # Phase 4: Web Servers
    if [[ "$ENABLE_WEB" == "true" ]]; then
        configure_web_servers
    fi
    
    # Phase 5: Database
    configure_database
    
    # Phase 6: Monitoring
    configure_monitoring
    
    # Phase 7: Application
    configure_application
    
    # Test SSH connections
    print_header "Final Review"
    echo "Ready to test SSH connections and save configuration"
    echo
    
    if ask_yes_no "Test SSH connections now?" "y"; then
        test_all_connections
    fi
    
    # Generate files
    print_header "Generating Configuration Files"
    
    generate_hosts_yml
    generate_group_vars
    
    # Create example files
    cp "${CONFIG[inventory_dir]}/hosts.yml" "${CONFIG[inventory_dir]}/hosts.yml.example"
    print_success "Generated: inventory/${ENVIRONMENT}/hosts.yml.example"
    
    # Final summary
    show_final_summary
    
    log "=== Setup completed successfully ==="
}

##############################################################################
# Entry Point
##############################################################################

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
