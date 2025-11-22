# Dry-Run Mode (Check Mode)

## Vue d'ensemble

Le mode dry-run a été implémenté pour améliorer la sécurité et la fiabilité des déploiements. Ansible exécute maintenant automatiquement une vérification en mode `--check --diff` avant chaque action réelle.

## Fonctionnement

### Provision

1. **Dry-run check** : Ansible vérifie ce qui serait changé sans modifier le système
   - Status: `Provisioning - Checking configuration (dry-run)...`
   - Options: `--check --diff`
   - Si le check échoue → Action annulée

2. **Exécution réelle** : Si le dry-run réussit, provisioning réel
   - Status: `Provisioning - Applying configuration...`
   - Application des changements

### Deploy

1. **Dry-run check** : Vérification du déploiement
   - Status: `Deploying - Checking deployment (dry-run)...`
   - Vérifie que toutes les ressources sont disponibles
   - Si le check échoue → Action annulée

2. **Exécution réelle** : Si le dry-run réussit, déploiement réel
   - Status: `Deploying - Deploying application...`
   - Déploiement de l'application

## Avantages

### 1. Détection précoce des erreurs
- Les problèmes de configuration sont détectés avant toute modification
- Évite les changements partiels qui peuvent casser le système

### 2. Idempotence garantie
- Ansible affiche clairement ce qui serait changé
- Si rien ne doit changer, le dry-run le montre
- Évite de relancer des installations inutiles

### 3. Sécurité accrue
- Aucune modification n'est appliquée si le dry-run échoue
- Réduit le risque de mettre le serveur dans un état invalide

### 4. Logs détaillés
- Les logs du dry-run sont sauvegardés avec le suffixe `_check`
- Format: `{server}_{action}_check_{timestamp}.log`
- Les logs réels restent séparés

## Différence avec l'approche précédente

### Avant
```
Provision → Ansible installe tout même si déjà installé
→ "changed" sur des tâches déjà appliquées
→ Perte de temps et confusion
```

### Maintenant
```
Provision → Dry-run check (--check --diff)
         → Si OK → Provision réelle
         → Seules les tâches nécessaires sont exécutées
         → "ok" pour les tâches déjà appliquées
         → "changed" seulement si modification nécessaire
```

## Exemple de sortie

### Dry-run détecte un problème
```
Status: Provisioning - Checking configuration (dry-run)...
🚀 Starting provision playbook (dry-run mode)...
⚙️  Collecting server information
  ❌ Error: SSH connection failed
Status: Failed - Dry-run check failed
```

### Dry-run réussit, exécution réelle
```
Status: Provisioning - Checking configuration (dry-run)...
🚀 Starting provision playbook (dry-run mode)...
✅ provision completed successfully

Status: Provisioning - Applying configuration...
🚀 Starting provision playbook...
⚙️  Updating package list
  ✓ Modified on docker-web-01
✅ provision completed successfully
Status: Provisioned
```

## Fichiers modifiés

### internal/ansible/executor.go
- Nouvelle fonction `RunPlaybookWithOptions()` avec paramètre `checkMode`
- Ajout de `--check --diff` quand `checkMode=true`
- Nouvelles fonctions `ProvisionCheck()` et `DeployCheck()`
- Les logs du dry-run ont le suffixe `_check`

### internal/ansible/orchestrator.go
- Exécution automatique du dry-run avant chaque action
- Si dry-run échoue → action annulée
- Si dry-run réussit → exécution réelle

## Notes techniques

### Flags Ansible utilisés
- `--check` : Mode dry-run, ne modifie rien
- `--diff` : Affiche les différences qui seraient appliquées

### Limitations
Certains modules Ansible ne supportent pas le mode `--check` :
- Commandes shell/command personnalisées
- Certaines actions de fichiers complexes

Dans ces cas, le module devrait avoir `check_mode: false` dans le playbook pour être ignoré pendant le dry-run.

## Prochaines améliorations possibles

1. **Afficher le diff dans l'UI** : Parser la sortie `--diff` pour montrer exactement ce qui changerait
2. **Mode manuel** : Option pour désactiver le dry-run automatique
3. **Statistiques** : Compter combien de tâches seraient modifiées
4. **Confirmation utilisateur** : Demander confirmation après le dry-run avant l'exécution réelle
