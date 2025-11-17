package ansible

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Executor struct {
	environment string
	logDir      string
}

func NewExecutor(environment string) *Executor {
	logDir := filepath.Join("logs", environment)
	os.MkdirAll(logDir, 0755)

	return &Executor{
		environment: environment,
		logDir:      logDir,
	}
}

type ExecutionResult struct {
	Success      bool
	ErrorMessage string
	LogFile      string
}

func (e *Executor) RunPlaybook(playbook string, serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	timestamp := time.Now().Format("20060102_150405")
	action := strings.TrimSuffix(filepath.Base(playbook), ".yml")
	logFile := filepath.Join(e.logDir, fmt.Sprintf("%s_%s_%s.log", serverName, action, timestamp))

	logWriter, err := os.Create(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logWriter.Close()

	inventoryPath := filepath.Join("inventory", e.environment, "hosts.yml")
	playbookPath := filepath.Join("playbooks", playbook)

	// Send initial progress
	if progressChan != nil {
		progressChan <- fmt.Sprintf("🚀 Starting %s playbook...", action)
	}

	cmd := exec.Command("ansible-playbook",
		"-i", inventoryPath,
		playbookPath,
		"--limit", serverName,
	)

	// Don't use JSON callback as it can hang - use default callback and parse text
	cmd.Env = append(os.Environ(), "ANSIBLE_FORCE_COLOR=false")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ansible: %w", err)
	}

	go e.streamOutput(stdout, logWriter, progressChan)
	go e.streamOutput(stderr, logWriter, nil)

	err = cmd.Wait()
	result := &ExecutionResult{
		Success: err == nil,
		LogFile: logFile,
	}

	if err != nil {
		result.ErrorMessage = err.Error()
		if progressChan != nil {
			progressChan <- fmt.Sprintf("❌ %s failed: %v", action, err)
		}
	} else {
		if progressChan != nil {
			progressChan <- fmt.Sprintf("✅ %s completed successfully", action)
		}
	}

	return result, nil
}

func (e *Executor) streamOutput(reader io.Reader, writer io.Writer, progressChan chan<- string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(writer, line)

		if progressChan != nil {
			e.parseProgress(line, progressChan)
		}
	}
}

func (e *Executor) parseProgress(line string, progressChan chan<- string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	// Parse Ansible text output
	switch {
	case strings.HasPrefix(line, "PLAY ["):
		playName := strings.TrimPrefix(line, "PLAY [")
		playName = strings.TrimSuffix(playName, "]")
		playName = strings.TrimSpace(strings.Split(playName, "*")[0])
		progressChan <- fmt.Sprintf("▶️  Starting: %s", playName)
		
	case strings.HasPrefix(line, "TASK ["):
		taskName := strings.TrimPrefix(line, "TASK [")
		taskName = strings.TrimSuffix(taskName, "]")
		taskName = strings.TrimSpace(strings.Split(taskName, "*")[0])
		
		// Translate common task names to French descriptions
		taskName = e.translateTaskName(taskName)
		
		if len(taskName) > 60 {
			taskName = taskName[:57] + "..."
		}
		progressChan <- fmt.Sprintf("⚙️  %s", taskName)
		
	case strings.HasPrefix(line, "ok:"):
		// Don't show "ok" - only show changes to reduce noise
		return
		
	case strings.HasPrefix(line, "changed:"):
		// Extract what was changed if possible
		parts := strings.Fields(line)
		if len(parts) > 1 {
			serverName := strings.Trim(parts[1], "[]")
			progressChan <- fmt.Sprintf("  ✓ Modified on %s", serverName)
		} else {
			progressChan <- "  ✓ Configuration updated"
		}
		
	case strings.HasPrefix(line, "failed:") || strings.HasPrefix(line, "fatal:"):
		// Try to extract error message
		parts := strings.SplitN(line, "=>", 2)
		if len(parts) == 2 {
			msg := strings.TrimSpace(parts[1])
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}
			progressChan <- fmt.Sprintf("  ❌ Error: %s", msg)
		} else {
			progressChan <- "  ❌ Task failed"
		}
		
	case strings.HasPrefix(line, "skipping:"):
		// Don't show skipped tasks to reduce noise
		return
		
	case strings.Contains(line, "UNREACHABLE"):
		progressChan <- "  ⚠️  Server unreachable - check SSH connection"
		
	case strings.HasPrefix(line, "PLAY RECAP"):
		progressChan <- "📊 Summary of execution"
		
	case strings.Contains(line, "WARNING") && !strings.Contains(line, "Skipping"):
		// Filter out skipping warnings as they're not important
		if !strings.Contains(line, "as it is not a mapping") && 
		   !strings.Contains(line, "as this is not a valid group") {
			warningMsg := strings.TrimPrefix(line, "[WARNING]:")
			warningMsg = strings.TrimSpace(warningMsg)
			if len(warningMsg) > 80 {
				warningMsg = warningMsg[:77] + "..."
			}
			progressChan <- fmt.Sprintf("⚠️  Warning: %s", warningMsg)
		}
		
	case strings.Contains(line, "ERROR"):
		errorMsg := strings.TrimPrefix(line, "[ERROR]:")
		errorMsg = strings.TrimSpace(errorMsg)
		if len(errorMsg) > 80 {
			errorMsg = errorMsg[:77] + "..."
		}
		progressChan <- fmt.Sprintf("❌ Error: %s", errorMsg)
	}
}

func (e *Executor) translateTaskName(taskName string) string {
	translations := map[string]string{
		"Gathering Facts": "Collecting server information",
		"Wait for system to become reachable": "Waiting for server connection",
		"Update apt cache": "Updating package list",
		"Install required packages": "Installing system packages",
		"Install Node.js": "Installing Node.js",
		"Install NVM": "Installing Node Version Manager",
		"Install PM2 globally": "Installing PM2 process manager",
		"Create deployment user": "Creating deployment user",
		"Setup Nginx": "Configuring web server",
		"Install Nginx": "Installing Nginx web server",
		"Configure Nginx": "Configuring web server",
		"Install UFW": "Installing firewall",
		"Configure UFW": "Configuring firewall",
		"Install Fail2ban": "Installing Fail2ban security",
		"Configure Fail2ban": "Configuring Fail2ban",
		"Clone repository": "Downloading application code",
		"Install dependencies": "Installing application dependencies",
		"Build application": "Building application",
		"Start application": "Starting application with PM2",
		"Restart Nginx": "Restarting web server",
		"Enable and start services": "Starting system services",
	}
	
	// Check for exact match first
	if translated, ok := translations[taskName]; ok {
		return translated
	}
	
	// Check for partial matches
	for key, value := range translations {
		if strings.Contains(taskName, key) {
			return value
		}
	}
	
	return taskName
}

func (e *Executor) Provision(serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybook("provision.yml", serverName, progressChan)
}

func (e *Executor) Deploy(serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybook("deploy.yml", serverName, progressChan)
}

func (e *Executor) HealthCheck(ip string, port int) error {
	url := fmt.Sprintf("http://%s:%d/", ip, port)
	log.Printf("[EXECUTOR] Health check: %s", url)
	
	cmd := exec.Command("curl", "-sf", "-m", "5", url)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Printf("[EXECUTOR] Health check failed: %v", err)
		return fmt.Errorf("curl failed: %w", err)
	}
	
	log.Printf("[EXECUTOR] Health check successful (%d bytes)", len(output))
	return nil
}
