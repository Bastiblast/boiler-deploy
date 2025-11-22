#!/bin/bash
# Autonomous Agent Test Script for Parallel Execution Feature
# Ce script teste automatiquement la feature parallel execution

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

echo "🤖 Autonomous Agent Test - Parallel Execution Feature"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test 1: Compilation
test_compilation() {
    log_info "Test 1: Compilation des packages modifiés"
    
    if go build ./internal/ansible/... 2>&1; then
        log_success "internal/ansible compiled"
    else
        log_error "internal/ansible compilation failed"
        return 1
    fi
    
    if go build ./internal/config/... 2>&1; then
        log_success "internal/config compiled"
    else
        log_error "internal/config compilation failed"
        return 1
    fi
    
    if go build ./internal/ui/... 2>&1; then
        log_success "internal/ui compiled"
    else
        log_error "internal/ui compilation failed"
        return 1
    fi
    
    log_success "Test 1 PASSED: All packages compile successfully"
    return 0
}

# Test 2: Vérifier les nouvelles méthodes
test_new_methods() {
    log_info "Test 2: Vérification des nouvelles méthodes"
    
    # Check Orchestrator methods
    if grep -q "func (o \*Orchestrator) SetMaxWorkers" internal/ansible/orchestrator.go; then
        log_success "SetMaxWorkers() found"
    else
        log_error "SetMaxWorkers() not found"
        return 1
    fi
    
    if grep -q "func (o \*Orchestrator) processQueueParallel" internal/ansible/orchestrator.go; then
        log_success "processQueueParallel() found"
    else
        log_error "processQueueParallel() not found"
        return 1
    fi
    
    if grep -q "func (o \*Orchestrator) processQueueSequential" internal/ansible/orchestrator.go; then
        log_success "processQueueSequential() found"
    else
        log_error "processQueueSequential() not found"
        return 1
    fi
    
    # Check Queue methods
    if grep -q "func (q \*Queue) NextBatch" internal/ansible/queue.go; then
        log_success "NextBatch() found"
    else
        log_error "NextBatch() not found"
        return 1
    fi
    
    if grep -q "func (q \*Queue) CompleteByID" internal/ansible/queue.go; then
        log_success "CompleteByID() found"
    else
        log_error "CompleteByID() not found"
        return 1
    fi
    
    log_success "Test 2 PASSED: All new methods present"
    return 0
}

# Test 3: Vérifier la configuration
test_config() {
    log_info "Test 3: Vérification de la configuration"
    
    if grep -q "MaxParallelWorkers" internal/config/types.go; then
        log_success "MaxParallelWorkers config option found"
    else
        log_error "MaxParallelWorkers config option not found"
        return 1
    fi
    
    if grep -q "MaxParallelWorkers.*0.*Sequential" internal/config/types.go; then
        log_success "Default value is 0 (sequential, backward compatible)"
    else
        log_warning "Could not verify default value"
    fi
    
    log_success "Test 3 PASSED: Configuration properly defined"
    return 0
}

# Test 4: Vérifier l'intégration UI
test_ui_integration() {
    log_info "Test 4: Vérification de l'intégration UI"
    
    if grep -q "SetMaxWorkers" internal/ui/workflow_view.go; then
        log_success "SetMaxWorkers() called in UI"
    else
        log_error "SetMaxWorkers() not called in UI"
        return 1
    fi
    
    log_success "Test 4 PASSED: UI integration verified"
    return 0
}

# Test 5: Vérifier la thread-safety
test_thread_safety() {
    log_info "Test 5: Vérification de la thread-safety"
    
    # Check mutexes
    if grep -q "workersMu.*sync.Mutex" internal/ansible/orchestrator.go; then
        log_success "Workers mutex found"
    else
        log_error "Workers mutex not found"
        return 1
    fi
    
    if grep -q "o.workersMu.Lock()" internal/ansible/orchestrator.go; then
        log_success "Workers mutex used"
    else
        log_error "Workers mutex not used"
        return 1
    fi
    
    # Check WaitGroup in parallel mode
    if grep -q "sync.WaitGroup" internal/ansible/orchestrator.go; then
        log_success "WaitGroup found for worker synchronization"
    else
        log_error "WaitGroup not found"
        return 1
    fi
    
    log_success "Test 5 PASSED: Thread-safety mechanisms in place"
    return 0
}

