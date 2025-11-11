package inventory

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateHostsYAML generates Ansible hosts.yml content
func (g *Generator) GenerateHostsYAML(env Environment) ([]byte, error) {
	// Group servers by type
	webServers := make(map[string]interface{})
	dbServers := make(map[string]interface{})
	monitoringServers := make(map[string]interface{})
	
	for _, server := range env.Servers {
		serverConfig := map[string]interface{}{
			"ansible_host":                server.IP,
			"ansible_user":                server.SSHUser,
			"ansible_port":                server.Port,
			"ansible_python_interpreter":  "/usr/bin/python3",
			"ansible_ssh_private_key_file": server.SSHKeyPath,
			"ansible_become":              server.AnsibleBecome,
		}
		
		if server.AppPort > 0 {
			serverConfig["app_port"] = server.AppPort
		}
		
		switch server.Type {
		case "web":
			webServers[server.Name] = serverConfig
		case "db":
			dbServers[server.Name] = serverConfig
		case "monitoring":
			monitoringServers[server.Name] = serverConfig
		}
	}
	
	// Build the structure
	hosts := map[string]interface{}{
		"all": map[string]interface{}{
			"children": map[string]interface{}{},
		},
	}
	
	children := hosts["all"].(map[string]interface{})["children"].(map[string]interface{})
	
	if len(webServers) > 0 {
		children["webservers"] = map[string]interface{}{
			"hosts": webServers,
		}
	}
	
	if len(dbServers) > 0 {
		children["dbservers"] = map[string]interface{}{
			"hosts": dbServers,
		}
	}
	
	if len(monitoringServers) > 0 {
		children["monitoring"] = map[string]interface{}{
			"hosts": monitoringServers,
		}
	}
	
	return yaml.Marshal(hosts)
}

// GenerateGroupVarsYAML generates group_vars/all.yml content
func (g *Generator) GenerateGroupVarsYAML(env Environment) ([]byte, error) {
	groupVars := map[string]interface{}{
		"app_name":       env.Config.AppName,
		"app_repo":       env.Config.AppRepo,
		"app_branch":     env.Config.AppBranch,
		"nodejs_version": env.Config.NodeJSVersion,
		"app_port":       env.Config.AppPort,
		"deploy_user":    env.Config.DeployUser,
		"timezone":       env.Config.Timezone,
	}
	
	return yaml.Marshal(groupVars)
}

// GenerateEnvironmentSummary generates a human-readable summary
func (g *Generator) GenerateEnvironmentSummary(env Environment) string {
	summary := fmt.Sprintf("Environment: %s\n", env.Name)
	summary += fmt.Sprintf("═══════════════════════════════\n\n")
	
	summary += "Services:\n"
	if env.Services.Web {
		summary += "  ✓ Web servers\n"
	}
	if env.Services.Database {
		summary += "  ✓ Database servers\n"
	}
	if env.Services.Monitoring {
		summary += "  ✓ Monitoring\n"
	}
	summary += "\n"
	
	summary += "Configuration:\n"
	summary += fmt.Sprintf("  App: %s\n", env.Config.AppName)
	summary += fmt.Sprintf("  Repo: %s\n", env.Config.AppRepo)
	summary += fmt.Sprintf("  Branch: %s\n", env.Config.AppBranch)
	summary += fmt.Sprintf("  Node.js: %s\n", env.Config.NodeJSVersion)
	summary += fmt.Sprintf("  Port: %s\n", env.Config.AppPort)
	summary += "\n"
	
	summary += fmt.Sprintf("Servers (%d total):\n", len(env.Servers))
	for _, server := range env.Servers {
		summary += fmt.Sprintf("  • %s (%s) - %s:%d\n", 
			server.Name, server.Type, server.IP, server.AppPort)
	}
	
	return summary
}
