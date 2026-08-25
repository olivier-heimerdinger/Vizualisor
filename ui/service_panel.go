package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/olivier-heimerdinger/Vizualisor/services"
)

// ServicePanel affiche les détails d'un service sélectionné.
type ServicePanel struct {
	app        *App
	container  *fyne.Container
	nameLabel  *widget.Label
	statusIcon *canvas.Circle
	statusText *widget.Label
	descLabel  *widget.Label
	detailText *widget.Label
	startBtn   *widget.Button
	stopBtn    *widget.Button
	restartBtn *widget.Button
	logsBtn    *widget.Button
	favBtn     *widget.Button

	currentServer  string
	currentService string
}

// NewServicePanel crée un nouveau panel de service.
func NewServicePanel(app *App) *ServicePanel {
	sp := &ServicePanel{app: app}
	sp.build()
	return sp
}

// Container retourne le conteneur du panel.
func (sp *ServicePanel) Container() fyne.CanvasObject {
	return sp.container
}

func (sp *ServicePanel) build() {
	// En-tête
	sp.nameLabel = widget.NewLabel("Sélectionnez un service")
	sp.nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	sp.statusIcon = canvas.NewCircle(color.NRGBA{R: 158, G: 158, B: 158, A: 255})
	sp.statusIcon.Resize(fyne.NewSize(16, 16))

	sp.statusText = widget.NewLabel("")
	sp.descLabel = widget.NewLabel("")
	sp.descLabel.Wrapping = fyne.TextWrapWord

	header := container.NewHBox(sp.statusIcon, sp.nameLabel, sp.statusText)

	// Boutons d'actions
	sp.startBtn = widget.NewButton("▶ Démarrer", sp.onStart)
	sp.startBtn.Importance = widget.SuccessImportance
	sp.startBtn.Disable()

	sp.stopBtn = widget.NewButton("■ Arrêter", sp.onStop)
	sp.stopBtn.Importance = widget.DangerImportance
	sp.stopBtn.Disable()

	sp.restartBtn = widget.NewButton("↻ Redémarrer", sp.onRestart)
	sp.restartBtn.Importance = widget.WarningImportance
	sp.restartBtn.Disable()

	sp.logsBtn = widget.NewButton("📋 Voir les logs", sp.onShowLogs)
	sp.logsBtn.Disable()

	sp.favBtn = widget.NewButton("☆ Favori", sp.onToggleFavorite)
	sp.favBtn.Disable()

	actionRow := container.NewHBox(sp.startBtn, sp.stopBtn, sp.restartBtn)
	toolRow := container.NewHBox(sp.logsBtn, sp.favBtn)

	// Détails
	sp.detailText = widget.NewLabel("Aucun service sélectionné.\n\nCliquez sur un service dans l'arborescence pour voir ses détails.")
	sp.detailText.Wrapping = fyne.TextWrapWord

	detailScroll := container.NewVScroll(sp.detailText)
	detailScroll.SetMinSize(fyne.NewSize(0, 300))

	sp.container = container.NewVBox(
		header,
		sp.descLabel,
		widget.NewSeparator(),
		actionRow,
		toolRow,
		widget.NewSeparator(),
		widget.NewLabel("Détails du service :"),
		detailScroll,
		layout.NewSpacer(),
	)
}

