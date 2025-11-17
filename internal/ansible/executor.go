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
		progressChan <- fmt.Sprintf("▶️  Play: %s", playName)
		
	case strings.HasPrefix(line, "TASK ["):
		taskName := strings.TrimPrefix(line, "TASK [")
		taskName = strings.TrimSuffix(taskName, "]")
		taskName = strings.TrimSpace(strings.Split(taskName, "*")[0])
		if len(taskName) > 60 {
			taskName = taskName[:57] + "..."
		}
		progressChan <- fmt.Sprintf("⚙️  %s", taskName)
		
	case strings.HasPrefix(line, "ok:"):
		progressChan <- "  ✓ OK"
		
	case strings.HasPrefix(line, "changed:"):
		progressChan <- "  ✓ Changed"
		
	case strings.HasPrefix(line, "failed:") || strings.HasPrefix(line, "fatal:"):
		// Try to extract error message
		parts := strings.SplitN(line, "=>", 2)
		if len(parts) == 2 {
			msg := strings.TrimSpace(parts[1])
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}
			progressChan <- fmt.Sprintf("  ❌ %s", msg)
		} else {
			progressChan <- "  ❌ Failed"
		}
		
	case strings.HasPrefix(line, "skipping:"):
		progressChan <- "  ⊘ Skipped"
		
	case strings.Contains(line, "UNREACHABLE"):
		progressChan <- "  ⚠️  Host unreachable"
		
	case strings.HasPrefix(line, "PLAY RECAP"):
		progressChan <- "📊 Execution summary"
		
	case strings.Contains(line, "WARNING"):
		if len(line) > 100 {
			line = line[:97] + "..."
		}
		progressChan <- fmt.Sprintf("⚠️  %s", line)
		
	case strings.Contains(line, "ERROR"):
		if len(line) > 100 {
			line = line[:97] + "..."
		}
		progressChan <- fmt.Sprintf("❌ %s", line)
	}
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
