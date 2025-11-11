# 🚀 Semaphore Quick Start - 5 Minutes Setup

## ⚡ Import Automatique (Méthode la plus rapide)

### Préparation (2 minutes)

```bash
# 1. Obtenir votre clé SSH privée
cat /home/basthook/.ssh/Hosting

# 2. Éditer le fichier d'import
nano import.yaml

# 3. Remplacer UNIQUEMENT ces valeurs:
#    ✏️ PASTE_YOUR_PRIVATE_KEY_CONTENT_HERE → Coller TOUTE votre clé
#    ✏️ YOUR_SERVER_IP_1 → IP de votre serveur production
#    ✏️ YOUR_DEV_IP → IP de dev (optionnel)
```

### Import dans Semaphore (3 minutes)

1. **Ouvrir** : http://localhost:3000
2. **Se connecter** : `admin` / `admin`
3. **Créer projet** : 
   - Projects → + New Project
   - Name: `boiler-deploy`
   - Create
4. **Copier** le contenu de `import.yaml`
5. **Configurer manuellement** (Semaphore n'a pas d'import direct, mais on va accélérer) :

---

## 📋 Configuration Accélérée (suivre dans l'ordre)

### 1️⃣ Key Store (Clé SSH)
- **Key Store** → + New Key
- Name: `deploy_key`
- Type: `SSH Key`
- Login: `root`
- Private Key: *Coller votre clé complète*
- **Create**

### 2️⃣ Repository
- **Repositories** → + New Repository
- Name: `local-playbooks`
- URL: `/ansible`
- Branch: `streamlit`
- Access Key: `None`
- **Create**

### 3️⃣ Inventory Production
- **Inventory** → + New Inventory
- Name: `production`
- Type: `Static`
- Content: *Copier depuis `import.yaml` section `inventory.production.content`*
- **Remplacer YOUR_SERVER_IP_1**
- **Create**

### 4️⃣ Environment
- **Environment** → + New Environment
- Name: `production-vars`
- Content (JSON):
```json
{
  "app_name": "myapp",
  "app_repo": "https://github.com/Bastiblast/ansible-next-test.git",
  "app_branch": "main",
  "nodejs_version": "20",
  "app_port": "3002",
  "deploy_user": "root",
  "timezone": "Europe/Paris",
  "pm2_instances": "2",
  "pm2_max_memory": "512M"
}
```
- **Create**

### 5️⃣ Task Templates

Créer 4 templates (**Task Templates** → + New Template) :

#### Template 1: Provision
- Name: `01 - Provision Server`
- Playbook: `playbooks/provision.yml`
- Inventory: `production`
- Environment: `production-vars`
- SSH Key: `deploy_key`
- Repository: `local-playbooks`
- **Create**

#### Template 2: Deploy
- Name: `02 - Deploy Application`
- Playbook: `playbooks/deploy.yml`
- *(même config que Provision)*
- **Create**

#### Template 3: Update
- Name: `03 - Update Application`
- Playbook: `playbooks/update.yml`
- *(même config)*
- **Create**

#### Template 4: Rollback
- Name: `04 - Rollback`
- Playbook: `playbooks/rollback.yml`
- *(même config)*
- **Create**

---

## ✅ Configuration Créée

Vous avez maintenant :
- ✅ 1 Clé SSH (`deploy_key`)
- ✅ 1 Repository (`local-playbooks`)
- ✅ 1 Inventory (`production`)
- ✅ 1 Environment (`production-vars`)
- ✅ 4 Task Templates (Provision, Deploy, Update, Rollback)

---

## 🎯 Premier Test

1. **Task Templates** → `01 - Provision Server`
2. Cliquer sur **▶ Run**
3. Version: `streamlit`
4. **Run**
5. Observer les logs ! 🎉

---

## 📦 Fichiers Fournis

```
├── import.yaml                    # ⭐ Configuration à copier/coller
├── semaphore-project-backup.json  # Backup JSON (référence)
├── QUICKSTART.md                  # 👈 Vous êtes ici
├── IMPORT_TO_SEMAPHORE.md        # Guide détaillé
└── SEMAPHORE_GUIDE.md             # Documentation complète
```

---

## 🔧 Ajouter Environnement DEV (optionnel)

Répéter les étapes 3-5 avec :
- Inventory: `dev` (remplacer IP)
- Environment: `dev-vars` (voir `import.yaml`)
- Templates: Même chose avec suffix `DEV -`

---

## ⚠️ Checklist Avant de Commencer

- [ ] Semaphore démarré : `docker ps | grep semaphore`
- [ ] Clé SSH existe : `ls -la /home/basthook/.ssh/Hosting`
- [ ] Permissions OK : `chmod 600 /home/basthook/.ssh/Hosting`
- [ ] Test SSH : `ssh -i /home/basthook/.ssh/Hosting root@YOUR_IP`
- [ ] `import.yaml` édité avec vos IPs

---

## 🆘 Aide Rapide

**Semaphore ne démarre pas ?**
```bash
docker compose -f docker-compose.semaphore.yml logs -f
```

**SSH échoue dans Semaphore ?**
```bash
# Tester manuellement
ssh -i /home/basthook/.ssh/Hosting root@YOUR_IP
```

**Playbook introuvable ?**
```bash
# Vérifier le montage Docker
docker exec -it semaphore-ui ls -la /ansible/playbooks/
```

---

## ⏱️ Temps Total

- **Préparation** : 2 min
- **Import/Config** : 3 min
- **Premier test** : 1 min
- **TOTAL** : ~6 minutes

---

## 📚 Prochaines Étapes

Après la config initiale :
1. ✅ Ajouter d'autres serveurs (éditer Inventory)
2. ✅ Créer des environnements multiples (staging, etc.)
3. ✅ Planifier des déploiements automatiques (Schedules)
4. ✅ Ajouter des utilisateurs (Team)

**Documentation complète** : `SEMAPHORE_GUIDE.md`

---

**Ready to deploy! 🚀**
