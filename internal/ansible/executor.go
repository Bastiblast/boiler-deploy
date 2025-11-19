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

	"github.com/bastiblast/boiler-deploy/internal/ssh"
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
	return e.RunPlaybookWithTags(playbook, serverName, "", progressChan)
}

func (e *Executor) RunPlaybookWithTags(playbook string, serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithOptions(playbook, serverName, tags, false, progressChan)
}

func (e *Executor) RunPlaybookWithOptions(playbook string, serverName string, tags string, checkMode bool, progressChan chan<- string) (*ExecutionResult, error) {
	timestamp := time.Now().Format("20060102_150405")
	action := strings.TrimSuffix(filepath.Base(playbook), ".yml")
	
	// Add "check" suffix to log file in check mode
	logSuffix := ""
	if checkMode {
		logSuffix = "_check"
	}
	logFile := filepath.Join(e.logDir, fmt.Sprintf("%s_%s%s_%s.log", serverName, action, logSuffix, timestamp))

	logWriter, err := os.Create(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logWriter.Close()

	inventoryPath := filepath.Join("inventory", e.environment, "hosts.yml")
	playbookPath := filepath.Join("playbooks", playbook)

	// Send initial progress
	if progressChan != nil {
		modeStr := ""
		if checkMode {
			modeStr = " (dry-run mode)"
		}
		if tags != "" {
			progressChan <- fmt.Sprintf("🚀 Starting %s playbook with tags: %s%s...", action, tags, modeStr)
		} else {
			progressChan <- fmt.Sprintf("🚀 Starting %s playbook%s...", action, modeStr)
		}
	}

	args := []string{
		"-i", inventoryPath,
		playbookPath,
		"--limit", serverName,
	}
	
	// Add tags if specified
	if tags != "" {
		args = append(args, "--tags", tags)
	}
	
	// Add check and diff flags for dry-run mode
	if checkMode {
		args = append(args, "--check", "--diff")
	}

	cmd := exec.Command("ansible-playbook", args...)

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
		// Remove everything after the closing bracket
		if idx := strings.Index(playName, "]"); idx != -1 {
			playName = playName[:idx]
		}
		playName = strings.TrimSpace(playName)
		progressChan <- fmt.Sprintf("▶️  Starting: %s", playName)
		
	case strings.HasPrefix(line, "TASK ["):
		taskName := strings.TrimPrefix(line, "TASK [")
		// Remove everything after the closing bracket
		if idx := strings.Index(taskName, "]"); idx != -1 {
			taskName = taskName[:idx]
		}
		taskName = strings.TrimSpace(taskName)
		
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
	return e.ProvisionWithTags(serverName, "", progressChan)
}

func (e *Executor) ProvisionWithTags(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithOptions("provision.yml", serverName, tags, false, progressChan)
}

func (e *Executor) ProvisionCheck(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithOptions("provision.yml", serverName, tags, true, progressChan)
}

func (e *Executor) Deploy(serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.DeployWithTags(serverName, "", progressChan)
}

func (e *Executor) DeployWithTags(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithOptions("deploy.yml", serverName, tags, false, progressChan)
}

func (e *Executor) DeployCheck(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithOptions("deploy.yml", serverName, tags, true, progressChan)
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

func (e *Executor) TestSSH(ip string, port int, user string, keyPath string) ssh.TestResult {
	log.Printf("[EXECUTOR] Testing SSH connection to %s:%d with user %s", ip, port, user)
	result := ssh.TestConnection(ip, port, user, keyPath)
	
	if result.Success {
		log.Printf("[EXECUTOR] SSH test successful: %s", result.Message)
	} else {
		log.Printf("[EXECUTOR] SSH test failed: %s", result.Message)
	}
	
	return result
}
