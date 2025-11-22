# Test Parallel Execution - Guide Pratique

## 🎯 Objectif

Vérifier que l'exécution parallèle fonctionne correctement après avoir activé `max_parallel_workers`.

## ⚙️ Configuration

### 1. Vérifier la config

```bash
# Vérifier que max_parallel_workers est défini
cat inventory/docker/config.yml | grep max_parallel_workers
# Devrait afficher: max_parallel_workers: 3

cat inventory/test-multi/config.yml | grep max_parallel_workers
# Devrait afficher: max_parallel_workers: 3
```

### 2. Vérifier les valeurs

Les configs doivent contenir:

```yaml
provisioning_strategy: parallel
deployment_strategy: rolling
max_parallel_workers: 3  # ← Cette ligne est essentielle!
```

⚠️ **Important**: Si `max_parallel_workers` est absent ou à 0, le mode reste **séquentiel**.

## 🧪 Tests

### Test 1: Vérifier le chargement de la config

```bash
# Lancer l'app
make run  # ou ./bin/inventory-manager

# Dans l'interface, aller dans Settings ou vérifier les logs
```

**Logs attendus** dans `debug.log`:

```
[ORCHESTRATOR] Max workers set to 3 (0=sequential, >0=parallel)
[ORCHESTRATOR] Running in PARALLEL mode with 3 workers
```

❌ **Si vous voyez** `Running in SEQUENTIAL mode` → Config pas chargée!

### Test 2: Déploiement sur plusieurs serveurs

1. **Préparer 3+ serveurs** dans un environnement (ex: test-multi)
2. **Sélectionner tous les serveurs**
3. **Lancer "Deploy"**
4. **Observer les logs**

**Logs attendus en mode parallèle**:

```
[ORCHESTRATOR] Running in PARALLEL mode with 3 workers
[ORCHESTRATOR] Worker 0 started
[ORCHESTRATOR] Worker 1 started
[ORCHESTRATOR] Worker 2 started
[ORCHESTRATOR] Worker 0 processing: deploy for server web-01 (active: 1/3)
[ORCHESTRATOR] Worker 1 processing: deploy for server web-02 (active: 2/3)
[ORCHESTRATOR] Worker 2 processing: deploy for server web-03 (active: 3/3)
```

**Indicateurs de succès**:
- ✅ Plusieurs `Worker X processing` en même temps
- ✅ Compteur `(active: 2/3)` ou `(active: 3/3)`
- ✅ Actions se terminent plus rapidement

**Indicateurs d'échec** (mode séquentiel):
- ❌ Une seule action à la fois
- ❌ Pas de mention de "Worker"
- ❌ `Running in SEQUENTIAL mode`

### Test 3: Mesurer la performance

**Setup**:
- 9 serveurs
- Action qui prend ~5 minutes par serveur

**Résultats attendus**:

| Mode | Workers | Temps total | Calcul |
|------|---------|-------------|--------|
| Séquentiel | 0 | 45 minutes | 9 × 5min |
| Parallèle | 3 | 15 minutes | (9/3) × 5min |
| Parallèle | 5 | 10 minutes | (9/5) × 5min + overhead |

**Commande pour chronométrer**:

```bash
time (deploy_action_here)
```

## 🔍 Debugging

### Problème: Mode séquentiel malgré config

**Diagnostic**:

```bash
# 1. Vérifier la config
cat inventory/YOUR_ENV/config.yml | grep max_parallel_workers

# 2. Vérifier les logs au démarrage
grep "Max workers set to" debug.log | tail -1

# 3. Vérifier le mode d'exécution
grep "Running in.*mode" debug.log | tail -1
```

**Solutions**:

1. **Config manquante**:
   ```bash
   # Ajouter à inventory/YOUR_ENV/config.yml
   max_parallel_workers: 3
   ```

2. **Config pas rechargée**:
   ```bash
   # Redémarrer l'app
   # Ou recréer l'orchestrator
   ```

3. **Valeur à 0**:
   ```bash
   # Modifier la valeur
   max_parallel_workers: 3  # Au lieu de 0
   ```

### Problème: Workers ne démarrent pas

**Vérifier**:

```bash
# Logs de démarrage des workers
grep "Worker.*started" debug.log

# Si vide, les workers ne sont pas créés
```

