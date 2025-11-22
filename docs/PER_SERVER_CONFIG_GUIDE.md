# 🔧 Configuration Par Serveur - Guide Complet

## 📋 Vue d'Ensemble

La nouvelle architecture permet de configurer **chaque serveur web indépendamment** avec son propre repository Git, version Node.js, et port d'application.

## 🎯 Pourquoi ce Changement ?

### Avant (Configuration Globale)
```
❌ Un seul repo pour tout l'environnement
❌ Même version Node.js partout
❌ Même port pour tous les serveurs
❌ Impossible de déployer des microservices différents
```

### Après (Configuration Par Serveur)
```
✅ Chaque serveur web a son propre repo
✅ Versions Node.js différentes possibles
✅ Ports d'application uniques
✅ Parfait pour microservices et applications multiples
```

---

## 🏗️ Nouvelle Structure

### Création d'Environnement (Simplifié)

```
Environment Form:
  ├─ Nom environnement       ✓ (ex: "production")
  └─ Services               ✓ (Web/Database/Monitoring)

  → Plus de config app globale!
```

### Ajout de Serveur (Étendu)

#### Pour TOUS les serveurs:
```
Common Fields:
  ├─ Nom                    ✓ (auto-généré ou manuel)
  ├─ IP                     ✓ (192.168.1.10)
  ├─ SSH Port               ✓ (22)
  ├─ SSH User               ✓ (root)
  ├─ SSH Key Path           ✓ (~/.ssh/id_rsa)
  └─ Type                   ✓ (Web/DB/Monitoring)
```

#### SEULEMENT pour serveurs Web:
```
Application Configuration:
  ├─ Application Port       ✓ (3000, 4000, 5000...)
  ├─ Git Repository         ✓ (https://github.com/user/frontend.git)
  ├─ Git Branch             ✓ (main, develop, v2...)
  └─ Node.js Version        ✓ (18, 20, 21...)
```

#### Pour DB/Monitoring:
```
→ Pas de configuration application
→ SSH uniquement
```

---

## 💡 Cas d'Usage Pratiques

### Exemple 1: Architecture Microservices

```yaml
Environment: production

Serveur: production-web-01
  Type: Web
  IP: 192.168.1.10
  App Port: 3000
  Git Repo: https://github.com/company/frontend.git
  Branch: main
  Node: 20

Serveur: production-web-02
  Type: Web
  IP: 192.168.1.11
  App Port: 4000
  Git Repo: https://github.com/company/api-v1.git    ← Différent!
  Branch: main
  Node: 18                                            ← Version différente!

Serveur: production-web-03
  Type: Web
  IP: 192.168.1.12
  App Port: 5000
  Git Repo: https://github.com/company/admin.git     ← Encore différent!
  Branch: develop                                     ← Branche différente!
  Node: 20

Serveur: production-db-01
  Type: Database
  IP: 192.168.1.20
  → Pas de config app
```

### Exemple 2: Test de Versions Node.js

```yaml
Environment: staging

Serveur: staging-web-01
  App Port: 3000
  Git Repo: https://github.com/company/app.git
  Branch: main
  Node: 18        ← Ancienne version stable

Serveur: staging-web-02
  App Port: 3001
  Git Repo: https://github.com/company/app.git      ← Même repo
  Branch: main
  Node: 20        ← Test nouvelle version
```

### Exemple 3: Branches de Développement

```yaml
Environment: dev

Serveur: dev-web-01
  App Port: 3000
  Git Repo: https://github.com/company/app.git
  Branch: feature-auth    ← Feature branch
  Node: 20

Serveur: dev-web-02
  App Port: 3001
  Git Repo: https://github.com/company/app.git
  Branch: feature-ui      ← Autre feature
  Node: 20
```

---

## 📁 Structure Générée (Option A: host_vars)

```
inventory/production/
├── hosts.yml              # Connexions SSH uniquement
├── config.yml             # État de l'environnement
├── group_vars/
│   └── all.yml           # Variables communes (timezone, deploy_user)
└── host_vars/
    ├── production-web-01.yml    # Config app web-01
    ├── production-web-02.yml    # Config app web-02
    └── production-web-03.yml    # Config app web-03
    # Pas de fichier pour DB/Monitoring
```

### Contenu de `hosts.yml`

```yaml
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: 192.168.1.10
          ansible_user: root
          ansible_port: 22
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          ansible_become: true
        production-web-02:
          ansible_host: 192.168.1.11
          ansible_user: root
          ansible_port: 22
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          ansible_become: true
    dbservers:
      hosts:
        production-db-01:
          ansible_host: 192.168.1.20
          ansible_user: root
          ansible_port: 22
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          ansible_become: true
```

