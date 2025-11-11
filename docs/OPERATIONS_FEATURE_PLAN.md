# Plan de Développement : Operations Manager (Working with Your Inventory)

## Vue d'ensemble
Nouvelle section principale de l'application permettant d'exécuter les playbooks Ansible (provision, deploy) avec:
- Statuts en temps réel des serveurs
- Gestion de file d'attente (FIFO)
- Logs détaillés
- Support multi-environnement avec navigation rapide

---

## 1. ARCHITECTURE DES DONNÉES

### 1.1 Structure de persistance des statuts
```
inventory/
├── dev/
│   ├── hosts.yml
│   └── .status/              # Nouveau dossier
│       ├── servers.json      # Statuts des serveurs
│       └── queue.json        # File d'attente des actions
├── staging/
│   └── .status/
└── prod/
    └── .status/
```

### 1.2 Modèle de données - ServerStatus
```go
type ServerStatus struct {
    ServerName   string        `json:"server_name"`
    Status       Status        `json:"status"`
    LastAction   string        `json:"last_action"`     // "provision", "deploy", "check"
    LastUpdate   time.Time     `json:"last_update"`
    ErrorMessage string        `json:"error_message,omitempty"`
    IsProvisioned bool         `json:"is_provisioned"`
    IsDeployed   bool          `json:"is_deployed"`
}

type Status string
const (
    StatusReady      Status = "ready"       // Toutes validations OK
    StatusNotReady   Status = "not_ready"   // Validations échouées
    StatusProvisioning Status = "provisioning"
    StatusDeploying  Status = "deploying"
    StatusVerifying  Status = "verifying"
    StatusSuccess    Status = "success"
    StatusFailed     Status = "failed"
    StatusInQueue    Status = "in_queue"
)
```

### 1.3 Modèle de données - ActionQueue
```go
type QueuedAction struct {
    ID          string    `json:"id"`           // UUID
    ServerName  string    `json:"server_name"`
    Action      string    `json:"action"`       // "provision", "deploy"
    Status      string    `json:"status"`       // "queued", "running", "completed", "failed"
    CreatedAt   time.Time `json:"created_at"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    Priority    int       `json:"priority"`     // Pour actions manuelles prioritaires
}

type ActionQueue struct {
    Actions []QueuedAction `json:"actions"`
}
```

---

## 2. COMPOSANTS UI (Bubbletea)

### 2.1 Vue principale : OperationsView
**Écran divisé en zones:**

```
┌─────────────────────────────────────────────────────────────────┐
│ 🚀 Working with Your Inventory - Environment: dev              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Servers Status                                  [Actions]       │
│ ┌───────────────────────────────────────────────────────────┐  │
│ │ [ ] web-01     192.168.1.10   ● Ready        Provision    │  │
│ │ [ ] web-02     192.168.1.11   ● Provisioned  Deploy       │  │
│ │ [✓] db-01      192.168.1.20   ⟳ Deploying...              │  │
│ │ [ ] mon-01     192.168.1.30   ✗ Failed       Retry        │  │
│ └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│ Queue (2 pending)                                               │
│ ┌───────────────────────────────────────────────────────────┐  │
│ │ 1. web-02 → Deploy (waiting for db-01)                    │  │
│ │ 2. mon-01 → Provision                                      │  │
│ └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│ Activity Logs                                   [View Full]     │
│ ┌───────────────────────────────────────────────────────────┐  │
│ │ [19:30:15] db-01: Starting deployment...                  │  │
│ │ [19:30:20] db-01: Pulling git repository...               │  │
│ │ [19:30:25] db-01: Installing dependencies... [75%]        │  │
│ └───────────────────────────────────────────────────────────┘  │
│                                                                 │
│ [Space] Select  [Enter] Action  [Tab] Switch Env  [q] Back     │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Composants détaillés

#### ServerListComponent
- Tableau scrollable des serveurs
- Checkboxes pour sélection multiple
- Colonnes: [Checkbox] Name | IP | Port | Status | Actions
- Statuts colorés avec icônes

#### StatusIndicator
```go
Symboles:
● Ready        (vert)
◐ Provisioned  (bleu)
◑ Deployed     (cyan)
⟳ En cours...  (jaune, animé)
✓ Success      (vert vif)
✗ Failed       (rouge)
⊙ In Queue     (gris)
```

