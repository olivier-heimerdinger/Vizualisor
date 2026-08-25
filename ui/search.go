package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SearchBar fournit une barre de recherche pour filtrer le TreeView.
type SearchBar struct {
	app   *App
	entry *widget.Entry
}

// NewSearchBar crée une nouvelle barre de recherche.
func NewSearchBar(app *App) *SearchBar {
	sb := &SearchBar{app: app}

	sb.entry = widget.NewEntry()
	sb.entry.SetPlaceHolder("🔍 Rechercher serveurs, services...")
	sb.entry.OnChanged = sb.onSearch

	return sb
}

// Container retourne le conteneur de la barre de recherche.
func (sb *SearchBar) Container() fyne.CanvasObject {
	return sb.entry
}

func (sb *SearchBar) onSearch(query string) {
	if sb.app.treeView != nil {
		sb.app.treeView.SetFilter(query)
	}
}
