#!/bin/bash

# Test Dry-Run Mode
# This script tests the dry-run functionality by monitoring logs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="$PROJECT_ROOT/logs/docker"

echo "🧪 Testing Dry-Run Mode"
echo "======================="
echo ""

# Check if container is running
if ! docker ps | grep -q boiler-test-vps; then
    echo "❌ Container not running. Run ./tests/test-docker-vps.sh setup first"
    exit 1
fi

echo "✓ Container is running"
echo ""

# Check if docker environment exists
if [ ! -f "$PROJECT_ROOT/inventory/docker/hosts.yml" ]; then
    echo "❌ Docker environment not found. Please create it in the inventory manager first"
    exit 1
fi

echo "✓ Docker environment found"
echo ""

# Create a simple Ansible test with check mode
echo "📋 Testing Ansible --check --diff flags..."
echo ""

cd "$PROJECT_ROOT"

# Test provision dry-run
echo "1️⃣  Testing provision dry-run..."
ansible-playbook \
    -i inventory/docker/hosts.yml \
    playbooks/provision.yml \
    --limit docker-web-01 \
    --check \
    --diff \
    2>&1 | tee /tmp/provision-check.log

if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo "✅ Provision dry-run completed successfully"
else
    echo "❌ Provision dry-run failed"
    exit 1
fi

echo ""
echo "📊 Dry-run Analysis:"
echo "-------------------"

# Count changes that would be made
CHANGES=$(grep -c "^changed:" /tmp/provision-check.log || echo "0")
OK=$(grep -c "^ok:" /tmp/provision-check.log || echo "0")
FAILED=$(grep -c "^failed:" /tmp/provision-check.log || echo "0")

echo "  Would change: $CHANGES tasks"
echo "  Already OK:   $OK tasks"
echo "  Would fail:   $FAILED tasks"
echo ""

if [ "$FAILED" -gt 0 ]; then
    echo "⚠️  Dry-run detected potential failures!"
    echo ""
    echo "Failed tasks:"
    grep "^failed:" /tmp/provision-check.log || true
    exit 1
fi

echo "✅ Dry-run validation passed!"
echo ""
echo "💡 This means:"
echo "   - All tasks can be executed safely"
echo "   - $CHANGES configurations will be modified"
echo "   - $OK configurations are already correct"
echo ""
echo "🚀 Safe to proceed with actual provisioning"
