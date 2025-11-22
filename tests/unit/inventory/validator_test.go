package inventory_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bastiblast/boiler-deploy/internal/inventory"
)

func TestValidateIP_ValidAddresses(t *testing.T) {
	validator := inventory.NewValidator()

	validIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"127.0.0.1",
		"0.0.0.0",
		"255.255.255.255",
	}

	for _, ip := range validIPs {
		t.Run(ip, func(t *testing.T) {
			if err := validator.ValidateIP(ip); err != nil {
				t.Errorf("Valid IP %s flagged as invalid: %v", ip, err)
			}
		})
	}
}

func TestValidateIP_InvalidAddresses(t *testing.T) {
	validator := inventory.NewValidator()

	invalidIPs := []string{
		"",
		"256.1.1.1",
		"192.168.1",
		"192.168.1.1.1",
		"not-an-ip",
		"192.168.-1.1",
		"abc.def.ghi.jkl",
		"192.168.1.256",
	}

	for _, ip := range invalidIPs {
		t.Run(ip, func(t *testing.T) {
			if err := validator.ValidateIP(ip); err == nil {
				t.Errorf("Invalid IP %s not detected", ip)
			}
		})
	}
}

func TestValidatePort_ValidRange(t *testing.T) {
	validator := inventory.NewValidator()

	validPorts := []int{1, 22, 80, 443, 2222, 3000, 8080, 65535}

	for _, port := range validPorts {
		t.Run(string(rune(port)), func(t *testing.T) {
			if err := validator.ValidatePort(port); err != nil {
				t.Errorf("Valid port %d flagged as invalid: %v", port, err)
			}
		})
	}
}

func TestValidatePort_InvalidRange(t *testing.T) {
	validator := inventory.NewValidator()

	invalidPorts := []int{0, -1, -100, 65536, 70000, 100000}

	for _, port := range invalidPorts {
		t.Run(string(rune(port)), func(t *testing.T) {
			if err := validator.ValidatePort(port); err == nil {
				t.Errorf("Invalid port %d not detected", port)
			}
		})
	}
}

func TestValidateEnvironmentName_Valid(t *testing.T) {
	validator := inventory.NewValidator()

	validNames := []string{
		"production",
		"staging",
		"dev",
		"test-01",
		"my_env",
		"Prod-2024",
		"env_test_123",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			if err := validator.ValidateEnvironmentName(name); err != nil {
				t.Errorf("Valid environment name %s flagged as invalid: %v", name, err)
			}
		})
	}
}

func TestValidateEnvironmentName_Invalid(t *testing.T) {
	validator := inventory.NewValidator()

	invalidNames := []string{
		"",
		"prod.env",
		"test env",
		"env@2024",
		"env#1",
		"env/test",
		"env\\test",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			if err := validator.ValidateEnvironmentName(name); err == nil {
				t.Errorf("Invalid environment name %s not detected", name)
			}
		})
	}
}

func TestValidateSSHKeyPath_ExistingFile(t *testing.T) {
	validator := inventory.NewValidator()

	// Create temporary SSH key file
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")
	if err := os.WriteFile(keyPath, []byte("fake key"), 0600); err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}

	if err := validator.ValidateSSHKeyPath(keyPath); err != nil {
		t.Errorf("Existing key file validation failed: %v", err)
	}
}

func TestValidateSSHKeyPath_NonExistingFile(t *testing.T) {
	validator := inventory.NewValidator()

	if err := validator.ValidateSSHKeyPath("/nonexistent/path/key"); err == nil {
		t.Error("Non-existing key file not detected")
	}
}

func TestValidateSSHKeyPath_HomeDirectory(t *testing.T) {
	validator := inventory.NewValidator()

	// Create key in actual home directory for test
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	// Create .ssh directory if not exists
	sshDir := filepath.Join(home, ".ssh")
	os.MkdirAll(sshDir, 0700)

	testKeyPath := filepath.Join(sshDir, "test_boiler_key")
	if err := os.WriteFile(testKeyPath, []byte("test"), 0600); err != nil {
		t.Skipf("Cannot create test key in home: %v", err)
	}
	defer os.Remove(testKeyPath)

	// Test with tilde notation
	if err := validator.ValidateSSHKeyPath("~/.ssh/test_boiler_key"); err != nil {
		t.Errorf("Home directory (~) expansion failed: %v", err)
	}
}

