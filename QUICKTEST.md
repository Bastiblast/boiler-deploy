# 🚀 Quick Test Guide

## Préparation (1 minute)

```bash
cd /home/basthook/devIronMenth/boiler-deploy

# 1. Vérifier que le container test est en cours
docker ps | grep boiler-test-vps

# Si pas de container, le créer :
./test-docker-vps.sh setup

# 2. Vérifier l'accès SSH
ssh -i ~/.ssh/boiler_test_rsa -p 2222 root@localhost -o StrictHostKeyChecking=no exit
echo "✅ SSH OK"
```

## Test de l'application (2 minutes)

```bash
# Lancer l'app
make run
```

### Dans l'interface :

#### 1️⃣ Tester la Navigation
- `Tab` → Switch entre environnements (docker, bast, etc.)
- `↑↓` → Naviguer entre serveurs
- `Space` → Sélectionner un serveur

#### 2️⃣ Tester la Validation
- Sélectionner `docker-web-01` (Space)
- Appuyer sur `v` (Validate)
- **Attendu** : Status change vers "✓ Ready"
- **Actuel** : Peut rester sur "Validating..." (bug connu)

#### 3️⃣ Tester le Provision
- Sélectionner le serveur (Space)
- Appuyer sur `p` (Provision)
- **Observer** :
  - Status → "⚡ Provisioning"
  - Section "📡 Live Output" apparaît en bas
  - Logs défilent en temps réel
  - Après ~2-5 min : Status → "✓ Provisioned"

#### 4️⃣ Tester le Deploy
- Serveur doit être "Provisioned"
- Sélectionner (Space)
- Appuyer sur `d` (Deploy)
- **Observer** :
  - Status → "⚡ Deploying"
  - Logs en temps réel
  - Puis → "🔍 Verifying"
  - Enfin → "✓ Deployed"

#### 5️⃣ Tester le Check
- Serveur doit être "Deployed"
- Appuyer sur `c` (Check)
- **Observer** :
  - Status → "🔍 Verifying"
  - Health check HTTP sur port 80
  - Résultat : "✓ Deployed" ou "✗ Failed"

#### 6️⃣ Tester les Logs
- Curseur sur un serveur
- Appuyer sur `l` (Logs)
- **Attendu** : Affiche les derniers logs
- `Esc` pour revenir

#### 7️⃣ Autres commandes
- `a` → Sélectionner tous les serveurs
- `r` → Refresh manuel des statuts
- `s` → Start/Stop l'orchestrator (queue)
- `x` → Clear la queue
- `q` → Quitter

## Vérification rapide

### Après Provision
```bash
# SSH dans le container
docker exec -it boiler-test-vps bash

# Vérifier les installations
which node nginx pm2 psql
systemctl status postgresql nginx
exit
```

### Après Deploy
```bash
# Tester l'application déployée
curl http://localhost:8080
# Doit retourner du HTML

# Ou dans le browser
firefox http://localhost:8080
```

## Logs de debug

```bash
# Si problème, voir les logs ansible
ls -la logs/docker/

# Dernier log
tail -100 logs/docker/*.log | tail -50

# Suivre en temps réel
tail -f logs/docker/*.log
```

## Test CLI (sans l'app)

```bash
# Provision manuel
./deploy.sh provision docker

# Deploy manuel
./deploy.sh deploy docker

# Check manuel
./deploy.sh check docker

# Tous avec --yes pour automation
./deploy.sh provision docker --yes
```

## Troubleshooting

### L'app freeze au lancement
```bash
# Vérifier les environnements
ls -la inventory/

# Au moins un doit avoir environment.json
cat inventory/docker/environment.json 2>/dev/null
```

### Pas de logs en temps réel
- Vérifier que `useScript = true` dans orchestrator
- Logs apparaissent dans section "📡 Live Output"
- Seulement pendant l'exécution (provision/deploy)

### SSH connection failed
```bash
# Re-créer les clés
rm ~/.ssh/boiler_test_rsa*
./test-docker-vps.sh cleanup
./test-docker-vps.sh setup
```

### Ansible errors
```bash
# Vérifier l'inventaire
ansible-inventory -i inventory/docker --list

# Test ping
ansible all -i inventory/docker -m ping
```

## Temps estimés

| Action     | Durée      | Description                    |
|------------|------------|--------------------------------|
| Setup      | 1-2 min    | Créer container + SSH          |
| Provision  | 3-5 min    | Installer tout sur serveur     |
| Deploy     | 1-2 min    | Déployer app + config          |
| Check      | < 5 sec    | Health check HTTP              |
| **Total**  | **5-10 min** | Workflow complet               |

## Checklist rapide

- [ ] Container running
- [ ] SSH fonctionne
- [ ] App compile (make build)
- [ ] App démarre (make run)
- [ ] Navigation OK (Tab, ↑↓)
- [ ] Selection OK (Space)
- [ ] Provision fonctionne (p)
- [ ] Logs en temps réel visibles
- [ ] Deploy fonctionne (d)
- [ ] Check fonctionne (c)
- [ ] App accessible (curl localhost:8080)

## Success = ✅ 

Quand tu vois :
```
docker-web-01    127.0.0.1    2222    web    ✓ Deployed    -
```

Et que `curl http://localhost:8080` retourne du HTML → **C'EST BON ! 🎉**
