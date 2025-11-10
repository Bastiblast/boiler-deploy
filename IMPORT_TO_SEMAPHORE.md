# 📋 Guide Complet : Importer votre Inventaire dans Semaphore UI

## 🎯 Vue d'ensemble

Vous avez actuellement :
- ✅ Semaphore UI démarré sur http://localhost:3000
- ✅ Configuration globale dans `group_vars/all.yml`
- ✅ Environnements vides : `dev`, `production`, `test-final`

## 📝 Étape par Étape : Configuration Complète

### 1️⃣ **Se Connecter à Semaphore**

1. Ouvrir : http://localhost:3000
2. Login : `admin`
3. Password : `admin`
4. ⚠️ **Changer le mot de passe immédiatement** : 
   - Cliquer sur l'avatar en haut à droite
   - **User Settings** → **Change Password**

---

### 2️⃣ **Créer un Projet**

1. Dans la sidebar gauche → **Projects**
2. Cliquer sur **+ New Project**
3. Remplir :
   - **Name:** `boiler-deploy`
   - **Alert Chat ID:** _(laisser vide)_
4. Cliquer sur **Create**

---

### 3️⃣ **Configurer les Clés SSH** (Key Store)

1. Dans votre projet → **Key Store** (menu latéral)
2. Cliquer sur **+ New Key**
3. Remplir :
   - **Name:** `deploy_key`
   - **Type:** `SSH Key`
   - **Login (Optional):** `root` (ou votre user SSH)
   
4. **Private Key:** Copier votre clé privée SSH
   ```bash
   # Dans votre terminal :
   cat /home/basthook/.ssh/Hosting
   ```
   Copier TOUT le contenu (de `-----BEGIN` à `-----END`)

5. Cliquer sur **Create**

---

### 4️⃣ **Créer le Repository Local**

1. Dans votre projet → **Repositories**
2. Cliquer sur **+ New Repository**
3. Remplir :
   - **Name:** `local-playbooks`
   - **URL:** `/ansible`
   - **Branch:** `streamlit` (notre branche actuelle)
   - **Access Key:** `None`

4. Cliquer sur **Create**

---

### 5️⃣ **Créer l'Inventaire Production**

#### Configuration actuelle détectée :
- **App:** myapp
- **Repo Git:** https://github.com/Bastiblast/ansible-next-test.git
- **Node.js:** v20
- **Port:** 3002
- **User:** root
- **SSH Key:** /home/basthook/.ssh/Hosting

#### Créer l'inventaire :

1. Dans votre projet → **Inventory**
2. Cliquer sur **+ New Inventory**
3. Remplir :
   - **Name:** `production`
   - **Type:** `Static`
   
4. **Inventory Content** (copier ce YAML) :

```yaml
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: YOUR_SERVER_IP_HERE
          ansible_user: root
          ansible_port: 22
          ansible_python_interpreter: /usr/bin/python3
          ansible_ssh_private_key_file: /home/basthook/.ssh/Hosting
          ansible_become: yes
          app_port: 3002
```

5. **⚠️ Remplacer `YOUR_SERVER_IP_HERE`** par l'IP de votre serveur
6. Cliquer sur **Create**

---

### 6️⃣ **Créer l'Environment (Variables)**

1. Dans votre projet → **Environment**
2. Cliquer sur **+ New Environment**
3. Remplir :
   - **Name:** `production-vars`
   
4. **Environment Variables** (format JSON) :

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

5. Cliquer sur **Create**

---

### 7️⃣ **Créer un Task Template**

1. Dans votre projet → **Task Templates**
2. Cliquer sur **+ New Template**
3. Remplir :
   - **Name:** `Provision Server`
   - **Playbook Filename:** `playbooks/provision.yml`
   - **Inventory:** `production`
   - **Environment:** `production-vars`
   - **SSH Key:** `deploy_key`
   
4. Cliquer sur **Create**

---

## 🔧 Ajouter un Serveur Supplémentaire

Pour ajouter `production-web-02` :

1. **Inventory** → `production` → **Edit**
2. Modifier le YAML :

```yaml
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: 192.168.1.10
          ansible_user: root
          app_port: 3002
        
        production-web-02:
          ansible_host: 192.168.1.11
          ansible_user: root
          app_port: 3003
```

3. **Save**

---

## 🗄️ Ajouter Database + Monitoring

```yaml
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: 192.168.1.10
          ansible_user: root
          app_port: 3002
    
    dbservers:
      hosts:
        production-db-01:
          ansible_host: 192.168.1.20
          ansible_user: root
    
    monitoring:
      hosts:
        production-monitoring-01:
          ansible_host: 192.168.1.30
          ansible_user: root
      vars:
        prometheus_targets:
          - targets:
              - '192.168.1.10:9100'
              - '192.168.1.20:9100'
            labels:
              job: 'node_exporter'
```

---

## 🆘 Dépannage Rapide

### Erreur SSH
```bash
# Tester manuellement
ssh -i /home/basthook/.ssh/Hosting root@YOUR_IP

# Vérifier permissions
chmod 600 /home/basthook/.ssh/Hosting
```

### Repository non trouvé
```bash
# Vérifier le montage Docker
docker exec -it semaphore-ui ls -la /ansible
```

---

## ✅ Checklist Configuration

- [ ] Se connecter à http://localhost:3000
- [ ] Changer mot de passe admin
- [ ] Créer projet `boiler-deploy`
- [ ] Ajouter clé SSH `deploy_key`
- [ ] Créer repository `local-playbooks`
- [ ] Créer inventaire `production`
- [ ] Remplacer les IPs
- [ ] Créer environment `production-vars`
- [ ] Créer template `Provision Server`
- [ ] Tester connexion SSH
- [ ] Exécuter premier playbook

---

**C'est prêt ! 🎉**

Voir aussi : `SEMAPHORE_GUIDE.md` pour plus de détails
