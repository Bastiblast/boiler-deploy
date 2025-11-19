package ansible

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/status"
)

type Orchestrator struct {
	statusMgr      *status.Manager
	queue          *Queue
	executor       *Executor
	scriptExecutor *ScriptExecutor
	environment    string
	mu             sync.RWMutex
	running        bool
	stopChan       chan struct{}
	progressCb     func(serverName, message string)
	useScript      bool // Use deploy.sh script instead of ansible directly
}

func NewOrchestrator(environment string, statusMgr *status.Manager) (*Orchestrator, error) {
	queue, err := NewQueue(environment)
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		statusMgr:      statusMgr,
		queue:          queue,
		executor:       NewExecutor(environment),
		scriptExecutor: NewScriptExecutor(environment),
		environment:    environment,
		stopChan:       make(chan struct{}),
		useScript:      false, // Use native Go ansible execution
	}, nil
}

func (o *Orchestrator) SetProgressCallback(cb func(serverName, message string)) {
	o.progressCb = cb
}

func (o *Orchestrator) ValidateInventory(servers []*inventory.Server) {
	for _, server := range servers {
		checks := o.statusMgr.ValidateServer(server)
		o.statusMgr.UpdateReadyChecks(server.Name, checks)
	}
}

func (o *Orchestrator) QueueProvision(serverNames []string, priority int) {
	o.QueueProvisionWithTags(serverNames, priority, "")
}

func (o *Orchestrator) QueueProvisionWithTags(serverNames []string, priority int, tags string) {
	log.Printf("[ORCHESTRATOR] QueueProvisionWithTags called with %d servers: %v, tags: %s", len(serverNames), serverNames, tags)
	for _, name := range serverNames {
		log.Printf("[ORCHESTRATOR] Adding provision action for server: %s with tags: %s", name, tags)
		item := o.queue.Add(name, status.ActionProvision, priority)
		item.Tags = tags
	}
	log.Printf("[ORCHESTRATOR] Queue size after adding provisions: %d", o.GetQueueSize())
}

func (o *Orchestrator) QueueDeploy(serverNames []string, priority int) {
	o.QueueDeployWithTags(serverNames, priority, "")
}

func (o *Orchestrator) QueueDeployWithTags(serverNames []string, priority int, tags string) {
	log.Printf("[ORCHESTRATOR] QueueDeploy called with %d servers: %v", len(serverNames), serverNames)
	for _, name := range serverNames {
		log.Printf("[ORCHESTRATOR] Adding deploy action for server: %s", name)
		item := o.queue.Add(name, status.ActionDeploy, priority)
		item.Tags = tags
	}
	log.Printf("[ORCHESTRATOR] Queue size after adding deploys: %d", o.GetQueueSize())
}

func (o *Orchestrator) QueueCheck(serverNames []string, priority int) {
	log.Printf("[ORCHESTRATOR] QueueCheck called with %d servers: %v", len(serverNames), serverNames)
	for _, name := range serverNames {
		log.Printf("[ORCHESTRATOR] Adding check action for server: %s", name)
		o.queue.Add(name, status.ActionCheck, priority)
	}
	log.Printf("[ORCHESTRATOR] Queue size after adding checks: %d", o.GetQueueSize())
}

func (o *Orchestrator) Start(servers []*inventory.Server) {
	o.mu.Lock()
	if o.running {
		log.Println("[ORCHESTRATOR] Already running, skipping start")
		o.mu.Unlock()
		return
	}
	
	// Reset stopChan if needed
	o.stopChan = make(chan struct{})
	o.running = true
	o.mu.Unlock()

	log.Println("[ORCHESTRATOR] Starting processQueue goroutine")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ORCHESTRATOR] PANIC in processQueue: %v", r)
			}
		}()
		o.processQueue(servers)
	}()
}

func (o *Orchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.running {
		close(o.stopChan)
		o.running = false
	}
}

func (o *Orchestrator) processQueue(servers []*inventory.Server) {
	log.Println("[ORCHESTRATOR] processQueue started")
	
	for {
		// Check for stop signal (non-blocking)
		select {
		case <-o.stopChan:
			log.Println("[ORCHESTRATOR] processQueue received stop signal")
			return
		default:
			// Continue processing
		}
		
		action := o.queue.Next()
		if action == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		log.Printf("[ORCHESTRATOR] Processing action: %s for server %s", action.Action, action.ServerName)
		o.executeAction(action, servers)
		o.queue.Complete()
		log.Printf("[ORCHESTRATOR] Completed action: %s for server %s", action.Action, action.ServerName)
	}
}

