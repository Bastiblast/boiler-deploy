package inventory

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// ValidateIP checks if the IP address format is valid
func (v *Validator) ValidateIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}
	return nil
}

// ValidatePort checks if port is in valid range
func (v *Validator) ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	return nil
}

// ValidateEnvironmentName checks if environment name is valid
func (v *Validator) ValidateEnvironmentName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("environment name cannot be empty")
	}
	
	// Only allow alphanumeric, dash, and underscore
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	if !matched {
		return fmt.Errorf("environment name can only contain letters, numbers, dash, and underscore")
	}
	
	return nil
}

// ValidateSSHKeyPath checks if SSH key file exists
func (v *Validator) ValidateSSHKeyPath(path string) error {
	// Expand home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %v", err)
		}
		path = strings.Replace(path, "~", home, 1)
	}
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("SSH key file not found: %s", path)
	}
	
	return nil
}

// CheckIPPortConflict checks for IP:Port conflicts
func (v *Validator) CheckIPPortConflict(servers []Server, newIP string, newPort int, excludeName string) error {
	for _, server := range servers {
		if server.Name == excludeName {
			continue
		}
		if server.IP == newIP && server.Port == newPort {
			return fmt.Errorf("IP:Port conflict with server %s (%s:%d)", 
				server.Name, server.IP, server.Port)
		}
	}
	return nil
}

// ValidateGitRepo basic validation of Git repository URL
func (v *Validator) ValidateGitRepo(repo string) error {
	if len(repo) == 0 {
		return fmt.Errorf("repository URL cannot be empty")
	}
	
	// Basic check for git URL patterns
	if !strings.HasPrefix(repo, "http://") && 
	   !strings.HasPrefix(repo, "https://") &&
	   !strings.HasPrefix(repo, "git@") {
		return fmt.Errorf("invalid repository URL format")
	}
	
	return nil
}

// ValidateServer validates all server fields
func (v *Validator) ValidateServer(server Server) []error {
	var errors []error
	
	if err := v.ValidateIP(server.IP); err != nil {
		errors = append(errors, err)
	}
	
	if err := v.ValidatePort(server.Port); err != nil {
		errors = append(errors, err)
	}
	
	if server.AppPort > 0 {
		if err := v.ValidatePort(server.AppPort); err != nil {
			errors = append(errors, fmt.Errorf("app_port: %v", err))
		}
	}
	
	if len(server.Name) == 0 {
		errors = append(errors, fmt.Errorf("server name cannot be empty"))
	}
	
	if len(server.SSHUser) == 0 {
		errors = append(errors, fmt.Errorf("SSH user cannot be empty"))
	}
	
	if err := v.ValidateSSHKeyPath(server.SSHKeyPath); err != nil {
		errors = append(errors, err)
	}
	
	return errors
}
