package config

import "time"

// ConfigOptions represents user-configurable runtime options
type ConfigOptions struct {
	// Provisioning options
	ProvisioningTags     []string      `yaml:"provisioning_tags"`
	ProvisioningStrategy string        `yaml:"provisioning_strategy"` // "sequential" or "parallel"
	
	// Deployment options
	DeploymentStrategy   string        `yaml:"deployment_strategy"`   // "rolling", "all_at_once", "blue_green"
	DeploymentTags       []string      `yaml:"deployment_tags"`
	
	// Health check options
	HealthCheckEnabled   bool          `yaml:"health_check_enabled"`
	HealthCheckTimeout   time.Duration `yaml:"health_check_timeout"`
	HealthCheckRetries   int           `yaml:"health_check_retries"`
	
	// Refresh and display options
	RefreshInterval      time.Duration `yaml:"refresh_interval"`     // Auto-refresh during operations
	LogRetention         int           `yaml:"log_retention_lines"`  // Number of log lines to keep
	
	// Retry options
	AutoRetryEnabled     bool          `yaml:"auto_retry_enabled"`
	MaxRetries           int           `yaml:"max_retries"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *ConfigOptions {
	return &ConfigOptions{
		ProvisioningTags:     []string{"all"},
		ProvisioningStrategy: "sequential",
		DeploymentStrategy:   "rolling",
		DeploymentTags:       []string{"all"},
		HealthCheckEnabled:   true,
		HealthCheckTimeout:   30 * time.Second,
		HealthCheckRetries:   3,
		RefreshInterval:      1 * time.Second,
		LogRetention:         100,
		AutoRetryEnabled:     false,
		MaxRetries:           3,
	}
}

// AvailableTags for provisioning
var ProvisioningTags = []string{
	"all",
	"base",
	"security",
	"firewall",
	"ssh",
	"nodejs",
	"nginx",
	"postgresql",
	"monitoring",
}

// AvailableTags for deployment
var DeploymentTags = []string{
	"all",
	"dependencies",
	"build",
	"deploy",
	"restart",
	"health_check",
}
