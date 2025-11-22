#!/bin/bash
# Quick check if parallel mode is configured and working

echo "🔍 Checking Parallel Execution Configuration"
echo "=============================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check configs
echo "📋 Configuration Files:"
echo ""

for config in inventory/*/config.yml; do
    if [ -f "$config" ]; then
        env=$(basename $(dirname "$config"))
        workers=$(grep "max_parallel_workers:" "$config" 2>/dev/null | awk '{print $2}')
        
        if [ -z "$workers" ]; then
            echo -e "  ${RED}✗${NC} $env: max_parallel_workers NOT FOUND"
        elif [ "$workers" = "0" ]; then
            echo -e "  ${YELLOW}⚠${NC} $env: max_parallel_workers = 0 (SEQUENTIAL)"
        else
            echo -e "  ${GREEN}✓${NC} $env: max_parallel_workers = $workers (PARALLEL)"
        fi
    fi
done

echo ""
echo "📊 Recent Logs:"
echo ""

if [ -f "debug.log" ]; then
    last_mode=$(grep "Running in.*mode" debug.log 2>/dev/null | tail -1 || true)
    if [ -n "$last_mode" ]; then
        if echo "$last_mode" | grep -q "PARALLEL"; then
            echo -e "  ${GREEN}✓${NC} Mode: PARALLEL"
        else
            echo -e "  ${RED}✗${NC} Mode: SEQUENTIAL"
        fi
    fi
    
    last_workers=$(grep "Max workers set to" debug.log 2>/dev/null | tail -1 || true)
    if [ -n "$last_workers" ]; then
        workers_val=$(echo "$last_workers" | grep -oP 'set to \K\d+' || echo "unknown")
        if [ "$workers_val" = "0" ]; then
            echo -e "  ${RED}✗${NC} Workers: $workers_val (sequential)"
        elif [ "$workers_val" != "unknown" ]; then
            echo -e "  ${GREEN}✓${NC} Workers: $workers_val"
        fi
    fi
else
    echo -e "  ${YELLOW}⚠${NC} No debug.log found (app not started?)"
fi

echo ""
echo "📖 For detailed testing guide, see: docs/TEST_PARALLEL_EXECUTION.md"
echo ""
