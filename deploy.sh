#!/bin/bash
#==============================================================================
# Unified Deployment Script for boiler-deploy
# VPS-agnostic deployment for Node.js applications
#==============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Parse arguments - syntax: ./deploy.sh ACTION ENVIRONMENT
ACTION=${1:-deploy}
ENVIRONMENT=${2:-hostinger}

# Available environments (auto-detected from inventory directory)
VALID_ENVS=$(ls -d inventory/*/ 2>/dev/null | xargs -n 1 basename | tr '\n' ' ')

#==============================================================================
# Helper Functions
#==============================================================================

print_header() {
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
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
    echo -e "${BLUE}→${NC} $1"
}

show_usage() {
    cat << EOF
Usage: $0 ACTION [ENVIRONMENT]

ACTIONS:
  provision   - Full server setup (first time only)
  deploy      - Deploy application
  update      - Quick update (pull + restart)
  rollback    - Rollback to previous version
  check       - Dry-run verification (no changes)
  status      - Show PM2 services status

ENVIRONMENTS:
  Available: $VALID_ENVS
  Default: hostinger

EXAMPLES:
  $0 provision hostinger     # First-time server setup
  $0 deploy                  # Deploy to hostinger (default)
  $0 deploy dev              # Deploy to dev
  $0 update hostinger        # Quick update
  $0 status dev              # Check PM2 status

For SSL configuration: ./configure-ssl.sh
For health check: ./health_check.sh [environment]

EOF
    exit 1
}

#==============================================================================
# Validation Functions
#==============================================================================

validate_environment() {
    if [[ ! " $VALID_ENVS " =~ " $ENVIRONMENT " ]]; then
        print_error "Invalid environment: $ENVIRONMENT"
        echo "Valid environments: $VALID_ENVS"
        exit 1
    fi
}

check_inventory() {
    local inventory_path="inventory/$ENVIRONMENT"
    
    if [ ! -d "$inventory_path" ]; then
        print_error "Inventory directory not found: $inventory_path"
        exit 1
    fi
    
    if [ ! -f "$inventory_path/hosts.yml" ]; then
        print_warning "Inventory file not found: $inventory_path/hosts.yml"
        
        if [ -f "$inventory_path/hosts.yml.example" ]; then
            echo ""
            print_info "Example file found. Creating from template..."
            cp "$inventory_path/hosts.yml.example" "$inventory_path/hosts.yml"
            print_success "Created: $inventory_path/hosts.yml"
            echo ""
            print_warning "Please edit the inventory file with your VPS details:"
            echo "  vim $inventory_path/hosts.yml"
            echo ""
            exit 1
        else
            print_error "No inventory file or template found"
            exit 1
        fi
    fi
}

check_connectivity() {
    local inventory_path="inventory/$ENVIRONMENT"
    
    print_info "Checking connectivity to $ENVIRONMENT servers..."
    
    if ansible all -i "$inventory_path" -m ping > /dev/null 2>&1; then
        print_success "Connection established"
        
        # Get server IP from inventory
        local server_ip=$(grep "ansible_host:" "$inventory_path/hosts.yml" | head -1 | awk '{print $2}')
        if [ -n "$server_ip" ]; then
            echo -e "${CYAN}  Server: $server_ip${NC}"
        fi
    else
        print_warning "Cannot connect to servers"
        echo ""
        echo "Troubleshooting:"
        echo "  1. Check SSH access to your VPS"
        echo "  2. Verify inventory file: $inventory_path/hosts.yml"
        echo "  3. Test manually: ssh user@your-vps-ip"
        echo ""
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

confirm_action() {
    local action_name="$1"
    local warning_msg="$2"
    
    echo ""
    print_warning "$warning_msg"
    echo ""
    read -p "Continue with $action_name on $ENVIRONMENT? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled."
        exit 0
    fi
}

#==============================================================================
# Main Actions
#==============================================================================

run_provision() {
    print_header "Provisioning $ENVIRONMENT Environment"
    
    print_info "This will install:"
    echo "  • PostgreSQL database"
    echo "  • Node.js runtime"
    echo "  • Nginx web server"
    echo "  • PM2 process manager"
    echo "  • Security packages (UFW, fail2ban)"
    echo "  • Monitoring stack (Prometheus, Grafana)"
    echo ""
    
    confirm_action "provision" "This will perform a full server setup"
    
    print_info "Running Ansible playbook..."
    ansible-playbook playbooks/provision.yml -i "inventory/$ENVIRONMENT"
}

run_deploy() {
    print_header "Deploying to $ENVIRONMENT"
    
    print_info "Running deployment playbook..."
    ansible-playbook playbooks/deploy.yml -i "inventory/$ENVIRONMENT"
}

run_update() {
    print_header "Updating $ENVIRONMENT"
    
    print_info "Pulling latest code and restarting services..."
    ansible-playbook playbooks/update.yml -i "inventory/$ENVIRONMENT"
}

run_rollback() {
    print_header "Rolling Back $ENVIRONMENT"
    
    confirm_action "rollback" "This will revert to the previous version"
    
    print_info "Running rollback playbook..."
    ansible-playbook playbooks/rollback.yml -i "inventory/$ENVIRONMENT"
}

run_check() {
    print_header "Checking $ENVIRONMENT Configuration"
    
    print_info "Running dry-run (no changes will be made)..."
    ansible-playbook playbooks/deploy.yml -i "inventory/$ENVIRONMENT" --check
}

run_status() {
    print_header "Status for $ENVIRONMENT"
    
    print_info "Querying PM2 status..."
    ansible webservers -i "inventory/$ENVIRONMENT" -a "pm2 status" -u deploy 2>/dev/null || {
        print_warning "PM2 not installed or not accessible"
        echo "Run 'provision' first if this is a new server"
    }
}

show_success() {
    local action="$1"
    
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ $action completed successfully!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    # Get server IP from inventory
    local server_ip=$(grep "ansible_host:" "inventory/$ENVIRONMENT/hosts.yml" | head -1 | awk '{print $2}')
    
    if [ -n "$server_ip" ]; then
        echo "Access your application:"
        echo -e "  ${CYAN}→${NC} http://$server_ip"
        echo ""
    fi
    
    echo "Useful commands:"
    if [ -n "$server_ip" ]; then
        echo -e "  ${CYAN}→${NC} View logs:    ssh deploy@$server_ip 'pm2 logs'"
        echo -e "  ${CYAN}→${NC} PM2 status:   ./deploy.sh status $ENVIRONMENT"
        echo -e "  ${CYAN}→${NC} Health check: ./health_check.sh $ENVIRONMENT"
    else
        echo -e "  ${CYAN}→${NC} PM2 status:   ./deploy.sh status $ENVIRONMENT"
        echo -e "  ${CYAN}→${NC} Health check: ./health_check.sh $ENVIRONMENT"
    fi
    
    if [ "$action" = "Provision" ]; then
        echo ""
        echo "Next steps:"
        echo -e "  ${CYAN}→${NC} Configure SSL: ./configure-ssl.sh"
        echo -e "  ${CYAN}→${NC} Deploy app:     ./deploy.sh deploy $ENVIRONMENT"
    fi
    
    echo ""
}

#==============================================================================
# Main Flow
#==============================================================================

# Handle help
if [ "$ACTION" = "help" ] || [ "$ACTION" = "--help" ] || [ "$ACTION" = "-h" ]; then
    show_usage
fi

# Validate inputs
validate_environment

# Show header
print_header "Deployment to $ENVIRONMENT"

# Pre-flight checks
check_inventory
check_connectivity

# Execute action
case $ACTION in
    provision)
        run_provision
        show_success "Provision"
        ;;
    deploy)
        run_deploy
        show_success "Deployment"
        ;;
    update)
        run_update
        show_success "Update"
        ;;
    rollback)
        run_rollback
        show_success "Rollback"
        ;;
    check)
        run_check
        echo -e "${GREEN}✓ Check completed${NC}"
        ;;
    status)
        run_status
        ;;
    *)
        print_error "Unknown action: $ACTION"
        echo ""
        show_usage
        ;;
esac
