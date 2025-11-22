#!/bin/bash
################################################################################
# Setup 3 Docker Containers for Testing
# Exact copy of test-docker-vps.sh logic for 3 containers
################################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CONTAINER_PREFIX="test-web"
IMAGE_NAME="boiler-test-ubuntu"
SSH_KEY_DIR="$HOME/.ssh"
SSH_PRIVATE_KEY="$SSH_KEY_DIR/boiler_test_rsa"
SSH_PUBLIC_KEY="$SSH_KEY_DIR/boiler_test_rsa.pub"
BASE_SSH_PORT=2222
BASE_HTTP_PORT=8080
BASE_HTTPS_PORT=8443
BASE_APP_PORT=3000

################################################################################
# Helper Functions
################################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

separator() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
}

################################################################################
# Cleanup Function
################################################################################

cleanup_containers() {
    log_info "Cleaning up existing containers..."
    
    for i in 1 2 3; do
        CONTAINER_NAME="${CONTAINER_PREFIX}-0${i}"
        if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            log_info "Removing container: ${CONTAINER_NAME}"
            docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
        fi
    done
    
    log_success "Cleanup completed"
}

################################################################################
# SSH Key Setup
################################################################################

setup_ssh_keys() {
    log_info "Setting up SSH keys..."
    
    if [ ! -f "$SSH_PRIVATE_KEY" ]; then
        log_info "Generating SSH key pair..."
        ssh-keygen -t rsa -b 4096 -f "$SSH_PRIVATE_KEY" -N "" -C "boiler-test-key"
        log_success "SSH keys generated at $SSH_PRIVATE_KEY"
    else
        log_info "Using existing SSH keys at $SSH_PRIVATE_KEY"
    fi
}

################################################################################
# Build Docker Image
################################################################################

build_image() {
    # Check if image already exists
    if docker images --format '{{.Repository}}' | grep -q "^${IMAGE_NAME}$"; then
        log_info "Image $IMAGE_NAME already exists"
        return 0
    fi
    
    log_info "Building Docker image..."
    
    # Create temporary directory for build
    BUILD_DIR=$(mktemp -d)
    trap "rm -rf $BUILD_DIR" EXIT
    
    # Create Dockerfile (exact copy from test-docker-vps.sh)
    cat > "$BUILD_DIR/Dockerfile" << 'EOF'
FROM jrei/systemd-ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install ONLY minimal packages for SSH and Ansible to work
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

# Expose ports
EXPOSE 22 80 443

# Start systemd
CMD ["/lib/systemd/systemd"]
EOF

    docker build -t "$IMAGE_NAME" "$BUILD_DIR" > /dev/null 2>&1
    log_success "Docker image built: $IMAGE_NAME"
}

################################################################################
# Container Creation
################################################################################

create_container() {
    local index=$1
    local container_name="${CONTAINER_PREFIX}-0${index}"
    local ssh_port=$((BASE_SSH_PORT + index - 1))
    local http_port=$((BASE_HTTP_PORT + index - 1))
    local https_port=$((BASE_HTTPS_PORT + index - 1))
    
    log_info "Creating container: ${container_name}"
    log_info "  Ports: SSH=${ssh_port}, HTTP=${http_port}, HTTPS=${https_port}"
    
    # Run systemd-enabled container (exact copy from test-docker-vps.sh)
    docker run -d \
        --name "$container_name" \
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
        "$IMAGE_NAME" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        log_success "Container ${container_name} created successfully"
        return 0
    else
        log_error "Failed to create container ${container_name}"
        return 1
    fi
}

################################################################################
# SSH Configuration
################################################################################

configure_ssh_access() {
    local container_name=$1
    
    log_info "Configuring SSH access for ${container_name}..."
    
    # Copy SSH public key using docker cp (most reliable method)
    docker exec "$container_name" mkdir -p /root/.ssh
    docker exec "$container_name" chmod 700 /root/.ssh
    docker cp "$SSH_PUBLIC_KEY" "$container_name:/root/.ssh/authorized_keys"
    docker exec "$container_name" chmod 600 /root/.ssh/authorized_keys
    docker exec "$container_name" chown root:root /root/.ssh/authorized_keys
    
    log_success "SSH key installed for ${container_name}"
}

################################################################################
# Validation
################################################################################

validate_ssh_connection() {
    local container_name=$1
    local ssh_port=$2
    
    log_info "Testing SSH connection to ${container_name}:${ssh_port}..."
    
    # Wait a bit for SSH to be fully ready
    sleep 2
    
    # Test SSH connection
    if ssh -o StrictHostKeyChecking=no \
           -o ConnectTimeout=10 \
           -i "$SSH_PRIVATE_KEY" \
           -p "$ssh_port" \
           root@localhost \
           "echo 'SSH test successful'" >/dev/null 2>&1; then
        log_success "SSH connection verified for ${container_name}"
        return 0
    else
        log_error "SSH connection failed for ${container_name}"
        return 1
    fi
}

################################################################################
# Inventory Generation
################################################################################

