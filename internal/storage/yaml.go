package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"github.com/bastiblast/boiler-deploy/internal/inventory"
)

type Storage struct {
	basePath string
}

func NewStorage(basePath string) *Storage {
	return &Storage{
		basePath: basePath,
	}
}

// SaveEnvironment saves an environment to disk
func (s *Storage) SaveEnvironment(env inventory.Environment) error {
	envPath := filepath.Join(s.basePath, "inventory", env.Name)
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(envPath, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %v", err)
	}
	
	// Save environment config
	configPath := filepath.Join(envPath, "config.yml")
	configData, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}
	
	// Generate and save hosts.yml
	generator := inventory.NewGenerator()
	hostsData, err := generator.GenerateHostsYAML(env)
	if err != nil {
		return fmt.Errorf("failed to generate hosts.yml: %v", err)
	}
	
	hostsPath := filepath.Join(envPath, "hosts.yml")
	if err := os.WriteFile(hostsPath, hostsData, 0644); err != nil {
		return fmt.Errorf("failed to write hosts.yml: %v", err)
	}
	
	// Generate and save group_vars
	groupVarsPath := filepath.Join(s.basePath, "group_vars")
	if err := os.MkdirAll(groupVarsPath, 0755); err != nil {
		return fmt.Errorf("failed to create group_vars directory: %v", err)
	}
	
	groupVarsData, err := generator.GenerateGroupVarsYAML(env)
	if err != nil {
		return fmt.Errorf("failed to generate group_vars: %v", err)
	}
	
	groupVarsFile := filepath.Join(groupVarsPath, fmt.Sprintf("%s.yml", env.Name))
	if err := os.WriteFile(groupVarsFile, groupVarsData, 0644); err != nil {
		return fmt.Errorf("failed to write group_vars: %v", err)
	}
	
	return nil
}

// LoadEnvironment loads an environment from disk
func (s *Storage) LoadEnvironment(name string) (*inventory.Environment, error) {
	configPath := filepath.Join(s.basePath, "inventory", name, "config.yml")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}
	
	var env inventory.Environment
	if err := yaml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	
	return &env, nil
}

// ListEnvironments lists all available environments
func (s *Storage) ListEnvironments() ([]string, error) {
	invPath := filepath.Join(s.basePath, "inventory")
	
	entries, err := os.ReadDir(invPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read inventory directory: %v", err)
	}
	
	var envs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if config.yml exists
			configPath := filepath.Join(invPath, entry.Name(), "config.yml")
			if _, err := os.Stat(configPath); err == nil {
				envs = append(envs, entry.Name())
			}
		}
	}
	
	return envs, nil
}

// EnvironmentExists checks if an environment exists
func (s *Storage) EnvironmentExists(name string) bool {
	configPath := filepath.Join(s.basePath, "inventory", name, "config.yml")
	_, err := os.Stat(configPath)
	return err == nil
}

// DeleteEnvironment deletes an environment
func (s *Storage) DeleteEnvironment(name string) error {
	envPath := filepath.Join(s.basePath, "inventory", name)
	return os.RemoveAll(envPath)
}
