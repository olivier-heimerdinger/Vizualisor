// Package config gère le chargement et la sauvegarde de la configuration YAML.
package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// AppConfig est la configuration racine de l'application.
type AppConfig struct {
	App         AppSettings          `yaml:"app"`
	Credentials CredentialsConfig    `yaml:"credentials"`
	Servers     []ServerConfig       `yaml:"servers"`
	Favorites   map[string][]string  `yaml:"favorites"`
	Alerts      AlertsConfig         `yaml:"alerts"`
}

// AppSettings contient les paramètres globaux de l'application.
type AppSettings struct {
	Name            string `yaml:"name"`
	RefreshInterval int    `yaml:"refresh_interval"`
	Theme           string `yaml:"theme"`
}

// CredentialsConfig définit comment résoudre les identifiants.
type CredentialsConfig struct {
	EnvFile  string         `yaml:"env_file"`
	KeePass  KeePassConfig  `yaml:"keepass"`
}

// KeePassConfig configure l'accès à une base KeePass.
type KeePassConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// ServerConfig définit un serveur à surveiller.
type ServerConfig struct {
	Name         string          `yaml:"name"`
	Host         string          `yaml:"host"`
	Port         int             `yaml:"port"`
	Username     string          `yaml:"username"`
	Password     string          `yaml:"password,omitempty"`
	AuthMethod   string          `yaml:"auth_method"`
	KeyPath      string          `yaml:"key_path,omitempty"`
	CustomLogs   []CustomLog     `yaml:"custom_logs,omitempty"`
	SudoCommands []SudoCommand   `yaml:"sudo_commands,omitempty"`
}

// CustomLog définit un fichier de log personnalisé.
type CustomLog struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// SudoCommand définit une commande sudo personnalisée.
type SudoCommand struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// AlertsConfig configure le système d'alertes.
type AlertsConfig struct {
	Enabled        bool `yaml:"enabled"`
	PollInterval   int  `yaml:"poll_interval"`
	WatchFavorites bool `yaml:"watch_favorites"`
}

// Manager gère la configuration avec un accès thread-safe.
type Manager struct {
	mu       sync.RWMutex
	config   *AppConfig
	filePath string
}

// NewManager crée un nouveau gestionnaire de configuration.
func NewManager(filePath string) *Manager {
	return &Manager{
		filePath: filePath,
	}
}

// Load charge la configuration depuis le fichier YAML.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	cfg := &AppConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}

	// Valeurs par défaut
	if cfg.App.RefreshInterval <= 0 {
		cfg.App.RefreshInterval = 30
	}
	if cfg.Alerts.PollInterval <= 0 {
		cfg.Alerts.PollInterval = 60
	}
	for i := range cfg.Servers {
		if cfg.Servers[i].Port <= 0 {
			cfg.Servers[i].Port = 22
		}
		if cfg.Servers[i].AuthMethod == "" {
			cfg.Servers[i].AuthMethod = "password"
		}
	}
	if cfg.Favorites == nil {
		cfg.Favorites = make(map[string][]string)
	}

	m.config = cfg
	return nil
}

// Save sauvegarde la configuration dans le fichier YAML.
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := yaml.Marshal(m.config)
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// Get retourne une copie de la configuration courante.
func (m *Manager) Get() AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return AppConfig{}
	}
	return *m.config
}

// GetServers retourne la liste des serveurs configurés.
func (m *Manager) GetServers() []ServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return nil
	}
	result := make([]ServerConfig, len(m.config.Servers))
	copy(result, m.config.Servers)
	return result
}

// GetFavorites retourne les favoris pour un serveur donné.
func (m *Manager) GetFavorites(serverName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.config == nil {
		return nil
	}
	favs, ok := m.config.Favorites[serverName]
	if !ok {
		return nil
	}
	result := make([]string, len(favs))
	copy(result, favs)
	return result
}

// ToggleFavorite ajoute ou retire un service des favoris.
func (m *Manager) ToggleFavorite(serverName, serviceName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Favorites == nil {
		m.config.Favorites = make(map[string][]string)
	}

	favs := m.config.Favorites[serverName]
	for i, f := range favs {
		if f == serviceName {
			// Retirer
			m.config.Favorites[serverName] = append(favs[:i], favs[i+1:]...)
			return false
		}
	}
	// Ajouter
	m.config.Favorites[serverName] = append(favs, serviceName)
	return true
}

// IsFavorite vérifie si un service est favori.
func (m *Manager) IsFavorite(serverName, serviceName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, f := range m.config.Favorites[serverName] {
		if f == serviceName {
			return true
		}
	}
	return false
}
