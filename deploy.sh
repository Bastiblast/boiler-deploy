#!/bin/bash
# Quick deployment script

set -e

ENVIRONMENT=${1:-dev}
ACTION=${2:-deploy}

if [ "$ENVIRONMENT" != "dev" ] && [ "$ENVIRONMENT" != "production" ] && [ "$ENVIRONMENT" != "hostinger" ]; then
    echo "Usage: $0 [dev|production|hostinger] [provision|deploy|update|rollback]"
    exit 1
fi

INVENTORY="inventory/$ENVIRONMENT"

case $ACTION in
    provision)
        echo "🚀 Provisioning $ENVIRONMENT environment..."
        ansible-playbook playbooks/provision.yml -i $INVENTORY
        ;;
    deploy)
        echo "🚢 Deploying to $ENVIRONMENT..."
        ansible-playbook playbooks/deploy.yml -i $INVENTORY
        ;;
    update)
        echo "🔄 Updating $ENVIRONMENT..."
        ansible-playbook playbooks/update.yml -i $INVENTORY
        ;;
    rollback)
        echo "⏪ Rolling back $ENVIRONMENT..."
        ansible-playbook playbooks/rollback.yml -i $INVENTORY
        ;;
    *)
        echo "Unknown action: $ACTION"
        echo "Usage: $0 [dev|production] [provision|deploy|update|rollback]"
        exit 1
        ;;
esac

echo "✅ Done!"
