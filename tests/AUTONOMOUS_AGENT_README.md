# Autonomous Agent Testing

## 🤖 Overview

Ce dossier contient des tests autonomes pour valider les fonctionnalités de boiler-deploy automatiquement. L'agent autonome peut être lancé localement ou dans une CI/CD.

## 📋 Tests Disponibles

### Parallel Execution Test

**Script:** `autonomous-agent-test.sh`  
**PR:** [#2 - Parallel action execution](https://github.com/Bastiblast/boiler-deploy/pull/2)

Teste automatiquement la feature d'exécution parallèle des actions.

**Tests inclus (8/8):**

1. ✅ **Compilation** - Vérifie que tous les packages compilent
2. ✅ **Nouvelles méthodes** - Vérifie présence de SetMaxWorkers, processQueueParallel, etc.
3. ✅ **Configuration** - Vérifie MaxParallelWorkers dans config
4. ✅ **Intégration UI** - Vérifie que SetMaxWorkers est appelé
5. ✅ **Thread-safety** - Vérifie mutexes et WaitGroup
6. ✅ **Documentation** - Vérifie sections requises
7. ✅ **Backward compatibility** - Vérifie mode séquentiel préservé
8. ✅ **Performance** - Vérifie documentation des gains

## 🚀 Utilisation

### Lancement local

```bash
# Depuis la racine du projet
./tests/autonomous-agent-test.sh

# Ou avec bash explicite
bash tests/autonomous-agent-test.sh
```

### Résultat attendu

```
🤖 Autonomous Agent Test - Parallel Execution Feature
==========================================

ℹ️  Starting autonomous agent tests...

✅ Test 1 PASSED: All packages compile successfully
✅ Test 2 PASSED: All new methods present
✅ Test 3 PASSED: Configuration properly defined
✅ Test 4 PASSED: UI integration verified
✅ Test 5 PASSED: Thread-safety mechanisms in place
✅ Test 6 PASSED: Documentation complete
✅ Test 7 PASSED: Backward compatible
✅ Test 8 PASSED: Performance documented

==========================================

✅ Passed: 8/8

✅ 🎉 ALL TESTS PASSED!
```

## 🔄 CI/CD Integration

### GitHub Actions

Le workflow `.github/workflows/test-parallel-execution.yml` s'exécute automatiquement sur chaque PR modifiant:
- `internal/ansible/orchestrator.go`
- `internal/ansible/queue.go`
- `internal/config/types.go`
- `internal/ui/workflow_view.go`

**Jobs exécutés:**
1. `test-compilation` - Build des packages
2. `test-sequential-mode` - Test backward compatibility
3. `test-parallel-mode` - Test mode parallèle
4. `code-quality` - gofmt et go vet
5. `documentation` - Vérifie docs complètes
6. `summary` - Résumé global

### Déclenchement manuel

```bash
# Via gh CLI
gh workflow run test-parallel-execution.yml

# Via interface GitHub
Actions > Test Parallel Execution > Run workflow
```

## 📊 Métriques

L'agent autonome mesure:
- **Code coverage** - Via présence de méthodes
- **Configuration** - Via grep dans fichiers config
- **Documentation** - Via sections requises
- **Compilation** - Via go build
- **Thread-safety** - Via patterns de synchronisation

## 🛠️ Création d'un nouveau test autonome

### Template

```bash
#!/bin/bash
set -euo pipefail

# Your test function
test_feature() {
    log_info "Test 1: Description"
    
    # Perform checks
    if condition; then
        log_success "Check passed"
        return 0
    else
        log_error "Check failed"
        return 1
    fi
}

# Run tests
run_all_tests() {
    tests=("test_feature")
    
    for test in "${tests[@]}"; do
        set +e
        $test
        result=$?
        set -e
        # Handle result
    done
}

main() {
    run_all_tests
}

main
```

### Bonnes pratiques

1. **Atomicité** - Chaque test doit être indépendant
2. **Logs clairs** - Utiliser log_info, log_success, log_error
3. **Exit codes** - 0 = succès, 1 = échec
4. **Performance** - Tests rapides (<60s total)
5. **Documentation** - Expliquer ce qui est testé

## 🔍 Debugging

### Test qui échoue

```bash
# Mode verbose
bash -x tests/autonomous-agent-test.sh

# Test spécifique
bash tests/autonomous-agent-test.sh 2>&1 | grep "Test 3"
```

### Logs GitHub Actions

```bash
# Via gh CLI
gh run list
gh run view <run_id>
gh run view <run_id> --log
```

## 📚 Ressources

- [PR #2 - Parallel Execution](https://github.com/Bastiblast/boiler-deploy/pull/2)
- [Documentation technique](../docs/PARALLEL_EXECUTION.md)
- [GitHub Actions Workflow](../.github/workflows/test-parallel-execution.yml)

## 🎯 Prochaines étapes

Pour ajouter un nouveau test autonome:

1. **Créer le script** dans `tests/`
2. **Ajouter workflow GitHub Actions** dans `.github/workflows/`
3. **Documenter** dans ce README
4. **Tester localement** avant de push
5. **Créer PR** avec tests intégrés

## ✅ Checklist nouveau test

- [ ] Script exécutable (`chmod +x`)
- [ ] Shebang `#!/bin/bash`
- [ ] Set flags (`set -euo pipefail`)
- [ ] Fonctions de logging (log_info, log_success, log_error)
- [ ] Tests indépendants
- [ ] Exit codes corrects
- [ ] Documentation dans README
- [ ] Workflow GitHub Actions (optionnel)
- [ ] Testé localement
- [ ] PR créée

---

**Maintenu par:** Boiler Expert Agent  
**Dernière mise à jour:** 2025-11-22
