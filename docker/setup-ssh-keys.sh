#!/bin/bash
# Setup SSH keys for Docker containers

set -e

echo "🔑 Setting up SSH keys for Docker containers..."

# Generate SSH key if it doesn't exist
if [ ! -f ~/.ssh/ansible_docker_rsa ]; then
    echo "📝 Generating new SSH key pair..."
    ssh-keygen -t rsa -b 4096 -f ~/.ssh/ansible_docker_rsa -N "" -C "ansible-docker-testing"
else
    echo "✓ SSH key already exists"
fi

# Copy public key to containers
echo ""
echo "📤 Copying SSH keys to containers..."

for container in ansible-web-01 ansible-web-02 ansible-db-01; do
    echo "  → $container"
    docker exec -i $container bash -c "mkdir -p /home/debian/.ssh && chmod 700 /home/debian/.ssh"
    docker exec -i $container bash -c "cat > /home/debian/.ssh/authorized_keys" < ~/.ssh/ansible_docker_rsa.pub
    docker exec -i $container bash -c "chown -R debian:debian /home/debian/.ssh && chmod 600 /home/debian/.ssh/authorized_keys"
done

echo ""
echo "✅ SSH keys configured!"
echo ""
echo "📝 Update your ansible.cfg with:"
echo "   private_key_file = ~/.ssh/ansible_docker_rsa"
echo ""
echo "🧪 Test connection:"
echo "   ansible all -i inventory/docker -m ping"
