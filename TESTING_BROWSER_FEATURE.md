# 🧪 Testing Browser Open Feature

## Problème Résolu

Le callback de déploiement réussi était appelé depuis un goroutine différent, mais les modifications d'état dans Bubble Tea doivent se faire via le système de messages. 

**Solution :** Utilisation d'un canal Go (`deploySuccessChan`) pour communiquer entre le goroutine de l'orchestrateur et la boucle principale de Bubble Tea.

## Architecture

```
[Orchestrator Goroutine]
         ↓
    onDeploySuccess()
         ↓
   deploySuccessChan (buffered channel)
         ↓
waitForDeploySuccess() (Bubble Tea Cmd)
         ↓
   deploySuccessMsg (Bubble Tea Message)
         ↓
    Update() handles message
         ↓
  showBrowserPrompt = true
```

## Comment Tester

### 1. Préparer l'environnement

```bash
# Nettoyer le log de debug
> debug.log

# Dans un terminal, surveiller les logs
tail -f debug.log | grep -E 'BROWSER|WORKFLOW|Deploy'
```

### 2. Lancer l'application

```bash
./bin/inventory-manager
```

### 3. Effectuer un déploiement

1. Sélectionnez votre environnement
2. Choisissez "Workflow"
3. Sélectionnez un serveur (Espace pour sélectionner)
4. Appuyez sur `d` pour déployer

### 4. Observer les logs

Vous devriez voir dans `debug.log` :

```
[ORCHESTRATOR] Completed action: deploy for server docker-web-01
[WORKFLOW] onDeploySuccess callback called: serverName=docker-web-01, serverIP=192.168.1.100
[WORKFLOW] Deploy success message sent to channel
[WORKFLOW] Received deploy success from channel: docker-web-01 -> 192.168.1.100
[WORKFLOW] Processing deploySuccessMsg: docker-web-01 -> 192.168.1.100
[WORKFLOW] Browser prompt activated: deployedServerIP=192.168.1.100, showBrowserPrompt=true
```

### 5. Dans l'interface TUI

Vous devriez voir apparaître dans les logs (en bas de l'écran) :

```
[docker-web-01] ✓ Deployment successful! Site: http://192.168.1.100
Press 'o' to open in browser, or any key to continue
```

### 6. Ouvrir le navigateur

**Appuyez sur la touche `o`**

Les logs devraient montrer :

```
[WORKFLOW] 'o' key pressed. showBrowserPrompt=true, deployedServerIP=192.168.1.100
[WORKFLOW] Opening browser for URL: http://192.168.1.100
[BROWSER] Attempting to open URL: http://192.168.1.100
[BROWSER] Detected OS: linux
[BROWSER] Using xdg-open
[BROWSER] Successfully started browser command
[WORKFLOW] Browser opened successfully
```

**Votre navigateur par défaut devrait s'ouvrir !**

## Vérifications

### ✅ Le callback est appelé ?
```bash
grep "onDeploySuccess callback called" debug.log
```

### ✅ Le message est envoyé au canal ?
```bash
grep "Deploy success message sent to channel" debug.log
```

### ✅ Le message est reçu par Bubble Tea ?
```bash
grep "Received deploy success from channel" debug.log
```

### ✅ Le prompt est activé ?
```bash
grep "Browser prompt activated" debug.log
```

### ✅ La touche 'o' est détectée ?
```bash
grep "'o' key pressed" debug.log
```

### ✅ La commande browser est lancée ?
```bash
grep "Successfully started browser command" debug.log
```

## Debugging

Si le navigateur ne s'ouvre toujours pas :

### 1. Vérifier xdg-open manuellement
```bash
xdg-open "http://google.com"
```

### 2. Vérifier les permissions
```bash
ls -l /usr/bin/xdg-open
which xdg-open
```

### 3. Tester avec strace
```bash
strace -e trace=execve xdg-open "http://google.com" 2>&1 | grep execve
```

### 4. Vérifier $DISPLAY (pour X11)
```bash
echo $DISPLAY
# Devrait afficher quelque chose comme :0 ou :1
```

### 5. En WSL
Si vous êtes sous WSL, installez wslu :
```bash
sudo apt install wslu
# Puis le code utilisera wslview automatiquement
```

## Commandes de Test Rapides

```bash
# Test complet automatique
./test_browser.sh

# Surveiller les logs en temps réel
tail -f debug.log | grep -E --color 'BROWSER|WORKFLOW.*browser|deploySuccess'

# Vérifier si xdg-open fonctionne
timeout 3s xdg-open "http://example.com" && echo "✅ Browser opened" || echo "❌ Failed"
```

## Notes

- Le canal est **buffered** (taille 10) pour éviter les blocages
- Le callback est **non-bloquant** (utilise `select` avec `default`)
- Le listener se **réinscrit** automatiquement après chaque message
- Les logs sont **détaillés** pour faciliter le debugging
