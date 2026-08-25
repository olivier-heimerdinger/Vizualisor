package logs

import (
	"fmt"

	"github.com/olivier-heimerdinger/Vizualisor/ssh"
)

// Streamer fournit un streaming temps réel de logs.
type Streamer struct {
	client     *ssh.Client
	serverName string
	stopFn     func()
	lines      <-chan string
	active     bool
}

// NewStreamer crée un nouveau streamer de logs.
func NewStreamer(client *ssh.Client, serverName string) *Streamer {
	return &Streamer{
		client:     client,
		serverName: serverName,
	}
}

// StreamServiceLog démarre le streaming de logs d'un service via journalctl -f.
func (s *Streamer) StreamServiceLog(serviceName string) (<-chan string, error) {
	if s.active {
		s.Stop()
	}

	cmd := fmt.Sprintf("journalctl -u %s.service -f --no-pager", serviceName)
	lines, stopFn, err := s.client.StartStream(cmd)
	if err != nil {
		return nil, fmt.Errorf("erreur streaming logs %s: %w", serviceName, err)
	}

	s.lines = lines
	s.stopFn = stopFn
	s.active = true

	return lines, nil
}

// StreamFileLog démarre le streaming d'un fichier de log via tail -f.
func (s *Streamer) StreamFileLog(filePath string) (<-chan string, error) {
	if s.active {
		s.Stop()
	}

	cmd := fmt.Sprintf("tail -f %s", filePath)
	lines, stopFn, err := s.client.StartStream(cmd)
	if err != nil {
		return nil, fmt.Errorf("erreur streaming fichier %s: %w", filePath, err)
	}

	s.lines = lines
	s.stopFn = stopFn
	s.active = true

	return lines, nil
}

// Stop arrête le streaming en cours.
func (s *Streamer) Stop() {
	if s.stopFn != nil {
		s.stopFn()
		s.stopFn = nil
		s.lines = nil
		s.active = false
	}
}

// IsActive indique si un stream est en cours.
func (s *Streamer) IsActive() bool {
	return s.active
}
