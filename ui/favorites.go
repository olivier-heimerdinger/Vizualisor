package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/olivier-heimerdinger/Vizualisor/services"
)

// FavoritesView affiche et gère les services favoris.
type FavoritesView struct {
	app       *App
	list      *widget.List
	container *fyne.Container
	showOnly  bool // Afficher uniquement les favoris
	toggle    *widget.Check
}

// NewFavoritesView crée une nouvelle vue de favoris.
func NewFavoritesView(app *App) *FavoritesView {
	fv := &FavoritesView{app: app}
	fv.build()
	return fv
}

func (fv *FavoritesView) build() {
	fv.toggle = widget.NewCheck("Favoris uniquement", func(checked bool) {
		fv.showOnly = checked
		if fv.app.treeView != nil {
			fv.app.treeView.Refresh()
		}
	})

	fv.container = container.NewVBox(fv.toggle)
}

// Container retourne le conteneur du composant.
func (fv *FavoritesView) Container() fyne.CanvasObject {
	return fv.container
}

// IsShowingFavoritesOnly retourne true si le filtre favoris est actif.
func (fv *FavoritesView) IsShowingFavoritesOnly() bool {
	return fv.showOnly
}

// GetFavoriteServices retourne les services favoris pour un serveur.
func (fv *FavoritesView) GetFavoriteServices(serverName string) []services.Service {
	favNames := fv.app.configMgr.GetFavorites(serverName)
	if len(favNames) == 0 {
		return nil
	}

	allServices := fv.app.getServicesForServer(serverName)
	favSet := make(map[string]bool, len(favNames))
	for _, name := range favNames {
		favSet[name] = true
	}

	var result []services.Service
	for _, svc := range allServices {
		if favSet[svc.Name] {
			result = append(result, svc)
		}
	}

	return result
}
