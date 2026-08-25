// Package logs gère la lecture de logs depuis des serveurs distants.
package logs

import (
	"fmt"
	"strings"

	"github.com/olivier-heimerdinger/Vizualisor/ssh"
)

// Reader lit des logs depuis un serveur distant.
type Reader struct {
	client     *ssh.Client
	serverName string
}

// NewReader crée un nouveau lecteur de logs.
func NewReader(client *ssh.Client, serverName string) *Reader {
	return &Reader{
		client:     client,
		serverName: serverName,
	}
}

// ReadServiceLog lit les dernières lignes de log d'un service via journalctl.
func (r *Reader) ReadServiceLog(serviceName string, lines int) (string, error) {
	if lines <= 0 {
		lines = 100
	}
	cmd := fmt.Sprintf("journalctl -u %s.service --no-pager -n %d", serviceName, lines)
	output, err := r.client.Execute(cmd)
	if err != nil {
		return "", fmt.Errorf("erreur lecture logs %s: %w", serviceName, err)
	}
	return output, nil
}

// ReadFileLog lit les dernières lignes d'un fichier de log.
func (r *Reader) ReadFileLog(filePath string, lines int) (string, error) {
	if lines <= 0 {
		lines = 100
	}
	cmd := fmt.Sprintf("tail -n %d %s", lines, filePath)
	output, err := r.client.Execute(cmd)
	if err != nil {
		return "", fmt.Errorf("erreur lecture fichier %s: %w", filePath, err)
	}
	return output, nil
}

// FilterLines filtre les lignes contenant un motif.
func FilterLines(content, filter string) string {
	if filter == "" {
		return content
	}

	filter = strings.ToLower(filter)
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), filter) {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}
