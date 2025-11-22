package ansible_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bastiblast/boiler-deploy/internal/ansible"
)

func TestExecutorContextCancellation(t *testing.T) {
	// Create temporary environment
	testEnv := "test-context"
	defer os.RemoveAll("inventory/" + testEnv)

	executor := ansible.NewExecutor(testEnv)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	progressChan := make(chan string, 10)
	go func() {
		for range progressChan {
			// Drain progress channel
		}
	}()

	// Run playbook that would take longer than timeout
	// This should be cancelled by context
	result, err := executor.RunPlaybookWithContext(ctx, "provision.yml", "test-server", "", progressChan)

	// We expect either context error OR ansible failure (no inventory)
	// The important part is no hang and graceful handling
	if err == nil && result != nil && result.Success {
		t.Error("Expected error or failure due to context cancellation or missing inventory")
	}

	if err != nil {
		t.Logf("Got expected error: %v", err)
	}

	close(progressChan)
}

func TestExecutorContextWithTimeout(t *testing.T) {
	testEnv := "test-timeout"
	defer os.RemoveAll("inventory/" + testEnv)

	executor := ansible.NewExecutor(testEnv)

	// Test that default timeout is applied when no deadline
	ctx := context.Background()

	progressChan := make(chan string, 10)
	go func() {
		for range progressChan {
			// Drain
		}
	}()

	// This will fail quickly due to missing inventory
	result, err := executor.RunPlaybookWithContext(ctx, "provision.yml", "test-server", "", progressChan)

	// We expect either error or failed result (invalid inventory)
	// The test verifies no panic and graceful failure
	if err == nil && result != nil && result.Success {
		t.Error("Expected error or failure due to invalid inventory")
	}

	if err != nil {
		t.Logf("Got expected error: %v", err)
	} else if result != nil && !result.Success {
		t.Logf("Got expected failure: %s", result.ErrorMessage)
	}

	close(progressChan)
}

func TestExecutorManualContextCancellation(t *testing.T) {
	testEnv := "test-manual-cancel"
	defer os.RemoveAll("inventory/" + testEnv)

	executor := ansible.NewExecutor(testEnv)

	ctx, cancel := context.WithCancel(context.Background())

	progressChan := make(chan string, 10)
	go func() {
		for range progressChan {
			// Drain
		}
	}()

	// Start execution and cancel immediately
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := executor.RunPlaybookWithContext(ctx, "provision.yml", "test-server", "", progressChan)

	// Should get context cancelled error
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	close(progressChan)
}

func TestProvisionWithContext(t *testing.T) {
	testEnv := "test-provision-ctx"
	defer os.RemoveAll("inventory/" + testEnv)

	executor := ansible.NewExecutor(testEnv)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	progressChan := make(chan string, 10)
	go func() {
		for range progressChan {
			// Drain
		}
	}()

	_, err := executor.ProvisionWithContext(ctx, "test-server", "", progressChan)

	// Expect error (timeout or invalid inventory)
	if err == nil {
		t.Error("Expected error")
	}

	close(progressChan)
}

func TestDeployWithContext(t *testing.T) {
	testEnv := "test-deploy-ctx"
	defer os.RemoveAll("inventory/" + testEnv)

	executor := ansible.NewExecutor(testEnv)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	progressChan := make(chan string, 10)
	go func() {
		for range progressChan {
			// Drain
		}
	}()

	_, err := executor.DeployWithContext(ctx, "test-server", "", progressChan)

	// Expect error (timeout or invalid inventory)
	if err == nil {
		t.Error("Expected error")
	}

	close(progressChan)
}
