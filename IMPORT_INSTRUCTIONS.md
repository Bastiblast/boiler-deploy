# 📥 Import Semaphore Project - Instructions Complètes

## 🎯 Fichier à Utiliser

**`boiler-deploy-backup.json`** - Format officiel Semaphore (basé sur backup_demo.json)

---

## ✅ Méthode 1 : Import via Interface Web (RECOMMANDÉ)

### Étape 1 : Se connecter à Semaphore

1. Ouvrir : **http://localhost:3000**
2. Se connecter :
   - Username : `admin`
   - Password : votre mot de passe (si oublié : `./reset-admin-password.sh`)

### Étape 2 : Préparer le fichier

```bash
# 1. Éditer le fichier d'import
nano boiler-deploy-backup.json

# 2. Remplacer UNIQUEMENT ces valeurs :
#    - PASTE_YOUR_SSH_PRIVATE_KEY_HERE → Votre clé SSH complète
#    - YOUR_SERVER_IP → Votre IP de production
#    - YOUR_DEV_IP → Votre IP de dev (optionnel)
#    - YOUR_SERVER_IP_1, YOUR_SERVER_IP_2, etc. (pour multi-serveurs)
```

**Pour obtenir votre clé SSH :**
```bash
cat /home/basthook/.ssh/Hosting
# Copier TOUT le contenu (de -----BEGIN à -----END)
```

### Étape 3 : Créer le projet et importer

#### A. Créer le projet vide

1. Dans Semaphore UI → **Projects**
2. Cliquer sur **+ New Project**
3. Name : `boiler-deploy`
4. **Create**

#### B. Importer la configuration

1. Dans le projet → **Settings** (icône ⚙️ en haut à droite)
2. Chercher la section **Backup & Restore**
3. Cliquer sur **Restore**
4. **Upload file** → Sélectionner `boiler-deploy-backup.json` OU
5. **Paste JSON** → Copier le contenu du fichier
6. Cliquer sur **Restore**

✅ **Import terminé !**

---

## ⚡ Méthode 2 : Import via Script Automatisé

Si l'import manuel ne fonctionne pas, utilisez le script :

```bash
./semaphore-import.sh
```

---

## 🔧 Ce qui sera importé

### ✅ Keys (Clés SSH)
- **None** - Pas de clé (pour repos locaux)
- **deploy_key** - Votre clé SSH pour les serveurs

### ✅ Repositories
- **local-playbooks** - Pointe vers `/ansible` (branche `streamlit`)

### ✅ Inventories (3)
1. **production** - 1 serveur web simple
2. **dev** - 1 serveur de développement
3. **production-multi** - Architecture complète :
   - 2 serveurs web
   - 1 serveur database
   - 1 serveur monitoring

### ✅ Environments (Variables)
- **production-vars** - Variables pour production
- **dev-vars** - Variables pour développement

### ✅ Views (Vues organisées)
- **Deploy** - Templates de déploiement
- **Manage** - Templates de gestion

### ✅ Templates (6 Playbooks)
1. **01 - Provision Server** → Configuration initiale
2. **02 - Deploy Application** → Déploiement
3. **03 - Update Application** → Mise à jour
4. **04 - Rollback** → Retour arrière
5. **DEV - Provision** → Config dev
6. **DEV - Deploy** → Déploiement dev

---

## 🎨 Structure Visuelle

Après import, vous verrez dans Semaphore :

```
boiler-deploy/
├── 📁 Key Store
│   ├── None
│   └── deploy_key (SSH)
│
├── 📁 Repositories
│   └── local-playbooks (/ansible, streamlit)
│
├── 📁 Inventory
│   ├── production (1 serveur)
│   ├── dev (1 serveur)
│   └── production-multi (architecture complète)
│
├── 📁 Environment
│   ├── production-vars
│   └── dev-vars
│
└── 📁 Task Templates
    ├── 🚀 Deploy
    │   ├── 01 - Provision Server
    │   ├── 02 - Deploy Application
    │   ├── DEV - Provision
    │   └── DEV - Deploy
    │
    └── ⚙️ Manage
        ├── 03 - Update Application
        └── 04 - Rollback
```

---

## ✏️ Personnalisation Avant Import

