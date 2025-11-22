# 📋 Plan Global : Gestionnaire d'Inventaire Ansible

## 🎯 Objectif

Créer un gestionnaire d'inventaire Ansible **léger, interactif et adapté** pour remplacer le script bash complexe de 1452 lignes.

---

## 📐 Architecture Globale

### Vue d'ensemble

```
┌─────────────────────────────────────────────────────────────┐
│                  Inventory Manager CLI                       │
│                    (Bubbletea TUI)                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Créer      │  │   Éditer     │  │   Valider    │    │
│  │ Environement │  │  Inventaire  │  │     SSH      │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Gérer      │  │   Exporter   │  │   Importer   │    │
│  │   Serveurs   │  │     YAML     │  │     YAML     │    │
│  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                    Couche Métier (Go)                       │
│  • Validation IP/Ports                                      │
│  • Test SSH                                                 │
│  • Génération YAML Ansible                                  │
│  • Gestion d'état                                           │
├─────────────────────────────────────────────────────────────┤
│              Stockage (Fichiers YAML)                       │
│  inventory/                                                 │
│    └── [env]/                                              │
│        ├── hosts.yml                                       │
│        └── config.yml                                      │
└─────────────────────────────────────────────────────────────┘
```

---

## 🏗️ Structure du Projet

```
boiler-deploy/
├── cmd/
│   └── inventory-manager/
│       └── main.go                    # Point d'entrée
│
├── internal/
│   ├── ui/                            # Interface Bubbletea
│   │   ├── models.go                  # Modèles de données
│   │   ├── views.go                   # Vues/Écrans
│   │   ├── components/                # Composants réutilisables
│   │   │   ├── menu.go
│   │   │   ├── form.go
│   │   │   ├── list.go
│   │   │   └── table.go
│   │   └── styles.go                  # Styles Lipgloss
│   │
│   ├── inventory/                     # Logique métier
│   │   ├── manager.go                 # Gestionnaire principal
│   │   ├── environment.go             # Gestion environnements
│   │   ├── server.go                  # Modèle serveur
│   │   ├── validator.go               # Validations
│   │   └── generator.go               # Génération YAML
│   │
│   ├── ssh/                           # Gestion SSH
│   │   ├── tester.go                  # Test connexions
│   │   └── keys.go                    # Gestion clés
│   │
│   └── storage/                       # Persistance
│       ├── yaml.go                    # Lecture/Écriture YAML
│       └── state.go                   # État application
│
├── go.mod
├── go.sum
└── Makefile
```

---

## 🎨 Interface Utilisateur (Bubbletea)

### Écran Principal - Menu

```
╔══════════════════════════════════════════════════════════════╗
║           Ansible Inventory Manager v1.0                     ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  📁 Environnements existants:                               ║
║     • production (3 serveurs)                               ║
║     • dev (1 serveur)                                       ║
║     • staging (2 serveurs)                                  ║
║                                                              ║
║  ┌────────────────────────────────────────────────────────┐ ║
║  │  > Créer un nouvel environnement                       │ ║
║  │    Gérer un environnement existant                     │ ║
║  │    Valider tous les inventaires                        │ ║
║  │    Exporter la configuration                           │ ║
║  │    Quitter                                             │ ║
║  └────────────────────────────────────────────────────────┘ ║
║                                                              ║
║  [↑↓] Naviguer  [Enter] Sélectionner  [q] Quitter          ║
╚══════════════════════════════════════════════════════════════╝
```

### Écran - Créer Environnement

```
╔══════════════════════════════════════════════════════════════╗
║           Nouvel Environnement                               ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  Nom de l'environnement: [production_______________]        ║
║                                                              ║
║  Services à déployer:                                        ║
║    [x] Serveurs Web                                         ║
║    [x] Base de données                                      ║
║    [ ] Monitoring (Prometheus + Grafana)                    ║
║                                                              ║
║  Configuration Git:                                          ║
║    Repository: [https://github.com/user/repo.git_______]   ║
║    Branche:    [main__________________________________]     ║
║                                                              ║
║  Configuration Node.js:                                      ║
║    Version: [20___]  Port: [3000]                          ║
║                                                              ║
║  [Tab] Champ suivant  [Enter] Continuer  [Esc] Retour      ║
╚══════════════════════════════════════════════════════════════╝
```

### Écran - Gestion Serveurs

