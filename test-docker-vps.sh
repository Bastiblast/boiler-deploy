#!/bin/bash

# Docker VPS Simulator for Testing
# This script creates a Docker container that simulates a VPS for testing provisioning and deployment

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_NAME="boiler-test-vps"
IMAGE_NAME="boiler-test-ubuntu"
SSH_PORT=2222
HTTP_PORT=8080
HTTPS_PORT=8443

# SSH key paths
SSH_KEY_DIR="$HOME/.ssh"
SSH_PRIVATE_KEY="$SSH_KEY_DIR/boiler_test_rsa"
SSH_PUBLIC_KEY="$SSH_KEY_DIR/boiler_test_rsa.pub"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to generate SSH keys if they don't exist
generate_ssh_keys() {
    if [[ ! -f "$SSH_PRIVATE_KEY" ]]; then
        print_info "Generating SSH key pair for testing..."
        ssh-keygen -t rsa -b 4096 -f "$SSH_PRIVATE_KEY" -N "" -C "boiler-test-key"
        print_success "SSH keys generated at $SSH_PRIVATE_KEY"
    else
        print_info "Using existing SSH keys at $SSH_PRIVATE_KEY"
    fi
}

# Function to stop and remove existing container
cleanup_container() {
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_info "Removing existing container..."
        docker stop "$CONTAINER_NAME" > /dev/null 2>&1 || true
        docker rm "$CONTAINER_NAME" > /dev/null 2>&1 || true
        print_success "Container removed"
    fi
}

# Function to build Docker image
build_image() {
    print_info "Building Docker image..."
    
    # Create temporary directory for Docker build context
    BUILD_DIR=$(mktemp -d)
    trap "rm -rf $BUILD_DIR" EXIT
    
    # Create Dockerfile - Systemd-enabled VPS simulator with SSH
    # Using jrei/systemd-ubuntu for proper systemd support
    # Ansible provision will install: Node.js, Nginx, UFW, Fail2ban, PostgreSQL, etc.
    cat > "$BUILD_DIR/Dockerfile" << 'EOF'
FROM jrei/systemd-ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install ONLY minimal packages for SSH and Ansible to work
# Everything else (Node.js, Nginx, UFW, Fail2ban, etc.) will be installed by Ansible provision
RUN apt-get update && apt-get install -y \
    openssh-server \
    sudo \
    python3 \
    python3-apt \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Configure SSH
RUN mkdir -p /var/run/sshd && \
    sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Create root .ssh directory
RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh

# Enable SSH service
RUN systemctl enable ssh

# Expose SSH port
EXPOSE 22 80 443

# Start systemd (this will start SSH and other services)
CMD ["/lib/systemd/systemd"]
EOF

    # Build the image
    docker build -t "$IMAGE_NAME" "$BUILD_DIR"
    print_success "Docker image built: $IMAGE_NAME"
}

# Function to start container
start_container() {
    print_info "Starting container..."
    
    # Run systemd-enabled container with proper privileges
    # Systemd needs: --privileged, tmpfs mounts, and specific stop signal
    docker run -d \
        --name "$CONTAINER_NAME" \
        --privileged \
        --tmpfs /tmp \
        --tmpfs /run \
        --tmpfs /run/lock \
        -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
        --cgroupns=host \
        --stop-signal SIGRTMIN+3 \
        -p "${SSH_PORT}:22" \
        -p "${HTTP_PORT}:80" \
        -p "${HTTPS_PORT}:443" \
        "$IMAGE_NAME"
    
    # Wait for systemd and SSH to be ready
    print_info "Waiting for systemd and SSH to initialize..."
    sleep 5
    
    print_success "Container started: $CONTAINER_NAME"
}

