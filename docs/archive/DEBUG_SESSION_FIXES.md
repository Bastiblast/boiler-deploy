# Session de Debug - 11 Novembre 2025

## Problèmes identifiés et résolus

### 1. ✅ Validation d'inventaire ne s'affiche pas
**Problème** : La validation ('v') ne donnait aucun feedback visuel
**Solution** : Ajout d'un feedback immédiat avant l'exécution de la validation en goroutine
- Mise à jour du status à "Validating..." immédiatement
- Refresh des statuts pour affichage instantané

**Fichiers modifiés** :
- `internal/ui/workflow_view.go` (ligne 178-198)

### 2. ✅ Check plante l'application
**Problème** : Le check ('c') faisait planter l'app car il utilisait le mauvais port
**Solution** : 
- Détection automatique localhost vs serveur distant
- Utilisation du port applicatif (AppPort) pour localhost
- Utilisation du port 80 (nginx) pour serveurs distants

**Fichiers modifiés** :
- `internal/ansible/orchestrator.go` (ligne 187-207)

### 3. ✅ Provisioning échoue sur timezone
**Problème** : `Europe/Paris` non disponible dans conteneurs légers
**Solution** : Rendre le timezone optionnel
- Condition `when: timezone is defined` dans role common
- Modification du générateur pour ne pas inclure timezone vide

**Fichiers modifiés** :
- `internal/inventory/generator.go` (ligne 94-106)
- `inventory/docker/group_vars/all.yml`

### 4. ✅ Variable deploy_user_groups manquante
**Problème** : Variable non définie dans role common
**Solution** : Création d'un fichier defaults pour le role common

**Fichiers créés** :
- `roles/common/defaults/main.yml`

### 5. ✅ SSH key obligatoire bloque le provisioning
**Problème** : La tâche "Add SSH key" cherchait toujours un fichier
**Solution** : Rendre la tâche conditionnelle avec `when: ssh_key_path is defined`

**Fichiers modifiés** :
- `roles/common/tasks/main.yml` (ligne 51-56)

### 6. ✅ UFW échoue dans conteneurs Docker
**Problème** : iptables nécessite des privilèges spéciaux dans Docker
**Solution** : Ajout d'une variable `enable_firewall` (défaut: true)
- Désactivation conditionnelle pour environnements Docker
- Toutes les tâches UFW et fail2ban sont conditionnelles

**Fichiers créés** :
- `roles/security/defaults/main.yml`

**Fichiers modifiés** :
- `roles/security/tasks/main.yml` (toutes les tâches UFW/fail2ban)
- `inventory/docker/group_vars/all.yml`

### 7. ✅ Handler systemd incompatible avec Docker
**Problème** : systemd non disponible dans conteneurs
**Solution** : 
- Changement de `systemd` vers `service` (plus compatible)
- Ajout de `ignore_errors: yes` pour SSH handler
- Condition sur fail2ban handler

**Fichiers modifiés** :
- `roles/security/handlers/main.yml`

### 8. ✅ Variable app_dir manquante dans nodejs role
**Problème** : Variable requise mais non définie
**Solution** : Création defaults pour role nodejs

**Fichiers créés** :
- `roles/nodejs/defaults/main.yml`

## État actuel

### ✅ Fonctionnel
1. Validation d'inventaire avec feedback visuel
2. Build de l'application Go sans erreurs
3. Génération d'inventaire Ansible
4. Provisioning jusqu'au role nodejs (en cours)

### 🔄 En cours de test
1. Provisioning complet du conteneur Docker
2. Déploiement de l'application

### 📋 À tester
1. Check de santé post-déploiement
2. Workflow complet : validate → provision → deploy → check
3. Multi-environnements
4. Logs Ansible en format JSON

## Logs ajoutés pour debugging

- `[ORCHESTRATOR]` : Orchestration des actions
- `[EXECUTOR]` : Exécution des playbooks et health checks
- `[STATUS]` : Gestion des statuts de serveurs

## Commandes de test

```bash
# Build
make build

# Test provisioning Docker
ansible-playbook -i inventory/docker/hosts.yml playbooks/provision.yml --limit docker-web-01

# Lancer l'app
make run

# Vérifier les logs
tail -f logs/docker/*.log
```

## Notes pour la suite

1. **Conteneurs de test** : Bien configurer `enable_firewall: false` pour éviter erreurs iptables
2. **SSH keys** : Optionnels pour tests, requis pour production
3. **Timezone** : Laisser vide pour conteneurs, définir pour serveurs réels
4. **Ports** : L'app détecte automatiquement localhost vs remote pour le health check

## Fichiers de configuration Docker actuels

- Environnement: `docker`
- Serveur: `docker-web-01` (127.0.0.1:2222)
- App: https://github.com/Bastiblast/portefolio
- Node: 20
- Firewall: désactivé
- SSH: root via ~/.ssh/boiler_test_rsa