#### ActionPanel
- Boutons contextuels selon statut serveur
- Actions disponibles:
  - **Validate Inventory** : Vérifier tous les serveurs
  - **Provision** : Lancer provision.yml
  - **Deploy** : Lancer deploy.yml (si provisionné)
  - **Stop Queue** : Arrêter la file d'attente
  - **Clear Queue** : Vider la file
  - **View Logs** : Voir logs détaillés

#### QueueComponent
- Liste FIFO des actions en attente
- Possibilité de supprimer une action spécifique
- Indicateur de progression

#### LogsComponent
- Logs en temps réel (tail -f style)
- Filtres par serveur
- Bouton pour ouvrir logs complets

---

## 3. VALIDATION D'INVENTAIRE

### 3.1 Critères de validation (Status: Ready)
```go
func ValidateServer(server inventory.Server) ValidationResult {
    checks := []Check{
        checkIPValid(server.IP),
        checkSSHKeyExists(server.SSHKeyPath),
        checkPortValid(server.Port),
        checkAllFieldsFilled(server),
    }
    
    // Pour web servers
    if server.Type == "web" {
        checks = append(checks,
            checkPortValid(server.AppPort),
            checkGitRepoFormat(server.GitRepo),
        )
    }
    
    return combineChecks(checks)
}
```

### 3.2 Checks implémentés
- IP valide (format IPv4)
- SSH key existe sur disque (fichier présent)
- Port valide (1-65535)
- Champs requis remplis (Name, Type, SSH User)
- Pour web: GitRepo non vide, AppPort valide
- Pour db: Port DB valide

---

## 4. EXÉCUTION ANSIBLE

### 4.1 Parser Ansible avec JSON callback

**Configuration Ansible:**
```ini
# ansible.cfg
[defaults]
stdout_callback = json
bin_ansible_callbacks = True
```

**Ou forcer à l'exécution:**
```bash
ANSIBLE_STDOUT_CALLBACK=json ansible-playbook playbooks/provision.yml -i inventory/dev
```

### 4.2 Structure de réponse JSON
```json
{
  "plays": [{
    "play": { "name": "Provision all servers" },
    "tasks": [{
      "task": { "name": "Install packages" },
      "hosts": {
        "web-01": {
          "action": "apt",
          "changed": true,
          "msg": "packages installed"
        }
      }
    }]
  }],
  "stats": {
    "web-01": {
      "ok": 15,
      "changed": 8,
      "unreachable": 0,
      "failed": 0
    }
  }
}
```

### 4.3 Exécuteur Ansible en Go
```go
type AnsibleExecutor struct {
    inventoryPath string
    playbookPath  string
    logWriter     io.Writer
}

func (e *AnsibleExecutor) RunPlaybook(ctx context.Context, options RunOptions) error {
    cmd := exec.CommandContext(ctx,
        "ansible-playbook",
        e.playbookPath,
        "-i", e.inventoryPath,
        "--limit", options.ServerLimit,  // Pour exécution individuelle
    )
    
    // Force JSON output
    cmd.Env = append(os.Environ(), "ANSIBLE_STDOUT_CALLBACK=json")
    
    // Capture stdout/stderr
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()
    
    // Start command
    cmd.Start()
    
    // Stream output en temps réel
    go e.streamOutput(stdout, e.logWriter)
    go e.streamOutput(stderr, e.logWriter)
    
    return cmd.Wait()
}
```

### 4.4 Parser de sortie JSON
```go
type AnsibleParser struct{}

func (p *AnsibleParser) ParseJSON(output []byte) (*AnsibleResult, error) {
    var result struct {
        Plays []struct {
            Tasks []struct {
                Hosts map[string]struct {
                    Changed bool   `json:"changed"`
                    Failed  bool   `json:"failed"`
                    Msg     string `json:"msg"`
                }
            }
        }
        Stats map[string]struct {
            Ok          int `json:"ok"`
            Changed     int `json:"changed"`
            Unreachable int `json:"unreachable"`
            Failed      int `json:"failed"`
        }
    }
    
    json.Unmarshal(output, &result)
    return convertToResult(result), nil
}
```

---

## 5. GESTION DE FILE D'ATTENTE

