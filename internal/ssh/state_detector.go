package ssh

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/status"
)

// StateDetectionResult represents the result of server state detection
type StateDetectionResult struct {
	State              status.ServerState
	Message            string
	ProvisioningChecks ProvisioningStatus
	DeploymentChecks   DeploymentStatus
}

// ProvisioningStatus contains the result of provisioning checks
type ProvisioningStatus struct {
	NodeInstalled    bool
	NginxInstalled   bool
	NVMInstalled     bool
	AppDirExists     bool
	AllProvisioned   bool
}

// DeploymentStatus contains the result of deployment checks
type DeploymentStatus struct {
	PM2Running        bool
	AppResponding     bool
	CurrentSymlink    bool
	AllDeployed       bool
}

// StateDetector detects the actual state of a server by connecting via SSH
type StateDetector struct{}

// NewStateDetector creates a new StateDetector instance
func NewStateDetector() *StateDetector {
	return &StateDetector{}
}

// DetectState connects to a server and detects its current state
func (sd *StateDetector) DetectState(server inventory.Server) StateDetectionResult {
	// Try to create SSH client
	client, err := sd.createSSHClient(server)
	if err != nil {
		return StateDetectionResult{
			State:   status.StateNotReady,
			Message: fmt.Sprintf("Offline - Cannot connect via SSH: %v", err),
		}
	}
	defer client.Close()

	// Check provisioning status
	provStatus := sd.checkProvisioning(client, server.SSHUser)
	
	// Check deployment status (only if provisioned)
	var deplStatus DeploymentStatus
	if provStatus.AllProvisioned {
		deplStatus = sd.checkDeployment(client, server.SSHUser, server.AppPort)
	}

	// Determine final state
	var finalState status.ServerState
	var message string

	if deplStatus.AllDeployed {
		finalState = status.StateDeployed
		message = "Application deployed and running"
	} else if provStatus.AllProvisioned {
		finalState = status.StateProvisioned
		message = "Server provisioned, ready for deployment"
	} else {
		finalState = status.StateReady
		message = "Server accessible but not provisioned"
	}

	return StateDetectionResult{
		State:              finalState,
		Message:            message,
		ProvisioningChecks: provStatus,
		DeploymentChecks:   deplStatus,
	}
}

// createSSHClient creates and returns an SSH client connection
func (sd *StateDetector) createSSHClient(server inventory.Server) (*ssh.Client, error) {
	keyPath := server.SSHKeyPath

	// Expand home directory if needed
	if len(keyPath) > 0 && keyPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot expand home directory: %w", err)
		}
		keyPath = home + keyPath[1:]
	}

	// Read private key
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read SSH key: %w", err)
	}

	// Parse private key
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("cannot parse SSH key: %w", err)
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: server.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect to server
	address := fmt.Sprintf("%s:%d", server.IP, server.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return client, nil
}

// executeCheck executes a command via SSH and returns the output
func (sd *StateDetector) executeCheck(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("cannot create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		// Command might return non-zero exit code, which is fine for checks
		return string(output), nil
	}

	return string(output), nil
}

// checkProvisioning checks if the server has been provisioned
func (sd *StateDetector) checkProvisioning(client *ssh.Client, user string) ProvisioningStatus {
	var status ProvisioningStatus

	// Check Node.js installed
	output, _ := sd.executeCheck(client, "command -v node >/dev/null 2>&1 && echo 'yes' || echo 'no'")
	status.NodeInstalled = strings.TrimSpace(output) == "yes"

	// Check Nginx installed
	output, _ = sd.executeCheck(client, "command -v nginx >/dev/null 2>&1 && echo 'yes' || echo 'no'")
	status.NginxInstalled = strings.TrimSpace(output) == "yes"

	// Check NVM installed
	// Handle root user (path is /root/.nvm, not /home/root/.nvm)
	nvmPath := fmt.Sprintf("/home/%s/.nvm", user)
	if user == "root" {
		nvmPath = "/root/.nvm"
	}
	output, _ = sd.executeCheck(client, fmt.Sprintf("test -d %s && echo 'yes' || echo 'no'", nvmPath))
	status.NVMInstalled = strings.TrimSpace(output) == "yes"

	// Check app directory exists
	output, _ = sd.executeCheck(client, "test -d /var/www && echo 'yes' || echo 'no'")
	status.AppDirExists = strings.TrimSpace(output) == "yes"

	// All checks must pass for provisioned state
	status.AllProvisioned = status.NodeInstalled && 
	                        status.NginxInstalled && 
	                        status.NVMInstalled && 
	                        status.AppDirExists

	return status
}

// checkDeployment checks if an application has been deployed
func (sd *StateDetector) checkDeployment(client *ssh.Client, user string, appPort int) DeploymentStatus {
	var status DeploymentStatus

	// Check PM2 running with at least one app
	// Handle root user (path is /root/.nvm, not /home/root/.nvm)
	nvmDir := fmt.Sprintf("/home/%s/.nvm", user)
	if user == "root" {
		nvmDir = "/root/.nvm"
	}
	nvmEnv := fmt.Sprintf("export NVM_DIR=%s && [ -s $NVM_DIR/nvm.sh ] && . $NVM_DIR/nvm.sh", nvmDir)
	pm2Command := fmt.Sprintf("%s && pm2 list 2>/dev/null | grep -q 'online' && echo 'yes' || echo 'no'", nvmEnv)
	output, _ := sd.executeCheck(client, pm2Command)
	status.PM2Running = strings.TrimSpace(output) == "yes"

	// Check app responding on port
	// Try curl first, fallback to wget, then nc (netcat)
	checkAppCommand := fmt.Sprintf(`
		if command -v curl >/dev/null 2>&1; then
			curl -s -o /dev/null -w '%%{http_code}' http://localhost:%d/ --max-time 3 2>/dev/null | grep -qE '200|307' && echo 'yes' || echo 'no'
		elif command -v wget >/dev/null 2>&1; then
			wget -q -O /dev/null --timeout=3 http://localhost:%d/ >/dev/null 2>&1 && echo 'yes' || echo 'no'
		elif command -v nc >/dev/null 2>&1; then
			echo "GET / HTTP/1.0" | nc -w 3 localhost %d >/dev/null 2>&1 && echo 'yes' || echo 'no'
		else
			echo 'no'
		fi
	`, appPort, appPort, appPort)
	output, _ = sd.executeCheck(client, checkAppCommand)
	status.AppResponding = strings.TrimSpace(output) == "yes"

	// Check current symlink exists
	output, _ = sd.executeCheck(client, "test -L /var/www/docker/current && echo 'yes' || echo 'no'")
	status.CurrentSymlink = strings.TrimSpace(output) == "yes"

	// All critical checks must pass for deployed state
	status.AllDeployed = status.PM2Running && status.AppResponding

	return status
}
