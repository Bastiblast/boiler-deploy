# Tests Directory

This directory contains all testing-related files for the Boiler Deploy project.

## Test Scripts

- **test-docker-vps.sh** - Main script to create and manage a Docker container simulating a VPS for testing
- **test_setup.sh** - Setup testing script
- **test_validation.sh** - Inventory validation testing
- **test_validation_debug.sh** - Debug version of validation testing
- **test_workflow.sh** - Workflow testing script
- **test_script_exec.sh** - Script execution testing
- **test_check_manual.sh** - Manual check testing

## Test Documentation

- **TEST_CONTAINER_GUIDE.md** - Guide for using test containers
- **TEST_ENVIRONMENT.md** - Test environment setup documentation
- **TEST_FIX_GUIDE.md** - Guide for fixing common test issues
- **QUICKSTART_TEST.md** - Quick start guide for testing
- **QUICKTEST.md** - Quick test procedures

## Usage

To run the Docker VPS test:
```bash
./tests/test-docker-vps.sh setup    # Create and start test container
./tests/test-docker-vps.sh status   # Check container status
./tests/test-docker-vps.sh cleanup  # Remove test container
```
