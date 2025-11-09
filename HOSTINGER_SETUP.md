# Configuration de déploiement Hostinger

## Informations serveur
- **IP**: 72.61.146.126
- **Branche**: hostinger
- **Environnement**: production

## Étapes de configuration

### 1. Configurer vos variables

Éditez les fichiers suivants avec vos informations :

#### `group_vars/all.yml`
```yaml
app_name: votre-app-name
app_port: 3000
app_repo: "https://github.com/votre-username/votre-repo.git"
app_branch: "main"  # ou la branche que vous voulez déployer
```

#### `group_vars/webservers.yml`
```yaml
ssl_certbot_email: "votre-email@example.com"
ssl_domains:
  - "votre-domaine.com"  # ou l'IP si pas de domaine
```

#### `group_vars/dbservers.yml`
```yaml
postgresql_users:
  - name: "{{ app_name }}_user"
    password: "CHANGEZ_CE_MOT_DE_PASSE_SECURISE"
    db: "{{ app_name }}_{{ environment }}"
    priv: "ALL"
```

### 2. Vérifier l'accès SSH

Testez la connexion SSH à votre serveur Hostinger :

```bash
ssh root@72.61.146.126
```

Si vous n'avez pas encore de clé SSH, générez-en une :
```bash
ssh-keygen -t rsa -b 4096 -C "votre-email@example.com"
```

Copiez votre clé sur le serveur :
```bash
ssh-copy-id root@72.61.146.126
```

### 3. Tester la connectivité Ansible

```bash
ansible all -i inventory/hostinger/hosts.yml -m ping
```

### 4. Déploiement complet (première fois)

Installez toutes les dépendances Ansible :
```bash
ansible-galaxy collection install -r requirements.yml
```

Lancez le provisioning complet :
```bash
ansible-playbook playbooks/provision.yml -i inventory/hostinger/hosts.yml
```

Cette commande va :
- Configurer les utilisateurs et la sécurité
- Installer PostgreSQL
- Installer Node.js et PM2
- Configurer Nginx
- Mettre en place le monitoring (optionnel)

### 5. Déployer votre application

```bash
ansible-playbook playbooks/deploy.yml -i inventory/hostinger/hosts.yml
```

### 6. Mise à jour rapide

Pour les déploiements suivants :
```bash
ansible-playbook playbooks/update.yml -i inventory/hostinger/hosts.yml
```

### 7. Rollback en cas de problème

```bash
ansible-playbook playbooks/rollback.yml -i inventory/hostinger/hosts.yml
```

## Vérification du déploiement

### Vérifier PM2
```bash
ssh deploy@72.61.146.126
pm2 status
pm2 logs
```

### Vérifier Nginx
```bash
ssh root@72.61.146.126
systemctl status nginx
nginx -t
```

### Vérifier PostgreSQL
```bash
ssh root@72.61.146.126
systemctl status postgresql
sudo -u postgres psql -l
```

## Accès aux services

### Application web
- HTTP: http://72.61.146.126
- HTTPS: https://72.61.146.126 (si SSL configuré)
- Ou via votre domaine si configuré

### Monitoring (si activé)
- Prometheus: http://72.61.146.126:9090
- Grafana: http://72.61.146.126:3001 (admin/admin)

## Notes importantes pour Hostinger

1. **Utilisateur par défaut**: Hostinger utilise généralement `root` comme utilisateur par défaut
2. **Firewall**: Vérifiez que les ports nécessaires sont ouverts (80, 443, 22)
3. **Domaine**: Si vous avez un domaine, pointez-le vers 72.61.146.126
4. **SSL**: Let's Encrypt nécessite un domaine valide (ne fonctionne pas avec IP uniquement)

## Configuration minimale recommandée

Si vous voulez une configuration minimaliste sans monitoring :

Éditez `playbooks/provision.yml` et commentez la section monitoring si nécessaire.

## Script de déploiement rapide

Utilisez le script `deploy.sh` :
```bash
./deploy.sh hostinger
```

## Troubleshooting

### Problème de connexion SSH
```bash
ssh -v root@72.61.146.126
```

### Problème Ansible
```bash
ansible-playbook playbooks/provision.yml -i inventory/hostinger/hosts.yml -vvv
```

### Logs de l'application
```bash
ssh deploy@72.61.146.126
pm2 logs --lines 100
```

### Logs Nginx
```bash
ssh root@72.61.146.126
tail -f /var/log/nginx/error.log
tail -f /var/log/nginx/access.log
```
