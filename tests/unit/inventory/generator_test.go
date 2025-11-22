package inventory_test

import (
	"testing"

	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"gopkg.in/yaml.v3"
)

func TestGenerateHostsYAML_SingleWebServer(t *testing.T) {
	gen := inventory.NewGenerator()

	env := inventory.Environment{
		Name: "test",
		Servers: []inventory.Server{
			{
				Name:          "web1",
				Type:          "web",
				IP:            "192.168.1.10",
				Port:          22,
				SSHUser:       "deploy",
				SSHKeyPath:    "/path/to/key",
				AppPort:       3000,
				AnsibleBecome: true,
			},
		},
	}

	data, err := gen.GenerateHostsYAML(env)
	if err != nil {
		t.Fatalf("Failed to generate hosts.yml: %v", err)
	}

	// Verify valid YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Generated invalid YAML: %v\nYAML content:\n%s", err, string(data))
	}

	// Verify structure: all > children > webservers > hosts > web1
	all, ok := result["all"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'all' key in generated YAML")
	}

	children, ok := all["children"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'children' key")
	}

	webservers, ok := children["webservers"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'webservers' group")
	}

	hosts, ok := webservers["hosts"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing hosts in webservers")
	}

	web1, exists := hosts["web1"]
	if !exists {
		t.Fatal("Expected web1 in webservers hosts")
	}

	// Verify server config
	web1Config := web1.(map[string]interface{})
	if web1Config["ansible_host"] != "192.168.1.10" {
		t.Errorf("Expected ansible_host 192.168.1.10, got %v", web1Config["ansible_host"])
	}
	if web1Config["ansible_user"] != "deploy" {
		t.Errorf("Expected ansible_user deploy, got %v", web1Config["ansible_user"])
	}
	if web1Config["ansible_port"] != 22 {
		t.Errorf("Expected ansible_port 22, got %v", web1Config["ansible_port"])
	}
}

func TestGenerateHostsYAML_MultipleServerTypes(t *testing.T) {
	gen := inventory.NewGenerator()

	env := inventory.Environment{
		Name: "production",
		Servers: []inventory.Server{
			{Name: "web1", Type: "web", IP: "192.168.1.10", Port: 22, SSHUser: "deploy", SSHKeyPath: "/key", AnsibleBecome: true},
			{Name: "web2", Type: "web", IP: "192.168.1.11", Port: 22, SSHUser: "deploy", SSHKeyPath: "/key", AnsibleBecome: true},
			{Name: "db1", Type: "db", IP: "192.168.1.20", Port: 22, SSHUser: "deploy", SSHKeyPath: "/key", AnsibleBecome: true},
			{Name: "mon1", Type: "monitoring", IP: "192.168.1.30", Port: 22, SSHUser: "deploy", SSHKeyPath: "/key", AnsibleBecome: true},
		},
	}

	data, err := gen.GenerateHostsYAML(env)
	if err != nil {
		t.Fatalf("Failed to generate hosts.yml: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Generated invalid YAML: %v", err)
	}

	children := result["all"].(map[string]interface{})["children"].(map[string]interface{})

	// Verify all groups exist
	if _, exists := children["webservers"]; !exists {
		t.Error("Missing webservers group")
	}
	if _, exists := children["dbservers"]; !exists {
		t.Error("Missing dbservers group")
	}
	if _, exists := children["monitoring"]; !exists {
		t.Error("Missing monitoring group")
	}

	// Verify server counts
	webHosts := children["webservers"].(map[string]interface{})["hosts"].(map[string]interface{})
	if len(webHosts) != 2 {
		t.Errorf("Expected 2 web servers, got %d", len(webHosts))
	}

	dbHosts := children["dbservers"].(map[string]interface{})["hosts"].(map[string]interface{})
	if len(dbHosts) != 1 {
		t.Errorf("Expected 1 db server, got %d", len(dbHosts))
	}
}

