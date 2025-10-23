#!/bin/bash
# Health check script for monitoring

set -e

INVENTORY=${1:-inventory/production}
OUTPUT_FILE="health_check_$(date +%Y%m%d_%H%M%S).log"

echo "==================================" | tee $OUTPUT_FILE
echo "Health Check Report" | tee -a $OUTPUT_FILE
echo "$(date)" | tee -a $OUTPUT_FILE
echo "==================================" | tee -a $OUTPUT_FILE
echo "" | tee -a $OUTPUT_FILE

# Ping all servers
echo "📡 Checking connectivity..." | tee -a $OUTPUT_FILE
ansible all -i $INVENTORY -m ping | tee -a $OUTPUT_FILE
echo "" | tee -a $OUTPUT_FILE

# Check system resources
echo "💻 Checking system resources..." | tee -a $OUTPUT_FILE
ansible all -i $INVENTORY -a "df -h /" --become | tee -a $OUTPUT_FILE
ansible all -i $INVENTORY -a "free -h" --become | tee -a $OUTPUT_FILE
echo "" | tee -a $OUTPUT_FILE

# Check services
echo "🔧 Checking services..." | tee -a $OUTPUT_FILE

echo "  Web Servers:" | tee -a $OUTPUT_FILE
ansible webservers -i $INVENTORY -a "systemctl is-active nginx" --become | tee -a $OUTPUT_FILE
ansible webservers -i $INVENTORY -a "pm2 list" --become-user deploy | tee -a $OUTPUT_FILE

echo "  Database Servers:" | tee -a $OUTPUT_FILE
ansible dbservers -i $INVENTORY -a "systemctl is-active postgresql" --become | tee -a $OUTPUT_FILE

echo "  Monitoring:" | tee -a $OUTPUT_FILE
ansible monitoring -i $INVENTORY -a "systemctl is-active prometheus" --become 2>/dev/null | tee -a $OUTPUT_FILE
ansible monitoring -i $INVENTORY -a "systemctl is-active grafana-server" --become 2>/dev/null | tee -a $OUTPUT_FILE

echo "" | tee -a $OUTPUT_FILE
echo "✅ Health check complete! Results saved to $OUTPUT_FILE" | tee -a $OUTPUT_FILE
