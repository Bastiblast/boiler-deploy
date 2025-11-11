package inventory

// Environment represents a deployment environment
type Environment struct {
	Name     string   `yaml:"name"`
	Services Services `yaml:"services"`
	Config   Config   `yaml:"config"`
	Servers  []Server `yaml:"servers"`
}

// Services represents enabled services
type Services struct {
	Web        bool `yaml:"web"`
	Database   bool `yaml:"database"`
	Monitoring bool `yaml:"monitoring"`
}

// Config represents environment configuration
type Config struct {
	AppName       string `yaml:"app_name"`
	AppRepo       string `yaml:"app_repo"`
	AppBranch     string `yaml:"app_branch"`
	NodeJSVersion string `yaml:"nodejs_version"`
	AppPort       string `yaml:"app_port"`
	DeployUser    string `yaml:"deploy_user"`
	Timezone      string `yaml:"timezone"`
}

// Server represents a single server
type Server struct {
	Name          string `yaml:"name"`
	IP            string `yaml:"ip"`
	Port          int    `yaml:"port"`
	SSHUser       string `yaml:"ssh_user"`
	SSHKeyPath    string `yaml:"ssh_key_path"`
	Type          string `yaml:"type"` // web, db, monitoring
	AppPort       int    `yaml:"app_port,omitempty"`
	AnsibleBecome bool   `yaml:"ansible_become"`
}

// ValidationResult holds validation results
type ValidationResult struct {
	Server  string
	Valid   bool
	Message string
}
