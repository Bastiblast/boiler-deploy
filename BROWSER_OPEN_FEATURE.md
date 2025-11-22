# 🌐 Browser Auto-Open Feature

## Overview

La nouvelle fonctionnalité d'ouverture automatique du navigateur permet d'accéder rapidement au site web déployé directement depuis l'interface de déploiement (TUI Go) ou via le script shell `deploy.sh`.

## ✨ Fonctionnalités

### 1. **Interface Go (TUI) - Principal** ⭐

Après un déploiement réussi via l'interface Inventory Manager :

```
[server-name] ✓ Deployment successful! Site: http://192.168.1.100
Press 'o' to open in browser, or any key to continue
```

- Appuyez sur **`o`** pour ouvrir le site dans votre navigateur
- Appuyez sur n'importe quelle autre touche pour continuer sans ouvrir

**Fichiers modifiés :**
- `internal/ui/browser.go` (nouveau) - Fonction d'ouverture cross-platform
- `internal/ui/workflow_view.go` - Gestion de l'événement et affichage du prompt
- `internal/ansible/orchestrator.go` - Callback de succès de déploiement

### 2. **Script Shell (Tests uniquement)**

Le script `deploy.sh` a également été mis à jour pour offrir la même fonctionnalité lors des tests :

```bash
./deploy.sh deploy production
```

Après un déploiement réussi :
```
========================================
✅ Deployment completed successfully!
========================================

Access your application:
  → http://192.168.1.100

Open site in browser? (Y/n)
```

**Fichiers modifiés :**
- `deploy.sh` - Ajout de la fonction `open_browser()` et prompt interactif

## 🖥️ Support Multi-Plateforme

La fonctionnalité détecte automatiquement votre système d'exploitation :

| OS | Commande utilisée |
|---|---|
| Linux | `xdg-open` (défaut), `gnome-open` (fallback) |
| WSL (Windows Subsystem for Linux) | `wslview` |
| macOS | `open` |
| Windows | `rundll32 url.dll,FileProtocolHandler` |

## 📝 Architecture

### Orchestrator Callback

L'orchestrateur Ansible a été enrichi avec un callback de succès :

```go
type Orchestrator struct {
    // ... autres champs
    deploySuccessCb func(serverName, serverIP string)
}

func (o *Orchestrator) SetDeploySuccessCallback(cb func(serverName, serverIP string)) {
    o.deploySuccessCb = cb
}
```

### Workflow View Handler

La vue workflow gère l'événement de succès :

```go
func (wv *WorkflowView) onDeploySuccess(serverName, serverIP string) {
    wv.deployedServerIP = serverIP
    wv.showBrowserPrompt = true
    // ... affichage du message
}
```

### Browser Opener

Fonction cross-platform pour ouvrir le navigateur :

```go
func OpenBrowser(url string) error {
    // Détection OS et sélection de la commande appropriée
    switch runtime.GOOS {
    case "linux":   // xdg-open, gnome-open, wslview
    case "darwin":  // open
    case "windows": // rundll32
    }
}
```

## 🎯 Utilisation

### Dans l'interface TUI

1. Lancez l'inventory manager : `./bin/inventory-manager`
2. Sélectionnez un environnement
3. Déployez sur un ou plusieurs serveurs
4. Attendez la fin du déploiement
5. Quand le message "Press 'o' to open in browser" apparaît, appuyez sur **`o`**
6. Votre navigateur par défaut s'ouvre automatiquement

### Via le script shell (tests)

```bash
# Déploiement avec prompt interactif
./deploy.sh deploy production

# Déploiement automatisé sans prompt (mode CI/CD)
./deploy.sh deploy production --yes
```

## 🔧 Configuration

Aucune configuration nécessaire ! La fonctionnalité :
- ✅ Se désactive automatiquement en mode `--yes` (automation)
- ✅ Gère les erreurs si aucun navigateur n'est disponible
- ✅ Affiche des messages d'erreur clairs en cas de problème

## 📊 État des Changements

| Fichier | Type | Description |
|---------|------|-------------|
| `internal/ui/browser.go` | Nouveau | Fonction d'ouverture multi-plateforme |
| `internal/ui/workflow_view.go` | Modifié | Gestion du prompt et événement 'o' |
| `internal/ansible/orchestrator.go` | Modifié | Callback de succès de déploiement |
| `deploy.sh` | Modifié | Fonction shell et prompt interactif |
| `INVENTORY_MANAGER_README.md` | Mis à jour | Documentation de la fonctionnalité |

## 🚀 Compilation

```bash
# Build
make build

# Ou directement
go build -o bin/inventory-manager ./cmd/inventory-manager
```

## ✅ Tests

La compilation réussit sans erreurs :
```bash
✓ Build successful!
Binary size: 9.5MB
```

## 💡 Notes

- La fonctionnalité est **non-intrusive** : elle ne force jamais l'ouverture
- Compatible avec tous les workflows d'automatisation (CI/CD)
- Fonctionne uniquement après un déploiement réussi avec health check validé
- Respecte les préférences utilisateur (peut être ignorée)
