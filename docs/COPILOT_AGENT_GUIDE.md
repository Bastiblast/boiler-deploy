# 🤖 Guide : Utiliser les Agents Copilot

## 🎯 Qu'est-ce qu'un Agent Copilot ?

Un **Agent Copilot** est une fonctionnalité qui permet de déléguer une **tâche complète et autonome** à un assistant IA spécialisé qui travaillera de manière indépendante.

### Différence : Conversation normale vs Agent

```
┌─────────────────────────────────────────────────────────────┐
│                  Conversation Normale                        │
├─────────────────────────────────────────────────────────────┤
│  Vous: "Crée un menu en Go avec Bubbletea"                 │
│  Copilot: "Voici le code..."                                │
│  Vous: "Ajoute la validation"                               │
│  Copilot: "Voici le code mis à jour..."                     │
│                                                              │
│  → Interactions multiples nécessaires                        │
│  → Vous guidez chaque étape                                 │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      Mode Agent                              │
├─────────────────────────────────────────────────────────────┤
│  Vous: "Crée une application complète de gestion           │
│         d'inventaire avec Bubbletea, incluant menu,        │
│         formulaires, validation et export YAML"             │
│                                                              │
│  Agent: [Travaille de manière autonome]                     │
│         → Crée la structure                                 │
│         → Code tous les composants                          │
│         → Teste                                             │
│         → Documente                                         │
│         → Fait un rapport final                             │
│                                                              │
│  → Une seule instruction                                    │
│  → L'agent gère tout le processus                           │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 Comment Lancer un Agent

### Dans l'interface Copilot CLI :

Il y a **plusieurs façons** selon votre interface :

### Option 1 : Via Commande Spéciale (si disponible)

```bash
# Syntaxe générale
@agent <description de la tâche>

# Exemple pour notre cas
@agent Créer un gestionnaire d'inventaire Ansible en Go avec Bubbletea. 
Structure modulaire avec menu interactif, formulaires de serveurs, 
validation IP/SSH, et export YAML. Suivre le plan dans 
docs/INVENTORY_MANAGER_PLAN.md
```

### Option 2 : Via Mention Explicite

```bash
# Demander explicitement
Je veux déléguer cette tâche à un agent autonome :

Tâche: Développer l'application inventory-manager en Go/Bubbletea
Contexte: Voir docs/INVENTORY_MANAGER_PLAN.md et docs/BUBBLETEA_VS_STREAMLIT.md
Objectif: MVP fonctionnel avec menu, ajout serveurs, export YAML
Temps estimé: 2-3 heures de développement

Peux-tu créer un agent pour cette tâche ?
```

### Option 3 : Via Interface Web Copilot

Si vous utilisez GitHub Copilot via interface web :

1. **Ouvrir le panneau Copilot**
2. **Chercher l'option "Create Agent Task" ou "Autonomous Mode"**
3. **Remplir le formulaire** :
   - Titre de la tâche
   - Description détaillée
   - Fichiers de contexte
   - Critères d'achèvement

---

## 📋 Anatomie d'une Bonne Instruction Agent

### Structure Recommandée :

```markdown
# 1. OBJECTIF CLAIR
Créer [quoi] pour [but] en utilisant [technologie]

# 2. CONTEXTE
- Fichiers de référence: docs/PLAN.md
- Contraintes: Binaire < 20MB, Go 1.21+
- Standards: Suivre structure interne/

# 3. LIVRABLES ATTENDUS
- [ ] Code source complet
- [ ] Tests unitaires
- [ ] Documentation
- [ ] Makefile
- [ ] README.md mis à jour

# 4. CRITÈRES DE SUCCÈS
- Compilation sans erreur
- Tests passent
- Interface TUI fonctionnelle
- Export YAML conforme Ansible

# 5. PRIORITÉS
1. Fonctionnel d'abord (MVP)
2. Propre et maintenable
3. Performant si possible
```

---

## 🎯 Exemple Concret pour Notre Projet

### Instruction Complète pour un Agent :

```markdown
# TÂCHE: Développer Inventory Manager Go/Bubbletea

## OBJECTIF
Créer un gestionnaire d'inventaire Ansible en Go avec interface TUI 
(Bubbletea) permettant de gérer des environnements multi-serveurs de 
manière interactive.

## CONTEXTE
- **Projet**: boiler-deploy (branche: streamlit)
- **Documentation**: 
  - Plan détaillé: docs/INVENTORY_MANAGER_PLAN.md
  - Comparaison techno: docs/BUBBLETEA_VS_STREAMLIT.md
