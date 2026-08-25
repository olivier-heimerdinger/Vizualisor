package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
)

// Couleurs thème pour les états des services.
var (
	// Couleurs d'état
	ColorActive   = colorToTheme(color.NRGBA{R: 76, G: 175, B: 80, A: 255})   // Vert
	ColorFailed   = colorToTheme(color.NRGBA{R: 244, G: 67, B: 54, A: 255})   // Rouge
	ColorInactive = colorToTheme(color.NRGBA{R: 158, G: 158, B: 158, A: 255}) // Gris
	ColorWarning  = colorToTheme(color.NRGBA{R: 255, G: 152, B: 0, A: 255})   // Orange
	ColorLog      = colorToTheme(color.NRGBA{R: 33, G: 150, B: 243, A: 255})  // Bleu

	// Couleurs directes (pour l'utilisation dans les canvas)
	DirectColorActive   = color.NRGBA{R: 76, G: 175, B: 80, A: 255}
	DirectColorFailed   = color.NRGBA{R: 244, G: 67, B: 54, A: 255}
	DirectColorInactive = color.NRGBA{R: 158, G: 158, B: 158, A: 255}
	DirectColorWarning  = color.NRGBA{R: 255, G: 152, B: 0, A: 255}
	DirectColorLog      = color.NRGBA{R: 33, G: 150, B: 243, A: 255}
)

// colorToTheme est un helper - pour Fyne on utilise directement les couleurs.
// Cette fonction existe pour montrer l'intention mais retourne une couleur Go directe.
func colorToTheme(c color.Color) color.Color {
	return c
}

// StatusIcon retourne une icône pour un état de service.
func StatusIcon(status string) fyne.Resource {
	// Utiliser les icônes intégrées de Fyne
	switch status {
	case "active":
		return resourceActiveIcon
	case "failed":
		return resourceFailedIcon
	case "inactive":
		return resourceInactiveIcon
	default:
		return resourceUnknownIcon
	}
}

// Icônes SVG inline pour les états des services.
var (
	resourceActiveIcon = fyne.NewStaticResource("active.svg", []byte(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#4CAF50">
  <circle cx="12" cy="12" r="10"/>
</svg>`))

	resourceFailedIcon = fyne.NewStaticResource("failed.svg", []byte(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#F44336">
  <circle cx="12" cy="12" r="10"/>
  <line x1="8" y1="8" x2="16" y2="16" stroke="white" stroke-width="2"/>
  <line x1="16" y1="8" x2="8" y2="16" stroke="white" stroke-width="2"/>
</svg>`))

	resourceInactiveIcon = fyne.NewStaticResource("inactive.svg", []byte(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#9E9E9E">
  <circle cx="12" cy="12" r="10"/>
</svg>`))

	resourceUnknownIcon = fyne.NewStaticResource("unknown.svg", []byte(`
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#FF9800">
  <circle cx="12" cy="12" r="10"/>
  <text x="12" y="16" text-anchor="middle" fill="white" font-size="14">?</text>
</svg>`))
)
