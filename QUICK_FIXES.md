# 🚀 Quick Fixes à Appliquer Immédiatement

## ✅ Fait (Automatique)

### 1. Git: Exclusion Runtime Files
```bash
✓ Ajouté à .gitignore:
  - inventory/*/.status/
  - inventory/*/.queue/
  - debug.log

✓ Retiré du tracking:
  - inventory/docker/.queue/actions.json
  - inventory/docker/.status/servers.json
```

**Commit:**
```bash
git add .gitignore
git commit -m "chore: exclude runtime state files from git"
```

---

## 🔴 Urgent: Sécurité (5 minutes)

### Fix 1: Deploy User (group_vars/all.yml)

**Avant:**
```yaml
deploy_user: root
allow_root_login: true
```

**Après:**
```yaml
deploy_user: deploy
deploy_user_groups:
  - sudo
  - www-data
allow_root_login: false  # Désactiver après première provision
```

**Note:** Pour transition douce, garder root temporairement en commentaire

---

## 🟡 Important: Tests (30 minutes)

### Créer Structure Tests

```bash
mkdir -p tests/{unit,integration,fixtures}
mkdir -p tests/unit/{ansible,inventory,ssh}
```

### Test Exemple: Queue

```go
// tests/unit/ansible/queue_test.go
package ansible_test

import (
    "testing"
    "github.com/bastiblast/boiler-deploy/internal/ansible"
    "github.com/bastiblast/boiler-deploy/internal/status"
)

func TestQueuePriority(t *testing.T) {
    q, _ := ansible.NewQueue("test")
    defer os.RemoveAll("inventory/test/.queue")
    
    q.Add("server1", status.ActionProvision, 1)
    q.Add("server2", status.ActionProvision, 10)
    
    next := q.Next()
    if next.ServerName != "server2" {
        t.Errorf("Expected server2 (priority 10), got %s", next.ServerName)
    }
}
```

**Lancer:**
```bash
go test ./tests/unit/... -v
```

---

## 🟢 Optionnel: Health Check Fallback (15 minutes)

### Ajouter HTTP Client Natif

```go
// internal/ansible/executor.go

import "net/http"

func (e *Executor) HealthCheck(ip string, port int) error {
    // 1. Essayer curl (existant)
    if err := e.healthCheckCurl(ip, port); err == nil {
        return nil
    }
    
    // 2. Fallback HTTP natif (nouveau)
    return e.healthCheckNative(ip, port)
}

func (e *Executor) healthCheckNative(ip string, port int) error {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }
    
    url := fmt.Sprintf("http://%s:%d/", ip, port)
    resp, err := client.Get(url)
    if err != nil {
        return fmt.Errorf("HTTP check failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Accepter 2xx-4xx (app répond)
    if resp.StatusCode >= 200 && resp.StatusCode < 500 {
        log.Printf("[EXECUTOR] ✓ Health check native OK (status: %d)", resp.StatusCode)
        return nil
    }
    
    return fmt.Errorf("bad status: %d", resp.StatusCode)
}

func (e *Executor) healthCheckCurl(ip string, port int) error {
    // Code existant
    url := fmt.Sprintf("http://%s:%d/", ip, port)
    cmd := exec.Command("curl", "-sf", "-m", "10", url)
    return cmd.Run()
}
```

---

## 📋 Checklist Application

- [x] .gitignore: runtime files exclus
- [ ] group_vars/all.yml: deploy_user → deploy
- [ ] Tests: structure créée + 1 test exemple
- [ ] Health check: fallback HTTP natif
- [ ] README: Badge tests (après CI)

**Temps Total:** ~50 minutes  
**Impact:** Sécurité +30pts, Qualité +20pts

---

## 🎯 Validation

```bash
# 1. Build check
make build

# 2. Run tests (quand créés)
make test

# 3. Security scan (optionnel)
gosec ./...

# 4. Git status
git status  # Devrait être clean sauf modifications voulues
```

---

**Généré par:** Boiler Expert Agent v2  
**Lire:** AUDIT_REPORT.md pour analyse complète
