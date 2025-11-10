# 🚀 Semaphore UI - Guide d'Installation

## 📋 Qu'est-ce que Semaphore UI ?

Semaphore est une interface web moderne pour gérer Ansible (et Terraform). Elle permet de :
- ✅ Gérer les inventaires visuellement
- ✅ Exécuter des playbooks depuis l'interface
- ✅ Planifier des déploiements
- ✅ Gérer les variables et secrets
- ✅ Contrôler les accès (RBAC)

## 🔧 Installation

### Démarrer Semaphore

```bash
# Dans le dossier boiler-deploy
docker compose -f docker-compose.semaphore.yml up -d
```

### Vérifier le statut

```bash
docker compose -f docker-compose.semaphore.yml ps
docker compose -f docker-compose.semaphore.yml logs -f semaphore
```

### Accéder à l'interface

Ouvrir dans votre navigateur : **http://localhost:3000**

**Credentials par défaut:**
- Username: `admin`
- Password: `admin`

⚠️ **Changez le mot de passe après la première connexion !**

## 📦 Configuration Initiale

### 1. Créer un Projet

1. Aller dans **Projects** → **New Project**
2. Nom: `boiler-deploy`
3. Description: Déploiement VPS multi-serveurs

### 2. Ajouter un Key Store (Clés SSH)

1. **Key Store** → **New Key**
2. Type: SSH Key
3. Name: `deploy_key`
4. Username: `deploy` (ou votre user SSH)
5. Private Key: Copier votre clé privée SSH
   - Par exemple: `~/.ssh/id_rsa`

### 3. Créer un Repository

1. **Repositories** → **New Repository**
2. Name: `local-playbooks`
3. URL: `/ansible` (point de montage Docker)
4. Branch: `main`
5. Access Key: (aucun si local)

### 4. Créer un Inventory

#### Option A: Via l'interface

1. **Inventory** → **New Inventory**
2. Name: `production`
3. Type: Static YAML
4. Inventory Content:
   ```yaml
   all:
     children:
       webservers:
         hosts:
           production-web-01:
             ansible_host: 192.168.1.10
             ansible_user: deploy
             ansible_ssh_private_key_file: ~/.ssh/id_rsa
             app_port: 3000
   ```

#### Option B: Importer depuis fichiers existants

Si vous avez déjà des inventaires dans `inventory/production/hosts.yml`:

1. Copier le contenu du fichier
2. Le coller dans l'interface Semaphore
3. Ajuster les chemins si nécessaire

### 5. Créer un Environment (Variables)

1. **Environment** → **New Environment**
2. Name: `production-vars`
3. Content (format JSON):
   ```json
   {
     "app_name": "myapp",
     "app_repo": "git@github.com:user/repo.git",
     "app_branch": "main",
     "nodejs_version": "20",
     "app_port": "3000"
   }
   ```

### 6. Créer un Task Template

1. **Task Templates** → **New Template**
2. Name: `Provision Servers`
3. Playbook: `playbooks/provision.yml`
4. Inventory: `production`
5. Environment: `production-vars`
6. Key: `deploy_key`

## 🎯 Utilisation

### Exécuter un Playbook

1. Aller dans **Task Templates**
2. Cliquer sur le template (ex: "Provision Servers")
3. Cliquer sur **Run**
4. Suivre les logs en temps réel

### Gérer l'Inventaire

1. **Inventory** → Sélectionner l'environnement
2. Éditer directement le YAML
3. **Save**

### Ajouter un Serveur

Éditer l'inventaire et ajouter:
```yaml
production-web-02:
  ansible_host: 192.168.1.11
  ansible_user: deploy
  ansible_ssh_private_key_file: ~/.ssh/id_rsa
  app_port: 3001
```

## 🛠️ Commandes Utiles

### Arrêter Semaphore

```bash
docker compose -f docker-compose.semaphore.yml down
```

### Redémarrer Semaphore

```bash
docker compose -f docker-compose.semaphore.yml restart
```

### Voir les logs

```bash
docker compose -f docker-compose.semaphore.yml logs -f
```

### Sauvegarder la configuration

Les données sont persistées dans des volumes Docker:
- `semaphore-mysql-data`: Base de données
- `semaphore-data`: Fichiers Semaphore

Pour sauvegarder:
```bash
docker compose -f docker-compose.semaphore.yml down
docker run --rm -v semaphore-mysql-data:/data -v $(pwd):/backup alpine tar czf /backup/semaphore-backup.tar.gz /data
```

### Réinitialiser Semaphore

```bash
docker compose -f docker-compose.semaphore.yml down -v
docker compose -f docker-compose.semaphore.yml up -d
```

## 🔐 Sécurité

### Changer le mot de passe admin

1. Se connecter en tant qu'admin
2. **User Settings** → **Change Password**

### Ajouter des utilisateurs

1. **Users** → **New User**
2. Attribuer des rôles par projet

### Variables sensibles

Utiliser la section **Environment** avec le type "Secret" pour les mots de passe, tokens, etc.

## 📚 Intégration avec votre workflow

### Utiliser avec vos scripts existants

Vos scripts `deploy.sh`, `setup.sh` fonctionnent toujours !

Semaphore est un **complément** qui offre:
- Une interface visuelle
- Un historique des exécutions
- Une planification des tâches
- Un contrôle d'accès multi-utilisateurs

### Workflow recommandé

1. **Configuration initiale:** Utiliser `setup.sh` OU Semaphore
2. **Gestion quotidienne:** Semaphore UI
3. **Automatisation CI/CD:** Scripts bash
4. **Debugging:** Logs Semaphore + SSH direct

## 🆘 Dépannage

### Semaphore ne démarre pas

Vérifier les logs:
```bash
docker compose -f docker-compose.semaphore.yml logs semaphore
```

### Erreur de connexion à la base de données

Attendre que MySQL soit prêt:
```bash
docker compose -f docker-compose.semaphore.yml restart semaphore
```

### Playbooks non trouvés

Vérifier que le volume est bien monté:
```bash
docker exec -it semaphore-ui ls -la /ansible
```

### Port 3000 déjà utilisé

Modifier le port dans `docker-compose.semaphore.yml`:
```yaml
ports:
  - "3001:3000"  # Change 3000 to 3001
```

## 📖 Ressources

- Documentation officielle: https://semaphoreui.com/
- GitHub: https://github.com/semaphoreui/semaphore
- Discussions: https://github.com/semaphoreui/semaphore/discussions

## 🎉 Prochaines Étapes

Maintenant que Semaphore est installé:

1. ✅ Créer votre premier projet
2. ✅ Importer vos inventaires existants
3. ✅ Configurer vos clés SSH
4. ✅ Créer des templates pour vos playbooks
5. ✅ Exécuter votre premier déploiement

**Bonne gestion d'infrastructure ! 🚀**
