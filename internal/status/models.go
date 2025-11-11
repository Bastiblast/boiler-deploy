package status

import "time"

type ServerState string

const (
	StateUnknown      ServerState = "unknown"
	StateNotReady     ServerState = "not_ready"
	StateReady        ServerState = "ready"
	StateProvisioning ServerState = "provisioning"
	StateProvisioned  ServerState = "provisioned"
	StateDeploying    ServerState = "deploying"
	StateDeployed     ServerState = "deployed"
	StateFailed       ServerState = "failed"
	StateVerifying    ServerState = "verifying"
)

type ActionType string

const (
	ActionValidate  ActionType = "validate"
	ActionProvision ActionType = "provision"
	ActionDeploy    ActionType = "deploy"
	ActionCheck     ActionType = "check"
)

type ServerStatus struct {
	Name          string      `json:"name"`
	State         ServerState `json:"state"`
	LastAction    ActionType  `json:"last_action,omitempty"`
	LastUpdate    time.Time   `json:"last_update"`
	ErrorMessage  string      `json:"error_message,omitempty"`
	ReadyChecks   ReadyChecks `json:"ready_checks"`
}

type ReadyChecks struct {
	IPValid       bool `json:"ip_valid"`
	SSHKeyExists  bool `json:"ssh_key_exists"`
	PortValid     bool `json:"port_valid"`
	AllFieldsFilled bool `json:"all_fields_filled"`
}

func (r ReadyChecks) IsReady() bool {
	return r.IPValid && r.SSHKeyExists && r.PortValid && r.AllFieldsFilled
}

type QueuedAction struct {
	ID          string     `json:"id"`
	ServerName  string     `json:"server_name"`
	Action      ActionType `json:"action"`
	Priority    int        `json:"priority"`
	QueuedAt    time.Time  `json:"queued_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
}

type ExecutionLog struct {
	ServerName string     `json:"server_name"`
	Action     ActionType `json:"action"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Status     string     `json:"status"`
	LogFile    string     `json:"log_file"`
}
