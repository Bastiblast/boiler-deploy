package ansible

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/ssh"
	"github.com/bastiblast/boiler-deploy/internal/status"
)

type Orchestrator struct {
	statusMgr           *status.Manager
	queue               *Queue
	executor            *Executor
	scriptExecutor      *ScriptExecutor
	environment         string
	mu                  sync.RWMutex
	running             bool
	stopChan            chan struct{}
	ctx                 context.Context
	cancel              context.CancelFunc
	progressCb          func(serverName, message string)
	deploySuccessCb     func(serverName, serverIP string) // Callback when deployment succeeds
	useScript           bool // Use deploy.sh script instead of ansible directly
	healthCheckEnabled  bool // Enable/disable health checks
	skipHealthCheck     bool // Skip health check for current deployment
	maxWorkers          int  // Number of parallel workers (0 = sequential)
	activeWorkers       int  // Current number of active workers
	workersMu           sync.Mutex // Mutex for activeWorkers counter
}

func NewOrchestrator(environment string, statusMgr *status.Manager) (*Orchestrator, error) {
	queue, err := NewQueue(environment)
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		statusMgr:          statusMgr,
		queue:              queue,
		executor:           NewExecutor(environment),
		scriptExecutor:     NewScriptExecutor(environment),
		environment:        environment,
		stopChan:           make(chan struct{}),
		useScript:          false, // Use native Go ansible execution
		healthCheckEnabled: true,  // Enable by default
		skipHealthCheck:    false,
		maxWorkers:         0,     // Sequential by default
		activeWorkers:      0,
	}, nil
}

func (o *Orchestrator) SetHealthCheckEnabled(enabled bool) {
	o.healthCheckEnabled = enabled
}

func (o *Orchestrator) SetMaxWorkers(workers int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if workers < 0 {
		workers = 0
	}
	o.maxWorkers = workers
	log.Printf("[ORCHESTRATOR] Max workers set to %d (0=sequential, >0=parallel)", workers)
}

func (o *Orchestrator) SkipNextHealthCheck() {
	o.skipHealthCheck = true
}

func (o *Orchestrator) SetProgressCallback(cb func(serverName, message string)) {
	o.progressCb = cb
}

func (o *Orchestrator) SetDeploySuccessCallback(cb func(serverName, serverIP string)) {
	o.deploySuccessCb = cb
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
	
	// Reset stopChan and create new context
	o.stopChan = make(chan struct{})
	o.ctx, o.cancel = context.WithCancel(context.Background())
	o.running = true
	o.mu.Unlock()

	log.Println("[ORCHESTRATOR] Starting processQueue goroutine with context")
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
		log.Println("[ORCHESTRATOR] Stopping: cancelling context and closing channels")
		if o.cancel != nil {
			o.cancel() // Cancel all running operations
		}
		close(o.stopChan)
		o.running = false
	}
}

func (o *Orchestrator) processQueue(servers []*inventory.Server) {
	log.Println("[ORCHESTRATOR] processQueue started")
	
	// Check if parallel or sequential mode
	o.mu.RLock()
	workers := o.maxWorkers
	o.mu.RUnlock()
	
	if workers <= 0 {
		// Sequential mode (original behavior)
		log.Println("[ORCHESTRATOR] Running in SEQUENTIAL mode")
		o.processQueueSequential(servers)
	} else {
		// Parallel mode with worker pool
		log.Printf("[ORCHESTRATOR] Running in PARALLEL mode with %d workers", workers)
		o.processQueueParallel(servers, workers)
	}
}

