// Package ui contient l'interface graphique Fyne de Vizualisor.
package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/olivier-heimerdinger/Vizualisor/config"
	"github.com/olivier-heimerdinger/Vizualisor/services"
	"github.com/olivier-heimerdinger/Vizualisor/ssh"
)

// App est l'application principale Vizualisor.
type App struct {
	fyneApp    fyne.App
	mainWindow fyne.Window
	configMgr  *config.Manager
	sshPool    *ssh.Pool
	credRes    *config.CredentialResolver

	// Composants UI
	treeView     *TreeView
	servicePanel *ServicePanel
	searchBar    *SearchBar
	alertsMgr    *AlertsManager

	// État
	currentServer  string
	currentService string
	serviceCache   map[string][]services.Service // serveur -> services
}

// NewApp crée une nouvelle application Vizualisor.
func NewApp(configMgr *config.Manager, sshPool *ssh.Pool, credRes *config.CredentialResolver) *App {
	fyneApp := app.NewWithID("com.vizualisor.app")

	cfg := configMgr.Get()
	if cfg.App.Theme == "dark" {
		fyneApp.Settings().SetTheme(theme.DarkTheme())
	} else {
		fyneApp.Settings().SetTheme(theme.LightTheme())
	}

	return &App{
		fyneApp:      fyneApp,
		configMgr:    configMgr,
		sshPool:      sshPool,
		credRes:      credRes,
		serviceCache: make(map[string][]services.Service),
	}
}

// Run démarre l'application.
func (a *App) Run() {
	a.mainWindow = a.fyneApp.NewWindow("Vizualisor - Monitoring Serveurs")
	a.mainWindow.Resize(fyne.NewSize(1200, 800))

	// Composants
	a.searchBar = NewSearchBar(a)
	a.treeView = NewTreeView(a)
	a.servicePanel = NewServicePanel(a)
	a.alertsMgr = NewAlertsManager(a)

	// Layout : Barre de recherche en haut, TreeView à gauche, Panel à droite
	leftPanel := container.NewBorder(
		a.searchBar.Container(), // Haut
		nil,                     // Bas
		nil,                     // Gauche
		nil,                     // Droite
		a.treeView.Container(),  // Centre
	)

	split := container.NewHSplit(leftPanel, a.servicePanel.Container())
	split.SetOffset(0.3) // 30% pour le tree, 70% pour le panel

	// Toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), a.refreshAll),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.SettingsIcon(), a.showSettings),
	)

	mainContent := container.NewBorder(
		toolbar, // Haut
		nil,     // Bas
		nil,     // Gauche
		nil,     // Droite
		split,   // Centre
	)

	a.mainWindow.SetContent(mainContent)
	a.mainWindow.SetMaster()

	// Démarrer le polling des alertes
	cfg := a.configMgr.Get()
	if cfg.Alerts.Enabled {
		a.alertsMgr.Start()
	}

	// Connecter les serveurs au démarrage
	go a.connectAllServers()

	a.mainWindow.ShowAndRun()
}

// connectAllServers se connecte à tous les serveurs configurés.
func (a *App) connectAllServers() {
	servers := a.configMgr.GetServers()
	for _, srv := range servers {
		_, err := a.sshPool.Connect(srv)
		if err != nil {
			fyne.LogError("Connexion SSH échouée pour "+srv.Name, err)
		}
	}
	// Rafraîchir le treeview après connexion
	a.treeView.Refresh()
}

// refreshAll rafraîchit toutes les données.
func (a *App) refreshAll() {
	a.serviceCache = make(map[string][]services.Service)
	go func() {
		a.connectAllServers()
		a.loadAllServices()
		a.treeView.Refresh()
	}()
}

// loadAllServices charge les services de tous les serveurs connectés.
func (a *App) loadAllServices() {
	servers := a.configMgr.GetServers()
	for _, srv := range servers {
		client, ok := a.sshPool.Get(srv.Name)
		if !ok {
			continue
		}
		mgr := services.NewManager(client, srv.Name)
		svcs, err := mgr.ListServices()
		if err != nil {
			fyne.LogError("Erreur listing services "+srv.Name, err)
			continue
		}
		a.serviceCache[srv.Name] = svcs
	}
}

// getServicesForServer retourne les services en cache pour un serveur.
func (a *App) getServicesForServer(serverName string) []services.Service {
	return a.serviceCache[serverName]
}

// showSettings affiche les paramètres (placeholder pour l'instant).
func (a *App) showSettings() {
	settingsWin := a.fyneApp.NewWindow("Paramètres")
	settingsWin.Resize(fyne.NewSize(400, 300))

	cfgPath := widget.NewLabel("Config: config.yaml")
	settingsWin.SetContent(container.NewVBox(
		widget.NewLabel("Paramètres Vizualisor"),
		widget.NewSeparator(),
		cfgPath,
	))
	settingsWin.Show()
}
