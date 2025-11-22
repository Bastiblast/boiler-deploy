package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/ssh"
	"github.com/bastiblast/boiler-deploy/internal/status"
	"github.com/bastiblast/boiler-deploy/internal/storage"
	"github.com/bastiblast/boiler-deploy/internal/ui"
)

func resetAllEnvironments() {
	stor := storage.NewStorage(".")
	envs, err := stor.ListEnvironments()
	if err != nil {
		log.Printf("Failed to list environments: %v", err)
		return
	}

	log.Printf("Resetting statuses for %d environments", len(envs))
	for _, envName := range envs {
		mgr, err := status.NewManager(envName)
		if err != nil {
			log.Printf("Failed to create status manager for %s: %v", envName, err)
			continue
		}
		log.Printf("Reset completed for environment: %s", envName)
		_ = mgr
	}
}

// validateAllServers checks the real state of all servers via SSH
func validateAllServers() {
	log.Println("[STARTUP] Starting server state validation...")
	
	stor := storage.NewStorage(".")
	envs, err := stor.ListEnvironments()
	if err != nil {
		log.Printf("[STARTUP] Failed to list environments: %v", err)
		return
	}
	
	detector := ssh.NewStateDetector()
	totalChecked := 0
	totalUpdated := 0
	
	for _, env := range envs {
		log.Printf("[STARTUP] Checking environment: %s", env)
		
		// Load servers from inventory
		servers, err := inventory.LoadServersForEnv(env)
		if err != nil {
			log.Printf("[STARTUP] Failed to load servers for %s: %v", env, err)
			continue
		}
		
		if len(servers) == 0 {
			log.Printf("[STARTUP] No servers in %s environment", env)
			continue
		}
		
		log.Printf("[STARTUP] Validating %d servers in %s...", len(servers), env)
		
		// Initialize status manager
		statusMgr, err := status.NewManager(env)
		if err != nil {
			log.Printf("[STARTUP] Failed to create status manager for %s: %v", env, err)
			continue
		}
		
		// Check each server
		for _, server := range servers {
			totalChecked++
			log.Printf("[STARTUP] Checking: %s (%s:%d)", server.Name, server.IP, server.Port)
			
			// Detect real state via SSH
			result := detector.DetectState(*server)
			
			// Get current status
			currentStatus := statusMgr.GetStatus(server.Name)
			
			// Update if state changed
			if currentStatus.State != result.State {
				log.Printf("[STARTUP] Updating %s: %s → %s (reason: %s)", 
					server.Name, currentStatus.State, result.State, result.Message)
				
				err := statusMgr.UpdateStatus(server.Name, result.State, "", result.Message)
				if err != nil {
					log.Printf("[STARTUP] Failed to update status for %s: %v", server.Name, err)
				} else {
					totalUpdated++
				}
			} else {
				log.Printf("[STARTUP] %s state confirmed: %s", server.Name, result.State)
			}
		}
	}
	
	log.Printf("[STARTUP] Validation complete: %d servers checked, %d statuses updated", totalChecked, totalUpdated)
}

func main() {
	// Setup debug logging
	logFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
	} else {
		defer logFile.Close()
		log.SetOutput(logFile)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		log.Println("============ Application Started ============")
	}

	// Reset all environments' in-progress statuses
	resetAllEnvironments()

	// Validate all servers and sync their real state
	log.Println("[STARTUP] Validating all servers state...")
	validateAllServers()

	p := tea.NewProgram(
		ui.NewMainMenu(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Printf("Application error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	log.Println("============ Application Exited ============")
}
