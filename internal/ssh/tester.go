package ssh

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// TestResult represents the result of an SSH connection test
type TestResult struct {
	Success bool
	Message string
	Latency time.Duration
}

// TestConnection tests SSH connectivity to a server
func TestConnection(host string, port int, user string, keyPath string) TestResult {
	start := time.Now()
	
	// Expand home directory if needed
	if len(keyPath) > 0 && keyPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return TestResult{
				Success: false,
				Message: fmt.Sprintf("Cannot expand home directory: %v", err),
			}
		}
		keyPath = home + keyPath[1:]
	}

	// Read private key
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return TestResult{
			Success: false,
			Message: fmt.Sprintf("Cannot read SSH key: %v", err),
		}
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return TestResult{
			Success: false,
			Message: fmt.Sprintf("Cannot parse SSH key: %v", err),
		}
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For local use only
		Timeout:         10 * time.Second,
	}

	// Connect to server
	address := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		// Check if it's a network error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return TestResult{
				Success: false,
				Message: "Connection timeout",
			}
		}
		return TestResult{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
		}
	}
	defer client.Close()

	// Test command execution
	session, err := client.NewSession()
	if err != nil {
		return TestResult{
			Success: false,
			Message: fmt.Sprintf("Cannot create session: %v", err),
		}
	}
	defer session.Close()

	// Run a simple test command
	output, err := session.CombinedOutput("echo OK")
	if err != nil {
		return TestResult{
			Success: false,
			Message: fmt.Sprintf("Command execution failed: %v", err),
		}
	}

	latency := time.Since(start)

	if string(output) != "OK\n" {
		return TestResult{
			Success: false,
			Message: "Unexpected command output",
			Latency: latency,
		}
	}

	return TestResult{
		Success: true,
		Message: fmt.Sprintf("Connected successfully (%dms)", latency.Milliseconds()),
		Latency: latency,
	}
}
