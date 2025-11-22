# 🏥 Health Check Implementation Guide

## Overview

Le health check vérifie que l'application déployée est accessible et fonctionne correctement avant de marquer le déploiement comme réussi.

## ✨ Fonctionnalités

### 1. **Multi-Port Health Check**

Le système essaie automatiquement plusieurs ports :
- **Port 80** : Nginx (proxy inverse)
- **Port d'application** : Port configuré dans inventory (ex: 3000)

```go
ports := []int{80}
if server.AppPort > 0 && server.AppPort != 80 {
    ports = append(ports, server.AppPort)
}
```

### 2. **Retry Logic avec Backoff**

Le health check réessaie jusqu'à 5 fois avec des délais croissants :

| Tentative | Délai avant |
|-----------|-------------|
| 1 | Immédiat |
| 2 | 2 secondes |
| 3 | 3 secondes |
| 4 | 5 secondes |
| 5 | 8 secondes |

**Total : ~30 secondes maximum**

### 3. **Diagnostic Intelligent**

Si curl échoue, le système vérifie avec `nc` (netcat) si le port est ouvert :
- ✅ Port ouvert → Problème HTTP/application
- ❌ Port fermé → Service non démarré ou firewall

```bash
nc -zv -w 3 192.168.1.100 80
```

### 4. **Configuration Flexible**

```yaml
# inventory/{env}/config.yml
health_check_enabled: true      # Enable/disable
health_check_timeout: 30s       # Timeout per check
health_check_retries: 5         # Number of retries
```

### 5. **Skip Health Check**

Possibilité de skip le health check via l'API :

```go
orchestrator.SkipNextHealthCheck()
orchestrator.QueueDeploy(serverNames, priority)
```

## 🔧 Configuration des Ports

### Dans l'Inventory

```yaml
# inventory/production/hosts.yml
servers:
  - name: web-01
    ip: 192.168.1.100
    port: 22
    ssh_user: deploy
    ssh_key_path: ~/.ssh/id_rsa
    type: web
    app_port: 3000        # ← Port de l'application
```

### Configuration Nginx

Assurez-vous que Nginx proxie correctement :

```nginx
# /etc/nginx/sites-available/default
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

## 🐛 Debugging

### Logs Détaillés

Le health check produit des logs détaillés dans `debug.log` :

```bash
# Surveiller en temps réel
tail -f debug.log | grep -E "HEALTH|Health|health"
```

**Exemple de logs :**

```
[EXECUTOR] Health check starting for: http://192.168.1.100:80
[EXECUTOR] Health check attempt 1/5 failed: exit status 7 (output: )
[EXECUTOR] Port 80 appears closed or unreachable: nc: connect to 192.168.1.100 port 80 (tcp) failed: Connection refused
[EXECUTOR] Health check retry 2/5 after 2s delay
[EXECUTOR] ✓ Health check successful on attempt 2 (1234 bytes)
```

### Erreurs Communes

#### ❌ Connection Refused
```
[EXECUTOR] Port 80 appears closed or unreachable
```

**Solutions :**
1. Vérifier que Nginx est démarré : `systemctl status nginx`
2. Vérifier le firewall : `sudo ufw status`
3. Ouvrir le port : `sudo ufw allow 80/tcp`

#### ❌ Connection Timeout
```
[EXECUTOR] Health check attempt 5/5 failed: curl: (28) Connection timed out
```

**Solutions :**
1. Vérifier que le serveur est accessible : `ping IP`
2. Vérifier la route réseau
3. Vérifier le firewall distant

#### ❌ Empty Reply from Server
```
[EXECUTOR] curl failed: Empty reply from server
```

**Solutions :**
1. L'application n'écoute pas sur le port
2. Vérifier PM2 : `pm2 list`
3. Vérifier les logs de l'app : `pm2 logs`

#### ❌ HTTP 502 Bad Gateway
```
[EXECUTOR] curl failed: HTTP 502
```

**Solutions :**
1. Nginx fonctionne mais l'application backend est down
2. Vérifier PM2 : `pm2 list`
3. Vérifier la config Nginx proxy_pass

### Tests Manuels

#### 1. Test Curl Direct
```bash
# Test port 80 (Nginx)
curl -v http://192.168.1.100:80

# Test port app direct (si firewall permet)
curl -v http://192.168.1.100:3000
```

#### 2. Test Netcat
```bash
# Vérifier si le port est ouvert
nc -zv 192.168.1.100 80

# Timeout personnalisé
nc -zv -w 3 192.168.1.100 80
```

#### 3. SSH et Test Local
```bash
# Se connecter au serveur
ssh deploy@192.168.1.100

# Tester en local
curl http://localhost:80
curl http://localhost:3000

# Vérifier les services
systemctl status nginx
pm2 list
pm2 logs --lines 50
```

## 🎯 Workflow Complet

### 1. Déploiement Normal
```
Deploy → Build → Restart PM2 → Wait 2s → Health Check (port 80) → Success ✓
                                              ↓ Failed
                                         Retry (port 3000) → Success ✓
```

### 2. Skip Health Check
```go
// Dans le code
orchestrator.SkipNextHealthCheck()
orchestrator.QueueDeploy([]string{"web-01"}, 0)

// Résultat
Deploy → Build → Restart PM2 → Mark as Deployed ✓
```

### 3. Health Check Désactivé
```yaml
# config.yml
health_check_enabled: false
```

```
Deploy → Build → Restart PM2 → Mark as Deployed ✓ (no check)
```

## 📊 Architecture du Code

```
orchestrator.go (Deploy Action)
       ↓
  performHealthCheck ?
       ↓ Yes
  Try ports [80, 3000]
       ↓
  executor.HealthCheck()
       ↓
  Retry up to 5 times with backoff
       ↓
  Success → StateDeployed → deploySuccessCb()
       ↓                            ↓
  Failed → StateFailed      Browser prompt: "Press 'o'"
```

## 🔐 Sécurité

### Firewall Configuration

```bash
# Sur le serveur distant
sudo ufw allow 80/tcp      # HTTP
sudo ufw allow 443/tcp     # HTTPS
sudo ufw allow 22/tcp      # SSH
sudo ufw enable

# Vérifier
sudo ufw status verbose
```

### Port Application (3000)

**Ne PAS exposer directement !** Utilisez Nginx comme proxy :
- ✅ Expose port 80/443 (Nginx)
- ❌ N'exposez pas port 3000
- Nginx proxie vers localhost:3000

## 💡 Tips

### 1. Développement Local
Pour tester sans serveur distant :
```go
orchestrator.SetHealthCheckEnabled(false)
```

### 2. Environnement de Staging
Augmenter les retries :
```yaml
health_check_retries: 10
health_check_timeout: 60s
```

### 3. Production
Configuration stricte :
```yaml
health_check_enabled: true
health_check_retries: 3
health_check_timeout: 30s
```

### 4. Monitoring
Ajouter un endpoint de health :
```javascript
// Dans votre app Node.js
app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});
```

Puis modifier le health check pour utiliser `/health` :
```go
url := fmt.Sprintf("http://%s:%d/health", ip, port)
```

## 🚀 Next Steps

Pour améliorer le health check :

1. **HTTP Status Check** : Vérifier le code de statut (200, 301, etc.)
2. **Content Validation** : Vérifier le contenu de la réponse
3. **SSL Support** : Support HTTPS avec certificats
4. **Custom Endpoints** : Configurer l'endpoint par serveur
5. **Metrics** : Collecter le temps de réponse

```yaml
# Future config
health_check:
  endpoint: "/api/health"
  expected_status: [200, 301]
  expected_content: "ok"
  ssl_verify: true
```