// ShowService affiche les détails d'un service.
func (sp *ServicePanel) ShowService(serverName, serviceName string) {
	sp.currentServer = serverName
	sp.currentService = serviceName

	sp.nameLabel.SetText(fmt.Sprintf("🔧 %s", serviceName))

	// Trouver le service dans le cache
	var svc *services.Service
	for _, s := range sp.app.getServicesForServer(serverName) {
		if s.Name == serviceName {
			svc = &s
			break
		}
	}

	if svc == nil {
		sp.statusText.SetText("Inconnu")
		sp.descLabel.SetText("")
		sp.detailText.SetText("Service non trouvé dans le cache.")
		return
	}

	// Statut
	sp.statusText.SetText(svc.StatusLabel())
	switch svc.Status {
	case services.StatusActive:
		sp.statusIcon.FillColor = DirectColorActive
	case services.StatusFailed:
		sp.statusIcon.FillColor = DirectColorFailed
	default:
		sp.statusIcon.FillColor = DirectColorInactive
	}
	sp.statusIcon.Refresh()

	sp.descLabel.SetText(svc.Description)

	// Activer les boutons
	sp.startBtn.Enable()
	sp.stopBtn.Enable()
	sp.restartBtn.Enable()
	sp.logsBtn.Enable()
	sp.favBtn.Enable()

	// Mettre à jour le bouton favori
	if sp.app.configMgr.IsFavorite(serverName, serviceName) {
		sp.favBtn.SetText("★ Retirer des favoris")
	} else {
		sp.favBtn.SetText("☆ Ajouter aux favoris")
	}

	// Charger les détails en arrière-plan
	go sp.loadServiceDetails(serverName, serviceName)
}

func (sp *ServicePanel) loadServiceDetails(serverName, serviceName string) {
	client, ok := sp.app.sshPool.Get(serverName)
	if !ok {
		sp.detailText.SetText("Erreur : non connecté au serveur " + serverName)
		return
	}

	mgr := services.NewManager(client, serverName)
	details, err := mgr.GetServiceStatus(serviceName)
	if err != nil {
		sp.detailText.SetText(fmt.Sprintf("Erreur : %v", err))
		return
	}

	sp.detailText.SetText(details)
}

func (sp *ServicePanel) onStart() {
	sp.executeServiceAction("start", func(mgr *services.Manager, pwd string) error {
		return mgr.StartService(sp.currentService, pwd)
	})
}

func (sp *ServicePanel) onStop() {
	sp.executeServiceAction("stop", func(mgr *services.Manager, pwd string) error {
		return mgr.StopService(sp.currentService, pwd)
	})
}

func (sp *ServicePanel) onRestart() {
	sp.executeServiceAction("restart", func(mgr *services.Manager, pwd string) error {
		return mgr.RestartService(sp.currentService, pwd)
	})
}

func (sp *ServicePanel) executeServiceAction(action string, fn func(*services.Manager, string) error) {
	if sp.currentServer == "" || sp.currentService == "" {
		return
	}

	// Demander le mot de passe sudo
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Mot de passe sudo")

	dialog.ShowForm(
		fmt.Sprintf("Mot de passe sudo pour %s %s", action, sp.currentService),
		"Exécuter",
		"Annuler",
		[]*widget.FormItem{
			widget.NewFormItem("Mot de passe", passwordEntry),
		},
		func(ok bool) {
			if !ok {
				return
			}
			go func() {
				client, ok := sp.app.sshPool.Get(sp.currentServer)
				if !ok {
					dialog.ShowError(fmt.Errorf("non connecté à %s", sp.currentServer), sp.app.mainWindow)
					return
				}

				mgr := services.NewManager(client, sp.currentServer)
				if err := fn(mgr, passwordEntry.Text); err != nil {
					dialog.ShowError(err, sp.app.mainWindow)
					return
				}

				dialog.ShowInformation("Succès",
					fmt.Sprintf("Service %s : %s effectué", sp.currentService, action),
					sp.app.mainWindow)

				// Rafraîchir
				sp.ShowService(sp.currentServer, sp.currentService)
				sp.app.treeView.Refresh()
			}()
		},
		sp.app.mainWindow,
	)
}

func (sp *ServicePanel) onShowLogs() {
	if sp.currentServer == "" || sp.currentService == "" {
		return
	}
	OpenLogWindow(sp.app, sp.currentServer, sp.currentService, "")
}

func (sp *ServicePanel) onToggleFavorite() {
	if sp.currentServer == "" || sp.currentService == "" {
		return
	}

	isFav := sp.app.configMgr.ToggleFavorite(sp.currentServer, sp.currentService)
	if isFav {
		sp.favBtn.SetText("★ Retirer des favoris")
	} else {
		sp.favBtn.SetText("☆ Ajouter aux favoris")
	}

	// Sauvegarder la config
	go sp.app.configMgr.Save()

	// Rafraîchir le treeview
	sp.app.treeView.Refresh()
}