func (o *Orchestrator) executeAction(action *status.QueuedAction, servers []*inventory.Server) {
	server := o.findServer(action.ServerName, servers)
	if server == nil {
		o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, "Server not found")
		return
	}

	progressChan := make(chan string, 100)
	go func() {
		for msg := range progressChan {
			if o.progressCb != nil {
				o.progressCb(action.ServerName, msg)
			}
		}
	}()

	var result *ExecutionResult
	var err error

	switch action.Action {
	case status.ActionProvision:
		o.statusMgr.UpdateStatus(action.ServerName, status.StateProvisioning, action.Action, "Provisioning server...")
		log.Printf("[ORCHESTRATOR] Running provision for %s with tags: %s", action.ServerName, action.Tags)
		
		if o.useScript {
			log.Printf("[ORCHESTRATOR] Using deploy.sh for provision")
			result, err = o.scriptExecutor.RunAction("provision", action.ServerName, progressChan)
		} else {
			log.Printf("[ORCHESTRATOR] Using ansible-playbook directly with tags: %s", action.Tags)
			if action.Tags != "" {
				result, err = o.executor.ProvisionWithTags(action.ServerName, action.Tags, progressChan)
			} else {
				result, err = o.executor.Provision(action.ServerName, progressChan)
			}
		}
		close(progressChan)

		if err != nil || !result.Success {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, result.ErrorMessage)
		} else {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateProvisioned, action.Action, "")
		}

	case status.ActionDeploy:
		currentStatus := o.statusMgr.GetStatus(action.ServerName)
		if currentStatus.State != status.StateProvisioned && currentStatus.State != status.StateDeployed {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, "Server must be provisioned first")
			close(progressChan)
			return
		}

		o.statusMgr.UpdateStatus(action.ServerName, status.StateDeploying, action.Action, "Deploying application...")
		log.Printf("[ORCHESTRATOR] Running deploy for %s with tags: %s", action.ServerName, action.Tags)
		
		if o.useScript {
			log.Printf("[ORCHESTRATOR] Using deploy.sh for deploy")
			result, err = o.scriptExecutor.RunAction("deploy", action.ServerName, progressChan)
		} else {
			log.Printf("[ORCHESTRATOR] Using ansible-playbook directly with tags: %s", action.Tags)
			if action.Tags != "" {
				result, err = o.executor.DeployWithTags(action.ServerName, action.Tags, progressChan)
			} else {
				result, err = o.executor.Deploy(action.ServerName, progressChan)
			}
		}
		close(progressChan)

		if err != nil || !result.Success {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, result.ErrorMessage)
		} else {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "Checking...")
			
			if err := o.executor.HealthCheck(server.IP, 80); err != nil {
				o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
					fmt.Sprintf("Health check failed: %v", err))
			} else {
				o.statusMgr.UpdateStatus(action.ServerName, status.StateDeployed, action.Action, "")
			}
		}

	case status.ActionCheck:
		log.Printf("[ORCHESTRATOR] Starting validation check for %s", action.ServerName)
		
		o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "Validating configuration...")
		
		// Step 1: Validate basic configuration
		checks := o.statusMgr.ValidateServer(server)
		if !checks.AllFieldsFilled {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
				"Configuration incomplete: missing required fields")
			close(progressChan)
			return
		}
		
		if !checks.IPValid {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
				"Invalid IP address format")
			close(progressChan)
			return
		}
		
		if !checks.PortValid {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
				"Invalid SSH port (must be 1-65535)")
			close(progressChan)
			return
		}
		
		if !checks.SSHKeyExists {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
				fmt.Sprintf("SSH key not found at: %s", server.SSHKeyPath))
			close(progressChan)
			return
		}
		
		// Step 2: Test SSH connection
		o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "Testing SSH connection...")
		log.Printf("[ORCHESTRATOR] Testing SSH connection to %s:%d", server.IP, server.Port)
		
		sshTest := o.executor.TestSSH(server.IP, server.Port, "root", server.SSHKeyPath)
		if !sshTest.Success {
			log.Printf("[ORCHESTRATOR] SSH test failed for %s: %s", action.ServerName, sshTest.Message)
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
				fmt.Sprintf("SSH connection failed: %s", sshTest.Message))
			close(progressChan)
			return
		}
		
		log.Printf("[ORCHESTRATOR] SSH test passed for %s", action.ServerName)
		
		// Step 3: Check if deployed (try HTTP health check)
		currentStatus := o.statusMgr.GetStatus(action.ServerName)
		if currentStatus.State == status.StateDeployed {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "Checking application...")
			
			checkPort := 80
			log.Printf("[ORCHESTRATOR] Checking HTTP on %s:%d", server.IP, checkPort)
			
			if err := o.executor.HealthCheck(server.IP, checkPort); err != nil {
				log.Printf("[ORCHESTRATOR] Application health check failed for %s: %v", action.ServerName, err)
				o.statusMgr.UpdateStatus(action.ServerName, status.StateProvisioned, action.Action, "")
			} else {
				log.Printf("[ORCHESTRATOR] Application is responding for %s", action.ServerName)
				o.statusMgr.UpdateStatus(action.ServerName, status.StateDeployed, action.Action, "")
			}
		} else {
			// Server is validated but not deployed yet
			log.Printf("[ORCHESTRATOR] Validation passed for %s (not deployed yet)", action.ServerName)
			o.statusMgr.UpdateStatus(action.ServerName, status.StateReady, action.Action, "")
		}
		
		close(progressChan)
	}
}

func (o *Orchestrator) findServer(name string, servers []*inventory.Server) *inventory.Server {
	for _, s := range servers {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (o *Orchestrator) GetQueueSize() int {
	return o.queue.Size()
}

func (o *Orchestrator) GetQueuedActions() []*status.QueuedAction {
	return o.queue.GetAll()
}

func (o *Orchestrator) ClearQueue() {
	o.queue.Clear()
}

func (o *Orchestrator) IsRunning() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.running
}
