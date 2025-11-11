# Corrections des problèmes de freeze et crash

## Problèmes identifiés

### 1. Freeze lors de la validation (touche 'v')
**Cause** : La validation s'exécutait de manière synchrone dans le thread principal de l'UI (via `tea.Cmd`), bloquant l'interface pendant l'opération.

**Solution** : Déplacement de la validation dans une goroutine séparée pour exécution asynchrone, permettant à l'UI de rester réactive.

### 2. Crash lors du check (touche 'c')
**Causes multiples** :
- Utilisation du port `AppPort` (3000) au lieu du port externe (80)
- L'application Node.js tourne sur le port 3000 en interne, mais est exposée via Nginx sur le port 80
- Aucun feedback visuel immédiat lors du lancement du check

**Solutions** :
- Port 80 utilisé pour le health check (accès externe via Nginx)
- Mise à jour immédiate du statut à "Verifying" avant le lancement du check
- Ajout de logs détaillés pour tracer l'exécution

### 3. Busy loop dans le processeur de queue
**Cause** : Boucle infinie sans pause quand la queue est vide, consommant inutilement du CPU.

**Solution** : Ajout d'un `time.Sleep(100ms)` quand la queue est vide.

## Fichiers modifiés

### `internal/ui/workflow_view.go`
- **Ligne 182-208** : Validation asynchrone dans goroutine
- **Ligne 198-213** : Mise à jour immédiate du statut avant check
- **Suppression** : Fonction `validateSelectedCmd()` devenue inutile

### `internal/ansible/orchestrator.go`
- **Ligne 3-9** : Ajout import `time`
- **Ligne 92-112** : Ajout sleep dans la boucle de traitement
- **Ligne 170-189** : Correction port 80 pour health check et amélioration gestion statuts

## Comportement après corrections

### Validation ('v')
1. Presse 'v' → UI reste réactive immédiatement
2. Validation s'exécute en arrière-plan
3. Statuts mis à jour progressivement (Ready/Not Ready)
4. Logs détaillés dans la console

### Check ('c')
1. Presse 'c' → Statut passe immédiatement à "Verifying"
2. Check ajouté à la queue et exécuté par l'orchestrator
3. Health check via `curl -sf -m 5 http://IP:80/`
4. Résultat : Deployed (succès) ou Failed (échec)
5. Logs détaillés de chaque étape

## Logs de débogage

Les logs suivent maintenant chaque étape :
```
[WORKFLOW] Key 'c' pressed - starting check
[WORKFLOW] Checking 1 selected servers: [bast-web-01]
[WORKFLOW] Setting bast-web-01 to Verifying state
[ORCHESTRATOR] Adding check action for server: bast-web-01
[ORCHESTRATOR] Queue size after adding checks: 1
[ORCHESTRATOR] Processing action: Check for server bast-web-01
[ORCHESTRATOR] Executing check for bast-web-01 (IP: 127.0.0.1, Port: 80)
[EXECUTOR] Running health check: curl -sf -m 5 http://127.0.0.1:80/
[EXECUTOR] Health check successful for http://127.0.0.1:80/
[ORCHESTRATOR] Health check PASSED for bast-web-01
```

## Tests recommandés

1. **Test validation** :
   - Sélectionner un serveur (espace)
   - Appuyer sur 'v'
   - Vérifier que l'UI ne freeze pas
   - Observer le changement de statut

2. **Test check** :
   - Sélectionner un serveur (espace)
   - Appuyer sur 'c'
   - Vérifier changement immédiat à "Verifying"
   - Observer le résultat final (Deployed/Failed)

3. **Test multiple** :
   - Sélectionner plusieurs serveurs ('a')
   - Tester validation et check
   - Vérifier traitement séquentiel

## Notes importantes

- **Port 80** : Utilisé pour tous les health checks (accès externe)
- **Port 3000** : Port interne de l'app Node.js (ne pas utiliser pour checks)
- **Asynchrone** : Toutes les opérations longues doivent être asynchrones
- **Feedback immédiat** : Le statut doit changer AVANT l'opération
