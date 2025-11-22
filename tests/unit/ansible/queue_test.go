package ansible_test

import (
	"os"
	"testing"

	"github.com/bastiblast/boiler-deploy/internal/ansible"
	"github.com/bastiblast/boiler-deploy/internal/status"
)

func TestQueueAddAndPriority(t *testing.T) {
	// Setup test environment
	testEnv := "test-queue"
	defer os.RemoveAll("inventory/" + testEnv)

	q, err := ansible.NewQueue(testEnv)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	// Add actions with different priorities
	q.Add("server1", status.ActionProvision, 1)
	q.Add("server2", status.ActionProvision, 10)
	q.Add("server3", status.ActionDeploy, 5)

	// Verify size
	if size := q.Size(); size != 3 {
		t.Errorf("Expected queue size 3, got %d", size)
	}

	// Verify priority ordering (highest first)
	next := q.Next()
	if next == nil {
		t.Fatal("Expected next action, got nil")
	}
	if next.ServerName != "server2" {
		t.Errorf("Expected server2 (priority 10), got %s (priority %d)", 
			next.ServerName, next.Priority)
	}

	q.Complete()

	// Next should be server3 (priority 5)
	next = q.Next()
	if next.ServerName != "server3" {
		t.Errorf("Expected server3 (priority 5), got %s", next.ServerName)
	}
}

func TestQueuePersistence(t *testing.T) {
	testEnv := "test-persistence"
	defer os.RemoveAll("inventory/" + testEnv)

	// Create queue and add actions
	q1, err := ansible.NewQueue(testEnv)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	q1.Add("server1", status.ActionProvision, 1)
	q1.Add("server2", status.ActionDeploy, 2)

	// Create new queue instance (simulates restart)
	q2, err := ansible.NewQueue(testEnv)
	if err != nil {
		t.Fatalf("Failed to load queue: %v", err)
	}

	// Verify persistence
	if size := q2.Size(); size != 2 {
		t.Errorf("Expected queue size 2 after reload, got %d", size)
	}

	actions := q2.GetAll()
	if len(actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(actions))
	}

	// Verify order maintained
	if actions[0].ServerName != "server2" || actions[0].Priority != 2 {
		t.Errorf("Expected server2 with priority 2, got %s with priority %d",
			actions[0].ServerName, actions[0].Priority)
	}
}

func TestQueueClear(t *testing.T) {
	testEnv := "test-clear"
	defer os.RemoveAll("inventory/" + testEnv)

	q, err := ansible.NewQueue(testEnv)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	q.Add("server1", status.ActionProvision, 1)
	q.Add("server2", status.ActionProvision, 2)

	if size := q.Size(); size != 2 {
		t.Errorf("Expected queue size 2, got %d", size)
	}

	q.Clear()

	if size := q.Size(); size != 0 {
		t.Errorf("Expected queue size 0 after clear, got %d", size)
	}
}

func TestQueueComplete(t *testing.T) {
	testEnv := "test-complete"
	defer os.RemoveAll("inventory/" + testEnv)

	q, err := ansible.NewQueue(testEnv)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	q.Add("server1", status.ActionProvision, 1)
	q.Add("server2", status.ActionProvision, 2)

	// Get first action
	action := q.Next()
	if action == nil {
		t.Fatal("Expected action, got nil")
	}

	initialSize := q.Size()

	// Complete should remove current action
	q.Complete()

	if newSize := q.Size(); newSize != initialSize-1 {
		t.Errorf("Expected size to decrease by 1, got %d -> %d", initialSize, newSize)
	}

	// Current should be nil after complete
	if current := q.GetCurrent(); current != nil {
		t.Errorf("Expected current to be nil after complete, got %v", current)
	}
}

func TestQueueGetAll(t *testing.T) {
	testEnv := "test-getall"
	defer os.RemoveAll("inventory/" + testEnv)

	q, err := ansible.NewQueue(testEnv)
	if err != nil {
		t.Fatalf("Failed to create queue: %v", err)
	}

	q.Add("server1", status.ActionProvision, 1)
	q.Add("server2", status.ActionDeploy, 5)
	q.Add("server3", status.ActionProvision, 3)

	actions := q.GetAll()

	if len(actions) != 3 {
		t.Errorf("Expected 3 actions, got %d", len(actions))
	}

	// Verify sorted by priority (descending)
	if actions[0].Priority < actions[1].Priority {
		t.Errorf("Actions not sorted by priority: %d < %d", 
			actions[0].Priority, actions[1].Priority)
	}
}
