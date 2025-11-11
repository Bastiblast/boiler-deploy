# 🚀 Import Automatique - Solution à l'Erreur

## ❌ Problème
L'erreur `Cannot set properties of undefined (setting 'name')` indique que Semaphore ne supporte pas l'import direct de fichiers YAML/JSON via l'interface.

## ✅ Solution : Script d'Import Automatisé

J'ai créé un script qui utilise l'API Semaphore pour tout configurer automatiquement !

---

## 🎯 Utilisation du Script (2 minutes)

### Lancer l'import automatique :

```bash
./semaphore-import.sh
```

### Le script va demander :

1. **Username** : `admin` (ou appuyez sur Entrée)
2. **Password** : `admin` (ou appuyez sur Entrée)
3. **Server IP** : Votre IP de production (ex: `192.168.1.10`)

**C'est tout !** Le script crée automatiquement :
- ✅ Le projet `boiler-deploy`
- ✅ La clé SSH `deploy_key`
- ✅ Le repository `local-playbooks`
- ✅ L'inventaire `production`
- ✅ L'environment `production-vars`
- ✅ 4 Task Templates (Provision, Deploy, Update, Rollback)

---

## 📋 Exemple d'Exécution

```bash
$ ./semaphore-import.sh

╔══════════════════════════════════════════════════════════╗
║     Semaphore Project Import - Automated Setup          ║
╚══════════════════════════════════════════════════════════╝

Step 1/7: Authentication
Enter Semaphore username [admin]: ⏎
Enter Semaphore password [admin]: ⏎
✓ Authenticated successfully

Step 2/7: Creating Project
✓ Project created (ID: 1)

Step 3/7: SSH Key Configuration
✓ SSH key loaded from /home/basthook/.ssh/Hosting
✓ SSH key created (ID: 1)

Step 4/7: Creating Repository
✓ Repository created (ID: 1)

Step 5/7: Server Configuration
Enter your production server IP: 192.168.1.10
✓ Inventory created (ID: 1)

Step 6/7: Creating Environment
✓ Environment created (ID: 1)

Step 7/7: Creating Task Templates
✓ Template 'Provision' created
✓ Template 'Deploy' created
✓ Template 'Update' created
✓ Template 'Rollback' created

╔══════════════════════════════════════════════════════════╗
║                    Import Complete! 🎉                   ║
╚══════════════════════════════════════════════════════════╝

✓ Project: boiler-deploy (ID: 1)
✓ SSH Key: deploy_key (ID: 1)
✓ Repository: local-playbooks (ID: 1)
✓ Inventory: production (ID: 1)
✓ Environment: production-vars (ID: 1)
✓ Templates: 4 task templates created

Next steps:
  1. Open: http://localhost:3000
  2. Go to project: boiler-deploy
  3. Run your first playbook!
```

---

## 🔧 Configuration Créée

Le script configure automatiquement :

### 1. **SSH Key** (`deploy_key`)
- Type: SSH
- Login: root
- Clé privée: Votre clé `/home/basthook/.ssh/Hosting`

### 2. **Repository** (`local-playbooks`)
- URL: `/ansible`
- Branch: `streamlit`

### 3. **Inventory** (`production`)
```yaml
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: [VOTRE_IP]
          ansible_user: root
          app_port: 3002
```

### 4. **Environment** (`production-vars`)
```json
{
  "app_name": "myapp",
  "app_repo": "https://github.com/Bastiblast/ansible-next-test.git",
  "app_branch": "main",
  "nodejs_version": "20",
  "app_port": "3002",
  "deploy_user": "root"
}
```

### 5. **Task Templates** (4)
- **01 - Provision Server** → `playbooks/provision.yml`
- **02 - Deploy Application** → `playbooks/deploy.yml`
- **03 - Update Application** → `playbooks/update.yml`
- **04 - Rollback** → `playbooks/rollback.yml`

---

## 🆘 Dépannage

### Erreur : "SSH key not found"
```bash
# Vérifier que la clé existe
ls -la /home/basthook/.ssh/Hosting

# Si elle est ailleurs, le script demandera le chemin
```

### Erreur : "Authentication failed"
```bash
# Vérifier que Semaphore est démarré
docker ps | grep semaphore

# Vérifier l'URL
curl http://localhost:3000
```

### Erreur : "API call failed"
```bash
# Vérifier les logs Semaphore
docker logs semaphore-ui --tail 50

# Relancer le script
./semaphore-import.sh
```

### Projet existe déjà
```bash
# Supprimer le projet dans Semaphore UI
# ou modifier PROJECT_NAME dans le script
nano semaphore-import.sh
# Changer: PROJECT_NAME="boiler-deploy-2"
```

---

## 🎯 Après l'Import

1. **Ouvrir Semaphore** : http://localhost:3000
2. **Aller dans le projet** : `boiler-deploy`
3. **Task Templates** → Sélectionner `01 - Provision Server`
4. **Cliquer sur Run** ▶
5. **Suivre les logs** en temps réel

---

## 📝 Ajouter un Serveur Supplémentaire

Après l'import initial, pour ajouter d'autres serveurs :

1. Dans Semaphore → **Inventory** → `production`
2. **Edit**
3. Ajouter dans le YAML :
```yaml
production-web-02:
  ansible_host: 192.168.1.11
  ansible_user: root
  app_port: 3003
```
4. **Save**

---

## 🔄 Créer l'Environnement DEV

Relancer le script en mode interactif ou modifier manuellement :

1. **Inventory** → New Inventory → `dev`
2. Copier la config production, changer l'IP
3. **Environment** → New → `dev-vars`
4. Ajuster les variables (ex: `pm2_instances: 1`)
5. **Templates** → Dupliquer et changer inventory/env

---

## ⚡ Commandes Rapides

```bash
# Import complet automatique
./semaphore-import.sh

# Vérifier Semaphore
docker ps | grep semaphore

# Voir les logs d'import (si erreur)
# Les erreurs API s'affichent directement

# Accéder à Semaphore
xdg-open http://localhost:3000
```

---

## 📚 Alternative : Configuration Manuelle

Si le script échoue, suivre le guide pas-à-pas :
→ Voir `IMPORT_TO_SEMAPHORE.md`

---

## ✅ Checklist Avant d'Exécuter

- [ ] Semaphore démarré : `docker ps | grep semaphore`
- [ ] Clé SSH existe : `ls -la /home/basthook/.ssh/Hosting`
- [ ] Permissions : `chmod 600 /home/basthook/.ssh/Hosting`
- [ ] Connaître l'IP du serveur
- [ ] Script exécutable : `chmod +x semaphore-import.sh`

---

## 🎉 Résultat Final

Après exécution, vous aurez un projet Semaphore **complètement configuré** avec :
- 1 projet
- 1 clé SSH
- 1 repository
- 1 inventaire
- 1 environment
- 4 playbooks exécutables

**Prêt à déployer en 2 minutes ! 🚀**

---

**Questions ? Consultez `SEMAPHORE_GUIDE.md` pour plus de détails**
