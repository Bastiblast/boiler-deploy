#!/bin/bash

##############################################################################
# Semaphore Project Import Script
# Automatically configures Semaphore via API
##############################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SEMAPHORE_URL="http://localhost:3000/api"
PROJECT_NAME="boiler-deploy"

echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Semaphore Project Import - Automated Setup          ║${NC}"
echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo

# Step 1: Get credentials
echo -e "${YELLOW}Step 1/7: Authentication${NC}"
echo
echo "First, let's check if you need to set up your admin password in Semaphore UI."
echo "Have you already logged into Semaphore at http://localhost:3000 ?"
echo -n "  [Y/n]: "
read -r ALREADY_SETUP
ALREADY_SETUP=${ALREADY_SETUP:-y}

if [[ ! "$ALREADY_SETUP" =~ ^[Yy] ]]; then
    echo
    echo -e "${YELLOW}⚠ Please complete the initial setup first:${NC}"
    echo "  1. Open http://localhost:3000 in your browser"
    echo "  2. Complete the setup wizard"
    echo "  3. Remember your admin password"
    echo "  4. Then run this script again"
    echo
    exit 0
fi

echo
echo -n "Enter Semaphore username [admin]: "
read -r SEMAPHORE_USER
SEMAPHORE_USER=${SEMAPHORE_USER:-admin}

echo -n "Enter Semaphore password: "
read -rs SEMAPHORE_PASS
echo

if [ -z "$SEMAPHORE_PASS" ]; then
    echo -e "${RED}✗ Password cannot be empty${NC}"
    exit 1
fi

# Login and get token
echo "Authenticating..."
LOGIN_RESPONSE=$(curl -s -X POST "${SEMAPHORE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"auth\":\"${SEMAPHORE_USER}\",\"password\":\"${SEMAPHORE_PASS}\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Authentication failed!${NC}"
    echo
    echo "Possible reasons:"
    echo "  1. Wrong password"
    echo "  2. Semaphore not fully initialized"
    echo
    echo "To reset admin password, run:"
    echo "  docker exec -it semaphore-ui semaphore user change-password --admin admin"
    echo
    exit 1
fi

echo -e "${GREEN}✓ Authenticated successfully${NC}"
echo

# Step 2: Create Project
echo -e "${YELLOW}Step 2/7: Creating Project${NC}"
PROJECT_RESPONSE=$(curl -s -X POST "${SEMAPHORE_URL}/projects" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${PROJECT_NAME}\",\"alert\":false}")

PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$PROJECT_ID" ]; then
    echo -e "${RED}✗ Project creation failed!${NC}"
    echo "Response: $PROJECT_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Project created (ID: ${PROJECT_ID})${NC}"
echo

# Step 3: Get SSH Key
echo -e "${YELLOW}Step 3/7: SSH Key Configuration${NC}"
SSH_KEY_PATH="/home/basthook/.ssh/Hosting"

if [ ! -f "$SSH_KEY_PATH" ]; then
    echo -e "${RED}✗ SSH key not found at ${SSH_KEY_PATH}${NC}"
    echo -n "Enter path to your SSH private key: "
    read -r SSH_KEY_PATH
fi

if [ ! -f "$SSH_KEY_PATH" ]; then
    echo -e "${RED}✗ SSH key still not found. Exiting.${NC}"
    exit 1
fi

SSH_KEY_CONTENT=$(cat "$SSH_KEY_PATH")
echo -e "${GREEN}✓ SSH key loaded from ${SSH_KEY_PATH}${NC}"

# Create SSH Key in Semaphore
KEY_RESPONSE=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/keys" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"deploy_key\",\"type\":\"ssh\",\"project_id\":${PROJECT_ID},\"login_password\":{\"login\":\"root\"},\"ssh\":{\"private_key\":\"${SSH_KEY_CONTENT}\"}}")

KEY_ID=$(echo "$KEY_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$KEY_ID" ]; then
    echo -e "${RED}✗ SSH key creation failed!${NC}"
    echo "Response: $KEY_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ SSH key created (ID: ${KEY_ID})${NC}"
echo

# Step 4: Create Repository
echo -e "${YELLOW}Step 4/7: Creating Repository${NC}"
REPO_RESPONSE=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/repositories" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"local-playbooks\",\"git_url\":\"/ansible\",\"git_branch\":\"streamlit\",\"project_id\":${PROJECT_ID}}")

REPO_ID=$(echo "$REPO_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$REPO_ID" ]; then
    echo -e "${RED}✗ Repository creation failed!${NC}"
    echo "Response: $REPO_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Repository created (ID: ${REPO_ID})${NC}"
echo

# Step 5: Get Server IPs
echo -e "${YELLOW}Step 5/7: Server Configuration${NC}"
echo -n "Enter your production server IP: "
read -r PROD_IP

