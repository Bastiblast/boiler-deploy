# Ansible Best Practices Review

## Executive Summary

This document analyzes the current Ansible project structure against official Ansible best practices documentation. Overall, the project follows many best practices but has several areas for improvement.

**Score: 7/10** - Good foundation with room for optimization.

---

## ✅ What's Done Well

### 1. Directory Structure ✓
- **Status**: EXCELLENT
- Clear separation of roles, playbooks, inventory
- Uses proper roles structure (tasks/, handlers/, defaults/, templates/)
- Group variables properly organized in `group_vars/`
- Environment-specific inventories in `inventory/` subdirectories

### 2. Roles Organization ✓
- **Status**: GOOD
- Roles are modular and focused on single responsibilities:
  - `common` - Base system setup
  - `security` - UFW, fail2ban, SSH hardening
  - `nodejs` - Node.js/NVM setup
  - `nginx` - Web server configuration
  - `postgresql` - Database setup
  - `monitoring` - Observability stack
  - `deploy-app` - Application deployment
- Each role has proper structure with defaults/, handlers/, tasks/

### 3. Handlers ✓
- **Status**: GOOD
- Properly defined in each role's `handlers/main.yml`
- Services restarted/reloaded via handlers (nginx, fail2ban, sshd, etc.)
- Using `notify` in tasks appropriately

### 4. Ansible Configuration ✓
- **Status**: EXCELLENT
- Well-configured `ansible.cfg` with:
  - SSH pipelining enabled for performance
  - Fact caching configured (jsonfile)
  - Smart gathering enabled
  - Proper callbacks (yaml, profile_tasks, timer)
  - Security settings (host_key_checking, SSH args)

### 5. Collections Management ✓
- **Status**: GOOD
- Uses `requirements.yml` with pinned versions
- Relevant collections imported:
  - `community.general`
  - `community.postgresql`
  - `ansible.posix`
  - `community.crypto`

### 6. Pre/Post Tasks ✓
- **Status**: GOOD
- Uses `pre_tasks` for connection validation in provision playbook
- Uses `post_tasks` for health checks in deploy playbook

### 7. Serial Deployment ✓
- **Status**: EXCELLENT
- Deploy playbook uses `serial: 1` for rolling deployments
- Prevents taking down all servers simultaneously

---

## ⚠️ Areas for Improvement

### 1. ❌ NO TAGS USAGE
- **Status**: CRITICAL MISSING
- **Issue**: Zero tags found in entire codebase
- **Impact**: Cannot run specific parts of playbooks selectively
- **Best Practice**: Tag tasks for selective execution

**Recommendation:**
```yaml
- name: Install UFW
  apt:
    name: ufw
    state: present
  tags:
    - security
    - firewall
    - packages

- name: Configure SSH
  lineinfile:
    path: /etc/ssh/sshd_config
    regexp: "{{ item.regexp }}"
    line: "{{ item.line }}"
  tags:
    - security
    - ssh
    - configuration
```

### 2. ❌ EXCESSIVE SHELL/COMMAND USAGE
- **Status**: MAJOR ISSUE
- **Issue**: 26 instances of `shell:` or `command:` modules
- **Impact**: Reduced idempotency, harder to maintain
- **Best Practice**: Use native Ansible modules when possible

**Problems Found:**
- Git operations using shell instead of `git` module
- Package manager operations via shell (npm, pnpm, yarn)
- PM2 operations via shell
- NVM operations require shell (acceptable, no module exists)

**Recommendation:**
```yaml
# BAD (current)
- name: Get latest commit hash
  shell: "git ls-remote {{ app_repo }} {{ app_branch }} | awk '{print $1}'"
  
# BETTER (but limited by no native module for ls-remote)
# For git clone, use git module:
- name: Clone repository
  git:
    repo: "{{ app_repo }}"
    dest: "{{ release_path }}"
    version: "{{ app_branch }}"
```

**Note**: Some shell usage is acceptable for NVM since no Ansible module exists, but should use `changed_when: false` and proper error handling.

### 3. ❌ NO ANSIBLE VAULT
- **Status**: SECURITY CONCERN
- **Issue**: No secrets management with Ansible Vault
- **Impact**: Potential security risks if sensitive data in repo
- **Best Practice**: Encrypt sensitive variables

**Recommendation:**
```bash
# Create encrypted file
ansible-vault create group_vars/all/vault.yml

# Content:
vault_db_password: "secret123"
vault_api_key: "xyz789"
```

### 4. ⚠️ INCONSISTENT BECOME USAGE
- **Status**: MINOR ISSUE
- **Issue**: Sometimes `become: yes` in playbook, sometimes in tasks
- **Impact**: Confusion about privilege escalation
- **Best Practice**: Consistent pattern

**Current Pattern:**
```yaml
# Playbook level
- name: Provision all servers
  hosts: all
  become: yes  # ✓ Good

# But also in tasks (redundant)
become: yes
become_user: "{{ deploy_user }}"
```

**Recommendation**: Keep `become` at playbook level unless specific tasks need different users.

### 5. ⚠️ HARD-CODED VALUES IN TASKS
- **Status**: MODERATE ISSUE
- **Issue**: Some values hard-coded instead of variables
- Examples:
  - Port numbers (80, 443, 9090, 3001, 9100)
  - Some paths
  - Timeout values

**Recommendation:**
```yaml
# In defaults/main.yml
http_port: 80
https_port: 443
prometheus_port: 9090
grafana_port: 3001

# In tasks
- name: Allow HTTP
  ufw:
    rule: allow
    port: "{{ http_port }}"
    proto: tcp
```

