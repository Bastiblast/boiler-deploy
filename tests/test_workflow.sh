#!/bin/bash

echo "==================================="
echo "Testing Workflow Validation & Check"
echo "==================================="
echo ""
echo "Instructions:"
echo "1. Navigate to 'Working with Inventory'"
echo "2. Select a server with SPACE"
echo "3. Press 'v' to validate"
echo "4. Press 'c' to check"
echo "5. Watch debug.log in another terminal:"
echo "   tail -f debug.log"
echo ""
echo "Press Enter to start the application..."
read

# Clear old debug log
> debug.log

# Run the application
make run

echo ""
echo "Application closed. Check debug.log for details:"
echo "  tail -50 debug.log"
