#!/bin/bash
# Deployment script spécifique pour Hostinger
# Usage: ./deploy-hostinger.sh [provision|deploy|update|rollback]

set -e

ACTION=${1:-deploy}
INVENTORY="inventory/hostinger"

# Couleurs pour l'affichage
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Déploiement Hostinger - IP: 72.61.146.126${NC}"
echo -e "${BLUE}========================================${NC}"

# Vérifier si le fichier d'inventaire existe
if [ ! -f "$INVENTORY" ]; then
    echo -e "${YELLOW}⚠️  Le fichier d'inventaire n'existe pas.${NC}"
    echo "Création depuis le template..."
    cp inventory/hostinger/hosts.yml.example $INVENTORY
    echo -e "${GREEN}✓ Fichier créé : $INVENTORY${NC}"
    echo "Veuillez le vérifier et ajuster si nécessaire."
fi

# Vérifier la connectivité
echo -e "\n${BLUE}📡 Vérification de la connectivité...${NC}"
if ansible all -i $INVENTORY -m ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Connexion au serveur établie${NC}"
else
    echo -e "${YELLOW}⚠️  Impossible de se connecter au serveur.${NC}"
    echo "Vérifiez votre accès SSH à 72.61.146.126"
    echo "Commande de test : ssh root@72.61.146.126"
    exit 1
fi

# Exécuter l'action demandée
case $ACTION in
    provision)
        echo -e "\n${BLUE}🚀 Provisioning complet du serveur Hostinger...${NC}"
        echo "Cela va installer : PostgreSQL, Node.js, Nginx, PM2, sécurité..."
        read -p "Continuer ? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ansible-playbook playbooks/provision.yml -i $INVENTORY
        else
            echo "Annulé."
            exit 0
        fi
        ;;
    deploy)
        echo -e "\n${BLUE}🚢 Déploiement de l'application sur Hostinger...${NC}"
        ansible-playbook playbooks/deploy.yml -i $INVENTORY
        ;;
    update)
        echo -e "\n${BLUE}🔄 Mise à jour rapide de l'application...${NC}"
        ansible-playbook playbooks/update.yml -i $INVENTORY
        ;;
    rollback)
        echo -e "\n${BLUE}⏪ Rollback vers la version précédente...${NC}"
        read -p "Êtes-vous sûr de vouloir rollback ? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ansible-playbook playbooks/rollback.yml -i $INVENTORY
        else
            echo "Annulé."
            exit 0
        fi
        ;;
    check)
        echo -e "\n${BLUE}🔍 Vérification de la configuration...${NC}"
        ansible-playbook playbooks/deploy.yml -i $INVENTORY --check
        ;;
    status)
        echo -e "\n${BLUE}📊 Statut des services...${NC}"
        ansible webservers -i $INVENTORY -a "pm2 status" -u deploy || echo "PM2 n'est peut-être pas encore installé"
        ;;
    *)
        echo "Action inconnue: $ACTION"
        echo "Usage: $0 [provision|deploy|update|rollback|check|status]"
        echo ""
        echo "Actions disponibles:"
        echo "  provision - Installation complète du serveur (première fois)"
        echo "  deploy    - Déploiement de l'application"
        echo "  update    - Mise à jour rapide (pull + restart)"
        echo "  rollback  - Retour à la version précédente"
        echo "  check     - Vérification sans exécution (dry-run)"
        echo "  status    - Afficher le statut de PM2"
        exit 1
        ;;
esac

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Opération terminée avec succès !${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Accès à votre application :"
echo "  → http://72.61.146.126"
echo ""
echo "Commandes utiles :"
echo "  → Voir les logs : ssh deploy@72.61.146.126 'pm2 logs'"
echo "  → Statut PM2   : ssh deploy@72.61.146.126 'pm2 status'"
echo "  → Redémarrer   : ssh deploy@72.61.146.126 'pm2 restart all'"
