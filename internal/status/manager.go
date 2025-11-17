package status

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/inventory"
)

type Manager struct {
	mu            sync.RWMutex
	statuses      map[string]*ServerStatus
	environment   string
	statusFile    string
}

func NewManager(environment string) (*Manager, error) {
	statusDir := filepath.Join("inventory", environment, ".status")
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create status directory: %w", err)
	}

	m := &Manager{
		statuses:    make(map[string]*ServerStatus),
		environment: environment,
		statusFile:  filepath.Join(statusDir, "servers.json"),
	}

	if err := m.Load(); err != nil {
		return m, nil
	}

	return m, nil
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.statusFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var statuses map[string]*ServerStatus
	if err := json.Unmarshal(data, &statuses); err != nil {
		return err
	}

	m.statuses = statuses
	
	// Reset any "in-progress" or "failed" states on load
	needsSave := false
	for _, status := range m.statuses {
		if status.State == StateProvisioning || 
		   status.State == StateDeploying || 
		   status.State == StateVerifying ||
		   status.State == StateFailed {
			log.Printf("[STATUS] Resetting state for %s from %v to unknown", status.Name, status.State)
			status.State = StateUnknown
			status.ErrorMessage = ""
			needsSave = true
		}
	}
	
	if needsSave {
		if err := m.save(); err != nil {
			log.Printf("[STATUS] Error saving reset statuses: %v", err)
		}
	}

	return nil
}

func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.save()
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.statuses, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.statusFile, data, 0644)
}

func (m *Manager) GetStatus(serverName string) *ServerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if status, ok := m.statuses[serverName]; ok {
		return status
	}

	return &ServerStatus{
		Name:       serverName,
		State:      StateUnknown,
		LastUpdate: time.Now(),
	}
}

func (m *Manager) UpdateStatus(serverName string, state ServerState, action ActionType, errorMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("[STATUS] Updating status for %s: state=%v action=%v error=%q", serverName, state, action, errorMsg)

	status := &ServerStatus{
		Name:         serverName,
		State:        state,
		LastAction:   action,
		LastUpdate:   time.Now(),
		ErrorMessage: errorMsg,
	}

	m.statuses[serverName] = status
	err := m.save()
	if err != nil {
		log.Printf("[STATUS] Error saving status for %s: %v", serverName, err)
	} else {
		log.Printf("[STATUS] Successfully saved status for %s", serverName)
	}
	return err
}

func (m *Manager) UpdateReadyChecks(serverName string, checks ReadyChecks) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	status, ok := m.statuses[serverName]
	if !ok {
		status = &ServerStatus{
			Name:       serverName,
			State:      StateUnknown,
			LastUpdate: time.Now(),
		}
	}

	status.ReadyChecks = checks
	status.LastUpdate = time.Now()

	// Only update state to Ready/NotReady if not already in a more advanced state
	// Don't overwrite Provisioned or Deployed states
	if status.State != StateProvisioned && status.State != StateDeployed {
		if checks.IsReady() {
			if status.State == StateUnknown || status.State == StateNotReady {
				status.State = StateReady
			}
		} else {
			status.State = StateNotReady
		}
	}

	m.statuses[serverName] = status
	return m.save()
}

func (m *Manager) ValidateServer(server *inventory.Server) ReadyChecks {
	checks := ReadyChecks{
		IPValid:       isValidIP(server.IP),
		SSHKeyExists:  fileExists(server.SSHKeyPath),
		PortValid:     server.Port > 0 && server.Port <= 65535,
		AllFieldsFilled: server.Name != "" && server.IP != "" && 
			server.SSHKeyPath != "" && server.GitRepo != "" &&
			server.AppPort > 0 && server.NodeVersion != "",
	}
	return checks
}

func isValidIP(ip string) bool {
	if ip == "" {
		return false
	}
	parts := []byte(ip)
	dots := 0
	for i := 0; i < len(parts); i++ {
		if parts[i] == '.' {
			dots++
		} else if parts[i] < '0' || parts[i] > '9' {
			return false
		}
	}
	return dots == 3
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	
	// Expand ~ to home directory
	expandedPath := expandTilde(path)
	
	_, err := os.Stat(expandedPath)
	return err == nil
}

func expandTilde(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	
	if len(path) == 1 {
		return homeDir
	}
	
	if path[1] == '/' {
		return filepath.Join(homeDir, path[2:])
	}
	
	return path
}

func (m *Manager) GetAllStatuses() map[string]*ServerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ServerStatus)
	for k, v := range m.statuses {
		result[k] = v
	}
	return result
}