if [ -z "$PROD_IP" ]; then
    echo -e "${YELLOW}⚠ No IP provided, using placeholder${NC}"
    PROD_IP="YOUR_SERVER_IP"
fi

# Create Inventory
INVENTORY_CONTENT="all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: ${PROD_IP}
          ansible_user: root
          ansible_port: 22
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: ${SSH_KEY_PATH}
          ansible_become: yes
          app_port: 3002"

INVENTORY_RESPONSE=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/inventory" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"production\",\"project_id\":${PROJECT_ID},\"inventory\":\"${INVENTORY_CONTENT}\",\"type\":\"static\"}")

INVENTORY_ID=$(echo "$INVENTORY_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$INVENTORY_ID" ]; then
    echo -e "${RED}✗ Inventory creation failed!${NC}"
    echo "Response: $INVENTORY_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Inventory created (ID: ${INVENTORY_ID})${NC}"
echo

# Step 6: Create Environment
echo -e "${YELLOW}Step 6/7: Creating Environment${NC}"
ENV_CONTENT='{
  "app_name": "myapp",
  "app_repo": "https://github.com/Bastiblast/ansible-next-test.git",
  "app_branch": "main",
  "nodejs_version": "20",
  "app_port": "3002",
  "deploy_user": "root",
  "timezone": "Europe/Paris",
  "pm2_instances": "2",
  "pm2_max_memory": "512M"
}'

ENV_RESPONSE=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/environment" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"production-vars\",\"project_id\":${PROJECT_ID},\"json\":\"${ENV_CONTENT}\"}")

ENV_ID=$(echo "$ENV_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$ENV_ID" ]; then
    echo -e "${RED}✗ Environment creation failed!${NC}"
    echo "Response: $ENV_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Environment created (ID: ${ENV_ID})${NC}"
echo

# Step 7: Create Task Templates
echo -e "${YELLOW}Step 7/7: Creating Task Templates${NC}"

# Template 1: Provision
TEMPLATE_1=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/templates" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"01 - Provision Server\",\"playbook\":\"playbooks/provision.yml\",\"project_id\":${PROJECT_ID},\"inventory_id\":${INVENTORY_ID},\"environment_id\":${ENV_ID},\"repository_id\":${REPO_ID},\"ssh_key_id\":${KEY_ID}}")

echo -e "${GREEN}✓ Template 'Provision' created${NC}"

# Template 2: Deploy
TEMPLATE_2=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/templates" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"02 - Deploy Application\",\"playbook\":\"playbooks/deploy.yml\",\"project_id\":${PROJECT_ID},\"inventory_id\":${INVENTORY_ID},\"environment_id\":${ENV_ID},\"repository_id\":${REPO_ID},\"ssh_key_id\":${KEY_ID}}")

echo -e "${GREEN}✓ Template 'Deploy' created${NC}"

# Template 3: Update
TEMPLATE_3=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/templates" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"03 - Update Application\",\"playbook\":\"playbooks/update.yml\",\"project_id\":${PROJECT_ID},\"inventory_id\":${INVENTORY_ID},\"environment_id\":${ENV_ID},\"repository_id\":${REPO_ID},\"ssh_key_id\":${KEY_ID}}")

echo -e "${GREEN}✓ Template 'Update' created${NC}"

# Template 4: Rollback
TEMPLATE_4=$(curl -s -X POST "${SEMAPHORE_URL}/project/${PROJECT_ID}/templates" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"04 - Rollback\",\"playbook\":\"playbooks/rollback.yml\",\"project_id\":${PROJECT_ID},\"inventory_id\":${INVENTORY_ID},\"environment_id\":${ENV_ID},\"repository_id\":${REPO_ID},\"ssh_key_id\":${KEY_ID}}")

echo -e "${GREEN}✓ Template 'Rollback' created${NC}"
echo

# Summary
echo
echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    Import Complete! 🎉                   ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════╝${NC}"
echo
echo -e "${GREEN}✓ Project:${NC} ${PROJECT_NAME} (ID: ${PROJECT_ID})"
echo -e "${GREEN}✓ SSH Key:${NC} deploy_key (ID: ${KEY_ID})"
echo -e "${GREEN}✓ Repository:${NC} local-playbooks (ID: ${REPO_ID})"
echo -e "${GREEN}✓ Inventory:${NC} production (ID: ${INVENTORY_ID})"
echo -e "${GREEN}✓ Environment:${NC} production-vars (ID: ${ENV_ID})"
echo -e "${GREEN}✓ Templates:${NC} 4 task templates created"
echo
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Open: http://localhost:3000"
echo "  2. Go to project: ${PROJECT_NAME}"
echo "  3. Run your first playbook!"
echo
echo -e "${YELLOW}Note:${NC} If server IP was not provided, edit the inventory in Semaphore UI"
echo