# Function to configure SSH in container
configure_ssh() {
    print_info "Configuring SSH access..."
    
    # Copy SSH public key to container
    docker exec "$CONTAINER_NAME" mkdir -p /root/.ssh
    docker exec "$CONTAINER_NAME" chmod 700 /root/.ssh
    docker cp "$SSH_PUBLIC_KEY" "$CONTAINER_NAME:/root/.ssh/authorized_keys"
    docker exec "$CONTAINER_NAME" chmod 600 /root/.ssh/authorized_keys
    docker exec "$CONTAINER_NAME" chown root:root /root/.ssh/authorized_keys
    
    print_success "SSH configured with key authentication"
}

# Function to test SSH connection
test_ssh() {
    print_info "Testing SSH connection..."
    
    # Wait a bit for SSH to be fully ready
    sleep 2
    
    if ssh -i "$SSH_PRIVATE_KEY" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p "$SSH_PORT" root@localhost "echo 'SSH connection successful'" > /dev/null 2>&1; then
        print_success "SSH connection test passed"
        return 0
    else
        print_error "SSH connection test failed"
        return 1
    fi
}

# Function to create test inventory
create_test_inventory() {
    print_info "Creating test environment in inventory manager..."
    
    # Check if inventory manager is built
    if [[ ! -f "bin/inventory-manager" ]]; then
        print_warning "Inventory manager not built. Building now..."
        make build
    fi
    
    # Create inventory directory structure
    INVENTORY_DIR="inventory/test-docker"
    mkdir -p "$INVENTORY_DIR"
    
    # Create hosts file
    cat > "$INVENTORY_DIR/hosts" << EOF
[webservers]
test-web-01 ansible_host=localhost ansible_port=${SSH_PORT} ansible_user=root

[dbservers]

[monitoring]

[all:vars]
ansible_ssh_private_key_file=${SSH_PRIVATE_KEY}
ansible_ssh_common_args='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'
EOF

    # Create environment-specific variables
    mkdir -p "group_vars"
    cat > "group_vars/test-docker.yml" << EOF
---
# Test Docker Environment Variables

# Override app port for testing
app_port: 3000

# Disable SSL for testing
ssl_enabled: false

# Test repository (you can change this)
app_repo: "https://github.com/Bastiblast/ansible-next-test.git"
app_branch: "main"

# Test app name
app_name: testapp
EOF

    print_success "Test inventory created at $INVENTORY_DIR"
}

# Function to display connection info
display_info() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         Docker VPS Test Environment Ready!                   ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Container Information:${NC}"
    echo -e "  Name:           $CONTAINER_NAME"
    echo -e "  SSH Port:       $SSH_PORT"
    echo -e "  HTTP Port:      $HTTP_PORT"
    echo -e "  HTTPS Port:     $HTTPS_PORT"
    echo ""
    echo -e "${BLUE}SSH Connection:${NC}"
    echo -e "  Command:        ${YELLOW}ssh -i $SSH_PRIVATE_KEY -p $SSH_PORT root@localhost${NC}"
    echo -e "  Private Key:    $SSH_PRIVATE_KEY"
    echo -e "  Public Key:     $SSH_PUBLIC_KEY"
    echo ""
    echo -e "${BLUE}Inventory Manager Configuration:${NC}"
    echo -e "  Environment:    test-docker"
    echo -e "  Server:         test-web-01"
    echo -e "  IP:             localhost (or 127.0.0.1)"
    echo -e "  Port:           22"
    echo -e "  SSH Port:       $SSH_PORT"
    echo -e "  SSH Key:        $SSH_PRIVATE_KEY"
    echo -e "  Type:           web"
    echo ""
    echo -e "${BLUE}Next Steps:${NC}"
    echo -e "  1. Run inventory manager:     ${YELLOW}make run${NC}"
    echo -e "  2. Create environment:        ${YELLOW}test-docker${NC}"
    echo -e "  3. Add server:                ${YELLOW}test-web-01${NC}"
    echo -e "     - IP: 127.0.0.1"
    echo -e "     - Port: $SSH_PORT"
    echo -e "     - SSH Key: $SSH_PRIVATE_KEY"
    echo -e "     - Repo: https://github.com/Bastiblast/ansible-next-test.git"
    echo -e "     - App Port: 3000"
    echo -e "  4. Validate inventory"
    echo -e "  5. Provision server"
    echo -e "  6. Deploy application"
    echo ""
    echo -e "${BLUE}Testing:${NC}"
    echo -e "  SSH Test:       ${YELLOW}ssh -i $SSH_PRIVATE_KEY -p $SSH_PORT root@localhost${NC}"
    echo -e "  View Logs:      ${YELLOW}docker logs $CONTAINER_NAME${NC}"
    echo -e "  Enter Container:${YELLOW}docker exec -it $CONTAINER_NAME bash${NC}"
    echo -e "  Stop Container: ${YELLOW}docker stop $CONTAINER_NAME${NC}"
    echo -e "  Start Container:${YELLOW}docker start $CONTAINER_NAME${NC}"
    echo -e "  Remove All:     ${YELLOW}./test-docker-vps.sh cleanup${NC}"
    echo ""
    echo -e "${BLUE}Access Deployed App:${NC}"
    echo -e "  After deployment: ${YELLOW}http://localhost:$HTTP_PORT${NC}"
    echo ""
}

