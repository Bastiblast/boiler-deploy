# 🔍 Détection Automatique du Type d'Application

Le système de déploiement détecte automatiquement le type d'application et adapte la configuration en conséquence.

## 📦 Types d'Applications Détectés

### Next.js
- **Détection** : Présence de `next` dans `dependencies` ou `devDependencies`
- **Configuration PM2** : 
  - Mode: `fork` (1 instance)
  - Script: `npm start`
  - Pas de cluster mode (géré nativement par Next.js)

### Nuxt.js
- **Détection** : Présence de `nuxt` dans les dépendances
- **Configuration PM2** : 
  - Mode: `fork` (1 instance)
  - Script: `npm start`

### Express
- **Détection** : Présence de `express` dans `dependencies`
- **Configuration PM2** :
  - Mode: `cluster` (instances multiples)
  - Script: fichier d'entrée détecté

### Fastify
- **Détection** : Présence de `fastify` dans `dependencies`
- **Configuration PM2** :
  - Mode: `cluster`
  - Script: fichier d'entrée détecté

### NestJS
- **Détection** : Présence de `nest` ou `@nestjs/core` dans `dependencies`
- **Configuration PM2** :
  - Mode: `cluster`
  - Script: fichier d'entrée détecté

### Node.js Standard
- **Détection** : Application Node.js sans framework spécifique
- **Configuration PM2** :
  - Mode: `cluster`
  - Script: fichier d'entrée détecté

## 🛠️ Gestionnaires de Paquets Détectés

Le système détecte automatiquement le gestionnaire de paquets utilisé :

### pnpm
- **Détection** : Présence de `pnpm-lock.yaml`
- **Commandes** :
  - Install: `pnpm install --prod`
  - Build: `pnpm run build`

### Yarn
- **Détection** : Présence de `yarn.lock`
- **Commandes** :
  - Install: `yarn install --production`
  - Build: `yarn build`

### npm (par défaut)
- **Détection** : Présence de `package-lock.json` ou défaut
- **Commandes** :
  - Install: `npm install --production`
  - Build: `npm run build`

## 🔧 Détection du Build

Le système vérifie si un script `build` existe dans `package.json` :

```json
{
  "scripts": {
    "build": "next build"  // Build détecté ✓
  }
}
```

Si présent, le build sera exécuté automatiquement avant le déploiement.

## 📍 Détection du Point d'Entrée

Pour les applications Node.js standard, le système recherche dans cet ordre :

1. `main` dans `package.json`
2. `index.js` à la racine
3. `server.js` à la racine
4. `app.js` à la racine
5. `src/index.js`
6. `src/server.js`
7. `src/app.js`

## 🎯 Configuration PM2 Adaptée

### Next.js / Nuxt.js
```javascript
{
  script: 'npm',
  args: 'start',
  instances: 1,
  exec_mode: 'fork'
}
```

### Node.js / Express / Fastify / NestJS
```javascript
{
  script: './index.js',  // ou autre point d'entrée détecté
  instances: 2,           // configurable
  exec_mode: 'cluster'
}
```

## 📊 Exemple de Détection

### Next.js avec pnpm

**Fichiers détectés** :
- `pnpm-lock.yaml` ✓
- `next` dans dependencies ✓
- Script `build` dans package.json ✓

**Configuration appliquée** :
- Package manager: `pnpm`
- Type: `nextjs`
- Build: Oui avec `pnpm run build`
- PM2: Mode fork, npm start

### Express avec npm

**Fichiers détectés** :
- `package-lock.json` ✓
- `express` dans dependencies ✓
- `index.js` à la racine ✓

**Configuration appliquée** :
- Package manager: `npm`
- Type: `express`
- Build: Non (pas de script build)
- PM2: Mode cluster, ./index.js

## 🔍 Voir les Informations Détectées

Lors du déploiement, les informations sont affichées :

```
TASK [deploy-app : Display detected configuration]
ok: [server] => 
  msg:
    - "Application Type: nextjs"
    - "Package Manager: pnpm"
    - "Needs Build: true"
    - "Entry File: N/A"
```

## 🛠️ Forcer un Type Spécifique

Si la détection automatique ne convient pas, vous pouvez forcer le type dans `group_vars/all.yml` :

```yaml
# Force le type d'application
app_type_override: "nodejs"  # ou "nextjs", "express", etc.

# Force le gestionnaire de paquets
package_manager_override: "pnpm"  # ou "yarn", "npm"

# Force le fichier d'entrée
app_entry_file_override: "dist/main.js"
```

## 📝 Ajouter un Nouveau Type

Pour ajouter un nouveau type d'application :

1. Modifiez `roles/deploy-app/tasks/detect-app-type.yml`
2. Ajoutez la détection dans la tâche "Detect application type"
3. Créez un template `ecosystem.config.VOTETYPE.js.j2`
4. Ajoutez la condition dans `main.yml`

## 🐛 Debugging

Si la détection ne fonctionne pas correctement :

```bash
# Voir les logs de détection
ansible-playbook playbooks/deploy.yml -i inventory/hostinger -vv

# Vérifier manuellement le package.json
ssh deploy@72.61.146.126 'cat /var/www/APP/current/package.json | jq .dependencies'
```

## 🎉 Avantages

✅ **Simplicité** : Aucune configuration manuelle nécessaire  
✅ **Flexibilité** : Supporte plusieurs frameworks et gestionnaires de paquets  
✅ **Intelligent** : Détecte automatiquement les besoins de build  
✅ **Adaptable** : Configuration PM2 optimisée selon le type d'app  
✅ **Maintenable** : Facile d'ajouter de nouveaux types  

## 📚 Exemples de Configurations

### Monorepo avec pnpm
```json
{
  "dependencies": {
    "next": "15.0.0"
  },
  "scripts": {
    "build": "next build",
    "start": "next start"
  }
}
```
→ Détection: Next.js + pnpm + build requis

### API Express simple
```json
{
  "dependencies": {
    "express": "4.18.0"
  },
  "main": "server.js"
}
```
→ Détection: Express + npm + pas de build + entry: server.js

### Application NestJS
```json
{
  "dependencies": {
    "@nestjs/core": "10.0.0"
  },
  "scripts": {
    "build": "nest build",
    "start:prod": "node dist/main"
  }
}
```
→ Détection: NestJS + npm + build requis + entry détecté
