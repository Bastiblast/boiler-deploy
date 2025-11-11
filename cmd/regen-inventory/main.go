package main

import (
"fmt"
"os"

"github.com/bastiblast/boiler-deploy/internal/inventory"
"github.com/bastiblast/boiler-deploy/internal/storage"
)

func main() {
if len(os.Args) < 2 {
fmt.Println("Usage: regen-inventory <environment>")
os.Exit(1)
}

envName := os.Args[1]

stor := storage.NewStorage(".")
env, err := stor.LoadEnvironment(envName)
if err != nil {
fmt.Fprintf(os.Stderr, "Error loading environment: %v\n", err)
os.Exit(1)
}

gen := inventory.NewGenerator()

// Regenerate group_vars
content, err := gen.GenerateGroupVarsYAML(*env)
if err != nil {
fmt.Fprintf(os.Stderr, "Error generating group_vars: %v\n", err)
os.Exit(1)
}

groupVarsFile := fmt.Sprintf("inventory/%s/group_vars/all.yml", envName)
if err := os.WriteFile(groupVarsFile, content, 0644); err != nil {
fmt.Fprintf(os.Stderr, "Error writing group_vars: %v\n", err)
os.Exit(1)
}

fmt.Printf("✓ Regenerated %s\n", groupVarsFile)
}
