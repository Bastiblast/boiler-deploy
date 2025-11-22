package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager handles configuration persistence
type Manager struct {
	inventoryPath string
}

// NewManager creates a new configuration manager
func NewManager(inventoryPath string) *Manager {
	return &Manager{
		inventoryPath: inventoryPath,
	}
}

// Load loads configuration for an environment
func (m *Manager) Load(envName string) (*ConfigOptions, error) {
	configPath := filepath.Join(m.inventoryPath, envName, "config.yml")
	
	// If config doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	var config ConfigOptions
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	return &config, nil
}

// Save saves configuration for an environment
func (m *Manager) Save(envName string, config *ConfigOptions) error {
	configPath := filepath.Join(m.inventoryPath, envName, "config.yml")
	
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}

// Delete removes configuration for an environment
func (m *Manager) Delete(envName string) error {
	configPath := filepath.Join(m.inventoryPath, envName, "config.yml")
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}
	
	return os.Remove(configPath)
}
