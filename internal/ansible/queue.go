package ansible

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/status"
	"github.com/google/uuid"
)

type Queue struct {
	mu          sync.RWMutex
	actions     []*status.QueuedAction
	current     *status.QueuedAction
	environment string
	queueFile   string
	stopChan    chan struct{}
	stopped     bool
}

func NewQueue(environment string) (*Queue, error) {
	queueDir := filepath.Join("inventory", environment, ".queue")
	if err := os.MkdirAll(queueDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %w", err)
	}

	q := &Queue{
		actions:     make([]*status.QueuedAction, 0),
		environment: environment,
		queueFile:   filepath.Join(queueDir, "actions.json"),
		stopChan:    make(chan struct{}),
	}

	if err := q.Load(); err != nil {
		return q, nil
	}

	return q, nil
}

func (q *Queue) Load() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	data, err := os.ReadFile(q.queueFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &q.actions); err != nil {
		return err
	}

	return nil
}

func (q *Queue) Save() error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.save()
}

func (q *Queue) save() error {
	data, err := json.MarshalIndent(q.actions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(q.queueFile, data, 0644)
}

func (q *Queue) Add(serverName string, action status.ActionType, priority int) *status.QueuedAction {
	q.mu.Lock()
	defer q.mu.Unlock()

	id := uuid.New().String()
	queuedAction := &status.QueuedAction{
		ID:         id,
		ServerName: serverName,
		Action:     action,
		Priority:   priority,
		QueuedAt:   time.Now(),
		Tags:       "",
	}

	log.Printf("[QUEUE] Adding action: %s for server %s (priority: %d, id: %s)", action, serverName, priority, id)
	q.actions = append(q.actions, queuedAction)
	q.sort()
	q.save()
	log.Printf("[QUEUE] Action added, queue size now: %d", len(q.actions))

	return queuedAction
}

func (q *Queue) sort() {
	for i := 0; i < len(q.actions)-1; i++ {
		for j := i + 1; j < len(q.actions); j++ {
			if q.actions[i].Priority < q.actions[j].Priority {
				q.actions[i], q.actions[j] = q.actions[j], q.actions[i]
			}
		}
	}
}

func (q *Queue) Next() *status.QueuedAction {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.actions) == 0 {
		return nil
	}

	action := q.actions[0]
	now := time.Now()
	action.StartedAt = &now
	q.current = action
	
	log.Printf("[QUEUE] Next action: %s for server %s (id: %s)", action.Action, action.ServerName, action.ID)

	return action
}

func (q *Queue) NextBatch(count int) []*status.QueuedAction {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.actions) == 0 {
		return nil
	}

	batchSize := count
	if batchSize > len(q.actions) {
		batchSize = len(q.actions)
	}

	batch := make([]*status.QueuedAction, batchSize)
	now := time.Now()
	
	for i := 0; i < batchSize; i++ {
		q.actions[i].StartedAt = &now
		batch[i] = q.actions[i]
	}
	
	log.Printf("[QUEUE] Next batch: %d actions", batchSize)
	return batch
}

func (q *Queue) Complete() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.actions) > 0 {
		completedAction := q.actions[0]
		log.Printf("[QUEUE] Completing action: %s for server %s", completedAction.Action, completedAction.ServerName)
		q.actions = q.actions[1:]
	}
	q.current = nil
	q.save()
	log.Printf("[QUEUE] Action completed, queue size now: %d", len(q.actions))
}

func (q *Queue) CompleteByID(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, action := range q.actions {
		if action.ID == id {
			log.Printf("[QUEUE] Completing action by ID: %s for server %s", action.Action, action.ServerName)
			q.actions = append(q.actions[:i], q.actions[i+1:]...)
			q.save()
			log.Printf("[QUEUE] Action completed, queue size now: %d", len(q.actions))
			return
		}
	}
	
	log.Printf("[QUEUE] Warning: action ID %s not found for completion", id)
}

func (q *Queue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.stopped {
		close(q.stopChan)
		q.stopped = true
	}
}

func (q *Queue) ShouldStop() bool {
	select {
	case <-q.stopChan:
		return true
	default:
		return false
	}
}

func (q *Queue) Resume() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		q.stopChan = make(chan struct{})
		q.stopped = false
	}
}

func (q *Queue) GetCurrent() *status.QueuedAction {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.current
}

func (q *Queue) GetAll() []*status.QueuedAction {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]*status.QueuedAction, len(q.actions))
	copy(result, q.actions)
	return result
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.actions = make([]*status.QueuedAction, 0)
	q.save()
}

func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.actions)
}