### Contenu de `group_vars/all.yml`

```yaml
deploy_user: root
timezone: Europe/Paris
```

### Contenu de `host_vars/production-web-01.yml`

```yaml
app_port: 3000
app_repo: https://github.com/company/frontend.git
app_branch: main
nodejs_version: "20"
deploy_user: root
```

### Contenu de `host_vars/production-web-02.yml`

```yaml
app_port: 4000
app_repo: https://github.com/company/api-v1.git
app_branch: main
nodejs_version: "18"
deploy_user: root
```

---

## 🎨 Formulaire dans l'Interface

### Création d'Environnement

```
╔════════════════════════════════════════════════════════════╗
║  📝 Create New Environment                                 ║
╚════════════════════════════════════════════════════════════╝

▶ Environment name:
  production

Services to enable:
▶ [✓] Web servers
  [ ] Database servers
  [ ] Monitoring

[Tab/↑↓] Navigate  [Space] Toggle  [Enter] Create  [Esc] Cancel
```

### Ajout d'un Serveur Web

```
╔════════════════════════════════════════════════════════════╗
║  ➕ Add New Server                                         ║
╚════════════════════════════════════════════════════════════╝

▶ Server name:
  production-web-01

  IP address:
  192.168.1.10

  SSH port:
  22

  SSH user:
  root

  SSH key path:
  ~/.ssh/id_rsa

─── Application Configuration ───

  Application port:
  3000

  Git repository:
  https://github.com/company/frontend.git

  Git branch:
  main

  Node.js version:
  20

  Server type:
  [Web]  Database   Monitoring

[Tab/↑↓] Navigate  [←→] Change type  [Enter] Save  [Esc] Cancel
```

### Ajout d'un Serveur DB (Plus Simple)

```
╔════════════════════════════════════════════════════════════╗
║  ➕ Add New Server                                         ║
╚════════════════════════════════════════════════════════════╝

▶ Server name:
  production-db-01

  IP address:
  192.168.1.20

  SSH port:
  22

  SSH user:
  root

  SSH key path:
  ~/.ssh/id_rsa

  Server type:
   Web  [Database]  Monitoring

[Tab/↑↓] Navigate  [←→] Change type  [Enter] Save  [Esc] Cancel

→ Pas de configuration application pour DB
```

---

## ⌨️ Workflow Complet

### 1. Créer un Environnement

```bash
# Lancer l'application
./bin/inventory-manager

# Dans le menu
→ Create new environment
  Nom: production
  Services: [✓] Web [✓] Database
→ Enter
```

**Résultat**: Structure vide créée

### 2. Ajouter un Frontend

```bash
→ Manage existing environment
→ production
→ Press 'a' (Add)

Remplir:
  Name: (vide pour auto: production-web-01)
  IP: 192.168.1.10
  SSH Port: 22
  SSH User: root
  SSH Key: ~/.ssh/id_rsa
  
  Type: [Web] ← (par défaut)
  
  App Port: 3000
  Git Repo: https://github.com/company/frontend.git
  Git Branch: main
  Node Version: 20

→ Enter pour sauver
```

### 3. Ajouter une API

```bash
→ Press 'a' (Add) encore

Remplir:
  Name: (vide pour auto: production-web-02)
  IP: 192.168.1.11
  SSH Port: 22
  
  Type: [Web]
  
  App Port: 4000                                    ← Port différent
  Git Repo: https://github.com/company/api.git     ← Repo différent
  Git Branch: v2                                    ← Branche différente
  Node Version: 18                                  ← Version différente

→ Enter
```

### 4. Ajouter une Base de Données

```bash
→ Press 'a' (Add)

Remplir:
  Name: production-db-01
  IP: 192.168.1.20
  SSH Port: 22
  
  Type: ←→ pour sélectionner [Database]

→ Enter

→ Pas de champs app, c'est normal!
```

### 5. Vérifier la Configuration

```bash
→ Press 'g' (Generate summary)
```

Affiche:
```
Environment: production
═══════════════════════════════

Services:
  ✓ Web servers
  ✓ Database servers

Configuration:
  Deploy user: root
  Timezone: Europe/Paris

Servers (3 total):
  • production-web-01 (web) - 192.168.1.10:3000
  • production-web-02 (web) - 192.168.1.11:4000
  • production-db-01 (db) - 192.168.1.20:0
```

### 6. Vérifier les Fichiers Générés

```bash
ls -la inventory/production/

# Résultat:
config.yml
hosts.yml
group_vars/
  └── all.yml
host_vars/
  ├── production-web-01.yml
  └── production-web-02.yml
  # Pas de fichier pour db-01
```

---

## 🔍 Validation et Erreurs

### Validation Automatique

