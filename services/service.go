// Package services gère le listing et le contrôle des services systemd.
package services

// Status représente l'état d'un service.
type Status string

const (
	StatusActive       Status = "active"
	StatusInactive     Status = "inactive"
	StatusFailed       Status = "failed"
	StatusActivating   Status = "activating"
	StatusDeactivating Status = "deactivating"
	StatusUnknown      Status = "unknown"
)

// Service représente un service systemd sur un serveur distant.
type Service struct {
	Name        string `json:"unit"`
	Description string `json:"description"`
	LoadState   string `json:"load"`
	ActiveState string `json:"active"`
	SubState    string `json:"sub"`
	Status      Status `json:"-"`
}

// IsRunning retourne true si le service est actif.
func (s *Service) IsRunning() bool {
	return s.Status == StatusActive
}

// IsFailed retourne true si le service est en erreur.
func (s *Service) IsFailed() bool {
	return s.Status == StatusFailed
}

// StatusLabel retourne un label court pour l'état.
func (s *Service) StatusLabel() string {
	switch s.Status {
	case StatusActive:
		return "● Actif"
	case StatusInactive:
		return "○ Inactif"
	case StatusFailed:
		return "✖ Erreur"
	case StatusActivating:
		return "▶ Démarrage..."
	case StatusDeactivating:
		return "■ Arrêt..."
	default:
		return "? Inconnu"
	}
}

// ParseStatus convertit une chaîne en Status.
func ParseStatus(s string) Status {
	switch s {
	case "active":
		return StatusActive
	case "inactive":
		return StatusInactive
	case "failed":
		return StatusFailed
	case "activating":
		return StatusActivating
	case "deactivating":
		return StatusDeactivating
	default:
		return StatusUnknown
	}
}
