#!/bin/bash

echo "=== Manual Check Test ==="
echo ""

# Ensure docker container is running
echo "1. Check if docker test container is running..."
docker ps | grep boiler-test-vps || {
    echo "Container not running, starting it..."
    cd /home/basthook/devIronMenth/boiler-deploy
    ./test-docker-vps.sh start
}

echo ""
echo "2. Test curl to container port 3000..."
curl -sf -m 5 http://127.0.0.1:3000/ | head -20 || echo "✗ Curl failed"

echo ""
echo "3. Testing validation..."
cd /home/basthook/devIronMenth/boiler-deploy

# Run app in background briefly to test
timeout 10 ./bin/inventory-manager <<EOF &
{down}{down}{enter}
v
q
EOF

wait
echo ""
echo "4. Checking status after validation..."
cat inventory/docker/.status/servers.json

echo ""
echo "=== Test Complete ==="
