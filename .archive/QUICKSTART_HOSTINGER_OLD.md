# 🚀 Démarrage rapide - Déploiement Hostinger

Configuration rapide pour déployer sur votre serveur Hostinger (IP: 72.61.146.126)

## ⚡ Installation en 5 minutes

### 1. Prérequis
```bash
# Installer Ansible
sudo apt update && sudo apt install ansible -y

# Vérifier l'installation
ansible --version
```

### 2. Configurer votre projet

```bash
# Cloner/Naviguer vers le projet
cd boiler-deploy
git checkout hostinger

# Installer les dépendances Ansible
ansible-galaxy collection install -r requirements.yml
```

### 3. Configurer l'accès SSH

```bash
# Tester la connexion SSH
ssh root@72.61.146.126

# Si vous n'avez pas de clé SSH, créez-en une
ssh-keygen -t rsa -b 4096

# Copiez votre clé sur le serveur
ssh-copy-id root@72.61.146.126
```

### 4. Configurer vos variables

Éditez `group_vars/all.yml` et changez au minimum :

```yaml
app_name: portefolio        # Nom de votre application
app_port: 3000                   # Port de votre app Node.js
app_repo: "https://github.com/Bastiblast/portefolio.git"  # Votre repo GitHub
```

Éditez `group_vars/dbservers.yml` pour le mot de passe de la base de données :

```yaml
postgresql_users:
  - name: "{{ app_name }}_user"
    password: "CHANGEZ_MOI_AVEC_MOT_DE_PASSE_SECURISE"  # ⚠️ IMPORTANT !
    db: "{{ app_name }}_{{ environment }}"
    priv: "ALL"
```

### 5. Déployer !

```bash
# Première installation (installe tout)
./deploy-hostinger.sh provision

# Déployer votre application
./deploy-hostinger.sh deploy
```

C'est tout ! Votre application est maintenant accessible sur http://72.61.146.126

---

## 📝 Commandes utiles

### Déploiement
```bash
./deploy-hostinger.sh provision  # Installation complète (première fois)
./deploy-hostinger.sh deploy     # Déployer l'application
./deploy-hostinger.sh update     # Mise à jour rapide
./deploy-hostinger.sh rollback   # Revenir à la version précédente
./deploy-hostinger.sh check      # Vérifier sans exécuter
./deploy-hostinger.sh status     # Voir le statut PM2
```

### SSH & Logs
```bash
# Se connecter au serveur
ssh deploy@72.61.146.126

# Voir les logs de l'application
ssh deploy@72.61.146.126 'pm2 logs'

# Voir le statut PM2
ssh deploy@72.61.146.126 'pm2 status'

# Redémarrer l'application
ssh deploy@72.61.146.126 'pm2 restart all'
```

### Base de données
```bash
# Se connecter à PostgreSQL
ssh root@72.61.146.126
sudo -u postgres psql

# Lister les bases de données
\l

# Se connecter à votre base
\c votre_app_hostinger
```

---

## 🎯 Checklist de déploiement

- [ ] Ansible installé sur votre machine locale
- [ ] Accès SSH configuré (ssh root@72.61.146.126)
- [ ] Variables configurées dans `group_vars/all.yml`
- [ ] Mot de passe DB changé dans `group_vars/dbservers.yml`
- [ ] Repository GitHub accessible
- [ ] Première installation : `./deploy-hostinger.sh provision`
- [ ] Déploiement de l'app : `./deploy-hostinger.sh deploy`
- [ ] Test de l'application : http://72.61.146.126

---

## 🔧 Configuration avancée

### Ajouter un domaine

Si vous avez un domaine pointant vers 72.61.146.126, éditez `group_vars/webservers.yml` :

```yaml
ssl_enabled: true
ssl_certbot_email: "votre-email@example.com"
ssl_domains:
  - "votre-domaine.com"
  - "www.votre-domaine.com"
```

Puis redéployez :
```bash
./deploy-hostinger.sh provision
```

### Variables d'environnement

Les variables d'environnement de votre app sont dans :
```
/var/www/votre-app/shared/config/.env
```

Pour les modifier :
```bash
ssh deploy@72.61.146.126
nano /var/www/votre-app/shared/config/.env
pm2 restart all
```

### Performance PM2

Éditez `group_vars/all.yml` pour ajuster :

```yaml
pm2_instances: 2          # Nombre d'instances (cluster mode)
pm2_max_memory: "512M"    # Redémarre si dépassé
```

---

## 🆘 Résolution de problèmes

### Erreur de connexion SSH
```bash
# Test verbose
ssh -v root@72.61.146.126

# Si timeout, vérifiez le firewall
# Vérifiez que le port 22 est ouvert chez Hostinger
```

### Application ne démarre pas
```bash
# Voir les logs
ssh deploy@72.61.146.126 'pm2 logs --lines 50'

# Redémarrer manuellement
ssh deploy@72.61.146.126 'pm2 restart all'

# Vérifier la config Nginx
ssh root@72.61.146.126 'nginx -t'
```

### Base de données inaccessible
```bash
# Vérifier PostgreSQL
ssh root@72.61.146.126 'systemctl status postgresql'

# Test de connexion
ssh root@72.61.146.126
sudo -u postgres psql -l
```

### Port déjà utilisé
```bash
# Voir les ports en écoute
ssh root@72.61.146.126 'netstat -tulpn | grep LISTEN'

# Changer le port de l'app dans group_vars/all.yml
app_port: 3001  # au lieu de 3000
```

---

## 📚 Documentation complète

Pour plus de détails, consultez :
- `HOSTINGER_SETUP.md` - Configuration détaillée
- `README.md` - Documentation complète du projet
- `TROUBLESHOOTING.md` - Guide de dépannage
- `DEPLOYMENT_CHECKLIST.md` - Checklist complète

---

## 🔐 Sécurité

⚠️ **Important** : 
- Changez TOUS les mots de passe par défaut
- Ne commitez JAMAIS les fichiers avec vos vraies IPs et mots de passe
- Utilisez `ansible-vault` pour les données sensibles en production

```bash
# Chiffrer un fichier sensible
ansible-vault encrypt group_vars/dbservers.yml

# Éditer un fichier chiffré
ansible-vault edit group_vars/dbservers.yml

# Déployer avec vault
./deploy-hostinger.sh deploy --ask-vault-pass
```

---

Bon déploiement ! 🎉
