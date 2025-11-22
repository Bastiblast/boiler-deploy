#!/bin/bash

##############################################################################
# Test script for setup.sh
# Validates the setup wizard without actual SSH connections
##############################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=================================="
echo "Testing setup.sh workflow"
echo "=================================="
echo

# Test 1: Check script is executable
echo "✓ Test 1: Script executable"
if [[ -x "${SCRIPT_DIR}/setup.sh" ]]; then
    echo "  ✓ setup.sh is executable"
else
    echo "  ✗ setup.sh is not executable"
    chmod +x "${SCRIPT_DIR}/setup.sh"
    echo "  → Fixed: Made executable"
fi
echo

# Test 2: Check prerequisites detection
echo "✓ Test 2: Prerequisites check"
if bash "${SCRIPT_DIR}/setup.sh" staging --help 2>&1 | grep -q "Multi-Server"; then
    echo "  ✓ Script loads correctly"
else
    echo "  ✗ Script failed to load"
fi
echo

# Test 3: Validate functions exist
echo "✓ Test 3: Function validation"
functions=(
    "validate_ip"
    "check_ip_conflict"
    "test_ssh_connection"
    "configure_web_servers"
    "configure_database"
    "configure_monitoring"
    "generate_hosts_yml"
    "generate_group_vars"
)

for func in "${functions[@]}"; do
    if grep -q "^${func}()" "${SCRIPT_DIR}/setup.sh"; then
        echo "  ✓ Function exists: $func"
    else
        echo "  ✗ Function missing: $func"
    fi
done
echo

# Test 4: Check IP validation logic
echo "✓ Test 4: IP validation (extracted logic test)"
validate_ip_test() {
    local ip="$1"
    local regex='^([0-9]{1,3}\.){3}[0-9]{1,3}$'
    
    if [[ ! $ip =~ $regex ]]; then
        return 1
    fi
    
    IFS='.' read -ra ADDR <<< "$ip"
    for i in "${ADDR[@]}"; do
        if ((i > 255)); then
            return 1
        fi
    done
    
    return 0
}

test_ips=(
    "192.168.1.1:valid"
    "10.0.0.1:valid"
    "256.1.1.1:invalid"
    "192.168.1:invalid"
    "abc.def.ghi.jkl:invalid"
)

for test_case in "${test_ips[@]}"; do
    IFS=':' read -r ip expected <<< "$test_case"
    if validate_ip_test "$ip"; then
        result="valid"
    else
        result="invalid"
    fi
    
    if [[ "$result" == "$expected" ]]; then
        echo "  ✓ IP validation: $ip → $result"
    else
        echo "  ✗ IP validation: $ip → $result (expected: $expected)"
    fi
done
echo

# Test 5: Check directory structure creation
echo "✓ Test 5: Directory structure"
required_dirs=(
    "inventory"
    "group_vars"
    "playbooks"
    "roles"
)

for dir in "${required_dirs[@]}"; do
    if [[ -d "${SCRIPT_DIR}/${dir}" ]]; then
        echo "  ✓ Directory exists: $dir"
    else
        echo "  ✗ Directory missing: $dir"
    fi
done
echo

# Test 6: Check ansible.cfg
echo "✓ Test 6: Ansible configuration"
if [[ -f "${SCRIPT_DIR}/ansible.cfg" ]]; then
    echo "  ✓ ansible.cfg exists"
    if grep -q "inventory" "${SCRIPT_DIR}/ansible.cfg"; then
        echo "  ✓ ansible.cfg has inventory config"
    fi
else
    echo "  ✗ ansible.cfg missing"
fi
echo

echo "=================================="
echo "All tests completed!"
echo "=================================="
echo
echo "To run full setup wizard:"
echo "  ./setup.sh"
echo
echo "To create a new environment:"
echo "  ./setup.sh staging"
echo
echo "To add servers to existing:"
echo "  ./setup.sh production --add-servers"
echo