### 5.1 Queue Manager
```go
type QueueManager struct {
    queue    *ActionQueue
    running  bool
    stopChan chan struct{}
}

func (qm *QueueManager) AddAction(action QueuedAction) {
    qm.queue.Actions = append(qm.queue.Actions, action)
    qm.save()
}

func (qm *QueueManager) ProcessQueue(ctx context.Context) {
    for qm.running {
        select {
        case <-qm.stopChan:
            return
        default:
            if action := qm.getNextAction(); action != nil {
                qm.executeAction(ctx, action)
            }
            time.Sleep(1 * time.Second)
        }
    }
}

func (qm *QueueManager) Stop() {
    qm.stopChan <- struct{}{}
}
```

### 5.2 Logique FIFO avec priorité
- Actions normales : FIFO standard
- Actions manuelles prioritaires : Insérées en tête
- Une seule action à la fois par serveur
- Si erreur : action marquée "failed", queue continue

---

## 6. SYSTÈME DE LOGS

### 6.1 Structure des logs
```
inventory/dev/.logs/
├── provision_web-01_20251111_193015.log    # Log brut Ansible
├── deploy_db-01_20251111_194520.log
└── latest/
    ├── provision.log -> ../provision_web-01_20251111_193015.log
    └── deploy.log    -> ../deploy_db-01_20251111_194520.log
```

### 6.2 Logger
```go
type OperationLogger struct {
    baseDir string
}

func (l *OperationLogger) CreateLog(action, server string) (*os.File, error) {
    timestamp := time.Now().Format("20060102_150405")
    filename := fmt.Sprintf("%s_%s_%s.log", action, server, timestamp)
    path := filepath.Join(l.baseDir, ".logs", filename)
    
    file, err := os.Create(path)
    
    // Créer symlink "latest"
    l.createLatestSymlink(action, path)
    
    return file, err
}

// Rotation : garder seulement 100 derniers logs
func (l *OperationLogger) RotateLogs() {
    // Supprimer logs > 100
}
```

---

## 7. DISTINCTION PROVISION / DEPLOY

### 7.1 État des serveurs
- **Not Ready** → peut tenter Provision
- **Ready** → peut Provision
- **Provisioned** → peut Deploy (provision OK)
- **Deployed** → peut Re-deploy

### 7.2 Règles métier
```go
func CanProvision(server Server, status ServerStatus) bool {
    return status.Status == StatusReady || status.Status == StatusNotReady
}

func CanDeploy(server Server, status ServerStatus) bool {
    return status.IsProvisioned && server.Type == "web"
}
```

### 7.3 Post-check automatique
Après deploy, exécuter:
```bash
curl http://<server-ip>:<app_port>/health
```
- Success (200) → Status = Success
- Failed → Status = Failed + erreur

---

## 8. MULTI-ENVIRONNEMENT

### 8.1 Navigation rapide
- **Tab** : Changer d'environnement (dev → staging → prod → dev)
- Chargement instantané du statut de l'environnement sélectionné
- Conservation de l'état de chaque environnement

### 8.2 Barre d'onglets
```
┌──────────────────────────────────────────────────────────┐
│ [ dev* ]  [ staging ]  [ prod ]                         │
└──────────────────────────────────────────────────────────┘
```

### 8.3 Isolation des données
Chaque environnement a:
- Son propre .status/
- Sa propre .logs/
- Sa propre queue

---

## 9. REFRESH AUTOMATIQUE

### 9.1 Stratégie de refresh
```go
type RefreshManager struct {
    ticker *time.Ticker
}

func (rm *RefreshManager) GetRefreshInterval(queueActive bool) time.Duration {
    if queueActive {
        return 3 * time.Second  // Queue en cours : refresh rapide
    }
    return 5 * time.Second      // Inactif : refresh lent
}

func (rm *RefreshManager) Start(updateFunc func()) {
    go func() {
        for range rm.ticker.C {
            updateFunc()
        }
    }()
}
```

### 9.2 Indicateur visuel
```
┌─────────────────────────────────────────┐
│ 🚀 Working... ⟳ (updated 2s ago)      │
└─────────────────────────────────────────┘
```

---

## 10. RETRY ET GESTION D'ERREURS

### 10.1 Retry manuel uniquement
- Pas de retry automatique
- Si action échoue → Status = Failed
- Afficher erreur à l'utilisateur
- Bouton "Retry" apparaît pour actions failed

