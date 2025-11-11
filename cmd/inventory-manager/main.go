package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bastiblast/boiler-deploy/internal/ui"
)

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