generate_inventory() {
    log_info "Generating Ansible inventory..."
    
    local inventory_dir="inventory/docker"
    local hosts_file="${inventory_dir}/hosts.yml"
    
    # Backup existing inventory
    if [ -f "${hosts_file}" ]; then
        local backup_file="${hosts_file}.backup.$(date +%Y%m%d_%H%M%S)"
        log_info "Backing up existing inventory to: ${backup_file}"
        cp "${hosts_file}" "${backup_file}"
    fi
    
    # Ensure directory exists
    mkdir -p "${inventory_dir}"
    
    # Generate hosts.yml
    cat > "${hosts_file}" << EOF
# Auto-generated by setup-3-servers.sh
# Date: $(date)

all:
  children:
    webservers:
      hosts:
EOF
    
    for i in 1 2 3; do
        local container_name="${CONTAINER_PREFIX}-0${i}"
        local ssh_port=$((BASE_SSH_PORT + i - 1))
        local http_port=$((BASE_HTTP_PORT + i - 1))
        local app_port=$((BASE_APP_PORT + i - 1))
        
        cat >> "${hosts_file}" << EOF
        ${container_name}:
          ansible_host: 127.0.0.1
          ansible_user: root
          ansible_port: ${ssh_port}
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ${SSH_PRIVATE_KEY}
          ansible_become: true
          app_port: ${app_port}
          http_port: ${http_port}  # External port (nginx proxy)
EOF
    done
    
    log_success "Inventory generated: ${hosts_file}"
    
    # Display inventory for verification
    separator
    log_info "Generated Inventory:"
    cat "${hosts_file}"
    separator
}

################################################################################
# Summary Display
################################################################################

display_summary() {
    separator
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         3 Docker Servers Setup - COMPLETE                   ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
    separator
    
    log_info "Container Details:"
    echo ""
    
    for i in 1 2 3; do
        local container_name="${CONTAINER_PREFIX}-0${i}"
        local ssh_port=$((BASE_SSH_PORT + i - 1))
        local http_port=$((BASE_HTTP_PORT + i - 1))
        local https_port=$((BASE_HTTPS_PORT + i - 1))
        local app_port=$((BASE_APP_PORT + i - 1))
        
        echo -e "  ${BLUE}${container_name}:${NC}"
        echo "    SSH:   ssh -i ${SSH_PRIVATE_KEY} -p ${ssh_port} root@localhost"
        echo "    HTTP:  http://localhost:${http_port}"
        echo "    HTTPS: https://localhost:${https_port}"
        echo "    APP:   http://localhost:${app_port}"
        echo ""
    done
    
    separator
    log_info "Quick Access Commands:"
    echo ""
    echo "  # SSH to containers:"
    echo "  ssh -i ${SSH_PRIVATE_KEY} -p 2222 root@localhost  # test-web-01"
    echo "  ssh -i ${SSH_PRIVATE_KEY} -p 2223 root@localhost  # test-web-02"
    echo "  ssh -i ${SSH_PRIVATE_KEY} -p 2224 root@localhost  # test-web-03"
    echo ""
    echo "  # Docker exec (direct):"
    echo "  docker exec -it test-web-01 bash"
    echo "  docker exec -it test-web-02 bash"
    echo "  docker exec -it test-web-03 bash"
    echo ""
    
    separator
    log_info "Next Steps:"
    echo ""
    echo "  1. Test SSH connections:"
    echo "     ssh -i ${SSH_PRIVATE_KEY} -p 2222 root@localhost"
    echo ""
    echo "  2. Run inventory manager:"
    echo "     ./bin/inventory-manager"
    echo ""
    echo "  3. Or provision directly with Ansible:"
    echo "     ansible-playbook -i inventory/docker/hosts.yml playbooks/provision.yml"
    echo ""
    
    separator
    log_success "Setup complete! All containers ready for testing."
    separator
}

################################################################################
# Main Execution
################################################################################

main() {
    separator
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║    Docker Test Environment Setup (3 Servers)                ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    separator
    
    # Step 1: Cleanup
    cleanup_containers
    separator
    
    # Step 2: Setup SSH keys
    setup_ssh_keys
    separator
    
    # Step 3: Build Docker image
    build_image
    separator
    
    # Step 4: Create containers
    log_info "Creating 3 containers..."
    echo ""
    
    for i in 1 2 3; do
        create_container $i
    done
    
    separator
    log_info "Waiting for systemd and SSH to initialize (10 seconds)..."
    sleep 10
    separator
    
    # Step 5: Configure SSH access
    log_info "Configuring SSH access for all containers..."
    echo ""
    
    for i in 1 2 3; do
        local container_name="${CONTAINER_PREFIX}-0${i}"
        configure_ssh_access "${container_name}"
    done
    
    separator
    
    # Step 6: Validate connections
    log_info "Validating SSH connections..."
    echo ""
    
    local all_success=true
    for i in 1 2 3; do
        local container_name="${CONTAINER_PREFIX}-0${i}"
        local ssh_port=$((BASE_SSH_PORT + i - 1))
        if ! validate_ssh_connection "${container_name}" "${ssh_port}"; then
            all_success=false
        fi
    done
    
    separator
    
    # Step 7: Generate inventory
    generate_inventory
    
    # Step 8: Display summary
    display_summary
    
    # Final status
    if [ "$all_success" = true ]; then
        exit 0
    else
        log_warning "Some SSH connections failed. Check logs above."
        exit 1
    fi
}

# Run main function
main "$@"
