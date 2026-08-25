package ui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/olivier-heimerdinger/Vizualisor/services"
)

// TreeView affiche la hiérarchie serveurs/services.
type TreeView struct {
	app    *App
	tree   *widget.Tree
	filter string
}

// NewTreeView crée un nouveau TreeView.
func NewTreeView(app *App) *TreeView {
	tv := &TreeView{app: app}
	tv.tree = tv.buildTree()
	return tv
}

// Container retourne le conteneur du TreeView.
func (tv *TreeView) Container() fyne.CanvasObject {
	return tv.tree
}

// SetFilter applique un filtre de recherche.
func (tv *TreeView) SetFilter(filter string) {
	tv.filter = strings.ToLower(filter)
	tv.tree.Refresh()
}

// Refresh rafraîchit le TreeView.
func (tv *TreeView) Refresh() {
	tv.tree.Refresh()
}

func (tv *TreeView) buildTree() *widget.Tree {
	return &widget.Tree{
		ChildUIDs: func(uid widget.TreeNodeID) []widget.TreeNodeID {
			if uid == "" {
				// Racine : liste des serveurs
				var serverIDs []widget.TreeNodeID
				for _, srv := range tv.app.configMgr.GetServers() {
					serverIDs = append(serverIDs, "server:"+srv.Name)
				}
				return serverIDs
			}

			if strings.HasPrefix(uid, "server:") {
				// Serveur : liste des services
				serverName := strings.TrimPrefix(uid, "server:")
				svcs := tv.app.getServicesForServer(serverName)
				var serviceIDs []widget.TreeNodeID

				for _, svc := range svcs {
					if tv.filter != "" &&
						!strings.Contains(strings.ToLower(svc.Name), tv.filter) &&
						!strings.Contains(strings.ToLower(svc.Description), tv.filter) {
						continue
					}
					serviceIDs = append(serviceIDs, fmt.Sprintf("service:%s:%s", serverName, svc.Name))
				}

				// Ajouter les logs custom
				for _, srv := range tv.app.configMgr.GetServers() {
					if srv.Name == serverName {
						for _, log := range srv.CustomLogs {
							if tv.filter != "" && !strings.Contains(strings.ToLower(log.Name), tv.filter) {
								continue
							}
							serviceIDs = append(serviceIDs, fmt.Sprintf("log:%s:%s", serverName, log.Path))
						}
						break
					}
				}

				return serviceIDs
			}

			return nil
		},

		IsBranch: func(uid widget.TreeNodeID) bool {
			return uid == "" || strings.HasPrefix(uid, "server:")
		},

		CreateNode: func(branch bool) fyne.CanvasObject {
			icon := canvas.NewCircle(DirectColorInactive)
			icon.Resize(fyne.NewSize(12, 12))
			label := widget.NewLabel("Template")
			favIcon := widget.NewLabel("")

			return container.NewHBox(icon, label, favIcon)
		},

		UpdateNode: func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			box := node.(*fyne.Container)
			icon := box.Objects[0].(*canvas.Circle)
			label := box.Objects[1].(*widget.Label)
			favLabel := box.Objects[2].(*widget.Label)

			if strings.HasPrefix(uid, "server:") {
				serverName := strings.TrimPrefix(uid, "server:")
				label.SetText("🖥 " + serverName)
				// Couleur : vert si connecté, gris sinon
				_, connected := tv.app.sshPool.Get(serverName)
				if connected {
					icon.FillColor = ColorActive
				} else {
					icon.FillColor = ColorInactive
				}
				favLabel.SetText("")
				icon.Refresh()

			} else if strings.HasPrefix(uid, "service:") {
				parts := strings.SplitN(strings.TrimPrefix(uid, "service:"), ":", 2)
				if len(parts) == 2 {
					serverName, svcName := parts[0], parts[1]

					// Trouver le service dans le cache
					for _, svc := range tv.app.getServicesForServer(serverName) {
						if svc.Name == svcName {
							label.SetText(svc.Name)
							icon.FillColor = statusColor(svc.Status)
							icon.Refresh()
							break
						}
					}

					// Icône favori
					if tv.app.configMgr.IsFavorite(serverName, svcName) {
						favLabel.SetText("⭐")
					} else {
						favLabel.SetText("")
					}
				}

			} else if strings.HasPrefix(uid, "log:") {
				parts := strings.SplitN(strings.TrimPrefix(uid, "log:"), ":", 2)
				if len(parts) == 2 {
					logPath := parts[1]
					// Trouver le nom du log
					logName := logPath
					for _, srv := range tv.app.configMgr.GetServers() {
						for _, cl := range srv.CustomLogs {
							if cl.Path == logPath {
								logName = cl.Name
								break
							}
						}
					}
					label.SetText("📄 " + logName)
					icon.FillColor = ColorLog
					icon.Refresh()
					favLabel.SetText("")
				}
			}
		},

		OnSelected: func(uid widget.TreeNodeID) {
			if strings.HasPrefix(uid, "service:") {
				parts := strings.SplitN(strings.TrimPrefix(uid, "service:"), ":", 2)
				if len(parts) == 2 {
					tv.app.currentServer = parts[0]
					tv.app.currentService = parts[1]
					tv.app.servicePanel.ShowService(parts[0], parts[1])
				}
			} else if strings.HasPrefix(uid, "log:") {
				parts := strings.SplitN(strings.TrimPrefix(uid, "log:"), ":", 2)
				if len(parts) == 2 {
					serverName, logPath := parts[0], parts[1]
					OpenLogWindow(tv.app, serverName, "", logPath)
				}
			} else if strings.HasPrefix(uid, "server:") {
				serverName := strings.TrimPrefix(uid, "server:")
				// Charger les services si pas en cache
				if _, ok := tv.app.serviceCache[serverName]; !ok {
					go func() {
						client, ok := tv.app.sshPool.Get(serverName)
						if !ok {
							return
						}
						mgr := services.NewManager(client, serverName)
						svcs, err := mgr.ListServices()
						if err != nil {
							fyne.LogError("Erreur listing services", err)
							return
						}
						tv.app.serviceCache[serverName] = svcs
						tv.tree.Refresh()
					}()
				}
			}
		},
	}
}

// statusColor retourne la couleur correspondant au statut.
func statusColor(status services.Status) color.Color {
	switch status {
	case services.StatusActive:
		return DirectColorActive
	case services.StatusFailed:
		return DirectColorFailed
	case services.StatusInactive:
		return DirectColorInactive
	default:
		return DirectColorInactive
	}
}
