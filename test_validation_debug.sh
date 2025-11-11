#!/bin/bash

# Test script to debug validation issue

echo "=== Testing Validation ==="
echo ""

# Check inventory exists
echo "1. Checking inventory files..."
ls -la inventory/docker/
echo ""

# Check config
echo "2. Checking docker config..."
cat inventory/docker/config.yml
echo ""

# Check status file
echo "3. Checking status before validation..."
cat inventory/docker/.status/servers.json 2>/dev/null || echo "No status file yet"
echo ""

# Try to validate using a simple test
echo "4. Testing server validation manually..."
cd /home/basthook/devIronMenth/boiler-deploy

# Create a simple Go test program
cat > /tmp/test_validate.go <<'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/status"
	"github.com/bastiblast/boiler-deploy/internal/storage"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	
	// Load docker environment
	stor := storage.NewStorage(".")
	env, err := stor.LoadEnvironment("docker")
	if err != nil {
		log.Fatalf("Failed to load environment: %v", err)
	}
	
	fmt.Printf("Loaded environment: %s\n", env.Name)
	fmt.Printf("Number of servers: %d\n", len(env.Servers))
	
	for _, server := range env.Servers {
		fmt.Printf("\nServer: %s\n", server.Name)
		fmt.Printf("  IP: %s\n", server.IP)
		fmt.Printf("  Port: %d\n", server.Port)
		fmt.Printf("  SSH Key: %s\n", server.SSHKeyPath)
		fmt.Printf("  Git Repo: %s\n", server.GitRepo)
		fmt.Printf("  App Port: %d\n", server.AppPort)
		fmt.Printf("  Node Version: %s\n", server.NodeVersion)
	}
	
	// Create status manager
	statusMgr, err := status.NewManager("docker")
	if err != nil {
		log.Fatalf("Failed to create status manager: %v", err)
	}
	
	// Validate first server
	if len(env.Servers) > 0 {
		server := &env.Servers[0]
		fmt.Printf("\n=== Validating %s ===\n", server.Name)
		
		checks := statusMgr.ValidateServer(server)
		fmt.Printf("Validation results:\n")
		fmt.Printf("  IP Valid: %v\n", checks.IPValid)
		fmt.Printf("  SSH Key Exists: %v\n", checks.SSHKeyExists)
		fmt.Printf("  Port Valid: %v\n", checks.PortValid)
		fmt.Printf("  All Fields Filled: %v\n", checks.AllFieldsFilled)
		fmt.Printf("  Is Ready: %v\n", checks.IsReady())
		
		// Update status
		err = statusMgr.UpdateReadyChecks(server.Name, checks)
		if err != nil {
			log.Printf("Failed to update ready checks: %v", err)
		} else {
			fmt.Println("\n✓ Status updated successfully")
		}
		
		// Read status file
		statusFile := filepath.Join("inventory", "docker", ".status", "servers.json")
		data, err := os.ReadFile(statusFile)
		if err != nil {
			log.Printf("Failed to read status file: %v", err)
		} else {
			fmt.Printf("\n=== Status file content ===\n")
			var statuses map[string]*status.ServerStatus
			json.Unmarshal(data, &statuses)
			jsonData, _ := json.MarshalIndent(statuses, "", "  ")
			fmt.Println(string(jsonData))
		}
	}
}
EOF

echo "5. Running validation test..."
cd /home/basthook/devIronMenth/boiler-deploy
go run /tmp/test_validate.go

echo ""
echo "=== Test Complete ==="
