#!/bin/bash
# Stop Docker testing environment

echo "🛑 Stopping Docker Testing Environment..."
echo ""

docker-compose down

echo ""
echo "✅ Docker environment stopped!"
echo ""
echo "💡 To remove all data (volumes):"
echo "   docker-compose down -v"
echo ""
echo "🚀 To start again:"
echo "   ./docker-start.sh"
