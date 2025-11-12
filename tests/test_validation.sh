#!/bin/bash

echo "Testing SSH Key Path Validation with Tilde Expansion"
echo "======================================================"
echo ""

# Clear debug log
> debug.log

# Create a simple test program to check validation
cat > /tmp/test_validation.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func expandTilde(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get home directory: %v", err)
		return path
	}
	
	if len(path) == 1 {
		return homeDir
	}
	
	if path[1] == '/' {
		return filepath.Join(homeDir, path[2:])
	}
	
	return path
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	
	expandedPath := expandTilde(path)
	fmt.Printf("Checking: %s\n", path)
	fmt.Printf("Expanded: %s\n", expandedPath)
	
	_, err := os.Stat(expandedPath)
	exists := err == nil
	fmt.Printf("Exists: %v\n", exists)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	return exists
}

func main() {
	testPaths := []string{
		"~/.ssh/Hostinger",
		"~/.ssh/id_rsa",
		"/home/basthook/.ssh/Hostinger",
		"/nonexistent/path",
		"",
	}
	
	for i, path := range testPaths {
		fmt.Printf("\nTest %d:\n", i+1)
		fmt.Printf("--------\n")
		result := fileExists(path)
		fmt.Printf("Result: %v\n", result)
	}
}
EOF

# Run the test
echo "Running standalone validation test..."
go run /tmp/test_validation.go
echo ""
echo "Test complete!"
echo ""
echo "Now checking the actual server validation in the app..."
echo "Check debug.log after running the app with 'make run'"