func TestGenerateGroupVarsYAML(t *testing.T) {
	gen := inventory.NewGenerator()

	env := inventory.Environment{
		Name: "staging",
		Config: inventory.Config{
			AppName:       "myapp",
			AppRepo:       "https://github.com/user/repo",
			AppBranch:     "develop",
			AppPort:       "3000",
			NodeJSVersion: "18",
			DeployUser:    "deploy",
			Timezone:      "UTC",
		},
		Servers: []inventory.Server{
			{
				Type:        "web",
				NodeVersion: "18",
				GitRepo:     "https://github.com/user/repo",
				GitBranch:   "develop",
				AppPort:     3000,
			},
		},
	}

	data, err := gen.GenerateGroupVarsYAML(env)
	if err != nil {
		t.Fatalf("Failed to generate group_vars: %v", err)
	}

	// Verify valid YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Generated invalid YAML: %v\nYAML content:\n%s", err, string(data))
	}

	// Verify required fields
	// Note: Generator uses env.Name as fallback for app_name
	requiredFields := map[string]interface{}{
		"deploy_user":    "deploy",
		"app_repo":       "https://github.com/user/repo",
		"app_branch":     "develop",
		"app_port":       3000,
		"nodejs_version": "18",
	}

	for field, expectedValue := range requiredFields {
		actualValue, exists := result[field]
		if !exists {
			t.Errorf("Missing required field: %s", field)
			continue
		}
		if actualValue != expectedValue {
			t.Errorf("Field %s: expected %v, got %v", field, expectedValue, actualValue)
		}
	}

	// app_name defaults to env.Name if Config.AppName empty
	if appName, ok := result["app_name"]; !ok {
		t.Error("Missing app_name field")
	} else if appName != "staging" { // env.Name used as fallback
		t.Errorf("Expected app_name 'staging' (env name), got %v", appName)
	}

	// Verify derived fields
	if appDir, ok := result["app_dir"]; !ok || appDir != "/var/www/staging" {
		t.Errorf("Expected app_dir '/var/www/staging', got %v", appDir)
	}
}

func TestGenerateHostVarsYAML_WebServer(t *testing.T) {
	gen := inventory.NewGenerator()

	server := inventory.Server{
		Name:        "web1",
		Type:        "web",
		AppPort:     3000,
		GitRepo:     "https://github.com/user/repo",
		GitBranch:   "main",
		NodeVersion: "20",
	}

	data, err := gen.GenerateHostVarsYAML(server)
	if err != nil {
		t.Fatalf("Failed to generate host_vars: %v", err)
	}

	if data == nil {
		t.Fatal("Expected host_vars data for web server, got nil")
	}

	// Verify valid YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Generated invalid YAML: %v", err)
	}

	// Verify web-specific fields
	expectedFields := map[string]interface{}{
		"app_port":       3000,
		"app_repo":       "https://github.com/user/repo",
		"app_branch":     "main",
		"nodejs_version": "20",
	}

	for field, expected := range expectedFields {
		if actual, exists := result[field]; !exists {
			t.Errorf("Missing field: %s", field)
		} else if actual != expected {
			t.Errorf("Field %s: expected %v, got %v", field, expected, actual)
		}
	}
}

func TestGenerateHostVarsYAML_NonWebServer(t *testing.T) {
	gen := inventory.NewGenerator()

	testCases := []struct {
		name       string
		serverType string
	}{
		{"db server", "db"},
		{"monitoring server", "monitoring"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := inventory.Server{
				Name: "server1",
				Type: tc.serverType,
			}

			data, err := gen.GenerateHostVarsYAML(server)
			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", tc.name, err)
			}

			// Non-web servers should return nil (no host_vars needed)
			if data != nil {
				t.Errorf("Expected nil for %s host_vars, got data", tc.name)
			}
		})
	}
}

func TestGenerateHostsYAML_EmptyServerList(t *testing.T) {
	gen := inventory.NewGenerator()

	env := inventory.Environment{
		Name:    "empty",
		Servers: []inventory.Server{},
	}

	data, err := gen.GenerateHostsYAML(env)
	if err != nil {
		t.Fatalf("Failed to generate hosts.yml: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Generated invalid YAML: %v", err)
	}

	// Should still have valid structure, just no groups
	all := result["all"].(map[string]interface{})
	children := all["children"].(map[string]interface{})

	if len(children) != 0 {
		t.Errorf("Expected no groups for empty server list, got %d", len(children))
	}
}

func TestGenerateHostsYAML_ServerWithoutAppPort(t *testing.T) {
	gen := inventory.NewGenerator()

	env := inventory.Environment{
		Name: "test",
		Servers: []inventory.Server{
			{
				Name:          "web1",
				Type:          "web",
				IP:            "192.168.1.10",
				Port:          22,
				SSHUser:       "deploy",
				SSHKeyPath:    "/key",
				AppPort:       0, // No app port
				AnsibleBecome: true,
			},
		},
	}

	data, err := gen.GenerateHostsYAML(env)
	if err != nil {
		t.Fatalf("Failed to generate hosts.yml: %v", err)
	}

	var result map[string]interface{}
	yaml.Unmarshal(data, &result)

	children := result["all"].(map[string]interface{})["children"].(map[string]interface{})
	hosts := children["webservers"].(map[string]interface{})["hosts"].(map[string]interface{})
	web1 := hosts["web1"].(map[string]interface{})

	// app_port should not be present if set to 0
	if _, exists := web1["app_port"]; exists {
		t.Error("app_port should not be present when set to 0")
	}
}
