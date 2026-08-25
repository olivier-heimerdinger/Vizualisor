// Vizualisor - Application de monitoring de serveurs Linux via SSH.
//
// Point d'entrée principal de l'application.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/olivier-heimerdinger/Vizualisor/config"
	"github.com/olivier-heimerdinger/Vizualisor/ssh"
	"github.com/olivier-heimerdinger/Vizualisor/ui"
)

const defaultConfigPath = "config.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "Chemin vers le fichier de configuration YAML")
	flag.Parse()

	// Charger la configuration
	cfgMgr := config.NewManager(*configPath)
	if err := cfgMgr.Load(); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Fichier de configuration '%s' non trouvé.\n", *configPath)
			fmt.Println("Créez un fichier config.yaml basé sur l'exemple fourni.")
			os.Exit(1)
		}
		log.Fatalf("Erreur chargement configuration: %v", err)
	}

	cfg := cfgMgr.Get()
	fmt.Printf("Vizualisor - %s\n", cfg.App.Name)
	fmt.Printf("Serveurs configurés: %d\n", len(cfg.Servers))

	// Initialiser le résolveur de credentials
	credRes := config.NewCredentialResolver()

	// Charger le fichier .env si configuré
	if cfg.Credentials.EnvFile != "" {
		if err := credRes.LoadEnvFile(cfg.Credentials.EnvFile); err != nil {
			log.Printf("Avertissement: %v", err)
		}
	}

	// Créer le pool SSH
	sshPool := ssh.NewPool(credRes)
	defer sshPool.DisconnectAll()

	// Lancer l'interface graphique
	application := ui.NewApp(cfgMgr, sshPool, credRes)
	application.Run()
}