- **Remplace**: Script bash setup.sh (1452 lignes)
- **Go version**: 1.25.0 (disponible)
- **Environnement**: Linux, déjà configuré

## STRUCTURE À CRÉER

```
boiler-deploy/
├── cmd/
│   └── inventory-manager/
│       └── main.go
├── internal/
│   ├── ui/
│   │   ├── menu.go          # Menu principal
│   │   ├── forms.go         # Formulaires
│   │   ├── styles.go        # Lipgloss styles
│   │   └── components/
│   │       ├── list.go
│   │       └── table.go
│   ├── inventory/
│   │   ├── manager.go       # Logique métier
│   │   ├── environment.go
│   │   ├── server.go
│   │   ├── validator.go     # Validation IP/Port
│   │   └── generator.go     # Génération YAML
│   ├── ssh/
│   │   └── tester.go        # Test connexions SSH
│   └── storage/
│       └── yaml.go          # Lecture/Écriture
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## FONCTIONNALITÉS MINIMALES (MVP)

### Phase 1: Structure & Menu
- [x] Initialiser go.mod avec dépendances Bubbletea
- [x] Menu principal avec navigation clavier
- [x] 4 options: Créer env, Gérer env, Valider, Quitter
- [x] Styles Lipgloss basiques

### Phase 2: Création Environnement
- [x] Formulaire: nom environnement
- [x] Checkboxes: services (web, db, monitoring)
- [x] Validation nom (alphanumerique, unique)
- [x] Création dossier inventory/[env]/

### Phase 3: Gestion Serveurs
- [x] Liste serveurs existants (table)
- [x] Formulaire ajout serveur:
  - Nom (auto-généré ou manuel)
  - IP (validation format)
  - Port (validation range)
  - User SSH (défaut: root)
  - Chemin clé SSH
- [x] Actions: Ajouter, Supprimer, Éditer

### Phase 4: Validation & Export
- [x] Validation IP format
- [x] Détection conflits IP:Port
- [x] Génération hosts.yml (format Ansible)
- [x] Génération group_vars/all.yml
- [x] Sauvegarde automatique

## DÉPENDANCES GO

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.17.1
    gopkg.in/yaml.v3 v3.0.1
)
```

## EXEMPLES DE DONNÉES

### Format hosts.yml généré:
```yaml
all:
  children:
    webservers:
      hosts:
        production-web-01:
          ansible_host: 192.168.1.10
          ansible_user: root
          ansible_port: 22
          app_port: 3000
```

### Format group_vars/all.yml:
```yaml
app_name: myapp
app_repo: https://github.com/user/repo.git
nodejs_version: "20"
app_port: "3000"
```

## CONTRAINTES

- **Binaire final**: < 20 MB
- **Pas de dépendances runtime**: Binaire statique
- **Compatible**: Linux, macOS (Windows bonus)
- **Performance**: Démarrage < 100ms
- **Qualité code**: gofmt, pas de warnings

## CRITÈRES DE SUCCÈS

1. ✅ Compilation: `go build` sans erreur
2. ✅ Lancement: `./inventory-manager` ouvre le TUI
3. ✅ Fonctionnel: Peut créer env + ajouter serveur + exporter
4. ✅ Valide: YAML généré compatible Ansible
5. ✅ Propre: Code structuré selon plan

## LIVRABLES

1. **Code source** dans structure définie
2. **go.mod/go.sum** avec dépendances
3. **Makefile** avec:
   - `make build`: Compiler
   - `make run`: Lancer
   - `make clean`: Nettoyer
4. **README.md** mis à jour avec:
   - Installation
   - Utilisation
   - Screenshots ASCII
5. **Documentation inline**: Commentaires sur fonctions publiques

## STYLE & CONVENTIONS

- Package names: lowercase, single word
- Exported: PascalCase
- Private: camelCase
- Errors: retourner plutôt que panic
- Context: passer en premier paramètre si async

## TESTS (Phase 2 - si temps)

- Tests unitaires: validator, generator
- Tests d'intégration: création environnement
- Coverage: > 70% sur logique métier

## NOTES

- Prioriser MVP fonctionnel sur code parfait
- Commenter les parties complexes (Bubbletea Update)
- Utiliser exemples Bubbletea officiels comme référence
- Git: Committer par feature (menu, forms, export, etc.)

## RESSOURCES

