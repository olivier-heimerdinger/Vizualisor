package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/olivier-heimerdinger/Vizualisor/ssh"
)

// Manager gère les opérations sur les services d'un serveur distant.
type Manager struct {
	client     *ssh.Client
	serverName string
}

// NewManager crée un nouveau gestionnaire de services.
func NewManager(client *ssh.Client, serverName string) *Manager {
	return &Manager{
		client:     client,
		serverName: serverName,
	}
}

// ListServices retourne la liste de tous les services systemd.
func (m *Manager) ListServices() ([]Service, error) {
	// Utiliser le format JSON de systemctl pour un parsing fiable
	output, err := m.client.Execute("systemctl list-units --type=service --all --output=json --no-pager")
	if err != nil {
		// Fallback : parsing texte si JSON non disponible
		return m.listServicesFallback()
	}

	var rawServices []Service
	if err := json.Unmarshal([]byte(output), &rawServices); err != nil {
		// Fallback si le JSON est invalide
		return m.listServicesFallback()
	}

	// Enrichir avec le Status parsé
	for i := range rawServices {
		rawServices[i].Status = ParseStatus(rawServices[i].ActiveState)
		// Nettoyer le nom du service
		rawServices[i].Name = strings.TrimSuffix(rawServices[i].Name, ".service")
	}

	return rawServices, nil
}

// listServicesFallback utilise le parsing texte comme backup.
func (m *Manager) listServicesFallback() ([]Service, error) {
	output, err := m.client.Execute("systemctl list-units --type=service --all --no-pager --no-legend")
	if err != nil {
		return nil, fmt.Errorf("erreur listing services sur %s: %w", m.serverName, err)
	}

	var services []Service
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: UNIT LOAD ACTIVE SUB DESCRIPTION
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := strings.TrimSuffix(fields[0], ".service")
		// Ignorer le ● au début si présent
		if name == "●" && len(fields) > 4 {
			name = strings.TrimSuffix(fields[1], ".service")
			fields = fields[1:]
		}

		svc := Service{
			Name:        name,
			LoadState:   fields[1],
			ActiveState: fields[2],
			SubState:    fields[3],
			Status:      ParseStatus(fields[2]),
		}

		if len(fields) > 4 {
			svc.Description = strings.Join(fields[4:], " ")
		}

		services = append(services, svc)
	}

	return services, nil
}

// SearchServices recherche des services par nom.
func (m *Manager) SearchServices(query string) ([]Service, error) {
	all, err := m.ListServices()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return all, nil
	}

	query = strings.ToLower(query)
	var results []Service
	for _, svc := range all {
		if strings.Contains(strings.ToLower(svc.Name), query) ||
			strings.Contains(strings.ToLower(svc.Description), query) {
			results = append(results, svc)
		}
	}

	return results, nil
}

// StartService démarre un service.
func (m *Manager) StartService(name, sudoPassword string) error {
	cmd := fmt.Sprintf("systemctl start %s.service", name)
	_, err := m.client.ExecuteSudo(cmd, sudoPassword)
	if err != nil {
		return fmt.Errorf("erreur démarrage %s: %w", name, err)
	}
	return nil
}

// StopService arrête un service.
func (m *Manager) StopService(name, sudoPassword string) error {
	cmd := fmt.Sprintf("systemctl stop %s.service", name)
	_, err := m.client.ExecuteSudo(cmd, sudoPassword)
	if err != nil {
		return fmt.Errorf("erreur arrêt %s: %w", name, err)
	}
	return nil
}

// RestartService redémarre un service.
func (m *Manager) RestartService(name, sudoPassword string) error {
	cmd := fmt.Sprintf("systemctl restart %s.service", name)
	_, err := m.client.ExecuteSudo(cmd, sudoPassword)
	if err != nil {
		return fmt.Errorf("erreur redémarrage %s: %w", name, err)
	}
	return nil
}

// GetServiceStatus retourne le statut détaillé d'un service.
func (m *Manager) GetServiceStatus(name string) (string, error) {
	output, err := m.client.Execute(fmt.Sprintf("systemctl status %s.service --no-pager", name))
	if err != nil {
		// systemctl status retourne exit code 3 pour un service inactif
		// On retourne quand même la sortie
		if output != "" {
			return output, nil
		}
		return "", err
	}
	return output, nil
}