**Causes possibles**:
- Queue vide (pas d'actions en attente)
- Orchestrator.Start() pas appelé
- Context cancelled prématurément

### Problème: Une seule action à la fois

**Diagnostic**:

```bash
# Vérifier le compteur active
grep "active:" debug.log | tail -10

# Si toujours (active: 1/3), une seule action traitée
```

**Causes possibles**:
- Queue.Next() au lieu de traitement parallèle
- Channel bloqué
- Mutex contention

## 📊 Monitoring

### Métriques clés

1. **Nombre de workers actifs**:
   ```bash
   grep "active:" debug.log | tail -20
   ```

2. **Temps par action**:
   ```bash
   grep "Worker.*completed" debug.log | awk '{print $1, $2, $9, $11, $12}'
   ```

3. **Throughput**:
   ```bash
   # Nombre d'actions complétées en 1 minute
   grep "Worker.*completed" debug.log | grep "$(date +%H:%M)" | wc -l
   ```

### Logs utiles

```bash
# Voir le pipeline complet d'une action
grep "web-01" debug.log | grep -E "(Queueing|processing|completed)"

# Voir tous les workers actifs
grep "Worker" debug.log | tail -50

# Voir les changements de mode
grep "Running in" debug.log
```

## ✅ Validation

**Checklist pour confirmer le mode parallèle**:

- [ ] `max_parallel_workers > 0` dans config.yml
- [ ] Log: `Max workers set to N` (N > 0)
- [ ] Log: `Running in PARALLEL mode with N workers`
- [ ] Log: `Worker 0/1/2... started`
- [ ] Log: Multiple `Worker X processing` simultanés
- [ ] Compteur `(active: 2/3)` ou plus
- [ ] Performance améliorée vs séquentiel

## 📈 Benchmarks

### Configuration recommandée par taille d'infra

| Serveurs | Workers | Justification |
|----------|---------|---------------|
| 1-3 | 0 (seq) | Overhead pas justifié |
| 4-10 | 3 | Optimal pour petite infra |
| 11-20 | 5 | Balance perf/resources |
| 21-50 | 7-10 | Grande infra |
| 50+ | 10+ | Attention aux limites SSH/network |

### Limites système

- **SSH connections**: Max ~100 simultanées (selon OS)
- **Network bandwidth**: Limite réelle souvent
- **Ansible processes**: Chaque worker = 1 processus

## 🚀 Exemples

### Exemple 1: Test simple (3 serveurs, 3 workers)

```bash
# 1. Config
echo "max_parallel_workers: 3" >> inventory/docker/config.yml

# 2. Lancer app
make run

# 3. Deploy sur 3 serveurs
# Observer: 3 workers traitent en parallèle

# Logs attendus:
# Worker 0 processing: deploy for server web-01
# Worker 1 processing: deploy for server web-02
# Worker 2 processing: deploy for server web-03
```

### Exemple 2: Test charge (9 serveurs, 3 workers)

```bash
# Config: max_parallel_workers: 3

# Déploiement sur 9 serveurs
# Vagues attendues:
# - Vague 1: web-01, web-02, web-03 (parallèle)
# - Vague 2: web-04, web-05, web-06 (parallèle)
# - Vague 3: web-07, web-08, web-09 (parallèle)

# Temps total ≈ 3 × temps_par_serveur
```

### Exemple 3: Mesure performance

```bash
# Test séquentiel
sed -i 's/max_parallel_workers: 3/max_parallel_workers: 0/' inventory/docker/config.yml
time deploy_action  # Note le temps

# Test parallèle
sed -i 's/max_parallel_workers: 0/max_parallel_workers: 3/' inventory/docker/config.yml
time deploy_action  # Compare le temps

# Calcul du gain
# Gain = (temps_seq - temps_para) / temps_seq × 100
```

## 📝 Rapport de test

**Template pour documenter les résultats**:

```markdown
### Test Parallel Execution - [DATE]

**Configuration**:
- Environment: docker / test-multi / production
- Servers: 9
- Workers: 3
- Action: deploy

**Résultats**:
- Mode: Parallel ✅ / Sequential ❌
- Temps total: 15 minutes
- Temps par serveur: 5 minutes
- Workers actifs: 3/3
- Gain vs séquentiel: 66%

**Logs clés**:
```
[ORCHESTRATOR] Running in PARALLEL mode with 3 workers
[ORCHESTRATOR] Worker 0 processing: deploy for server web-01 (active: 3/3)
```

**Conclusion**: Exécution parallèle fonctionnelle ✅
```

---

**Dernière mise à jour**: 2025-11-22  
**Version**: 1.0  
**Maintenu par**: Boiler Expert Team