```
╔══════════════════════════════════════════════════════════════╗
║           Environnement: production                          ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  Serveurs Web:                                              ║
║  ┌────────────────────────────────────────────────────────┐ ║
║  │ Nom              IP            Port   Status            │ ║
║  ├────────────────────────────────────────────────────────┤ ║
║  │ > prod-web-01    192.168.1.10  3000   ✓ SSH OK        │ ║
║  │   prod-web-02    192.168.1.11  3001   ⚠ Non testé     │ ║
║  │   prod-web-03    192.168.1.12  3002   ✗ Erreur SSH    │ ║
║  └────────────────────────────────────────────────────────┘ ║
║                                                              ║
║  Base de données:                                           ║
║  ┌────────────────────────────────────────────────────────┐ ║
║  │   prod-db-01     192.168.1.20  5432   ✓ SSH OK        │ ║
║  └────────────────────────────────────────────────────────┘ ║
║                                                              ║
║  [a] Ajouter  [e] Éditer  [d] Supprimer  [t] Tester SSH   ║
║  [s] Sauvegarder  [Esc] Retour                             ║
╚══════════════════════════════════════════════════════════════╝
```

### Écran - Ajouter Serveur

```
╔══════════════════════════════════════════════════════════════╗
║           Ajouter un Serveur Web                             ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  Nom du serveur:     [prod-web-04________________]          ║
║                      (auto: production-web-04)              ║
║                                                              ║
║  Adresse IP:         [192.168.1.13________________]         ║
║                      ✓ Format IP valide                     ║
║                                                              ║
║  Port application:   [3003]                                 ║
║                      ⚠ Conflit possible avec web-02         ║
║                                                              ║
║  User SSH:           [root__________________________]       ║
║                                                              ║
║  Clé SSH:            [~/.ssh/id_rsa_________________]       ║
║                      ✓ Clé trouvée                          ║
║                                                              ║
║  Hostname (opt):     [web04.prod.example.com________]       ║
║                                                              ║
║  [Enter] Valider  [t] Tester SSH  [Esc] Annuler            ║
╚══════════════════════════════════════════════════════════════╝
```

---

## 🧩 Composants Clés

### 1. Menu Principal (ui/menu.go)

```go
type MenuModel struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
    envs     []Environment
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (m MenuModel) View() string
```

### 2. Formulaire Serveur (ui/components/form.go)

```go
type ServerForm struct {
    name        textinput.Model
    ip          textinput.Model
    port        textinput.Model
    user        textinput.Model
    sshKey      textinput.Model
    focusIndex  int
    validation  ValidationResult
}

func (f *ServerForm) Validate() error
func (f *ServerForm) TestSSH() (bool, error)
```

### 3. Liste Interactive (ui/components/list.go)

```go
type ServerList struct {
    items    []Server
    cursor   int
    selected int
    filter   string
}

func (l ServerList) View() string
func (l *ServerList) HandleKey(key string) tea.Cmd
```

### 4. Gestionnaire Inventaire (inventory/manager.go)

```go
type Manager struct {
    storage     storage.Storage
    validator   Validator
    sshTester   ssh.Tester
}

func (m *Manager) CreateEnvironment(name string, config Config) error
func (m *Manager) AddServer(env string, server Server) error
func (m *Manager) GenerateInventory(env string) ([]byte, error)
func (m *Manager) ValidateAll() ([]ValidationResult, error)
```

### 5. Validateur (inventory/validator.go)

```go
type Validator struct {}

func (v *Validator) ValidateIP(ip string) error
func (v *Validator) ValidatePort(port int) error
func (v *Validator) CheckIPConflict(servers []Server, ip string, port int) error
func (v *Validator) ValidateGitRepo(repo, branch string) error
```

### 6. Testeur SSH (ssh/tester.go)

```go
type Tester struct {
    timeout time.Duration
}

func (t *Tester) TestConnection(server Server) (bool, error)
func (t *Tester) TestAllServers(servers []Server) map[string]bool
func (t *Tester) CheckPython3(server Server) (bool, error)
```

### 7. Générateur YAML (inventory/generator.go)

```go
type Generator struct {}

func (g *Generator) GenerateHostsYAML(env Environment) ([]byte, error)
func (g *Generator) GenerateGroupVarsYAML(env Environment) ([]byte, error)
func (g *Generator) GenerateAnsibleCfg() ([]byte, error)
```

---

## 📊 Flux de Données

```
┌──────────────┐
│  Utilisateur │
└──────┬───────┘
       │ Interaction (clavier)
       ▼
┌──────────────────┐
│  Bubbletea UI    │
│  (View/Update)   │
└──────┬───────────┘
       │ Commandes
       ▼
┌──────────────────┐
│  Manager         │ ◄──► Validator
│  (Logique)       │ ◄──► SSH Tester
└──────┬───────────┘
       │ Données
       ▼
┌──────────────────┐
│  Storage         │
│  (YAML Files)    │
└──────────────────┘
```

---

## 🎯 Fonctionnalités Principales

### Phase 1 : MVP (Minimum Viable Product)

- [x] ✅ Menu principal interactif
- [x] ✅ Créer un environnement
- [x] ✅ Ajouter des serveurs web
- [x] ✅ Validation IP/Port
- [x] ✅ Génération hosts.yml
- [x] ✅ Sauvegarde fichiers

### Phase 2 : Fonctionnalités Avancées

