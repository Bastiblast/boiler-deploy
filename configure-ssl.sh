#!/bin/bash

#==============================================================================
# SSL Configuration Script for boiler-deploy
# Interactively configures HTTPS with Let's Encrypt
#==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${SCRIPT_DIR}/.ssl_backup"

# Global variables
DETECTED_APPS=()
DETECTED_IPS=()
DETECTED_ENVS=()
SSL_ENABLED=""
CURRENT_DOMAINS=()
CURRENT_EMAIL=""
SELECTED_ENV=""
SELECTED_APP=""
SELECTED_IP=""
NEW_DOMAINS=()
NEW_EMAIL=""
ENABLE_REDIRECT=true

#==============================================================================
# Helper Functions
#==============================================================================

print_header() {
    echo ""
    echo -e "${CYAN}╔═══════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}  $1"
    echo -e "${CYAN}╚═══════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo -e "${BLUE}→${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

prompt_user() {
    local prompt="$1"
    local default="$2"
    local response
    
    if [ -n "$default" ]; then
        read -p "$(echo -e "${CYAN}?${NC} ${prompt} [${default}]: ")" response
        echo "${response:-$default}"
    else
        read -p "$(echo -e "${CYAN}?${NC} ${prompt}: ")" response
        echo "$response"
    fi
}

prompt_yes_no() {
    local prompt="$1"
    local default="$2"
    local response
    
    if [ "$default" = "y" ]; then
        read -p "$(echo -e "${CYAN}?${NC} ${prompt} (Y/n): ")" response
        response="${response:-y}"
    else
        read -p "$(echo -e "${CYAN}?${NC} ${prompt} (y/N): ")" response
        response="${response:-n}"
    fi
    
    [[ "$response" =~ ^[Yy]$ ]]
}

validate_email() {
    local email="$1"
    if [[ "$email" =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
        return 0
    else
        return 1
    fi
}

validate_domain() {
    local domain="$1"
    if [[ "$domain" =~ ^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$ ]]; then
        return 0
    else
        return 1
    fi
}

#==============================================================================
# Detection Functions
#==============================================================================

detect_current_config() {
    print_step "Detecting deployed applications..."
    
    # Find inventory files
    local inventory_files=$(find inventory/ -name "hosts.yml" -type f 2>/dev/null)
    
    if [ -z "$inventory_files" ]; then
        print_error "No inventory files found"
        exit 1
    fi
    
    # Parse inventory files for deployed servers
    while IFS= read -r inv_file; do
        local env_name=$(basename $(dirname "$inv_file"))
        
        # Extract IPs from inventory (simple parsing)
        local ips=$(grep "ansible_host:" "$inv_file" | awk '{print $2}' | sort -u)
        
        if [ -n "$ips" ]; then
            while IFS= read -r ip; do
                DETECTED_IPS+=("$ip")
                DETECTED_ENVS+=("$env_name")
            done <<< "$ips"
        fi
    done <<< "$inventory_files"
    
    # Read app name from group_vars
    if [ -f "group_vars/all.yml" ]; then
        local app_name=$(grep "^app_name:" group_vars/all.yml | awk '{print $2}' | tr -d '"' | tr -d "'")
        if [ -n "$app_name" ]; then
            DETECTED_APPS+=("$app_name")
        fi
    fi
    
    # Read current SSL configuration
    if [ -f "group_vars/webservers.yml" ]; then
        SSL_ENABLED=$(grep "^ssl_enabled:" group_vars/webservers.yml | awk '{print $2}')
        CURRENT_EMAIL=$(grep "^ssl_certbot_email:" group_vars/webservers.yml | awk '{print $2}' | tr -d '"' | tr -d "'")
        
        # Read domains (YAML list parsing)
        local in_domains=false
        while IFS= read -r line; do
            if [[ "$line" =~ ^ssl_domains: ]]; then
                in_domains=true
                continue
            elif $in_domains; then
                if [[ "$line" =~ ^[[:space:]]*-[[:space:]]*\"?([^\"]+)\"? ]]; then
                    local domain="${BASH_REMATCH[1]}"
                    CURRENT_DOMAINS+=("$domain")
                elif [[ ! "$line" =~ ^[[:space:]]*- ]]; then
                    break
                fi
            fi
        done < group_vars/webservers.yml
    fi
    
    # Show detected configuration
    if [ ${#DETECTED_APPS[@]} -gt 0 ]; then
        print_success "Found application: ${DETECTED_APPS[0]}"
    fi
    
    if [ ${#DETECTED_IPS[@]} -gt 0 ]; then
        for i in "${!DETECTED_IPS[@]}"; do
            print_success "Found server: ${DETECTED_IPS[$i]} (${DETECTED_ENVS[$i]})"
        done
    fi
    
    if [ "$SSL_ENABLED" = "true" ]; then
        print_info "SSL Status: Enabled"
        if [ ${#CURRENT_DOMAINS[@]} -gt 0 ]; then
            print_info "Current domains: ${CURRENT_DOMAINS[*]}"
        fi
        if [ -n "$CURRENT_EMAIL" ]; then
            print_info "Current email: $CURRENT_EMAIL"
        fi
    else
        print_info "SSL Status: Disabled"
    fi
    
    if [ ${#DETECTED_IPS[@]} -eq 0 ]; then
        print_error "No servers detected in inventory"
        exit 1
    fi
}

#==============================================================================
# Prompt Functions
#==============================================================================

prompt_environment() {
    print_step "Select environment to configure"
    
    if [ ${#DETECTED_ENVS[@]} -eq 1 ]; then
        SELECTED_ENV="${DETECTED_ENVS[0]}"
        SELECTED_IP="${DETECTED_IPS[0]}"
        print_info "Using: ${SELECTED_ENV} (${SELECTED_IP})"
        return
    fi
    
    echo ""
    for i in "${!DETECTED_ENVS[@]}"; do
        echo "  $((i+1))) ${DETECTED_ENVS[$i]} - ${DETECTED_IPS[$i]}"
    done
    echo ""
    
    local choice
    while true; do
        choice=$(prompt_user "Select environment [1-${#DETECTED_ENVS[@]}]" "1")
        if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#DETECTED_ENVS[@]}" ]; then
            SELECTED_ENV="${DETECTED_ENVS[$((choice-1))]}"
            SELECTED_IP="${DETECTED_IPS[$((choice-1))]}"
            break
        else
            print_error "Invalid selection"
        fi
    done
}

prompt_domain_info() {
    print_step "Configure domains for SSL certificate"
    
    # Primary domain
    local default_domain=""
    if [ ${#CURRENT_DOMAINS[@]} -gt 0 ]; then
        default_domain="${CURRENT_DOMAINS[0]}"
    fi
    
    while true; do
        local domain=$(prompt_user "Primary domain (e.g., myapp.com)" "$default_domain")
        if validate_domain "$domain"; then
            NEW_DOMAINS+=("$domain")
            break
        else
            print_error "Invalid domain format"
        fi
    done
    
    # Add www subdomain
    if prompt_yes_no "Add www subdomain? (www.${NEW_DOMAINS[0]})" "y"; then
        NEW_DOMAINS+=("www.${NEW_DOMAINS[0]}")
    fi
    
    # Additional domains
    while prompt_yes_no "Add another domain?" "n"; do
        local domain=$(prompt_user "Additional domain")
        if validate_domain "$domain"; then
            NEW_DOMAINS+=("$domain")
        else
            print_error "Invalid domain format, skipping"
        fi
    done
    
    print_info "Domains to configure: ${NEW_DOMAINS[*]}"
}

prompt_email_info() {
    print_step "Configure Let's Encrypt notification email"
    
    local default_email="$CURRENT_EMAIL"
    
    while true; do
        NEW_EMAIL=$(prompt_user "Email for Let's Encrypt notifications" "$default_email")
        if validate_email "$NEW_EMAIL"; then
            break
        else
            print_error "Invalid email format"
        fi
    done
}

prompt_redirect_option() {
    if prompt_yes_no "Enable HTTP → HTTPS redirect?" "y"; then
        ENABLE_REDIRECT=true
    else
        ENABLE_REDIRECT=false
    fi
}

#==============================================================================
# Validation Functions
#==============================================================================

check_prerequisites() {
    print_step "Checking prerequisites..."
    
    # Check Ansible
    if ! command -v ansible >/dev/null 2>&1; then
        print_error "Ansible not found. Please install Ansible first."
        exit 1
    fi
    print_success "Ansible found"
    
    # Check SSH connectivity
    if ! ansible all -i "inventory/${SELECTED_ENV}" -m ping >/dev/null 2>&1; then
        print_warning "Cannot connect to server via SSH"
        if ! prompt_yes_no "Continue anyway?" "n"; then
            exit 1
        fi
    else
        print_success "SSH connectivity OK"
    fi
    
    # Check if port 80 is accessible
    print_info "Port 80 must be accessible for Let's Encrypt validation"
}

validate_dns() {
    print_step "Validating DNS configuration..."
    
    local all_valid=true
    
    for domain in "${NEW_DOMAINS[@]}"; do
        print_step "Checking DNS for $domain..."
        
        # Use dig to check DNS
        if command -v dig >/dev/null 2>&1; then
            local resolved_ip=$(dig +short "$domain" | tail -n1)
            
            if [ -z "$resolved_ip" ]; then
                print_warning "No A record found for $domain"
                print_info "Please add: $domain → $SELECTED_IP"
                all_valid=false
            elif [ "$resolved_ip" = "$SELECTED_IP" ]; then
                print_success "$domain → $resolved_ip"
            else
                print_warning "$domain points to $resolved_ip (expected $SELECTED_IP)"
                all_valid=false
            fi
        else
            print_warning "dig not installed, skipping DNS check"
            print_info "Please ensure $domain points to $SELECTED_IP"
        fi
    done
    
    if ! $all_valid; then
        echo ""
        print_warning "DNS configuration issues detected"
        print_info "Let's Encrypt requires domains to point to your server"
        echo ""
        
        if prompt_yes_no "Continue anyway? (Certificate may fail)" "n"; then
            print_info "Continuing..."
        else
            print_info "Waiting for DNS propagation..."
            read -p "Press Enter when DNS is configured..."
            validate_dns
        fi
    fi
}

#==============================================================================
# Configuration Functions
#==============================================================================

show_configuration_summary() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}CONFIGURATION SUMMARY${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "Environment:     $SELECTED_ENV"
    echo "Server:          $SELECTED_IP"
    echo "Application:     ${DETECTED_APPS[0]:-unknown}"
    echo "Domains:         ${NEW_DOMAINS[*]}"
    echo "Email:           $NEW_EMAIL"
    echo "Redirect HTTP:   $([ "$ENABLE_REDIRECT" = true ] && echo "Yes" || echo "No")"
    echo ""
    echo "Changes to apply:"
    echo "  • Update group_vars/webservers.yml"
    echo "  • Install SSL certificate via Let's Encrypt"
    echo "  • Configure Nginx for HTTPS"
    echo "  • Enable certificate auto-renewal"
    echo "  • Open port 443 in firewall"
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

backup_configuration() {
    print_step "Backing up current configuration..."
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_path="${BACKUP_DIR}/${timestamp}"
    
    mkdir -p "$backup_path"
    
    if [ -f "group_vars/webservers.yml" ]; then
        cp "group_vars/webservers.yml" "${backup_path}/webservers.yml"
        print_success "Backup saved to ${backup_path}"
    fi
}

update_ansible_vars() {
    print_step "Updating Ansible variables..."
    
    local webservers_file="group_vars/webservers.yml"
    
    if [ ! -f "$webservers_file" ]; then
        print_error "File not found: $webservers_file"
        exit 1
    fi
    
    # Create temporary file
    local temp_file=$(mktemp)
    
    # Update ssl_enabled
    sed 's/^ssl_enabled:.*/ssl_enabled: true/' "$webservers_file" > "$temp_file"
    mv "$temp_file" "$webservers_file"
    
    # Update ssl_certbot_email
    temp_file=$(mktemp)
    sed "s/^ssl_certbot_email:.*/ssl_certbot_email: \"$NEW_EMAIL\"/" "$webservers_file" > "$temp_file"
    mv "$temp_file" "$webservers_file"
    
    # Update ssl_domains (replace entire list)
    temp_file=$(mktemp)
    awk -v domains="${NEW_DOMAINS[*]}" '
    BEGIN { in_domains=0 }
    /^ssl_domains:/ { 
        print "ssl_domains:"
        split(domains, arr, " ")
        for (i in arr) {
            print "  - \"" arr[i] "\""
        }
        in_domains=1
        next
    }
    in_domains && /^[[:space:]]*-/ { next }
    in_domains && !/^[[:space:]]*-/ { in_domains=0 }
    !in_domains { print }
    ' "$webservers_file" > "$temp_file"
    mv "$temp_file" "$webservers_file"
    
    print_success "Variables updated"
}

apply_ssl_config() {
    print_step "Applying SSL configuration..."
    
    # Run Ansible playbook for nginx role
    print_info "Running Ansible playbook (this may take a few minutes)..."
    
    if ansible-playbook playbooks/provision.yml \
        -i "inventory/${SELECTED_ENV}" \
        --tags nginx \
        -e "ssl_enabled=true" \
        -e "ssl_certbot_email=${NEW_EMAIL}" \
        -e "ssl_domains=[$(printf '"%s",' "${NEW_DOMAINS[@]}" | sed 's/,$//')]"; then
        print_success "SSL configuration applied"
    else
        print_error "Failed to apply SSL configuration"
        print_info "Check the error above and run again"
        exit 1
    fi
}

verify_ssl() {
    print_step "Verifying SSL configuration..."
    
    # Test HTTPS access
    local primary_domain="${NEW_DOMAINS[0]}"
    
    print_info "Testing HTTPS access to $primary_domain..."
    
    if command -v curl >/dev/null 2>&1; then
        if curl -sI "https://${primary_domain}" >/dev/null 2>&1; then
            print_success "HTTPS is working!"
        else
            print_warning "HTTPS test failed (may need a few minutes to propagate)"
        fi
    fi
}

show_success_report() {
    echo ""
    print_header "✓ SSL Configuration Complete!"
    
    echo -e "${GREEN}Certificate Details:${NC}"
    echo "  • Domains: ${NEW_DOMAINS[*]}"
    echo "  • Issuer: Let's Encrypt"
    echo "  • Email: $NEW_EMAIL"
    echo "  • Auto-renewal: Enabled (via cron)"
    echo ""
    echo -e "${GREEN}Access your application:${NC}"
    for domain in "${NEW_DOMAINS[@]}"; do
        echo "  • https://${domain}"
    done
    echo ""
    echo -e "${CYAN}Next Steps:${NC}"
    echo "  • Test your application over HTTPS"
    echo "  • Update any hardcoded HTTP URLs to HTTPS"
    echo "  • Certificate will auto-renew 30 days before expiry"
    echo ""
    print_info "Backup saved in: ${BACKUP_DIR}"
    print_info "To rollback: restore files from backup and run provision again"
    echo ""
}

#==============================================================================
# Main Flow
#==============================================================================

main() {
    print_header "SSL Configuration Helper for boiler-deploy"
    
    echo "This script will help you configure HTTPS for your application using Let's Encrypt."
    echo ""
    
    # Detection phase
    detect_current_config
    
    # If SSL already enabled, ask for update
    if [ "$SSL_ENABLED" = "true" ]; then
        echo ""
        print_info "SSL is already configured"
        if ! prompt_yes_no "Update SSL configuration?" "y"; then
            echo "Exiting..."
            exit 0
        fi
    fi
    
    # Interactive prompts
    prompt_environment
    prompt_domain_info
    prompt_email_info
    prompt_redirect_option
    
    # Validation
    check_prerequisites
    validate_dns
    
    # Show summary and confirm
    show_configuration_summary
    
    if ! prompt_yes_no "Proceed with SSL configuration?" "y"; then
        echo "Cancelled by user"
        exit 0
    fi
    
    # Apply configuration
    backup_configuration
    update_ansible_vars
    apply_ssl_config
    verify_ssl
    
    # Success report
    show_success_report
}

# Run main
main "$@"
