package ansible

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/logger"
	"github.com/bastiblast/boiler-deploy/internal/ssh"
	"github.com/rs/zerolog"
)

type Executor struct {
	environment string
	logDir      string
	log         zerolog.Logger
}

func NewExecutor(environment string) *Executor {
	logDir := filepath.Join("logs", environment)
	os.MkdirAll(logDir, 0755)

	return &Executor{
		environment: environment,
		logDir:      logDir,
		log:         logger.Get("executor"),
	}
}

type ExecutionResult struct {
	Success      bool
	ErrorMessage string
	LogFile      string
}

func (e *Executor) RunPlaybook(playbook string, serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContext(context.Background(), playbook, serverName, "", progressChan)
}

func (e *Executor) RunPlaybookWithTags(playbook string, serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContext(context.Background(), playbook, serverName, tags, progressChan)
}

func (e *Executor) RunPlaybookWithOptions(playbook string, serverName string, tags string, checkMode bool, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContextAndOptions(context.Background(), playbook, serverName, tags, checkMode, progressChan)
}

// RunPlaybookWithContext runs playbook with context for cancellation
func (e *Executor) RunPlaybookWithContext(ctx context.Context, playbook string, serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContextAndOptions(ctx, playbook, serverName, tags, false, progressChan)
}

func (e *Executor) RunPlaybookWithContextAndOptions(ctx context.Context, playbook string, serverName string, tags string, checkMode bool, progressChan chan<- string) (*ExecutionResult, error) {
	// Add timeout if none specified
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()
	}
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

	// Use CommandContext for cancellation support
	cmd := exec.CommandContext(ctx, "ansible-playbook", args...)

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

	// Stream output in goroutines with context awareness
	outputDone := make(chan struct{}, 2)
	go func() {
		e.streamOutput(stdout, logWriter, progressChan)
		outputDone <- struct{}{}
	}()
	go func() {
		e.streamOutput(stderr, logWriter, nil)
		outputDone <- struct{}{}
	}()

	// Wait for command with context cancellation monitoring
	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	var cmdErr error
	select {
	case <-ctx.Done():
		// Context cancelled - kill process
		if cmd.Process != nil {
			log.Printf("[EXECUTOR] Context cancelled, killing ansible process")
			cmd.Process.Kill()
		}
		cmdErr = ctx.Err()
		// Wait for process to be killed
		<-waitDone
	case cmdErr = <-waitDone:
		// Command completed normally
	}

	// Wait for output goroutines to finish
	<-outputDone
	<-outputDone
	result := &ExecutionResult{
		Success: cmdErr == nil,
		LogFile: logFile,
	}

	if cmdErr != nil {
		result.ErrorMessage = cmdErr.Error()
		if progressChan != nil {
			progressChan <- fmt.Sprintf("❌ %s failed: %v", action, cmdErr)
		}
		// Return error for context cancellation
		if cmdErr == context.DeadlineExceeded || cmdErr == context.Canceled {
			return result, cmdErr
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
	return e.ProvisionWithContext(context.Background(), serverName, "", progressChan)
}

func (e *Executor) ProvisionWithTags(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.ProvisionWithContext(context.Background(), serverName, tags, progressChan)
}

func (e *Executor) ProvisionWithContext(ctx context.Context, serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContextAndOptions(ctx, "provision.yml", serverName, tags, false, progressChan)
}

func (e *Executor) ProvisionCheck(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContextAndOptions(context.Background(), "provision.yml", serverName, tags, true, progressChan)
}

func (e *Executor) Deploy(serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.DeployWithContext(context.Background(), serverName, "", progressChan)
}

func (e *Executor) DeployWithTags(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.DeployWithContext(context.Background(), serverName, tags, progressChan)
}

func (e *Executor) DeployWithContext(ctx context.Context, serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContextAndOptions(ctx, "deploy.yml", serverName, tags, false, progressChan)
}

func (e *Executor) DeployCheck(serverName string, tags string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybookWithContextAndOptions(context.Background(), "deploy.yml", serverName, tags, true, progressChan)
}

func (e *Executor) HealthCheck(ip string, port int) error {
	// For local development (127.0.0.1), use direct HTTP check
	// For remote servers, this will still work if ports are properly forwarded
	url := fmt.Sprintf("http://%s:%d/", ip, port)
	log.Printf("[EXECUTOR] Health check starting for: %s", url)
	
	// Retry logic: try multiple times with increasing delays
	maxRetries := 5
	retryDelays := []time.Duration{2 * time.Second, 3 * time.Second, 5 * time.Second, 8 * time.Second, 10 * time.Second}
	
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("[EXECUTOR] Health check retry %d/%d after %v delay", i+1, maxRetries, retryDelays[i-1])
			time.Sleep(retryDelays[i-1])
		}
		
		// Try curl first (if available)
		if err := e.healthCheckCurl(url); err == nil {
			log.Printf("[EXECUTOR] ✓ Health check successful via curl on attempt %d", i+1)
			return nil
		} else {
			lastErr = err
			log.Printf("[EXECUTOR] Health check curl attempt %d/%d failed: %v", i+1, maxRetries, err)
		}
		
		// Fallback: Native Go HTTP client (always available, no external dependency)
		if err := e.healthCheckNative(url); err == nil {
			log.Printf("[EXECUTOR] ✓ Health check successful via native HTTP on attempt %d", i+1)
			return nil
		} else {
			lastErr = err
			log.Printf("[EXECUTOR] Health check native attempt %d/%d failed: %v", i+1, maxRetries, err)
		}
		
		// Check if service exists at all with nc/telnet (diagnostic only)
		if i == 0 {
			tcpCheckCmd := exec.Command("nc", "-zv", "-w", "3", ip, fmt.Sprintf("%d", port))
			tcpOutput, tcpErr := tcpCheckCmd.CombinedOutput()
			if tcpErr != nil {
				log.Printf("[EXECUTOR] Port %d appears closed or unreachable: %s", port, strings.TrimSpace(string(tcpOutput)))
			} else {
				log.Printf("[EXECUTOR] Port %d is open, but HTTP request failed", port)
			}
		}
	}
	
	log.Printf("[EXECUTOR] ✗ Health check failed after %d attempts", maxRetries)
	return fmt.Errorf("health check failed after %d attempts: %w", maxRetries, lastErr)
}

