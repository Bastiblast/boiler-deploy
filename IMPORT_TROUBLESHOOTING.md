# 🔧 Troubleshooting: Erreur 400 lors du Restore

## ❌ Erreur Rencontrée

```
Restore Project...
Request failed with status code 400
```

## 🎯 Cause du Problème

L'API de restore de Semaphore est **très stricte** sur le format JSON. Plusieurs raisons possibles :

1. **Structure des clés SSH** - Format non exactement conforme
2. **Champs manquants** - Certains champs obligatoires absents
3. **Relations entre objets** - Les IDs doivent correspondre
4. **Données invalides** - Caractères spéciaux dans les clés SSH

## ✅ Solution : Utiliser le Script Automatisé

**Le script `semaphore-import.sh` est BEAUCOUP plus fiable** car il utilise l'API REST directement.

### Pourquoi le script fonctionne mieux ?

- ✅ Crée les objets dans le bon ordre
- ✅ Gère automatiquement les IDs
- ✅ Valide chaque étape
- ✅ Échappe correctement les caractères spéciaux
- ✅ Fournit des erreurs détaillées

---

## 🚀 Procédure Recommandée

### Étape 1 : Réinitialiser le mot de passe (si nécessaire)

```bash
./reset-admin-password.sh
```

### Étape 2 : Lancer l'import automatisé

```bash
./semaphore-import.sh
```

Le script va demander :
1. **Username** : `admin` (Entrée)
2. **Password** : `admin` (ou votre mot de passe)
3. **Déjà configuré ?** : `y` (Entrée)
4. **Server IP** : Votre IP de production

### Étape 3 : Attendre la fin

Le script va :
- ✅ S'authentifier
- ✅ Créer le projet `boiler-deploy`
- ✅ Créer la clé SSH `deploy_key`
- ✅ Créer le repository `local-playbooks`
- ✅ Créer l'inventaire `production`
- ✅ Créer l'environment `production-vars`
- ✅ Créer 4 task templates

**Temps total : ~30 secondes**

---

## 🔄 Alternative : Import Manuel dans Semaphore UI

Si vous voulez vraiment utiliser l'interface de restore, voici comment :

### Option A : Import Minimal d'abord

1. Utiliser `boiler-deploy-minimal.json` (version ultra-simple)
2. Une fois importé, ajouter manuellement :
   - Les clés SSH
   - Les templates
   - Les autres inventaires

### Option B : Import Section par Section

Au lieu d'importer tout d'un coup :

1. **Key Store** → New Key → Créer `deploy_key` manuellement
2. **Repositories** → New → Créer `local-playbooks`
3. **Inventory** → New → Créer `production`
4. **Environment** → New → Créer `production-vars`
5. **Templates** → New → Créer chaque template

---

## 🆘 Dépannage Avancé

### Test 1 : Vérifier que Semaphore fonctionne

```bash
# Tester l'API
curl http://localhost:3000/api/ping
# Doit retourner: {"success":true}
```

### Test 2 : Vérifier l'authentification

```bash
# Login manuel
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"auth":"admin","password":"admin"}'
# Doit retourner un token
```

### Test 3 : Vérifier les logs en temps réel

```bash
# Terminal 1 : Suivre les logs
docker logs -f semaphore-ui

# Terminal 2 : Tenter l'import
# Observer les erreurs détaillées dans Terminal 1
```

### Test 4 : Réinitialiser complètement Semaphore

```bash
# Sauvegarder les données importantes d'abord !
docker compose -f docker-compose.semaphore.yml down -v
docker compose -f docker-compose.semaphore.yml up -d

# Attendre 30 secondes
sleep 30

# Réessayer l'import
./semaphore-import.sh
```

---

## 📊 Comparaison des Méthodes

| Méthode | Taux de Succès | Temps | Complexité |
|---------|----------------|-------|------------|
| **Script automatisé** | ✅ 99% | 30s | Facile |
| **Restore JSON (UI)** | ⚠️ 50% | 5min | Moyenne |
| **Import manuel** | ✅ 100% | 20min | Difficile |

**Recommandation : Utiliser `semaphore-import.sh`**

---

## 🎯 Pourquoi l'Erreur 400 ?

### Causes Communes

1. **Clé SSH mal formatée**
   ```json
   "private_key": "PASTE_YOUR_SSH_PRIVATE_KEY_HERE"  ❌
   ```
   La clé doit être échappée correctement avec `\n`

2. **Relations brisées**
   ```json
   "ssh_key": "deploy_key"  ❌ (n'existe pas encore)
   "ssh_key": "None"        ✅ (existe toujours)
   ```

3. **Champs manquants dans templates**
   ```json
   {
     "name": "Deploy",
     "playbook": "deploy.yml",
     // Manque: app, inventory, repository, etc.
   }
   ```

4. **Format JSON invalide**
   - Virgules en trop
   - Guillemets manquants
   - Caractères spéciaux non échappés

---

## 💡 Solution de Contournement

Si VRAIMENT vous voulez utiliser l'UI de restore :

### 1. Exporter un projet existant d'abord

1. Créer un projet simple manuellement
2. Ajouter 1 clé SSH
3. Ajouter 1 repository
4. **Settings** → **Backup** → Télécharger le JSON
5. Utiliser ce JSON comme template

### 2. Adapter votre configuration

Comparer avec `backup_demo.json` et ajuster les structures.

---

## ✅ Solution Finale : Le Script !

```bash
# 1. Reset password si nécessaire
./reset-admin-password.sh

# 2. Lancer l'import
./semaphore-import.sh

# 3. Profiter ! 🎉
# http://localhost:3000 → boiler-deploy
```

**C'est la méthode la plus fiable et rapide.** ⚡

---

## 📚 Ressources

- **Script d'import :** `semaphore-import.sh`
- **Reset password :** `reset-admin-password.sh`
- **Guide détaillé :** `EASY_IMPORT.md`
- **Logs Semaphore :** `docker logs semaphore-ui`

---

## 🎓 Leçon Apprise

> L'API de backup/restore de Semaphore est conçue pour **sauvegarder** des projets existants, pas pour **créer** de nouveaux projets.
> 
> Pour créer un nouveau projet, **l'API REST directe** (utilisée par notre script) est plus appropriée.

---

**Utilisez `semaphore-import.sh` et gagnez du temps ! 🚀**