func (o *Orchestrator) processQueueSequential(servers []*inventory.Server) {
	for {
		select {
		case <-o.stopChan:
			log.Println("[ORCHESTRATOR] processQueueSequential received stop signal")
			return
		default:
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

func (o *Orchestrator) processQueueParallel(servers []*inventory.Server, maxWorkers int) {
	var wg sync.WaitGroup
	actionChan := make(chan *status.QueuedAction, maxWorkers)
	
	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("[ORCHESTRATOR] Worker %d started", workerID)
			
			for action := range actionChan {
				o.workersMu.Lock()
				o.activeWorkers++
				currentActive := o.activeWorkers
				o.workersMu.Unlock()
				
				log.Printf("[ORCHESTRATOR] Worker %d processing: %s for server %s (active: %d/%d)", 
					workerID, action.Action, action.ServerName, currentActive, maxWorkers)
				
				o.executeAction(action, servers)
				
				o.workersMu.Lock()
				o.activeWorkers--
				o.workersMu.Unlock()
				
				log.Printf("[ORCHESTRATOR] Worker %d completed: %s for server %s", 
					workerID, action.Action, action.ServerName)
			}
			
			log.Printf("[ORCHESTRATOR] Worker %d stopped", workerID)
		}(i)
	}
	
	// Feed actions to workers
	for {
		select {
		case <-o.stopChan:
			log.Println("[ORCHESTRATOR] processQueueParallel received stop signal")
			close(actionChan)
			wg.Wait()
			return
		default:
		}
		
		action := o.queue.Next()
		if action == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		
		// Remove from queue BEFORE sending to worker to prevent duplicate processing
		o.queue.Complete()
		
		log.Printf("[ORCHESTRATOR] Queueing action for workers: %s for server %s", action.Action, action.ServerName)
		actionChan <- action
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
			log.Printf("[ORCHESTRATOR] Using ansible-playbook directly with context and tags: %s", action.Tags)
			// Use context for cancellation support
			result, err = o.executor.ProvisionWithContext(o.ctx, action.ServerName, action.Tags, progressChan)
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
			log.Printf("[ORCHESTRATOR] Using ansible-playbook directly with context and tags: %s", action.Tags)
			// Use context for cancellation support
			result, err = o.executor.DeployWithContext(o.ctx, action.ServerName, action.Tags, progressChan)
		}
		close(progressChan)

		if err != nil || !result.Success {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, result.ErrorMessage)
		} else {
			// Check if health check should be performed
			performHealthCheck := o.healthCheckEnabled && !o.skipHealthCheck
			o.skipHealthCheck = false // Reset for next deployment
			
			if !performHealthCheck {
				log.Printf("[ORCHESTRATOR] Health check skipped (disabled or skip requested)")
				o.statusMgr.UpdateStatus(action.ServerName, status.StateDeployed, action.Action, "")
				
				// Trigger deploy success callback
				if o.deploySuccessCb != nil {
					o.deploySuccessCb(action.ServerName, server.IP)
				}
			} else {
				o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "Checking...")
				
				// Determine if we need remote health check (SSH-based)
				// Use remote check if:
				// - Server IP is 127.0.0.1 (localhost/Docker container)
				// - Server has SSH credentials configured
				useRemoteCheck := server.IP == "127.0.0.1" && server.Port > 0 && server.SSHKeyPath != ""
				
				var healthCheckErr error
				healthCheckPassed := false
				
				if useRemoteCheck {
					// Remote health check via SSH (for Docker containers or localhost servers)
					log.Printf("[ORCHESTRATOR] Using remote health check via SSH for %s (port %d)", action.ServerName, server.AppPort)
					
					if server.AppPort > 0 {
						if err := o.executor.HealthCheckRemote(server.IP, server.Port, server.SSHUser, server.SSHKeyPath, server.AppPort); err == nil {
							log.Printf("[ORCHESTRATOR] Remote health check passed on port %d", server.AppPort)
							healthCheckPassed = true
						} else {
							healthCheckErr = err
							log.Printf("[ORCHESTRATOR] Remote health check failed: %v", err)
						}
					} else {
						healthCheckErr = fmt.Errorf("app_port not configured for server")
						log.Printf("[ORCHESTRATOR] Cannot perform health check: %v", healthCheckErr)
					}
				} else {
					// Standard health check (direct HTTP from orchestrator to server IP)
					log.Printf("[ORCHESTRATOR] Using direct health check for %s:%d", server.IP, server.AppPort)
					
					// Try health check on multiple ports: 80 (nginx), 443 (https), app port
					ports := []int{80}
					if server.AppPort > 0 && server.AppPort != 80 {
						ports = append(ports, server.AppPort)
					}
					
					for _, port := range ports {
						log.Printf("[ORCHESTRATOR] Trying health check on %s:%d", server.IP, port)
						if err := o.executor.HealthCheck(server.IP, port); err == nil {
							log.Printf("[ORCHESTRATOR] Health check passed on port %d", port)
							healthCheckPassed = true
							break
						} else {
							healthCheckErr = err
							log.Printf("[ORCHESTRATOR] Health check failed on port %d: %v", port, err)
						}
					}
				}
				
				if !healthCheckPassed {
					errMsg := fmt.Sprintf("Health check failed: %v", healthCheckErr)
					log.Printf("[ORCHESTRATOR] %s", errMsg)
					log.Printf("[ORCHESTRATOR] Tip: Check if application is running on server, nginx is configured, and ports are open")
					o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, errMsg)
					
					// Trigger callback even on health check failure (allow browser access attempt)
					if o.deploySuccessCb != nil {
						log.Printf("[ORCHESTRATOR] Triggering deploy success callback despite health check failure (app may still be accessible)")
						o.deploySuccessCb(action.ServerName, server.IP)
					}
				} else {
					o.statusMgr.UpdateStatus(action.ServerName, status.StateDeployed, action.Action, "")
					
					// Trigger deploy success callback
					if o.deploySuccessCb != nil {
						o.deploySuccessCb(action.ServerName, server.IP)
					}
				}
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
		
		// Step 3: Detect actual server state (provisioned/deployed)
		o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "Detecting server state...")
		log.Printf("[ORCHESTRATOR] Detecting state for %s using State Detector", action.ServerName)
		
		detector := ssh.NewStateDetector()
		stateResult := detector.DetectState(*server)
		
		log.Printf("[ORCHESTRATOR] State detected for %s: %s - %s", 
			action.ServerName, stateResult.State, stateResult.Message)
		
		// Log detailed checks for debugging
		if stateResult.ProvisioningChecks.AllProvisioned {
			log.Printf("[ORCHESTRATOR] %s: Provisioning checks passed (Node: %v, Nginx: %v, NVM: %v, AppDir: %v)",
				action.ServerName,
				stateResult.ProvisioningChecks.NodeInstalled,
				stateResult.ProvisioningChecks.NginxInstalled,
				stateResult.ProvisioningChecks.NVMInstalled,
				stateResult.ProvisioningChecks.AppDirExists)
		}
		
		if stateResult.DeploymentChecks.AllDeployed {
			log.Printf("[ORCHESTRATOR] %s: Deployment checks passed (PM2: %v, App: %v, Symlink: %v)",
				action.ServerName,
				stateResult.DeploymentChecks.PM2Running,
				stateResult.DeploymentChecks.AppResponding,
				stateResult.DeploymentChecks.CurrentSymlink)
		}
		
		// Update status with detected state
		o.statusMgr.UpdateStatus(action.ServerName, stateResult.State, action.Action, stateResult.Message)
		
		// Save status to file
		if err := o.statusMgr.Save(); err != nil {
			log.Printf("[ORCHESTRATOR] Warning: Could not save status for %s: %v", action.ServerName, err)
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
