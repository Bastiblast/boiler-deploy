#!/bin/bash

# Test script to simulate UI interactions with the inventory manager
# This helps debug the provisioning freeze issue

echo "Starting test of provision UI flow..."
echo ""
echo "Steps to test manually:"
echo "1. Run: make run"
echo "2. Select 'Working with Inventory'"
echo "3. Select 'docker' environment"
echo "4. Use arrow keys to select docker-web-01 server"
echo "5. Press SPACE to check the server"
echo "6. Press 'p' for provision"
echo "7. Observe if tag selector appears (should NOT show 'Loading...')"
echo "8. Press ENTER to confirm tags"
echo "9. Check if provision starts (look for debug logs)"
echo ""
echo "Expected behavior:"
echo "- Tag selector should display immediately with tags"
echo "- After pressing ENTER, provision should start"
echo "- Status should update to show provisioning progress"
echo ""
echo "Debug logs will appear at the bottom showing:"
echo "- [DEBUG] 'p' key pressed"
echo "- [DEBUG] Opening tag selector"
echo "- [DEBUG] Tag selector confirmed"
echo "- [DEBUG] executeActionWithTags called"
echo ""
echo "Press ENTER to continue with manual test..."
read

cd /home/basthook/devIronMenth/boiler-deploy
make run
