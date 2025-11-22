# Migration vers NVM (Node Version Manager)

## Changements effectués

### 1. Provisioning (rôle `nodejs`)
- **AVANT** : Installation de NodeJS via le dépôt Nodesource avec une version fixe
- **APRÈS** : Installation de NVM uniquement, sans NodeJS

#### Avantages :
- Le serveur reste léger lors du provisioning
- Possibilité d'avoir plusieurs versions de Node sur le même serveur
- Chaque projet peut utiliser sa propre version de Node

### 2. Deployment (rôle `deploy-app`)
- Détection automatique de la version Node requise depuis `package.json` (champ `engines.node`)
- Version par défaut : Node 20 si non spécifiée
- Installation automatique de la version Node via NVM lors du premier déploiement
- Installation des gestionnaires de packages (npm, pnpm, yarn) et PM2 avec la bonne version de Node
- Tous les scripts (install, build, PM2) utilisent maintenant la version Node correcte via NVM

### 3. Fichiers modifiés
- `roles/nodejs/tasks/main.yml` - Installation NVM au lieu de NodeJS
- `roles/deploy-app/tasks/detect-app-type.yml` - Détection de la version Node requise
- `roles/deploy-app/tasks/main.yml` - Utilisation de NVM pour toutes les commandes Node

## Comment ça fonctionne

1. **Provisioning** : Installe NVM sur le serveur
2. **Premier deploy** : 
   - Lit `package.json` pour détecter la version Node requise
   - Installe cette version via `nvm install XX`
   - Installe PM2 et les package managers
   - Deploy l'application normalement

3. **Deploys suivants** : Utilise la version Node déjà installée

## Exemple de package.json

```json
{
  "name": "mon-app",
  "version": "1.0.0",
  "engines": {
    "node": ">=18.0.0"
  }
}
```
Sera détecté comme Node 18.

## Test

Le conteneur Docker de test a été recréé. Pour tester :
1. Lancer l'application : `make run`
2. Sélectionner l'environnement `docker`
3. Valider l'inventaire (V)
4. Provisionner le serveur (P) - Installera NVM
5. Déployer l'application (D) - Installera Node 20 et déploiera

## Rollback possible

Le backup de l'ancien rôle nodejs est dans `roles/nodejs.bak/`
