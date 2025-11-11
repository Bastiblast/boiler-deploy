# Bug Fix Summary - Validation Freeze & Check Crash

## Issues Reported

1. **Validation Freeze**: L'application se fige quand on sélectionne un serveur (avec ESPACE) et qu'on appuie sur 'v' pour valider
2. **Check Crash**: L'application plante quand on appuie sur 'c' pour faire un check

## Solution Appliquée

### 1. Ajout de Logs Détaillés

J'ai ajouté des logs de debug complets dans tous les composants pour tracer l'exécution :

- **Workflow View** : Actions utilisateur, sélections, flux de validation
- **Orchestrator** : Traitement de la queue, exécution des actions
- **Queue** : Ajout et complétion des actions
- **Status Manager** : Mises à jour de statut, vérifications de validation
- **Executor** : Exécution des health checks avec sortie complète

Tous les logs sont écrits dans `debug.log` avec timestamps, microseconds et emplacements de fichiers.

### 2. Corrections de Code

**Validation (touche 'v')** :
- Vérification si des serveurs sont sélectionnés avant de valider
- Logs pour tracer chaque étape du processus de validation
- Gestion d'erreurs améliorée lors de la sauvegarde des statuts

**Check (touche 'c')** :
- Vérification si des serveurs sont sélectionnés avant de lancer le check
- Démarrage automatique de l'orchestrator s'il n'est pas en cours d'exécution
- Capture de la sortie complète de curl pour le debugging
- Meilleure gestion des erreurs de health check

### 3. Améliorations Générales

- Null safety : Vérifications ajoutées pour éviter les nil pointer dereferences
- Feedback utilisateur : Les logs montrent clairement ce qui se passe
- Gestion des erreurs : Toutes les erreurs sont loggées avec contexte

## Comment Tester

### Test Manuel

1. **Démarrer l'application** :
   ```bash
   make run
   ```

2. **Dans un autre terminal, surveiller les logs** :
   ```bash
   tail -f debug.log
   ```

3. **Tester la validation** :
   - Aller à "Working with Inventory"
   - Sélectionner un serveur avec ESPACE
   - Appuyer sur 'v'
   - Vérifier les logs pour voir le flux d'exécution

4. **Tester le check** :
   - Sélectionner un serveur avec ESPACE
   - Appuyer sur 'c'
   - Vérifier les logs pour voir l'exécution du health check

### Script de Test

Un script de test a été créé pour faciliter le debugging :

```bash
./test_workflow.sh
```

Ce script :
- Nettoie le fichier debug.log
- Affiche les instructions
- Lance l'application
- Permet de consulter facilement les logs après test

## Prochaines Étapes

### Si la Validation Freeze Toujours

Vérifier dans les logs :
1. Est-ce que `validateSelectedCmd()` est appelé ?
2. Est-ce que `validationCompleteMsg` est reçu ?
3. Y a-t-il des erreurs dans `UpdateReadyChecks` ?
4. Le I/O fichier bloque-t-il ? (vérifier espace disque, permissions)

Regarder spécifiquement les lignes avec `[WORKFLOW]` et `[STATUS]`.

### Si le Check Plante Toujours

Vérifier dans les logs :
1. Où exactement ça plante ? (dernière entrée de log avant le crash)
2. L'orchestrator est-il démarré ?
3. La queue est-elle correctement initialisée ?
4. Y a-t-il des problèmes de nil pointer ?

Regarder spécifiquement les lignes avec `[ORCHESTRATOR]`, `[QUEUE]`, et `[EXECUTOR]`.

## Commandes de Debugging Utiles

```bash
# Surveiller les logs en temps réel
tail -f debug.log

# Filtrer par composant
grep "\[WORKFLOW\]" debug.log
grep "\[ORCHESTRATOR\]" debug.log
grep "\[QUEUE\]" debug.log
grep "\[STATUS\]" debug.log
grep "\[EXECUTOR\]" debug.log

# Trouver les erreurs
grep -i "error\|failed\|panic" debug.log

# Voir les 50 dernières lignes
tail -50 debug.log

# Nettoyer le log avant de tester
> debug.log
```

## Séquences de Log Attendues

### Validation Réussie

```
[WORKFLOW] Key 'v' pressed - starting validation
[WORKFLOW] Validating 1 selected servers
[STATUS] UpdateReadyChecks for bast-web-01: IP=true SSH=true Port=true Fields=true
[STATUS] Server bast-web-01 is ready, updating state to Ready
[WORKFLOW] Validation complete, sending message
```

### Check Réussi (avec service non accessible)

```
[WORKFLOW] Key 'c' pressed - starting check
[ORCHESTRATOR] QueueCheck called with 1 servers: [bast-web-01]
[QUEUE] Adding action: check for server bast-web-01
[ORCHESTRATOR] Executing check for bast-web-01
[EXECUTOR] Running health check: curl http://127.0.0.1:3000/
[EXECUTOR] Health check failed: curl: (7) Failed to connect...
[STATUS] Updating status: state=failed
```

## Documentation Complète

Pour plus de détails, consulter `DEBUG_GUIDE.md` qui contient :
- Analyse détaillée des problèmes
- Toutes les modifications apportées
- Instructions de test complètes
- Patterns d'erreurs courants
- Améliorations futures recommandées

## Fichiers Modifiés

- `cmd/inventory-manager/main.go` : Ajout du système de logging
- `internal/ui/workflow_view.go` : Logs + vérifications de sélection
- `internal/ansible/orchestrator.go` : Logs + auto-démarrage
- `internal/ansible/queue.go` : Logs détaillés de queue
- `internal/ansible/executor.go` : Logs de health check avec sortie complète
- `internal/status/manager.go` : Logs de mise à jour de statut

## Nouveaux Fichiers

- `DEBUG_GUIDE.md` : Guide complet de debugging
- `test_workflow.sh` : Script de test interactif
- `debug.log` : Fichier de logs (créé automatiquement)

---

**Note** : Les logs sont maintenant très verbeux pour faciliter le debugging. Une fois les problèmes résolus, on pourra réduire le niveau de logging ou ajouter un flag de debug.