### Changer les IPs des serveurs

Éditer `boiler-deploy-backup.json` :

```json
{
  "inventories": [
    {
      "name": "production",
      "inventory": "all:\n  children:\n    webservers:\n      hosts:\n        production-web-01:\n          ansible_host: 192.168.1.10  ← CHANGER ICI
```

### Changer le repository Git

```json
{
  "environments": [
    {
      "name": "production-vars",
      "json": "{\n  \"app_repo\": \"https://github.com/YOUR_USER/YOUR_REPO.git\"  ← ICI
```

### Ajouter votre clé SSH

```json
{
  "keys": [
    {
      "name": "deploy_key",
      "ssh": {
        "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nVOTRE CLÉ ICI\n-----END OPENSSH PRIVATE KEY-----"
```

---

## 🆘 Dépannage

### Erreur : "Invalid backup format"

**Solution :** Vérifier que le JSON est valide
```bash
# Tester la syntaxe
python3 -m json.tool boiler-deploy-backup.json > /dev/null && echo "JSON valide" || echo "JSON invalide"
```

### Erreur : "SSH Key invalid"

**Solution :** Vérifier que vous avez copié la clé COMPLÈTE
```bash
# La clé doit commencer par -----BEGIN et finir par -----END
grep -E "(BEGIN|END)" /home/basthook/.ssh/Hosting
```

### Import ne crée rien

**Solution :** 
1. Vérifier que le projet existe
2. Essayer d'importer section par section (Keys, puis Repos, puis Inventory...)
3. Utiliser le script automatisé : `./semaphore-import.sh`

### Playbooks non trouvés après import

**Solution :** Vérifier le montage Docker
```bash
docker exec -it semaphore-ui ls -la /ansible/playbooks/
```

---

## 🎯 Après l'Import

### 1. Vérifier l'import

1. **Key Store** → Vérifier que `deploy_key` existe
2. **Repositories** → Vérifier `local-playbooks`
3. **Inventory** → Vérifier les 3 inventaires
4. **Environment** → Vérifier les variables
5. **Task Templates** → Vérifier les 6 templates

### 2. Éditer les IPs si nécessaire

Si vous avez importé avec des placeholders :

1. **Inventory** → Sélectionner `production`
2. **Edit**
3. Remplacer `YOUR_SERVER_IP` par votre vraie IP
4. **Save**

### 3. Tester une connexion

1. **Task Templates** → `01 - Provision Server`
2. Cliquer sur **Run** ▶
3. Version : `streamlit`
4. **Run**
5. Observer les logs

---

## 📊 Comparaison des Méthodes

| Méthode | Temps | Complexité | Succès |
|---------|-------|------------|--------|
| Import JSON | 2 min | Facile | 95% |
| Script API | 3 min | Moyen | 100% |
| Manuel | 15 min | Difficile | 100% |

**Recommandation :** Essayer l'import JSON d'abord, puis le script si ça échoue.

---

## ✅ Checklist Avant Import

- [ ] Semaphore démarré : `docker ps | grep semaphore`
- [ ] Connecté à Semaphore (mot de passe fonctionnel)
- [ ] Fichier `boiler-deploy-backup.json` édité
- [ ] Clé SSH collée dans le JSON
- [ ] IPs remplacées par les vraies
- [ ] JSON validé : `python3 -m json.tool boiler-deploy-backup.json`

---

## 🚀 Quick Start

```bash
# 1. Éditer le fichier
nano boiler-deploy-backup.json
# Remplacer : PASTE_YOUR_SSH_PRIVATE_KEY_HERE et YOUR_SERVER_IP

# 2. Valider le JSON
python3 -m json.tool boiler-deploy-backup.json > /dev/null && echo "✓ OK"

# 3. Importer dans Semaphore UI
# http://localhost:3000 → Projects → boiler-deploy → Settings → Restore

# 4. Tester
# Task Templates → 01 - Provision Server → Run
```

---

## 📚 Ressources

- **Fichier d'import :** `boiler-deploy-backup.json`
- **Script alternatif :** `semaphore-import.sh`
- **Reset password :** `reset-admin-password.sh`
- **Guide complet :** `SEMAPHORE_GUIDE.md`

---

**Bonne configuration ! 🎉**