// HealthCheckRemote performs health check via SSH on the remote server
// This is needed when the app listens only on localhost inside the server
func (e *Executor) HealthCheckRemote(sshHost string, sshPort int, sshUser string, sshKeyPath string, appPort int) error {
	log.Printf("[EXECUTOR] Remote health check via SSH to %s:%d checking localhost:%d", sshHost, sshPort, appPort)
	
	maxRetries := 5
	retryDelays := []time.Duration{2 * time.Second, 3 * time.Second, 5 * time.Second, 8 * time.Second, 10 * time.Second}
	
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Printf("[EXECUTOR] Remote health check retry %d/%d after %v delay", i+1, maxRetries, retryDelays[i-1])
			time.Sleep(retryDelays[i-1])
		}
		
		// Try curl on remote server (check localhost from inside)
		cmd := fmt.Sprintf("curl -sf -m 5 http://localhost:%d/ > /dev/null 2>&1 && echo 'OK' || echo 'FAIL'", appPort)
		result := ssh.ExecuteCommand(sshHost, sshPort, sshUser, sshKeyPath, cmd)
		
		if result.Success && strings.TrimSpace(result.Output) == "OK" {
			log.Printf("[EXECUTOR] ✓ Remote health check successful on attempt %d", i+1)
			return nil
		}
		
		log.Printf("[EXECUTOR] Remote health check attempt %d/%d failed: %s", i+1, maxRetries, result.Message)
	}
	
	log.Printf("[EXECUTOR] ✗ Remote health check failed after %d attempts", maxRetries)
	return fmt.Errorf("remote health check failed after %d attempts", maxRetries)
}

// healthCheckCurl uses curl command (traditional method)
func (e *Executor) healthCheckCurl(url string) error {
	cmd := exec.Command("curl", "-sf", "-m", "10", "--connect-timeout", "5", url)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		e.log.Debug().Str("method", "curl").Str("url", url).Err(err).Msg("Health check failed")
		return fmt.Errorf("curl failed: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}
	
	e.log.Info().Str("method", "curl").Int("bytes", len(output)).Msg("Health check successful")
	return nil
}

// healthCheckNative uses Go's native HTTP client (no external dependencies)
func (e *Executor) healthCheckNative(url string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConns:        1,
			IdleConnTimeout:     5 * time.Second,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
	
	resp, err := client.Get(url)
	if err != nil {
		e.log.Debug().Str("method", "native_http").Str("url", url).Err(err).Msg("Health check failed")
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Accept 2xx-4xx status codes (app is responding)
	// 5xx means server error but app is alive
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		e.log.Info().Str("method", "native_http").Int("status", resp.StatusCode).Msg("Health check successful")
		return nil
	}
	
	e.log.Warn().Str("method", "native_http").Int("status", resp.StatusCode).Msg("Health check bad status")
	return fmt.Errorf("bad HTTP status: %d (%s)", resp.StatusCode, resp.Status)
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
