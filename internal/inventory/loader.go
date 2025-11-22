package inventory

import (
"fmt"
"os"
"path/filepath"

"gopkg.in/yaml.v3"
)

// LoadServersForEnv loads servers from inventory for a specific environment
func LoadServersForEnv(environment string) ([]*Server, error) {
inventoryPath := filepath.Join("inventory", environment, "hosts.yml")

// Check if file exists
if _, err := os.Stat(inventoryPath); os.IsNotExist(err) {
return nil, fmt.Errorf("inventory not found for environment %s", environment)
}

return LoadServers(inventoryPath)
}

// LoadServers loads servers from an inventory file
// Supports both formats:
// 1. all.children.*.hosts (Ansible standard)
// 2. all.hosts (flat, legacy)
func LoadServers(inventoryPath string) ([]*Server, error) {
data, err := os.ReadFile(inventoryPath)
if err != nil {
return nil, fmt.Errorf("failed to read inventory: %w", err)
}

// Parse as generic map first to detect format
var raw map[string]interface{}
if err := yaml.Unmarshal(data, &raw); err != nil {
return nil, fmt.Errorf("failed to parse inventory: %w", err)
}

var servers []*Server

// Try format 1: all.children.*.hosts (standard Ansible)
if all, ok := raw["all"].(map[string]interface{}); ok {
if children, ok := all["children"].(map[string]interface{}); ok {
// Iterate all groups (webservers, dbservers, etc.)
for _, group := range children {
if groupMap, ok := group.(map[string]interface{}); ok {
if hosts, ok := groupMap["hosts"].(map[string]interface{}); ok {
for name, hostData := range hosts {
if server := parseHost(name, hostData); server != nil {
servers = append(servers, server)
}
}
}
}
}
if len(servers) > 0 {
return servers, nil
}
}

// Try format 2: all.hosts (flat)
if hosts, ok := all["hosts"].(map[string]interface{}); ok {
for name, hostData := range hosts {
if server := parseHost(name, hostData); server != nil {
servers = append(servers, server)
}
}
if len(servers) > 0 {
return servers, nil
}
}
}

return nil, fmt.Errorf("no servers found in inventory (check format)")
}

// parseHost converts raw YAML host data to Server struct
func parseHost(name string, hostData interface{}) *Server {
hostMap, ok := hostData.(map[string]interface{})
if !ok {
return nil
}

server := &Server{Name: name}

if ip, ok := hostMap["ansible_host"].(string); ok {
server.IP = ip
}
if port, ok := hostMap["ansible_port"].(int); ok {
server.Port = port
}
if user, ok := hostMap["ansible_user"].(string); ok {
server.SSHUser = user
}
if appPort, ok := hostMap["app_port"].(int); ok {
server.AppPort = appPort
}
if gitRepo, ok := hostMap["git_repo"].(string); ok {
server.GitRepo = gitRepo
}
if sshKey, ok := hostMap["ansible_ssh_private_key_file"].(string); ok {
server.SSHKeyPath = sshKey
}

return server
}