# Function to cleanup everything
cleanup_all() {
    print_info "Cleaning up test environment..."
    
    # Stop and remove container
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        docker stop "$CONTAINER_NAME" > /dev/null 2>&1 || true
        docker rm "$CONTAINER_NAME" > /dev/null 2>&1 || true
        print_success "Container removed"
    fi
    
    # Remove image
    if docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
        docker rmi "$IMAGE_NAME" > /dev/null 2>&1 || true
        print_success "Image removed"
    fi
    
    # Optionally remove SSH keys (commented out for safety)
    # rm -f "$SSH_PRIVATE_KEY" "$SSH_PUBLIC_KEY"
    
    print_success "Cleanup complete"
}

# Function to show status
show_status() {
    echo -e "${BLUE}Docker VPS Test Environment Status:${NC}"
    echo ""
    
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        print_success "Container is running"
        echo ""
        echo "Container Details:"
        docker ps --filter "name=$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""
        
        # Test SSH
        if ssh -i "$SSH_PRIVATE_KEY" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p "$SSH_PORT" root@localhost "echo 'OK'" > /dev/null 2>&1; then
            print_success "SSH is accessible"
        else
            print_error "SSH is not accessible"
        fi
    else
        print_warning "Container is not running"
    fi
    
    if docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
        print_success "Image exists"
    else
        print_warning "Image does not exist"
    fi
    
    if [[ -f "$SSH_PRIVATE_KEY" ]]; then
        print_success "SSH keys exist"
    else
        print_warning "SSH keys do not exist"
    fi
}

# Main function
main() {
    case "${1:-setup}" in
        setup)
            echo -e "${GREEN}Setting up Docker VPS test environment...${NC}"
            echo ""
            
            check_docker
            generate_ssh_keys
            cleanup_container
            build_image
            start_container
            configure_ssh
            test_ssh
            create_test_inventory
            display_info
            ;;
        
        cleanup)
            cleanup_all
            ;;
        
        status)
            show_status
            ;;
        
        restart)
            print_info "Restarting container..."
            docker restart "$CONTAINER_NAME"
            sleep 3
            print_success "Container restarted"
            ;;
        
        logs)
            docker logs -f "$CONTAINER_NAME"
            ;;
        
        ssh)
            ssh -i "$SSH_PRIVATE_KEY" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -p "$SSH_PORT" root@localhost
            ;;
        
        exec)
            docker exec -it "$CONTAINER_NAME" bash
            ;;
        
        help|*)
            echo "Docker VPS Test Environment Manager"
            echo ""
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  setup      - Setup the complete test environment (default)"
            echo "  cleanup    - Remove container and image"
            echo "  status     - Show current status"
            echo "  restart    - Restart the container"
            echo "  logs       - Show container logs (follow mode)"
            echo "  ssh        - SSH into the container"
            echo "  exec       - Execute bash in the container"
            echo "  help       - Show this help message"
            echo ""
            ;;
    esac
}

# Run main function
main "$@"