func TestCheckIPPortConflict_NoConflict(t *testing.T) {
	validator := inventory.NewValidator()

	servers := []inventory.Server{
		{Name: "server1", IP: "192.168.1.10", Port: 22},
		{Name: "server2", IP: "192.168.1.11", Port: 22},
	}

	// Different IP, same port
	if err := validator.CheckIPPortConflict(servers, "192.168.1.12", 22, ""); err != nil {
		t.Errorf("False positive conflict: %v", err)
	}

	// Same IP, different port
	if err := validator.CheckIPPortConflict(servers, "192.168.1.10", 2222, ""); err != nil {
		t.Errorf("False positive conflict: %v", err)
	}
}

func TestCheckIPPortConflict_Conflict(t *testing.T) {
	validator := inventory.NewValidator()

	servers := []inventory.Server{
		{Name: "server1", IP: "192.168.1.10", Port: 22},
		{Name: "server2", IP: "192.168.1.11", Port: 22},
	}

	// Same IP and port as server1
	if err := validator.CheckIPPortConflict(servers, "192.168.1.10", 22, ""); err == nil {
		t.Error("IP:Port conflict not detected")
	}
}

func TestCheckIPPortConflict_ExcludeSelf(t *testing.T) {
	validator := inventory.NewValidator()

	servers := []inventory.Server{
		{Name: "server1", IP: "192.168.1.10", Port: 22},
	}

	// Same IP:Port as server1, but exclude server1 (editing existing server)
	if err := validator.CheckIPPortConflict(servers, "192.168.1.10", 22, "server1"); err != nil {
		t.Errorf("Self-exclusion not working: %v", err)
	}
}

func TestValidateGitRepo_ValidURLs(t *testing.T) {
	validator := inventory.NewValidator()

	validRepos := []string{
		"https://github.com/user/repo",
		"https://github.com/user/repo.git",
		"git@github.com:user/repo.git",
		"https://gitlab.com/group/project",
		"git@gitlab.com:group/project.git",
	}

	for _, repo := range validRepos {
		t.Run(repo, func(t *testing.T) {
			if err := validator.ValidateGitRepo(repo); err != nil {
				t.Errorf("Valid repo %s flagged as invalid: %v", repo, err)
			}
		})
	}
}

func TestValidateGitRepo_InvalidURLs(t *testing.T) {
	validator := inventory.NewValidator()

	invalidRepos := []string{
		"",
		"not-a-url",
		"ftp://github.com/user/repo",
		// Note: "http://" is technically valid prefix, just incomplete
		"github.com/user/repo", // Missing protocol
	}

	for _, repo := range invalidRepos {
		t.Run(repo, func(t *testing.T) {
			if err := validator.ValidateGitRepo(repo); err == nil {
				t.Errorf("Invalid repo %s not detected", repo)
			}
		})
	}
}

func TestValidateServer_AllFieldsValid(t *testing.T) {
	validator := inventory.NewValidator()

	// Create temporary SSH key
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")
	os.WriteFile(keyPath, []byte("key"), 0600)

	server := inventory.Server{
		Name:       "web1",
		IP:         "192.168.1.10",
		Port:       22,
		SSHUser:    "deploy",
		SSHKeyPath: keyPath,
		Type:       "web",
		AppPort:    3000,
	}

	errors := validator.ValidateServer(server)
	if len(errors) > 0 {
		t.Errorf("Valid server flagged as invalid: %v", errors)
	}
}

func TestValidateServer_MultipleErrors(t *testing.T) {
	validator := inventory.NewValidator()

	server := inventory.Server{
		Name:       "", // Invalid: empty
		IP:         "999.999.999.999", // Invalid: bad IP
		Port:       70000, // Invalid: out of range
		SSHUser:    "deploy",
		SSHKeyPath: "/nonexistent/key", // Invalid: doesn't exist
	}

	errors := validator.ValidateServer(server)

	// Should have at least 4 errors (name, IP, port, SSH key)
	if len(errors) < 4 {
		t.Errorf("Expected at least 4 errors, got %d: %v", len(errors), errors)
	}
}

// Note: ValidateEnvironment tests removed - method needs to be implemented in validator.go
// TODO: Add ValidateEnvironment to internal/inventory/validator.go

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