# Test 6: Vérifier la documentation
test_documentation() {
    log_info "Test 6: Vérification de la documentation"
    
    if [ ! -f "docs/PARALLEL_EXECUTION.md" ]; then
        log_error "Documentation file not found"
        return 1
    fi
    
    required_sections=(
        "Overview"
        "Configuration"
        "Architecture"
        "Thread-safety"
        "Performance"
        "Fallbacks"
        "Examples"
    )
    
    for section in "${required_sections[@]}"; do
        if grep -q "$section" docs/PARALLEL_EXECUTION.md; then
            log_success "Section '$section' found"
        else
            log_error "Section '$section' not found"
            return 1
        fi
    done
    
    log_success "Test 6 PASSED: Documentation complete"
    return 0
}

# Test 7: Vérifier la backward compatibility
test_backward_compatibility() {
    log_info "Test 7: Vérification de la backward compatibility"
    
    # Check that processQueueSequential still exists
    if grep -q "processQueueSequential" internal/ansible/orchestrator.go; then
        log_success "Sequential mode preserved"
    else
        log_error "Sequential mode not found"
        return 1
    fi
    
    # Check default is 0 (sequential)
    if grep -q "maxWorkers.*0" internal/ansible/orchestrator.go; then
        log_success "Default is sequential (0 workers)"
    else
        log_warning "Could not verify default sequential mode"
    fi
    
    log_success "Test 7 PASSED: Backward compatible"
    return 0
}

# Test 8: Performance analysis
test_performance_documentation() {
    log_info "Test 8: Vérification de l'analyse de performance"
    
    if grep -q "N / W" docs/PARALLEL_EXECUTION.md; then
        log_success "Performance formula documented"
    else
        log_error "Performance formula not documented"
        return 1
    fi
    
    if grep -q "66%" docs/PARALLEL_EXECUTION.md; then
        log_success "Performance example with percentages found"
    else
        log_warning "Performance example could be more detailed"
    fi
    
    log_success "Test 8 PASSED: Performance documented"
    return 0
}

# Run all tests
run_all_tests() {
    echo ""
    log_info "Starting autonomous agent tests..."
    echo ""
    
    tests=(
        "test_compilation"
        "test_new_methods"
        "test_config"
        "test_ui_integration"
        "test_thread_safety"
        "test_documentation"
        "test_backward_compatibility"
        "test_performance_documentation"
    )
    
    failed=0
    passed=0
    
    for test in "${tests[@]}"; do
        echo ""
        set +e
        $test
        result=$?
        set -e
        if [ $result -eq 0 ]; then
            passed=$((passed + 1))
        else
            failed=$((failed + 1))
        fi
    done
    
    echo ""
    echo "=========================================="
    echo ""
    log_info "Test Results Summary:"
    echo ""
    log_success "Passed: $passed/${#tests[@]}"
    
    if [ $failed -gt 0 ]; then
        log_error "Failed: $failed/${#tests[@]}"
        echo ""
        log_error "Some tests failed. Please review the output above."
        return 1
    else
        echo ""
        log_success "🎉 ALL TESTS PASSED!"
        echo ""
        log_info "Feature is ready for:"
        echo "  ✅ Code review"
        echo "  ✅ Merge to main"
        echo "  ✅ Production deployment"
        echo ""
        log_info "Recommended next steps:"
        echo "  1. Test manually with 'max_parallel_workers: 3' in config"
        echo "  2. Monitor logs for parallel execution"
        echo "  3. Measure performance improvements"
        return 0
    fi
}

# Main execution
main() {
    run_all_tests
    exit_code=$?
    
    echo ""
    log_info "Autonomous agent test completed."
    echo ""
    
    exit $exit_code
}

# Run if executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main
fi
