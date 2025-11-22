# Ansible Tags Guide

## Overview

Le système de tags Ansible permet d'exécuter des parties spécifiques des playbooks, offrant une flexibilité et une rapidité accrues lors du provisioning et du déploiement.

## Utilisation dans l'application

Lors du lancement d'une action de **Provision** ou de **Deploy** :

1. Sélectionnez les serveurs à traiter (avec la touche `espace`)
2. Appuyez sur `p` (provision) ou `d` (deploy)
3. Une interface de sélection de tags s'affiche automatiquement
4. Parcourez les catégories et sélectionnez les tags souhaités
5. Confirmez avec `Enter` pour lancer l'action

### Raccourcis clavier (Tag Selector)

- `↑/↓` ou `k/j` : Naviguer entre les tags
- `Espace` : Cocher/Décocher un tag
- `a` : Sélectionner tous les tags
- `n` : Désélectionner tous les tags
- `Enter` : Confirmer et lancer l'action
- `Esc` : Annuler et revenir

## Tags Disponibles

### Provision

#### System Base
Configuration système de base et packages

- **common** : Toutes les tâches communes (apt update, packages, utilisateurs)
- **packages** : Installation et mise à jour des packages
- **apt** : Opérations APT spécifiques
- **upgrade** : Mise à jour du système (désactivé par défaut)
- **users** : Gestion des utilisateurs (création deploy user)
- **config** : Configuration système (timezone, journald)

#### Security
Pare-feu, SSH et durcissement de la sécurité

- **security** : Toutes les tâches de sécurité
- **firewall** : Configuration du pare-feu UFW
- **ufw** : Configuration spécifique UFW
- **fail2ban** : Installation et configuration de Fail2ban
- **ssh** : Configuration SSH (désactivation password auth)
- **hardening** : Durcissement de la sécurité

#### Runtime & Services
Runtime applicatif et services web

- **nodejs** : Installation de Node.js via NVM
- **nginx** : Installation et configuration du serveur web Nginx
- **postgresql** : Installation de PostgreSQL (pour serveurs DB)

#### Monitoring
Outils de monitoring et observabilité

- **monitoring** : Outils de monitoring (désactivé par défaut)

### Deploy

#### Application
Déploiement de l'application

- **deploy** : Toutes les tâches de déploiement
- **code** : Déploiement du code (clone, build, install)
- **health** : Health checks post-déploiement

## Exemples d'utilisation

### Cas d'usage courants

#### 1. Installation complète (défaut)
Tous les tags par défaut sont sélectionnés.
- Installe : système de base, sécurité, runtime, sans upgrade ni monitoring

#### 2. Mise à jour rapide de la sécurité
Tags sélectionnés : `security`, `firewall`, `ssh`
- Met à jour uniquement la configuration de sécurité
- Gain de temps : ~2-3 minutes au lieu de 10-15 minutes

#### 3. Installation Node.js uniquement
Tags sélectionnés : `nodejs`
- Installe uniquement Node.js via NVM
- Utile pour changer la version de Node.js

#### 4. Configuration Nginx
Tags sélectionnés : `nginx`
- Reconfigure uniquement Nginx
- Utile après modification des variables Nginx

#### 5. Déploiement sans health check
Tags sélectionnés : `deploy`, `code`
- Déploie l'application sans vérifier la santé
- Plus rapide pour les tests

#### 6. Upgrade système complet
Tags sélectionnés : `packages`, `apt`, `upgrade`
- Met à jour tous les packages système
- À faire en maintenance programmée

## Avantages des Tags

### 🚀 Rapidité
- Exécution ciblée = temps réduit
- Idéal pour itérations rapides
- Correction rapide de configurations spécifiques

### 🎯 Précision
- Modification chirurgicale
- Moins de risques d'effets secondaires
- Meilleur contrôle sur les changements

### 🔧 Flexibilité
- Adaptation aux besoins spécifiques
- Personnalisation par environnement
- Tests de composants isolés

### 📊 Efficacité
- Moins de ressources utilisées
- Actions parallèles possibles
- Maintenance simplifiée

## Architecture des Tags

### Dans les Playbooks

Les tags sont définis à trois niveaux :

1. **Niveau Play** : Tag appliqué à tout le play
```yaml
- name: Provision all servers
  hosts: all
  tags: [always]
```

2. **Niveau Role** : Tag appliqué à tout le rôle
```yaml
roles:
  - role: security
    tags: [security, firewall, ssh]
```

3. **Niveau Task** : Tag appliqué à une tâche spécifique
```yaml
- name: Install UFW
  apt:
    name: ufw
  tags: [firewall, ufw, install]
```

### Tag Spécial : `always`

Le tag `always` est exécuté quels que soient les tags sélectionnés.
Utilisé pour les tâches critiques comme :
- Connexion au serveur
- Collecte des facts
- Vérifications pré-déploiement

## Bonnes Pratiques

### ✅ À Faire

1. **Tester sur environnement de dev** avant production
2. **Sélectionner les tags appropriés** pour le contexte
3. **Utiliser "Select All"** pour une installation complète
4. **Documenter** les combinaisons de tags utilisées

### ❌ À Éviter

1. **Ne pas désélectionner tous les tags** : rien ne sera exécuté
2. **Ne pas oublier les dépendances** : nginx nécessite common
3. **Ne pas faire d'upgrade** en production sans test
4. **Ne pas mélanger** tags incompatibles (ex: firewall sans common)

## Dépendances entre Tags

Certains tags dépendent d'autres pour fonctionner correctement :

- `nginx` → nécessite `common` (utilisateur deploy)
- `nodejs` → nécessite `common` (packages de base)
- `postgresql` → nécessite `common` (packages de base)
- `fail2ban` → nécessite `ufw` (pour la configuration)
- `deploy` → nécessite `nodejs` et `nginx` (provisionnés)

## Commandes Ansible directes

Pour utiliser les tags en ligne de commande :

```bash
# Provision avec tags spécifiques
ansible-playbook -i inventory/docker/hosts.yml playbooks/provision.yml \
  --tags "common,security,nodejs" --limit docker-web-01

# Provision en excluant certains tags
ansible-playbook -i inventory/docker/hosts.yml playbooks/provision.yml \
  --skip-tags "monitoring,upgrade" --limit docker-web-01

# Lister les tags disponibles
ansible-playbook playbooks/provision.yml --list-tags
```

## Support et Ajout de Tags

Pour ajouter de nouveaux tags :

1. **Modifier les playbooks** : Ajouter les tags dans `playbooks/*.yml`
2. **Mettre à jour la définition** : Éditer `internal/ansible/tags.go`
3. **Rebuild** : Recompiler l'application avec `make build`

## Changelog Tags

### v1.0 - Initial Release
- Tags pour provision (system, security, runtime, monitoring)
- Tags pour deploy (application, code, health)
- Interface de sélection interactive
- Sélection par défaut intelligente

---

**Note** : Cette fonctionnalité utilise les tags natifs d'Ansible pour une compatibilité maximale et des performances optimales.