- [ ] 🔄 Test connexion SSH
- [ ] 🔄 Édition serveurs existants
- [ ] 🔄 Suppression serveurs
- [ ] 🔄 Détection conflits IP/Port
- [ ] 🔄 Import inventaire existant
- [ ] 🔄 Support base de données
- [ ] 🔄 Support monitoring

### Phase 3 : Polish

- [ ] 🎨 Thèmes de couleurs
- [ ] 🎨 Animation de chargement
- [ ] 🎨 Barre de progression
- [ ] 🎨 Aide contextuelle
- [ ] 🎨 Raccourcis clavier personnalisables

---

## 📦 Dépendances Go

```go
// go.mod
module github.com/bastiblast/inventory-manager

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.17.1
    golang.org/x/crypto v0.17.0         // SSH
    gopkg.in/yaml.v3 v3.0.1             // YAML
)
```

---

## 🚀 Installation & Utilisation

### Installation

```bash
# Compiler
cd boiler-deploy
go build -o bin/inventory-manager ./cmd/inventory-manager

# Ou via Makefile
make build

# Installation globale
make install
```

### Utilisation

```bash
# Lancer l'interface
./bin/inventory-manager

# Ou depuis n'importe où si installé
inventory-manager

# Mode CLI (non-interactif)
inventory-manager create production --web=3 --db=1
inventory-manager add server production web-01 192.168.1.10
inventory-manager validate production
inventory-manager export production
```

---

## 🎨 Avantages de Bubbletea vs Streamlit

| Aspect | Bubbletea (Go) | Streamlit (Python) |
|--------|----------------|-------------------|
| **Performance** | ⚡ Très rapide | 🐌 Plus lent |
| **Déploiement** | 📦 Binaire unique | 🐍 Python + dépendances |
| **Interface** | 🖥️ Terminal (TUI) | 🌐 Navigateur web |
| **Dépendances** | ✅ Aucune (binaire) | ❌ Python, pip, browser |
| **Portabilité** | ✅ Linux/Mac/Win | ⚠️ Nécessite Python |
| **Ressources** | 💚 < 10MB RAM | 💛 > 100MB RAM |
| **SSH Direct** | ✅ Natif | ⚠️ Via subprocess |
| **Offline** | ✅ Fonctionne | ⚠️ Besoin localhost:8501 |
| **Installation** | ✅ Copier binaire | ❌ pip install + setup |

**Verdict : Bubbletea est plus adapté pour un outil DevOps CLI.**

---

## 🔄 Comparaison avec le Script Bash

| Script Bash (1452 lignes) | Inventory Manager (Go) |
|---------------------------|------------------------|
| ❌ Difficile à maintenir | ✅ Structure modulaire |
| ❌ Erreurs cryptiques | ✅ Messages clairs |
| ❌ Pas de validation temps réel | ✅ Validation instantanée |
| ❌ Interface textuelle linéaire | ✅ Interface interactive |
| ❌ Pas de sauvegarde état | ✅ État persistant |
| ❌ Reprise difficile | ✅ Reprise automatique |

---

## 📝 Exemple de Code

### main.go (Simplifié)

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/bastiblast/inventory-manager/internal/ui"
)

func main() {
    p := tea.NewProgram(ui.NewMainMenu())
    
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

### ui/menu.go (Simplifié)

```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type MainMenu struct {
    choices []string
    cursor  int
}

func NewMainMenu() MainMenu {
    return MainMenu{
        choices: []string{
            "Créer un environnement",
            "Gérer un environnement",
            "Valider inventaires",
            "Quitter",
        },
    }
}

func (m MainMenu) Init() tea.Cmd {
    return nil
}

func (m MainMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }
        case "enter":
            // Action selon choix
            return m, nil
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m MainMenu) View() string {
    s := "╔══════════════════════════════╗\n"
    s += "║  Ansible Inventory Manager   ║\n"
    s += "╠══════════════════════════════╣\n\n"
    
    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        s += fmt.Sprintf(" %s %s\n", cursor, choice)
    }
    
    s += "\n[↑↓] Navigate [Enter] Select [q] Quit\n"
    return s
}
```

---

## 🎯 Prochaines Étapes

1. **Initialiser le projet Go** ✅
2. **Créer la structure de base** ✅
3. **Implémenter le menu principal** 🔄
4. **Développer les formulaires** 🔄
5. **Ajouter validation** 🔄
6. **Implémenter génération YAML** 🔄
7. **Tests** ⏳
8. **Documentation** ⏳

---

## 📚 Ressources

- **Bubbletea Docs**: https://github.com/charmbracelet/bubbletea
- **Lipgloss (Styling)**: https://github.com/charmbracelet/lipgloss
- **Bubbles (Components)**: https://github.com/charmbracelet/bubbles
- **Exemples**: https://github.com/charmbracelet/bubbletea/tree/master/examples

---

**Prêt à coder ! 🚀**