**Pour serveurs Web:**
```
✓ Application port requis (erreur si vide)
✓ Git repository requis (erreur si vide)
✓ Git branch (défaut: main si vide)
✓ Node.js version (défaut: 20 si vide)
```

**Pour DB/Monitoring:**
```
✓ Pas de validation app (champs non affichés)
```

### Exemples d'Erreurs

```
❌ "application port is required for web servers"
   → Vous n'avez pas rempli le port

❌ "git repository is required for web servers"
   → Vous n'avez pas rempli le repo

❌ "IP:Port conflict with server production-web-01 (192.168.1.10:22)"
   → Conflit détecté
```

---

## 🎯 Bonnes Pratiques

### Nommage des Serveurs

```
✅ Bon:
  production-frontend-01
  production-api-01
  production-admin-01
  
❌ Éviter:
  web1, web2 (pas de contexte)
  server-a, server-b (ambigü)
```

### Organisation des Ports

```
Frontend:     3000-3099
APIs:         4000-4099
Admin:        5000-5099
Monitoring:   9000-9099
```

### Gestion des Branches

```
Production:   main, master
Staging:      develop, staging
Dev:          feature-*, develop
```

### Versions Node.js

```
Stable LTS:   18, 20
Latest:       21
Legacy:       16 (à éviter)
```

---

## 🐛 Dépannage

### "Champs app ne s'affichent pas"

**Cause**: Type de serveur n'est pas "Web"

**Solution**: Utilisez ←→ pour sélectionner "Web"

### "Trop de champs dans le formulaire"

**Cause**: Vous avez sélectionné "Web" alors que vous voulez DB

**Solution**: Changez le type avec ←→

### "host_vars/ vide"

**Cause**: Aucun serveur web configuré

**Solution**: Ajoutez au moins un serveur de type "Web"

### "Variables non trouvées par Ansible"

**Vérifiez**:
```bash
# Les fichiers doivent exister
ls inventory/production/host_vars/production-web-01.yml

# Le nom du serveur doit correspondre
grep production-web-01 inventory/production/hosts.yml
```

---

## 📊 Comparaison Avant/Après

### Avant (Global)

```yaml
# group_vars/production.yml
app_name: production-app
app_repo: https://github.com/user/repo.git    ← UN SEUL REPO
app_branch: main
nodejs_version: "20"                           ← UNE VERSION
app_port: "3000"                               ← UN PORT

# Tous les serveurs utilisent la même config
```

### Après (Par Serveur)

```yaml
# host_vars/production-web-01.yml
app_port: 3000
app_repo: https://github.com/user/frontend.git
app_branch: main
nodejs_version: "20"

# host_vars/production-web-02.yml
app_port: 4000
app_repo: https://github.com/user/api.git      ← DIFFÉRENT!
app_branch: v2                                  ← DIFFÉRENT!
nodejs_version: "18"                            ← DIFFÉRENT!
```

---

## 🚀 Exemples Complets

### Cas 1: Application Monolithique

```
Environment: monolith
  └─ monolith-web-01
      ├─ Repo: https://github.com/company/app.git
      ├─ Branch: main
      ├─ Node: 20
      └─ Port: 3000
```

### Cas 2: Microservices

```
Environment: microservices
  ├─ microservices-frontend-01
  │   ├─ Repo: https://github.com/company/frontend.git
  │   ├─ Branch: main
  │   ├─ Node: 20
  │   └─ Port: 3000
  │
  ├─ microservices-auth-01
  │   ├─ Repo: https://github.com/company/auth-service.git
  │   ├─ Branch: main
  │   ├─ Node: 18
  │   └─ Port: 4000
  │
  ├─ microservices-payment-01
  │   ├─ Repo: https://github.com/company/payment-service.git
  │   ├─ Branch: v2
  │   ├─ Node: 20
  │   └─ Port: 5000
  │
  └─ microservices-db-01
      └─ (SSH seulement)
```

### Cas 3: Multi-tenant

```
Environment: multi-tenant
  ├─ tenant-client-a-01
  │   ├─ Repo: https://github.com/company/app.git
  │   ├─ Branch: client-a-custom
  │   ├─ Node: 20
  │   └─ Port: 3000
  │
  └─ tenant-client-b-01
      ├─ Repo: https://github.com/company/app.git
      ├─ Branch: client-b-custom
      ├─ Node: 20
      └─ Port: 4000
```

---

## 📚 Ressources

- [Main README](../INVENTORY_MANAGER_README.md)
- [Server Management Guide](SERVER_MANAGEMENT_GUIDE.md)
- [Architecture Plan](INVENTORY_MANAGER_PLAN.md)

---

**Configuration flexible pour architectures modernes! 🎉**
