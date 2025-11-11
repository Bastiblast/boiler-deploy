package ansible

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

	cmd := exec.Command("ansible-playbook",
		"-i", inventoryPath,
		playbookPath,
		"--limit", serverName,
	)

	cmd.Env = append(os.Environ(), "ANSIBLE_STDOUT_CALLBACK=json", "ANSIBLE_FORCE_COLOR=false")

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
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return
	}

	if eventType, ok := event["event"].(string); ok {
		switch eventType {
		case "playbook_on_task_start":
			if task, ok := event["event_data"].(map[string]interface{}); ok {
				if taskName, ok := task["name"].(string); ok {
					progressChan <- fmt.Sprintf("Task: %s", taskName)
				}
			}
		case "runner_on_ok":
			progressChan <- "✓ Task completed"
		case "runner_on_failed":
			if result, ok := event["event_data"].(map[string]interface{}); ok {
				if msg, ok := result["msg"].(string); ok {
					progressChan <- fmt.Sprintf("✗ Failed: %s", msg)
				}
			}
		}
	}
}

func (e *Executor) Provision(serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybook("provision.yml", serverName, progressChan)
}

func (e *Executor) Deploy(serverName string, progressChan chan<- string) (*ExecutionResult, error) {
	return e.RunPlaybook("deploy.yml", serverName, progressChan)
}

func (e *Executor) HealthCheck(ip string, port int) error {
	cmd := exec.Command("curl", "-sf", "-m", "5", fmt.Sprintf("http://%s:%d/", ip, port))
	return cmd.Run()
}
