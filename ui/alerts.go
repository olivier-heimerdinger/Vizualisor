package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"

	"github.com/olivier-heimerdinger/Vizualisor/services"
)

// AlertsManager surveille l'état des services favoris et envoie des alertes.
type AlertsManager struct {
	app    *App
	stopCh chan struct{}
}

// NewAlertsManager crée un nouveau gestionnaire d'alertes.
func NewAlertsManager(app *App) *AlertsManager {
	return &AlertsManager{
		app: app,
	}
}

// Start démarre le polling des alertes.
func (am *AlertsManager) Start() {
	cfg := am.app.configMgr.Get()
	if !cfg.Alerts.Enabled {
		return
	}

	interval := time.Duration(cfg.Alerts.PollInterval) * time.Second
	am.stopCh = make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-am.stopCh:
				return
			case <-ticker.C:
				am.checkAlerts()
			}
		}
	}()
}

// Stop arrête le polling.
func (am *AlertsManager) Stop() {
	if am.stopCh != nil {
		close(am.stopCh)
	}
}

func (am *AlertsManager) checkAlerts() {
	cfg := am.app.configMgr.Get()
	if !cfg.Alerts.WatchFavorites {
		return
	}

	for _, srv := range cfg.Servers {
		favs := am.app.configMgr.GetFavorites(srv.Name)
		if len(favs) == 0 {
			continue
		}

		client, ok := am.app.sshPool.Get(srv.Name)
		if !ok {
			// Serveur non connecté - alerte
			am.sendNotification(
				fmt.Sprintf("Serveur déconnecté : %s", srv.Name),
				"Impossible de se connecter au serveur",
			)
			continue
		}

		mgr := services.NewManager(client, srv.Name)
		svcs, err := mgr.ListServices()
		if err != nil {
			continue
		}

		// Mettre à jour le cache
		am.app.serviceCache[srv.Name] = svcs

		// Vérifier les favoris
		svcMap := make(map[string]services.Service)
		for _, s := range svcs {
			svcMap[s.Name] = s
		}

		for _, favName := range favs {
			if svc, ok := svcMap[favName]; ok {
				if svc.IsFailed() {
					am.sendNotification(
						fmt.Sprintf("⚠ Service en erreur : %s", favName),
						fmt.Sprintf("Le service %s sur %s est en état 'failed'", favName, srv.Name),
					)
				} else if !svc.IsRunning() {
					am.sendNotification(
						fmt.Sprintf("Service arrêté : %s", favName),
						fmt.Sprintf("Le service %s sur %s est inactif", favName, srv.Name),
					)
				}
			}
		}
	}

	// Rafraîchir le TreeView
	if am.app.treeView != nil {
		am.app.treeView.Refresh()
	}
}

func (am *AlertsManager) sendNotification(title, content string) {
	notification := fyne.NewNotification(title, content)
	am.app.fyneApp.SendNotification(notification)
}
