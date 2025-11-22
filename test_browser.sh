#!/bin/bash
# Test script for browser opening functionality

echo "=== Browser Open Test ==="
echo ""

# Test 1: Check available browser openers
echo "1. Checking available browser openers:"
for cmd in xdg-open gnome-open wslview open; do
    if command -v $cmd &> /dev/null; then
        echo "   ✓ $cmd found at: $(which $cmd)"
    else
        echo "   ✗ $cmd not found"
    fi
done
echo ""

# Test 2: Test xdg-open directly
echo "2. Testing xdg-open with a test URL:"
if command -v xdg-open &> /dev/null; then
    echo "   Attempting to open http://example.com..."
    xdg-open "http://example.com" 2>&1 &
    sleep 2
    echo "   Command executed. Check if browser opened."
else
    echo "   xdg-open not available"
fi
echo ""

# Test 3: Check debug log
echo "3. Checking debug.log for browser events:"
if [ -f debug.log ]; then
    echo "   Last 10 browser-related logs:"
    grep -i "\[BROWSER\]\|\[WORKFLOW\].*browser\|deploySuccess" debug.log | tail -10 || echo "   No browser logs found"
else
    echo "   debug.log not found"
fi
echo ""

# Test 4: Instructions
echo "=== Instructions ==="
echo "To test the browser feature in the app:"
echo "1. Run: ./bin/inventory-manager"
echo "2. Deploy to a server"
echo "3. After successful deployment, look for: 'Press 'o' to open in browser'"
echo "4. Press 'o' key"
echo "5. Check debug.log for detailed logs"
echo ""
echo "To monitor logs in real-time:"
echo "   tail -f debug.log | grep -E 'BROWSER|WORKFLOW'"
