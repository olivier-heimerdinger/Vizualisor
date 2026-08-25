# Vizualisor

## Objectif

J'ai une série de serveur Linux différents et je veux pouvoir visualiser les informations suivantes :

- Les services en cours d'exécution et leurs états.
- Visualiser les logs des services.
- Visualiser d'autres logs spécifiques à chaque serveurs
- Pouvoir faire des recherches de services et de logs.
- Avoir des favoris pour les services associés à chaque service afin d'avoir un état des lieux voulus.
- Avoir des alertes en cas de problème.
- Un affichage en forme de treeview pour les serveurs et les services.
- Je dois pouvoir avoir la possibilité d'arrêter, de démarrer et de redémarrer les services.
- La configuration doit être stockée dans un fichier JSON et doit être facilement modifiable.
- L'application doit être légère et rapide.
- L'application doit être open source.
- L'application doit pouvoir jouer des sudos avec des opérations spécifiques propre à un serveur
- Tout doit être paramétrable et stocké dans un fichier yaml
- Les logs doivent être ouverts dans une fenêtre nouvelle, pouvoir être lues comme un tail -f et filtrables
- L'application doit chercher les infos des serveurs via ssh
- L'application peut chercher les mots de passes et nom d'utilisateurs dans un fichier d'environnement ou un fichier Keepass en priorité
- Pour l'affichage des services, il faut utiliser des icônes et des couleurs pour différencier les services et leurs états.
- Utilise le repo git initialisé pour toi


## Architecture

- L'application doit être composée de plusieurs modules.
- Les modules doivent être indépendants les uns des autres.
- Les modules doivent être facilement remplaçables.

## Technologies

- L'application doit être développée en Go.
- L'application doit utiliser Fyne pour l'interface graphique.

## Etapes

1. Création de l'architecture de l'application.
2. Création de l'interface graphique.
3. Création des modules.
4. Création des tests.
5. Création de la documentation.