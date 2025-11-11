#!/bin/bash
# Test if ScriptExecutor can run ansible commands

ENV="docker"
ACTION="check"

echo "Testing script execution..."
echo "Environment: $ENV"
echo "Action: $ACTION"
echo ""

# Run ansible check directly (skip deploy.sh interactive prompts)
ansible-playbook playbooks/deploy.yml -i "inventory/$ENV" --check 2>&1

echo ""
echo "Exit code: $?"
