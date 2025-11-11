#!/bin/bash

##############################################################################
# Reset Semaphore Admin Password
##############################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Reset Semaphore Admin Password                  ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════╝${NC}"
echo

# Check if Semaphore is running
if ! docker ps | grep -q semaphore-ui; then
    echo -e "${RED}✗ Semaphore is not running!${NC}"
    echo "Start it with: docker compose -f docker-compose.semaphore.yml up -d"
    exit 1
fi

echo "This will reset the admin password to 'admin'"
echo -n "Continue? [y/N]: "
read -r CONFIRM

if [[ ! "$CONFIRM" =~ ^[Yy] ]]; then
    echo "Cancelled."
    exit 0
fi

echo
echo "Resetting password..."

# Method 1: Using Semaphore CLI
docker exec -it semaphore-ui semaphore user change-password --admin admin 2>/dev/null << 'EOF'
admin
admin
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Password reset successfully!${NC}"
    echo
    echo "You can now login with:"
    echo "  Username: admin"
    echo "  Password: admin"
    echo
    echo "⚠️ Change this password in the UI after logging in!"
else
    echo -e "${YELLOW}CLI method failed, trying database method...${NC}"
    
    # Method 2: Direct database update
    # Generate bcrypt hash for "admin" (cost 10)
    HASH='$2a$10$N8OdxuOC/6Nt2Hj/6mw.dOJVSVPj7dK1i6cU5TmHpBfMlm5p3YNCC'
    
    docker exec semaphore-mysql mysql -u semaphore -psemaphore_pass semaphore -e "UPDATE user SET password='${HASH}' WHERE username='admin';" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Password reset via database!${NC}"
        echo
        echo "Password is now: admin"
        echo "Restart Semaphore:"
        echo "  docker restart semaphore-ui"
    else
        echo -e "${RED}✗ Failed to reset password${NC}"
        exit 1
    fi
fi
