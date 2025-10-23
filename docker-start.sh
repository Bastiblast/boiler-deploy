#!/bin/bash
# Start Docker testing environment

set -e

echo "🐳 Starting Docker Testing Environment for Ansible"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running"
    echo "   Please start Docker and try again"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Error: docker-compose is not installed"
    echo "   Install with: sudo apt install docker-compose"
    exit 1
fi

# Build and start containers
echo "🏗️  Building Docker images..."
docker-compose build

echo ""
echo "🚀 Starting containers..."
docker-compose up -d

echo ""
echo "⏳ Waiting for containers to be ready..."
sleep 5

# Check container status
echo ""
echo "📊 Container Status:"
docker-compose ps

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Docker environment is ready!"
echo ""
echo "📋 Available containers:"
echo "   • ansible-web-01  (172.28.0.11) - SSH: localhost:2201"
echo "   • ansible-web-02  (172.28.0.12) - SSH: localhost:2202"
echo "   • ansible-db-01   (172.28.0.21) - SSH: localhost:2203"
echo ""
echo "🔑 Next steps:"
echo "   1. Setup SSH keys: ./docker/setup-ssh-keys.sh"
echo "   2. Test connection: ansible all -i inventory/docker -m ping"
echo "   3. Provision servers: ./deploy.sh docker provision"
echo ""
echo "📚 Access containers:"
echo "   docker exec -it ansible-web-01 bash"
echo "   docker exec -it ansible-db-01 bash"
echo ""
echo "🛑 Stop environment: ./docker-stop.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
