#!/bin/bash

# Multi-Container Test Environment Setup
# Creates 3 Docker containers to test parallel provisioning and deployment

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
IMAGE_NAME="boiler-test-ubuntu"
SSH_KEY_DIR="$HOME/.ssh"
SSH_PRIVATE_KEY="$SSH_KEY_DIR/boiler_test_rsa"
SSH_PUBLIC_KEY="$SSH_KEY_DIR/boiler_test_rsa.pub"

# Container configurations
declare -A CONTAINERS=(
    ["test-web-01"]="2222:8080:8443:3000"
    ["test-web-02"]="2223:8081:8444:3001"
    ["test-web-03"]="2224:8082:8445:3002"
)

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[✓]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[!]${NC} $1"; }
print_error() { echo -e "${RED}[✗]${NC} $1"; }

# Check if image exists
check_image() {
    if ! docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
        print_error "Image $IMAGE_NAME not found. Please run tests/test-docker-vps.sh setup first."
        exit 1
    fi
    print_success "Image $IMAGE_NAME found"
}

# Create and start container
start_container() {
    local name=$1
    local ports=$2
    IFS=':' read -r ssh_port http_port https_port app_port <<< "$ports"
    
    # Check if container already exists
    if docker ps -a --format '{{.Names}}' | grep -q "^${name}$"; then
        if docker ps --format '{{.Names}}' | grep -q "^${name}$"; then
            print_warning "Container $name already running"
            return 0
        else
            print_info "Starting existing container $name..."
            docker start "$name" > /dev/null
            sleep 3
            print_success "Container $name started"
            return 0
        fi
    fi
    
    print_info "Creating container $name (SSH:$ssh_port, HTTP:$http_port)..."
    
    docker run -d \
        --name "$name" \
        --privileged \
        --tmpfs /tmp \
        --tmpfs /run \
        --tmpfs /run/lock \
        -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
        --cgroupns=host \
        --stop-signal SIGRTMIN+3 \
        -p "${ssh_port}:22" \
        -p "${http_port}:80" \
        -p "${https_port}:443" \
        "$IMAGE_NAME" > /dev/null
    
    sleep 5
    print_success "Container $name created and started"
}

# Configure SSH for container
configure_ssh() {
    local name=$1
    
    print_info "Configuring SSH for $name..."
    
    # Copy SSH public key
    docker exec "$name" bash -c "mkdir -p /root/.ssh && chmod 700 /root/.ssh"
    docker cp "$SSH_PUBLIC_KEY" "${name}:/root/.ssh/authorized_keys"
    docker exec "$name" bash -c "chmod 600 /root/.ssh/authorized_keys"
    docker exec "$name" bash -c "chown -R root:root /root/.ssh"
    
    # Restart SSH
    docker exec "$name" systemctl restart ssh > /dev/null 2>&1
    
    print_success "SSH configured for $name"
}

# Test SSH connection
test_ssh() {
    local name=$1
    local ports=$2
    IFS=':' read -r ssh_port _ _ _ <<< "$ports"
    
    print_info "Testing SSH connection to $name..."
    
    if ssh -i "$SSH_PRIVATE_KEY" \
           -o StrictHostKeyChecking=no \
           -o UserKnownHostsFile=/dev/null \
           -o ConnectTimeout=5 \
           -p "$ssh_port" \
           root@localhost "echo 'SSH OK'" > /dev/null 2>&1; then
        print_success "SSH connection to $name successful"
        return 0
    else
        print_error "SSH connection to $name failed"
        return 1
    fi
}

# Create all containers
setup_all() {
    echo -e "${GREEN}=== Multi-Container Test Environment Setup ===${NC}"
    echo ""
    
    check_image
    
    for name in "${!CONTAINERS[@]}"; do
        start_container "$name" "${CONTAINERS[$name]}"
        configure_ssh "$name"
        test_ssh "$name" "${CONTAINERS[$name]}"
        echo ""
    done
    
    print_success "All containers are ready!"
    echo ""
    display_info
}

# Stop all containers
stop_all() {
    print_info "Stopping all test containers..."
    for name in "${!CONTAINERS[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "^${name}$"; then
            docker stop "$name" > /dev/null
            print_success "Stopped $name"
        fi
    done
}

# Start all containers
start_all() {
    print_info "Starting all test containers..."
    for name in "${!CONTAINERS[@]}"; do
        if docker ps -a --format '{{.Names}}' | grep -q "^${name}$"; then
            docker start "$name" > /dev/null
            sleep 2
            print_success "Started $name"
        fi
    done
}

# Remove all containers
cleanup_all() {
    print_info "Removing all test containers..."
    for name in "${!CONTAINERS[@]}"; do
        if docker ps -a --format '{{.Names}}' | grep -q "^${name}$"; then
            docker stop "$name" > /dev/null 2>&1 || true
            docker rm "$name" > /dev/null 2>&1 || true
            print_success "Removed $name"
        fi
    done
}

# Show status
show_status() {
    echo -e "${BLUE}=== Multi-Container Test Environment Status ===${NC}"
    echo ""
    
    for name in "${!CONTAINERS[@]}"; do
        IFS=':' read -r ssh_port http_port https_port app_port <<< "${CONTAINERS[$name]}"
        
        if docker ps --format '{{.Names}}' | grep -q "^${name}$"; then
            print_success "$name is running"
            echo "  SSH:  ssh -i $SSH_PRIVATE_KEY -p $ssh_port root@localhost"
            echo "  HTTP: http://localhost:$http_port (after deploy)"
            
            # Test SSH
            if ssh -i "$SSH_PRIVATE_KEY" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=2 -p "$ssh_port" root@localhost "echo 'OK'" > /dev/null 2>&1; then
                echo "  SSH Status: ✓ Connected"
            else
                echo "  SSH Status: ✗ Failed"
            fi
        else
            print_warning "$name is not running"
        fi
        echo ""
    done
}

# Display connection info
display_info() {
    echo -e "${BLUE}=== Container Connection Information ===${NC}"
    echo ""
    
    for name in "${!CONTAINERS[@]}"; do
        IFS=':' read -r ssh_port http_port https_port app_port <<< "${CONTAINERS[$name]}"
        echo -e "${YELLOW}$name:${NC}"
        echo "  SSH Port:   $ssh_port"
        echo "  HTTP Port:  $http_port"
        echo "  HTTPS Port: $https_port"
        echo "  App Port:   $app_port"
        echo ""
    done
    
    echo -e "${BLUE}Next Steps:${NC}"
    echo "  1. Run inventory manager:  make run"
    echo "  2. Create environment:     test-multi"
    echo "  3. Add servers:            test-web-01, test-web-02, test-web-03"
    echo "  4. Provision & Deploy in parallel"
    echo ""
}

# Main
main() {
    case "${1:-setup}" in
        setup)
            setup_all
            ;;
        stop)
            stop_all
            ;;
        start)
            start_all
            ;;
        cleanup)
            cleanup_all
            ;;
        status)
            show_status
            ;;
        help|*)
            echo "Multi-Container Test Environment Manager"
            echo ""
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  setup   - Create and start all containers (default)"
            echo "  stop    - Stop all containers"
            echo "  start   - Start all containers"
            echo "  cleanup - Remove all containers"
            echo "  status  - Show status of all containers"
            echo "  help    - Show this message"
            ;;
    esac
}

main "$@"
