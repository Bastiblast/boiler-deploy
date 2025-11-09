# 🎯 Configuration Hostinger - Résumé

Ce document résume la configuration du projet pour le déploiement sur Hostinger.

## 📋 Informations du serveur

| Paramètre | Valeur |
|-----------|--------|
| **Serveur** | Hostinger VPS |
| **IP** | 72.61.146.126 |
| **Branche Git** | `hostinger` |
| **Utilisateur SSH** | `root` (par défaut) |
| **Environnement** | production |

## 🚀 Commande de déploiement rapide

```bash
# Depuis votre machine locale, sur la branche hostinger
./deploy-hostinger.sh provision  # Première fois uniquement
./deploy-hostinger.sh deploy     # Déployer votre app
```

## 📁 Fichiers de configuration

### Fichiers créés spécifiquement pour Hostinger :

1. **`inventory/hostinger/hosts.yml`** - Inventaire Ansible avec l'IP du serveur
2. **`deploy-hostinger.sh`** - Script de déploiement simplifié
3. **`HOSTINGER_SETUP.md`** - Guide de configuration détaillé
4. **`QUICKSTART_HOSTINGER.md`** - Guide de démarrage rapide (⭐ Commencez ici!)

### Fichiers à configurer (dans `group_vars/`) :

- **`all.yml`** - Variables globales (nom app, repo GitHub, etc.)
- **`webservers.yml`** - Configuration Nginx et SSL
- **`dbservers.yml`** - Configuration PostgreSQL (⚠️ changez le mot de passe!)

## 🔑 Accès et connexions

### SSH
```bash
ssh root@72.61.146.126
ssh deploy@72.61.146.126  # Après provisioning
```

### Application web
- **HTTP**: http://72.61.146.126
- **HTTPS**: https://72.61.146.126 (si domaine configuré avec SSL)

### Services (après provisioning)
- **Prometheus**: http://72.61.146.126:9090
- **Grafana**: http://72.61.146.126:3001 (admin/admin)
- **PostgreSQL**: Port 5432 (localhost uniquement par défaut)

## 📝 Workflow de déploiement typique

### Première installation
```bash
# 1. Vérifier la connexion SSH
ssh root@72.61.146.126

# 2. Configurer vos variables
nano group_vars/all.yml
nano group_vars/dbservers.yml

# 3. Provisionner le serveur (~ 10-15 min)
./deploy-hostinger.sh provision

# 4. Déployer l'application
./deploy-hostinger.sh deploy
```

### Mises à jour régulières
```bash
# Mise à jour rapide (pull + restart)
./deploy-hostinger.sh update

# Ou déploiement complet
./deploy-hostinger.sh deploy
```

### En cas de problème
```bash
# Revenir à la version précédente
./deploy-hostinger.sh rollback

# Voir les logs
ssh deploy@72.61.146.126 'pm2 logs'

# Vérifier le statut
./deploy-hostinger.sh status
```

## 🔧 Configuration minimale requise

Avant le premier déploiement, éditez ces valeurs dans `group_vars/all.yml` :

```yaml
app_name: mon-app                    # ⚠️ À changer
app_repo: "https://github.com/user/repo.git"  # ⚠️ À changer
```

Et dans `group_vars/dbservers.yml` :

```yaml
postgresql_users:
  - password: "MOT_DE_PASSE_FORT"    # ⚠️ À changer absolument!
```

## 📊 Structure après déploiement

```
/var/www/mon-app/
├── current -> releases/20250109_050000_abc1234  # Symlink vers version active
├── releases/
│   ├── 20250109_050000_abc1234/                 # Release actuelle
│   ├── 20250108_120000_def5678/                 # Release précédente
│   └── ...
└── shared/
    ├── logs/                                     # Logs de l'application
    └── config/.env                               # Variables d'environnement
```

## 🎓 Prochaines étapes

1. ✅ **Vous êtes ici** : Configuration initiale terminée
2. 📖 Lire le guide de démarrage rapide : `QUICKSTART_HOSTINGER.md`
3. 🔧 Configurer vos variables dans `group_vars/`
4. 🚀 Lancer le provisioning : `./deploy-hostinger.sh provision`
5. 📦 Déployer votre app : `./deploy-hostinger.sh deploy`
6. 🎉 Profiter !

## 📚 Documentation

| Document | Description |
|----------|-------------|
| **`QUICKSTART_HOSTINGER.md`** | ⭐ Guide de démarrage rapide (commencez ici) |
| **`HOSTINGER_SETUP.md`** | Configuration détaillée et troubleshooting |
| **`README.md`** | Documentation complète du projet Ansible |
| **`DEPLOYMENT_CHECKLIST.md`** | Checklist complète de déploiement |
| **`TROUBLESHOOTING.md`** | Guide de résolution de problèmes |

## 🆘 Besoin d'aide ?

### Problèmes courants

**Impossible de se connecter en SSH**
```bash
ssh -v root@72.61.146.126  # Mode verbose pour diagnostic
```

**L'application ne démarre pas**
```bash
ssh deploy@72.61.146.126 'pm2 logs --lines 100'
```

**Erreur Ansible**
```bash
ansible all -i inventory/hostinger/hosts.yml -m ping
```

### Commandes de diagnostic

```bash
# Test de connectivité
ansible all -i inventory/hostinger/hosts.yml -m ping

# Voir la configuration détectée
ansible-inventory -i inventory/hostinger/hosts.yml --list

# Mode dry-run (ne fait rien, montre ce qui serait fait)
./deploy-hostinger.sh check
```

## 🔐 Sécurité - Rappels importants

- ⚠️ Changez tous les mots de passe par défaut
- ⚠️ Ne commitez jamais `group_vars/all.yml` ou `dbservers.yml` avec vos vraies valeurs
- ⚠️ Utilisez des clés SSH, pas de mots de passe
- ✅ Le fichier `inventory/hostinger/hosts.yml` est déjà ignoré par git
- ✅ Seuls les fichiers `.example` sont versionnés

## 💡 Tips & Astuces

### Alias utiles
Ajoutez à votre `~/.bashrc` ou `~/.zshrc` :

```bash
alias deploy-h='./deploy-hostinger.sh deploy'
alias update-h='./deploy-hostinger.sh update'
alias logs-h='ssh deploy@72.61.146.126 "pm2 logs"'
alias status-h='ssh deploy@72.61.146.126 "pm2 status"'
```

### Surveillance continue
```bash
# Suivre les logs en temps réel
ssh deploy@72.61.146.126 'pm2 logs --lines 0'

# Monitoring avec watch
watch -n 5 'ssh deploy@72.61.146.126 "pm2 status"'
```

---

**Branche**: `hostinger`  
**Dernière mise à jour**: 2025-01-09  
**Status**: ✅ Prêt pour le déploiement
