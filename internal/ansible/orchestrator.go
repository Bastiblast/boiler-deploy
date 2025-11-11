package ansible

import (
	"fmt"
	"sync"

	"github.com/bastiblast/boiler-deploy/internal/inventory"
	"github.com/bastiblast/boiler-deploy/internal/status"
)

type Orchestrator struct {
	statusMgr   *status.Manager
	queue       *Queue
	executor    *Executor
	environment string
	mu          sync.RWMutex
	running     bool
	stopChan    chan struct{}
	progressCb  func(serverName, message string)
}

func NewOrchestrator(environment string, statusMgr *status.Manager) (*Orchestrator, error) {
	queue, err := NewQueue(environment)
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		statusMgr:   statusMgr,
		queue:       queue,
		executor:    NewExecutor(environment),
		environment: environment,
		stopChan:    make(chan struct{}),
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
	for _, name := range serverNames {
		o.queue.Add(name, status.ActionProvision, priority)
	}
}

func (o *Orchestrator) QueueDeploy(serverNames []string, priority int) {
	for _, name := range serverNames {
		o.queue.Add(name, status.ActionDeploy, priority)
	}
}

func (o *Orchestrator) QueueCheck(serverNames []string, priority int) {
	for _, name := range serverNames {
		o.queue.Add(name, status.ActionCheck, priority)
	}
}

func (o *Orchestrator) Start(servers []*inventory.Server) {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return
	}
	o.running = true
	o.mu.Unlock()

	go o.processQueue(servers)
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
	for {
		select {
		case <-o.stopChan:
			return
		default:
			action := o.queue.Next()
			if action == nil {
				continue
			}

			o.executeAction(action, servers)
			o.queue.Complete()
		}
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
		o.statusMgr.UpdateStatus(action.ServerName, status.StateProvisioning, action.Action, "")
		result, err = o.executor.Provision(action.ServerName, progressChan)
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

		o.statusMgr.UpdateStatus(action.ServerName, status.StateDeploying, action.Action, "")
		result, err = o.executor.Deploy(action.ServerName, progressChan)
		close(progressChan)

		if err != nil || !result.Success {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, result.ErrorMessage)
		} else {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "")
			
			if err := o.executor.HealthCheck(server.IP, server.AppPort); err != nil {
				o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
					fmt.Sprintf("Health check failed: %v", err))
			} else {
				o.statusMgr.UpdateStatus(action.ServerName, status.StateDeployed, action.Action, "")
			}
		}

	case status.ActionCheck:
		o.statusMgr.UpdateStatus(action.ServerName, status.StateVerifying, action.Action, "")
		close(progressChan)

		if err := o.executor.HealthCheck(server.IP, server.AppPort); err != nil {
			o.statusMgr.UpdateStatus(action.ServerName, status.StateFailed, action.Action, 
				fmt.Sprintf("Health check failed: %v", err))
		} else {
			currentStatus := o.statusMgr.GetStatus(action.ServerName)
			if currentStatus.State == status.StateDeploying || currentStatus.State == status.StateVerifying {
				o.statusMgr.UpdateStatus(action.ServerName, status.StateDeployed, action.Action, "")
			}
		}
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
