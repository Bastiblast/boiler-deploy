# Debug du Problème de Freeze lors du Provisioning

## Symptôme
Lorsqu'on appuie sur 'p' pour provisionner, l'application affiche "Loading..." puis ne fait plus rien.

## Analyse
Le problème venait de l'initialisation du TagSelector qui n'avait pas accès aux dimensions du terminal (`width` et `height`).

## Corrections Appliquées

### 1. Gestion des Dimensions du Terminal dans WorkflowView
- Ajout des champs `width` et `height` dans la structure `WorkflowView`
- Capture et stockage des dimensions lors de la réception de `tea.WindowSizeMsg`
- Les messages `WindowSizeMsg` sont maintenant traités en priorité avant le TagSelector

### 2. Initialisation du TagSelector avec Dimensions
- Lors de la création du TagSelector (touche 'p' ou 'd'), on initialise maintenant `selector.width` et `selector.height` avec les dimensions actuelles du terminal
- Ajout d'une dimension par défaut (80) dans le `View()` du TagSelector au cas où les dimensions ne seraient pas encore définies

### 3. Ajout de Logs de Debug
Des logs de debug ont été ajoutés pour tracer l'exécution :
- Dans `handleMainKeys` : Quand 'p' est pressé
- Dans `Update` du workflow : Quand le TagSelector est confirmé/annulé
- Dans `executeActionWithTags` : Quand l'action est exécutée
- Dans `QueueProvisionWithTags` de l'orchestrateur : Quand les tâches sont ajoutées à la queue

## Fichiers Modifiés
1. `internal/ui/workflow_view.go`
   - Ajout de champs `width` et `height`
   - Réorganisation de la gestion des messages WindowSize
   - Initialisation du TagSelector avec les dimensions
   - Ajout de logs de debug

2. `internal/ui/tag_selector.go`
   - Remplacement de "Loading..." par une dimension par défaut

3. `internal/ansible/orchestrator.go`
   - Ajout de logs dans `QueueProvisionWithTags`

## Test Manuel

### Prérequis
Assurez-vous que le conteneur Docker de test est en cours d'exécution :
```bash
docker ps | grep boiler-test-vps
```

### Étapes de Test
1. Lancer l'application :
   ```bash
   make run
   ```

2. Navigation :
   - Sélectionner "Working with Inventory"
   - Sélectionner l'environnement "docker"

3. Sélection du serveur :
   - Utiliser les flèches pour naviguer vers `docker-web-01`
   - Appuyer sur ESPACE pour sélectionner le serveur

4. Test du Provisioning :
   - Appuyer sur 'p'
   - **VÉRIFIER** : Le sélecteur de tags doit s'afficher avec la liste des tags
   - **NE DOIT PAS** : Afficher "Loading..." indéfiniment

5. Confirmation :
   - Appuyer sur ENTER pour confirmer (tous les tags sont sélectionnés par défaut)
   - **VÉRIFIER** : Le provisioning doit démarrer
   - **OBSERVER** : Les logs de debug en bas de l'écran

### Logs Attendus
En bas de l'écran, vous devriez voir :
```
[DEBUG] 'p' key pressed
[DEBUG] Opening tag selector (width=XXX, height=YYY)
[DEBUG] Tag selector confirmed. Tags: [liste des tags]
[DEBUG] executeActionWithTags called: action=provision, tags=[...], servers=[docker-web-01]
[DEBUG] Starting orchestrator
[DEBUG] Queueing provision with tags: [...]
[ORCHESTRATOR] QueueProvisionWithTags called with 1 servers: [docker-web-01], tags: [...]
[ORCHESTRATOR] Adding provision action for server: docker-web-01with tags: [...]
```

### Résultats Attendus
1. Le TagSelector s'affiche immédiatement (pas de "Loading...")
2. Les tags sont affichés organisés par catégories
3. Après confirmation, le provisioning démarre
4. Le status du serveur passe à "Provisioning"
5. Les logs Ansible apparaissent en temps réel en bas de l'écran

## Nettoyage des Logs de Debug
Une fois le problème confirmé résolu, les logs de debug (lignes avec `fmt.Printf("[DEBUG]...)`) peuvent être supprimés pour une version de production.

## Script de Test
Utilisez le script de test automatisé :
```bash
./tests/test_provision_ui.sh
```
