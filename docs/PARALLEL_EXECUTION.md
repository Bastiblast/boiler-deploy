# Parallel Action Execution

## Overview

Cette fonctionnalité permet d'exécuter plusieurs actions (provision/deploy/check) en parallèle sur différents serveurs, accélérant significativement les déploiements multi-serveurs.

## Configuration

### Fichier de configuration

Ajoutez dans `inventory/<environment>/config.yml`:

```yaml
max_parallel_workers: 3  # Nombre de workers parallèles (0 = séquentiel)
```

### Valeurs recommandées

- **0**: Mode séquentiel (comportement par défaut)
- **3-5**: Optimal pour la plupart des infrastructures
- **10+**: Pour infrastructures très larges (attention aux limites réseau/SSH)

## Architecture

### Mode Séquentiel (workers = 0)

```
Queue → Action 1 → Action 2 → Action 3
        [Execute] [Execute] [Execute]
```

### Mode Parallèle (workers = 3)

```
Queue → [Worker 1] → Action 1
     ↘ [Worker 2] → Action 2
     ↘ [Worker 3] → Action 3
```

## Implémentation

### Orchestrator

**Nouveaux champs:**
- `maxWorkers`: Nombre de workers (0 = séquentiel)
- `activeWorkers`: Compteur de workers actifs
- `workersMu`: Mutex pour la synchronisation

**Nouvelles méthodes:**
- `SetMaxWorkers(workers int)`: Configure le nombre de workers
- `processQueueParallel()`: Traite la queue avec worker pool
- `processQueueSequential()`: Mode séquentiel original

### Queue

**Nouvelles méthodes:**
- `NextBatch(count int)`: Récupère N actions
- `CompleteByID(id string)`: Complète une action par ID (thread-safe)

### Workflow

La configuration est chargée automatiquement au démarrage:

```go
orchestrator.SetMaxWorkers(configOpts.MaxParallelWorkers)
```

## Gestion de la concurrence

### Thread-safety

- **Queue**: Mutex RWLock pour toutes les opérations
- **Status Manager**: Mutex pour les updates
- **Worker counter**: Mutex dédié pour `activeWorkers`

### Logs

Chaque worker log ses actions:

```
[ORCHESTRATOR] Worker 1 processing: deploy for server web-01 (active: 2/3)
[ORCHESTRATOR] Worker 2 processing: deploy for server web-02 (active: 3/3)
```

## Tests

### Test manuel

1. Configurer 3+ serveurs dans un environnement
2. Activer parallel execution: `max_parallel_workers: 3`
3. Lancer provision/deploy sur tous les serveurs
4. Observer les logs pour vérifier l'exécution parallèle

### Comportement attendu

- **Séquentiel**: Actions se terminent une par une
- **Parallèle**: Plusieurs actions en cours simultanément

## Limitations

### Actuelles

1. **Ansible limitation**: Chaque worker lance un processus ansible séparé
2. **SSH connections**: Limité par les ressources système
3. **Network bandwidth**: Peut être un goulot d'étranglement

### Recommandations

- Ne pas dépasser 10 workers sauf infrastructure dédiée
- Surveiller la charge système lors de déploiements parallèles
- Utiliser tags pour limiter les actions lourdes

## Fallbacks

Le code gère automatiquement:
- **Workers < 0**: Reset à 0 (séquentiel)
- **Queue vide**: Sleep 100ms, retry
- **Stop signal**: Graceful shutdown de tous les workers

## Monitoring

### Logs clés

```bash
[ORCHESTRATOR] Running in PARALLEL mode with 3 workers
[ORCHESTRATOR] Worker 1 started
[ORCHESTRATOR] Worker 1 processing: deploy for server web-01 (active: 2/3)
[ORCHESTRATOR] Worker 1 completed: deploy for server web-01
```

### Métriques

- **Active workers**: Visible dans les logs
- **Queue size**: `orchestrator.GetQueueSize()`
- **Completion time**: Comparer avec mode séquentiel

## Examples

### Configuration minimale

```yaml
# inventory/production/config.yml
max_parallel_workers: 3
```

### Configuration avancée

```yaml
# inventory/production/config.yml
max_parallel_workers: 5
health_check_enabled: true
health_check_timeout: 30s
```

### Utilisation programmatique

```go
// Création orchestrator
orchestrator, _ := ansible.NewOrchestrator(env, statusMgr)
orchestrator.SetMaxWorkers(5)

// Queue actions
orchestrator.QueueDeploy(serverNames, 1)
orchestrator.Start(servers)
```

## Performances

### Gains attendus

Pour N serveurs avec T temps de déploiement:

- **Séquentiel**: N × T
- **Parallèle (W workers)**: (N / W) × T

**Exemple:** 9 serveurs, 3 workers, 5min/serveur
- Séquentiel: 45 minutes
- Parallèle: 15 minutes (gain de 66%)

## Migration

### De séquentiel vers parallèle

1. **Tester sur environnement de dev**
2. **Commencer avec 2-3 workers**
3. **Surveiller les logs et performances**
4. **Augmenter progressivement si nécessaire**

### Rollback

Mettre `max_parallel_workers: 0` dans la config.

---

**Version:** 1.0  
**Date:** 2025-11-22  
**Auteur:** Boiler Expert Agent