### 6. ⚠️ NO CHECK MODE SUPPORT
- **Status**: MODERATE ISSUE
- **Issue**: Many shell tasks don't work with `--check` mode
- **Impact**: Cannot do dry-run validation
- **Best Practice**: Support check mode

**Recommendation:**
```yaml
- name: Check something
  shell: "some command"
  check_mode: no  # Skip in check mode
  changed_when: false
```

### 7. ⚠️ LIMITED ERROR HANDLING
- **Status**: MINOR ISSUE
- **Issue**: Some tasks use `ignore_errors: yes` broadly
- **Best Practice**: Specific error handling with `failed_when`

**Current:**
```yaml
- name: Install unattended-upgrades
  apt:
    name:
      - unattended-upgrades
    state: present
  ignore_errors: yes  # Too broad
```

**Better:**
```yaml
- name: Install unattended-upgrades
  apt:
    name:
      - unattended-upgrades
    state: present
  register: upgrade_result
  failed_when:
    - upgrade_result.rc != 0
    - "'Unable to locate package' not in upgrade_result.msg"
```

### 8. ⚠️ NO BLOCK/RESCUE PATTERN
- **Status**: MINOR ISSUE
- **Issue**: No use of block/rescue for error handling
- **Best Practice**: Use for complex operations

**Recommendation:**
```yaml
- name: Deploy application
  block:
    - name: Install dependencies
      shell: npm install
    - name: Build app
      shell: npm run build
  rescue:
    - name: Rollback
      file:
        path: "{{ release_path }}"
        state: absent
    - name: Notify failure
      debug:
        msg: "Deployment failed, rolled back"
```

### 9. ⚠️ INVENTORY STRUCTURE
- **Status**: COULD BE IMPROVED
- **Issue**: Each environment duplicates structure
- **Best Practice**: Use inventory plugins or dynamic inventory

**Current:**
```
inventory/
├── docker/
│   ├── hosts.yml
│   └── group_vars/
├── dev/
│   ├── hosts.yml
│   └── group_vars/
```

**Better**: Use inventory plugin or single inventory with groups.

### 10. ⚠️ NO META/MAIN.YML IN ROLES
- **Status**: MINOR ISSUE
- **Issue**: Roles don't declare dependencies
- **Best Practice**: Use `meta/main.yml` for role metadata

**Recommendation:**
```yaml
# roles/deploy-app/meta/main.yml
---
dependencies:
  - role: nodejs
  - role: nginx

galaxy_info:
  author: Your Name
  description: Deploy Node.js applications
  min_ansible_version: "2.9"
```

---

## 📊 Best Practices Checklist

| Practice | Status | Priority |
|----------|--------|----------|
| Directory structure | ✅ | - |
| Roles organization | ✅ | - |
| Handlers usage | ✅ | - |
| ansible.cfg | ✅ | - |
| Collections management | ✅ | - |
| Pre/post tasks | ✅ | - |
| Serial deployment | ✅ | - |
| **Tags** | ❌ | **HIGH** |
| **Ansible Vault** | ❌ | **HIGH** |
| **Shell module usage** | ⚠️ | **MEDIUM** |
| Become consistency | ⚠️ | LOW |
| Variable usage | ⚠️ | LOW |
| Check mode support | ⚠️ | MEDIUM |
| Error handling | ⚠️ | LOW |
| Block/rescue | ⚠️ | LOW |
| Inventory structure | ⚠️ | LOW |
| Role metadata | ⚠️ | LOW |

---

## 🎯 Priority Recommendations

### IMMEDIATE (Do First)

1. **Add Tags Throughout**
   - Tag all tasks with relevant categories
   - Enable selective playbook execution
   - Estimated effort: 2-3 hours

2. **Implement Ansible Vault**
   - Encrypt any sensitive variables
   - Set up vault password file
   - Estimated effort: 1 hour

### SHORT TERM (Next Sprint)

3. **Reduce Shell Module Usage**
   - Replace shell commands with native modules where possible
   - Add `changed_when` and `failed_when` to remaining shell tasks
   - Estimated effort: 4-6 hours

4. **Add Check Mode Support**
   - Test playbooks with `--check` flag
   - Fix tasks that fail in check mode
   - Estimated effort: 2 hours

### LONG TERM (Future Improvement)

5. **Implement Block/Rescue**
   - Add error handling to critical operations
   - Implement rollback mechanisms
   - Estimated effort: 4 hours

6. **Improve Inventory Structure**
   - Consider dynamic inventory or inventory plugins
   - Reduce duplication
   - Estimated effort: 3 hours

7. **Add Role Metadata**
   - Create meta/main.yml for each role
   - Document dependencies
   - Estimated effort: 1 hour

---

## 📚 References

- [Ansible Best Practices](https://docs.ansible.com/ansible/latest/tips_tricks/ansible_tips_tricks.html)
- [Sample Directory Layout](https://docs.ansible.com/ansible/latest/tips_tricks/sample_setup.html)
- [Working with Playbooks](https://docs.ansible.com/ansible/latest/playbook_guide/index.html)
- [Ansible Vault](https://docs.ansible.com/ansible/latest/vault_guide/index.html)
- [Using Handlers](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_handlers.html)
- [Tags](https://docs.ansible.com/ansible/latest/playbook_guide/playbooks_tags.html)

---

## 🔍 Conclusion

The project demonstrates a **solid understanding of Ansible fundamentals** with excellent directory structure, role organization, and configuration. However, to reach production-grade quality, focus on:

1. **Adding comprehensive tags** for operational flexibility
2. **Implementing Ansible Vault** for security
3. **Reducing shell module usage** for better idempotency
4. **Improving error handling** and check mode support

These improvements will make the project more maintainable, secure, and production-ready.

---

*Generated on: 2025-11-19*
*Project: boiler-deploy*
*Ansible Version: 2.9+*