- Bubbletea examples: https://github.com/charmbracelet/bubbletea/tree/master/examples
- Plan détaillé: docs/INVENTORY_MANAGER_PLAN.md
- Script actuel: setup.sh (pour comprendre logique)

---

**Temps estimé**: 2-3 heures
**Priorité**: Haute
**Complexité**: Moyenne
```

---

## 🎨 Comment l'Agent Travaille

### Workflow Typique d'un Agent :

```
┌─────────────────────────────────────────┐
│  1. ANALYSE                             │
│     → Lit les documents de contexte     │
│     → Comprend les exigences            │
│     → Planifie les étapes               │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  2. CRÉATION STRUCTURE                  │
│     → Initialise go.mod                 │
│     → Crée tous les dossiers            │
│     → Fichiers vides avec TODO          │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  3. DÉVELOPPEMENT                       │
│     → Code chaque module                │
│     → Teste au fur et à mesure          │
│     → Compile régulièrement             │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  4. INTÉGRATION                         │
│     → Assemble les modules              │
│     → Tests end-to-end                  │
│     → Correction bugs                   │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  5. DOCUMENTATION                       │
│     → README.md                         │
│     → Commentaires code                 │
│     → Exemples d'utilisation            │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  6. LIVRAISON                           │
│     → Commit final                      │
│     → Rapport de ce qui a été fait      │
│     → Liste des limitations             │
│     → Suggestions d'amélioration        │
└─────────────────────────────────────────┘
```

---

## ⚡ Avantages du Mode Agent

### ✅ Pour Vous :

1. **Gain de temps** : Une instruction → Résultat complet
2. **Moins d'itérations** : L'agent anticipe les besoins
3. **Cohérence** : Structure uniforme
4. **Focus** : Vous restez sur la vision, pas les détails

### ✅ Pour le Projet :

1. **Rapidité** : Développement en quelques heures
2. **Qualité** : Code structuré dès le départ
3. **Complet** : Tests, docs, tout inclus
4. **Maintenable** : Standards respectés

---

## 🔄 Interaction avec l'Agent

Pendant que l'agent travaille, vous pouvez :

### 1. **Suivre la Progression**
L'agent vous tient informé :
```
Agent: [1/6] Création de la structure...
Agent: [2/6] Implémentation du menu principal...
Agent: [3/6] Développement des formulaires...
```

### 2. **Intervenir si Nécessaire**
```
Vous: Stop, change la couleur du menu en bleu
Agent: Compris, je modifie les styles...
Agent: Reprise du développement...
```

### 3. **Demander des Clarifications**
```
Agent: Question: Pour la validation SSH, 
       dois-je tester la connexion ou juste 
       vérifier que la clé existe ?
       
Vous: Juste vérifier que le fichier existe

Agent: Ok, je continue...
```

---

## 🎯 Quand Utiliser un Agent ?

### ✅ Bon Usage :

- Développement d'une feature complète
- Migration de code (bash → Go)
- Création de structure projet
- Refactoring important
- Génération de documentation

### ❌ Mauvais Usage :

- Petites modifications ponctuelles
- Expérimentation rapide
- Debugging interactif
- Apprentissage d'une techno

---

## 💡 Pour Notre Projet Inventory Manager

### Je vous recommande :

**Option 1 : Agent Complet** ⭐ RECOMMANDÉ
```
Lancez un agent avec l'instruction complète ci-dessus.
En 2-3h, vous aurez une application fonctionnelle.
```

**Option 2 : Itératif avec Moi**
```
On développe ensemble, feature par feature.
Plus pédagogique, vous comprenez chaque étape.
Temps : 1 journée avec interactions
```

**Option 3 : Hybride**
```
Agent fait la structure + menu (1h)
Puis on développe ensemble les features (2h)
Bon compromis apprentissage/vitesse
```

---

## 🚀 Prêt à Lancer ?

### Pour Lancer un Agent, Dites :

```
Je veux lancer un agent autonome pour créer l'inventory manager.
Utilise l'instruction complète du fichier COPILOT_AGENT_GUIDE.md
section "Exemple Concret pour Notre Projet".

Contexte:
- Branche: streamlit
- Docs: docs/INVENTORY_MANAGER_PLAN.md
- Go: 1.25.0 installé

Commence quand tu es prêt !
```

### Ou Continuons Ensemble :

```
On fait ça ensemble, étape par étape.
Commence par créer la structure et le menu principal.
```

---

**Quelle approche préférez-vous ? 🤔**