### 10.2 Affichage des erreurs
```
┌──────────────────────────────────────────────────────────────┐
│ ✗ web-01 - Deployment Failed                                │
│                                                              │
│ Error: Connection timeout during git clone                  │
│                                                              │
│ Last 10 lines of log:                                       │
│   fatal: unable to access repository                        │
│   Connection timed out after 60 seconds                     │
│                                                              │
│ [View Full Log]  [Retry]  [Dismiss]                         │
└──────────────────────────────────────────────────────────────┘
```

---

## 11. IMPLÉMENTATION PAR PHASES

### Phase 1: Structure de base ✓ TODO
1. Créer modèles ServerStatus, ActionQueue
2. Créer package `internal/operations/`
3. Créer status persistence (JSON)
4. Créer logs directory structure

### Phase 2: Validation ✓ TODO
1. Implémenter validation complète des serveurs
2. Fonction SetStatus(server, status)
3. UI: Affichage liste serveurs avec statuts

### Phase 3: Exécution Ansible ✓ TODO
1. AnsibleExecutor avec JSON parsing
2. Logger intégré
3. Tester avec provision.yml

### Phase 4: Queue System ✓ TODO
1. QueueManager avec FIFO
2. Actions simultanées (via sélection multiple)
3. Stop/Start queue

### Phase 5: UI Complète ✓ TODO
1. OperationsView avec tous composants
2. Navigation multi-env (Tab)
3. Refresh automatique
4. Actions panel contextuel

### Phase 6: Post-checks ✓ TODO
1. Health check automatique après deploy
2. Mise à jour statut selon résultat
3. Retry UI

### Phase 7: Polish ✓ TODO
1. Animations (spinner lors exécution)
2. Logs viewer détaillé
3. Statistiques (temps moyen, taux succès)
4. Shortcuts clavier avancés

---

## 12. FICHIERS À CRÉER

```
internal/
├── operations/
│   ├── models.go          # ServerStatus, ActionQueue
│   ├── status.go          # Persistence des statuts
│   ├── validator.go       # Validation serveurs
│   ├── ansible.go         # AnsibleExecutor
│   ├── parser.go          # JSON parser
│   ├── queue.go           # QueueManager
│   ├── logger.go          # OperationLogger
│   └── checker.go         # Health checks
│
└── ui/
    ├── operations_view.go      # Vue principale
    ├── components/
    │   ├── server_list.go      # Liste serveurs
    │   ├── action_panel.go     # Panel actions
    │   ├── queue_view.go       # Vue queue
    │   ├── logs_view.go        # Logs viewer
    │   └── env_tabs.go         # Onglets environnements
    └── operations_menu.go      # Menu depuis main
```

---

## 13. QUESTIONS À CLARIFIER

### ✅ Clarifiées
1. **Statuts** : Ready, Provisioning, Provisioned, Deploying, Deployed, Success, Failed, In Queue
2. **Logs** : Format texte brut Ansible, 100 logs max, par environnement
3. **Queue** : FIFO, stop continue autres serveurs, pas de retry auto
4. **Multi-env** : Vues séparées, navigation Tab rapide
5. **Validation** : IP, SSH key exists, ports, champs requis
6. **Actions** : Checkboxes pour multiple, possibilité de priorité manuelle
7. **Refresh** : 3s pendant exécution, 5s inactif, automatique

### ⚠️ À confirmer avec toi
1. **Ansible JSON callback** : Est-ce que forcer `ANSIBLE_STDOUT_CALLBACK=json` convient ?
2. **Health check endpoint** : Tous tes apps ont `/health` ou faut-il configurable ?
3. **Temps max d'exécution** : Timeout pour provision (30min?) et deploy (15min?) ?
4. **Notifications** : Besoin de sons/alertes quand action terminée ?
5. **Rollback** : Intégrer le playbook rollback.yml dans cette interface ?

---

## 14. ESTIMATION

- **Phase 1-2** (Structure + Validation) : ~4h
- **Phase 3** (Ansible executor) : ~3h
- **Phase 4** (Queue) : ~2h
- **Phase 5** (UI complète) : ~5h
- **Phase 6** (Post-checks) : ~2h
- **Phase 7** (Polish) : ~2h

**Total estimé : ~18-20h de développement**

---

## PRÊT À COMMENCER ?

Le plan est complet et détaillé. Dis-moi :
1. ✅ **Validation du plan** : Est-ce que ce plan répond à tous tes besoins ?
2. ❓ **Réponses aux questions** : Les 5 questions de la section 13 ?
3. 🚀 **Go/No-Go** : On commence par quelle phase ?
