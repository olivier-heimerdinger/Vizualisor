package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tobischo/gokeepasslib/v3"
)

// CredentialResolver résout les identifiants selon l'ordre de priorité :
// 1. Fichier .env
// 2. KeePass
// 3. Configuration YAML directe
type CredentialResolver struct {
	envVars        map[string]string
	keepassDB      *gokeepasslib.Database
	keepassEntries map[string]keepassEntry
}

type keepassEntry struct {
	Username string
	Password string
}

// NewCredentialResolver crée un nouveau résolveur de credentials.
func NewCredentialResolver() *CredentialResolver {
	return &CredentialResolver{
		envVars:        make(map[string]string),
		keepassEntries: make(map[string]keepassEntry),
	}
}

// LoadEnvFile charge les variables depuis un fichier .env.
func (cr *CredentialResolver) LoadEnvFile(path string) error {
	if path == "" {
		return nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Fichier optionnel
	}

	envMap, err := godotenv.Read(path)
	if err != nil {
		return fmt.Errorf("erreur lecture .env: %w", err)
	}
	cr.envVars = envMap
	return nil
}

// LoadKeePass charge les entrées depuis une base KeePass.
func (cr *CredentialResolver) LoadKeePass(path, masterPassword string) error {
	if path == "" {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("erreur ouverture KeePass: %w", err)
	}
	defer file.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(masterPassword)

	if err := gokeepasslib.NewDecoder(file).Decode(db); err != nil {
		return fmt.Errorf("erreur décodage KeePass: %w", err)
	}

	db.UnlockProtectedEntries()
	cr.keepassDB = db

	// Indexer les entrées par titre
	for _, group := range db.Content.Root.Groups {
		cr.indexKeePassGroup(group)
	}

	return nil
}

func (cr *CredentialResolver) indexKeePassGroup(group gokeepasslib.Group) {
	for _, entry := range group.Entries {
		title := entry.GetTitle()
		if title != "" {
			cr.keepassEntries[strings.ToLower(title)] = keepassEntry{
				Username: entry.GetContent("UserName"),
				Password: entry.GetPassword(),
			}
		}
	}
	for _, subGroup := range group.Groups {
		cr.indexKeePassGroup(subGroup)
	}
}

// ResolvePassword résout le mot de passe pour un serveur.
// Ordre : .env (PASS_<serverKey>) → KeePass (titre = nom serveur) → config directe
func (cr *CredentialResolver) ResolvePassword(server ServerConfig) string {
	// 1. Fichier .env
	envKey := "PASS_" + sanitizeEnvKey(server.Name)
	if pass, ok := cr.envVars[envKey]; ok && pass != "" {
		return pass
	}

	// 2. KeePass
	if entry, ok := cr.keepassEntries[strings.ToLower(server.Name)]; ok && entry.Password != "" {
		return entry.Password
	}

	// 3. Config directe
	return server.Password
}

// ResolveUsername résout le nom d'utilisateur pour un serveur.
func (cr *CredentialResolver) ResolveUsername(server ServerConfig) string {
	// 1. Fichier .env
	envKey := "USER_" + sanitizeEnvKey(server.Name)
	if user, ok := cr.envVars[envKey]; ok && user != "" {
		return user
	}

	// 2. KeePass
	if entry, ok := cr.keepassEntries[strings.ToLower(server.Name)]; ok && entry.Username != "" {
		return entry.Username
	}

	// 3. Config directe
	return server.Username
}

// sanitizeEnvKey transforme un nom de serveur en clé d'env valide.
func sanitizeEnvKey(name string) string {
	r := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return strings.ToUpper(r.Replace(name))
}
