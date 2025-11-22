package ansible

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ScriptExecutor runs deploy.sh and streams output
type ScriptExecutor struct {
	environment string
	logDir      string
	scriptPath  string
}

func NewScriptExecutor(environment string) *ScriptExecutor {
	logDir := filepath.Join("logs", environment)
	os.MkdirAll(logDir, 0755)

	return &ScriptExecutor{
		environment: environment,
		logDir:      logDir,
		scriptPath:  "./deploy.sh",
	}
}

// RunAction runs deploy.sh with specified action (provision, deploy, check)
func (e *ScriptExecutor) RunAction(action string, serverName string, outputChan chan<- string) (*ExecutionResult, error) {
	timestamp := time.Now().Format("20060102_150405")
	logFile := filepath.Join(e.logDir, fmt.Sprintf("%s_%s_%s.log", serverName, action, timestamp))

	logWriter, err := os.Create(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logWriter.Close()

	// Build command: ./deploy.sh ACTION ENVIRONMENT --yes
	// The --yes flag skips interactive prompts for automation
	cmd := exec.Command(e.scriptPath, action, e.environment, "--yes")
	cmd.Dir = "." // Run from project root

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start deploy.sh: %w", err)
	}

	// Stream output in real-time
	done := make(chan bool, 2)
	go e.streamLines(stdout, logWriter, outputChan, done)
	go e.streamLines(stderr, logWriter, outputChan, done)

	// Wait for streaming to complete
	<-done
	<-done

	// Wait for command to finish
	err = cmd.Wait()
	result := &ExecutionResult{
		Success: err == nil,
		LogFile: logFile,
	}

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("exit status %v", err)
	}

	return result, nil
}

// streamLines reads lines and sends them to both log file and output channel
func (e *ScriptExecutor) streamLines(reader io.Reader, writer io.Writer, outputChan chan<- string, done chan<- bool) {
	defer func() { done <- true }()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Write to log file
		fmt.Fprintln(writer, line)
		
		// Send to UI if channel provided
		if outputChan != nil {
			// Clean ANSI color codes for UI display
			cleanLine := stripAnsiCodes(line)
			if strings.TrimSpace(cleanLine) != "" {
				outputChan <- cleanLine
			}
		}
	}
}

// stripAnsiCodes removes ANSI escape sequences from a string
func stripAnsiCodes(str string) string {
	// Simple ANSI code removal (handles most common cases)
	var result strings.Builder
	inEscape := false
	
	for _, r := range str {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	
	return result.String()
}

// ValidateInventory checks if inventory is ready for deployment
func (e *ScriptExecutor) ValidateInventory(outputChan chan<- string) error {
	inventoryPath := filepath.Join("inventory", e.environment, "hosts.yml")
	
	if _, err := os.Stat(inventoryPath); os.IsNotExist(err) {
		return fmt.Errorf("inventory file not found: %s", inventoryPath)
	}

	// Run ansible inventory check
	cmd := exec.Command("ansible-inventory",
		"-i", inventoryPath,
		"--list",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if outputChan != nil {
			outputChan <- fmt.Sprintf("Inventory validation failed: %v", err)
		}
		return fmt.Errorf("invalid inventory: %w", err)
	}

	if outputChan != nil {
		outputChan <- "✓ Inventory file is valid"
		outputChan <- fmt.Sprintf("  Location: %s", inventoryPath)
	}

	// Parse and show servers
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" && outputChan != nil {
			outputChan <- "  " + line
		}
	}

	return nil
}

// CheckConnectivity tests SSH connectivity to servers
func (e *ScriptExecutor) CheckConnectivity(outputChan chan<- string) error {
	inventoryPath := filepath.Join("inventory", e.environment)
	
	cmd := exec.Command("ansible",
		"all",
		"-i", inventoryPath,
		"-m", "ping",
	)

	output, err := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) != "" && outputChan != nil {
			outputChan <- line
		}
	}

	if err != nil {
		return fmt.Errorf("connectivity check failed: %w", err)
	}

	return nil
}
