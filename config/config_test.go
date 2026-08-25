package config

import (
	"os"
	"path/filepath"
	"testing"
)

const testConfigYAML = `
app:
  name: "Test Vizualisor"
  refresh_interval: 15
  theme: "dark"

credentials:
  env_file: ""

servers:
  - name: "Test Server"
    host: "192.168.1.100"
    port: 2222
    username: "testuser"
    password: "testpass"
    auth_method: "password"
    custom_logs:
      - name: "App Log"
        path: "/var/log/app.log"
    sudo_commands:
      - name: "Restart App"
        command: "systemctl restart app"

  - name: "Server 2"
    host: "10.0.0.1"
    port: 22
    username: "admin"
    auth_method: "key"
    key_path: "~/.ssh/id_rsa"

favorites:
  "Test Server":
    - "nginx"
    - "myapp"

alerts:
  enabled: true
  poll_interval: 30
  watch_favorites: true
`

func createTempConfig(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(testConfigYAML), 0644); err != nil {
		t.Fatalf("erreur création config test: %v", err)
	}
	return path, func() { os.RemoveAll(dir) }
}

func TestLoadConfig(t *testing.T) {
	path, cleanup := createTempConfig(t)
	defer cleanup()

	mgr := NewManager(path)
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() erreur: %v", err)
	}

	cfg := mgr.Get()

	if cfg.App.Name != "Test Vizualisor" {
		t.Errorf("App.Name = %q, want %q", cfg.App.Name, "Test Vizualisor")
	}
	if cfg.App.RefreshInterval != 15 {
		t.Errorf("App.RefreshInterval = %d, want 15", cfg.App.RefreshInterval)
	}
	if cfg.App.Theme != "dark" {
		t.Errorf("App.Theme = %q, want %q", cfg.App.Theme, "dark")
	}
}

func TestLoadConfigServers(t *testing.T) {
	path, cleanup := createTempConfig(t)
	defer cleanup()

	mgr := NewManager(path)
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() erreur: %v", err)
	}

	servers := mgr.GetServers()
	if len(servers) != 2 {
		t.Fatalf("len(servers) = %d, want 2", len(servers))
	}

	srv := servers[0]
	if srv.Name != "Test Server" {
		t.Errorf("srv.Name = %q, want %q", srv.Name, "Test Server")
	}
	if srv.Host != "192.168.1.100" {
		t.Errorf("srv.Host = %q, want %q", srv.Host, "192.168.1.100")
	}
	if srv.Port != 2222 {
		t.Errorf("srv.Port = %d, want 2222", srv.Port)
	}
	if len(srv.CustomLogs) != 1 {
		t.Errorf("len(srv.CustomLogs) = %d, want 1", len(srv.CustomLogs))
	}
	if len(srv.SudoCommands) != 1 {
		t.Errorf("len(srv.SudoCommands) = %d, want 1", len(srv.SudoCommands))
	}
}

func TestDefaultValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := `
app:
  name: "Test"
servers:
  - name: "DefaultServer"
    host: "1.2.3.4"
`
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	mgr := NewManager(path)
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() erreur: %v", err)
	}

	cfg := mgr.Get()
	if cfg.App.RefreshInterval != 30 {
		t.Errorf("Default RefreshInterval = %d, want 30", cfg.App.RefreshInterval)
	}
	if cfg.Alerts.PollInterval != 60 {
		t.Errorf("Default PollInterval = %d, want 60", cfg.Alerts.PollInterval)
	}

	servers := mgr.GetServers()
	if servers[0].Port != 22 {
		t.Errorf("Default Port = %d, want 22", servers[0].Port)
	}
	if servers[0].AuthMethod != "password" {
		t.Errorf("Default AuthMethod = %q, want 'password'", servers[0].AuthMethod)
	}
}

func TestFavorites(t *testing.T) {
	path, cleanup := createTempConfig(t)
	defer cleanup()

	mgr := NewManager(path)
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() erreur: %v", err)
	}

	// Vérifier les favoris chargés
	favs := mgr.GetFavorites("Test Server")
	if len(favs) != 2 {
		t.Fatalf("len(favs) = %d, want 2", len(favs))
	}
	if favs[0] != "nginx" {
		t.Errorf("favs[0] = %q, want %q", favs[0], "nginx")
	}

	// Toggle favori
	if !mgr.IsFavorite("Test Server", "nginx") {
		t.Error("nginx devrait être favori")
	}

	// Retirer
	result := mgr.ToggleFavorite("Test Server", "nginx")
	if result != false {
		t.Error("ToggleFavorite devrait retourner false (retiré)")
	}
	if mgr.IsFavorite("Test Server", "nginx") {
		t.Error("nginx ne devrait plus être favori")
	}

	// Ajouter
	result = mgr.ToggleFavorite("Test Server", "redis")
	if result != true {
		t.Error("ToggleFavorite devrait retourner true (ajouté)")
	}
	if !mgr.IsFavorite("Test Server", "redis") {
		t.Error("redis devrait être favori")
	}
}

func TestSaveConfig(t *testing.T) {
	path, cleanup := createTempConfig(t)
	defer cleanup()

	mgr := NewManager(path)
	if err := mgr.Load(); err != nil {
		t.Fatalf("Load() erreur: %v", err)
	}

	// Modifier et sauvegarder
	mgr.ToggleFavorite("Test Server", "newservice")
	if err := mgr.Save(); err != nil {
		t.Fatalf("Save() erreur: %v", err)
	}

	// Recharger et vérifier
	mgr2 := NewManager(path)
	if err := mgr2.Load(); err != nil {
		t.Fatalf("Load() après Save erreur: %v", err)
	}
	if !mgr2.IsFavorite("Test Server", "newservice") {
		t.Error("Le favori ajouté n'a pas été persisté")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	mgr := NewManager("/path/inexistant/config.yaml")
	err := mgr.Load()
	if err == nil {
		t.Error("Load() devrait retourner une erreur pour un fichier inexistant")
	}
}

func TestGetEmptyConfig(t *testing.T) {
	mgr := NewManager("")
	cfg := mgr.Get()
	if cfg.App.Name != "" {
		t.Errorf("Empty config App.Name = %q, want empty", cfg.App.Name)
	}
	servers := mgr.GetServers()
	if servers != nil {
		t.Errorf("Empty config servers = %v, want nil", servers)
	}
}
