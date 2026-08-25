// Package ssh fournit un client SSH pour la connexion aux serveurs distants.
package ssh

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client représente une connexion SSH vers un serveur distant.
type Client struct {
	Host       string
	Port       int
	Username   string
	conn       *ssh.Client
	connTime   time.Time
}

// ConnectConfig contient les paramètres de connexion SSH.
type ConnectConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	AuthMethod string // "password" ou "key"
	KeyPath    string
}

// NewClient crée et connecte un nouveau client SSH.
func NewClient(cfg ConnectConfig) (*Client, error) {
	authMethods, err := buildAuthMethods(cfg)
	if err != nil {
		return nil, fmt.Errorf("erreur auth SSH: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: configurer known_hosts
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("erreur connexion SSH à %s: %w", addr, err)
	}

	return &Client{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		conn:     conn,
		connTime: time.Now(),
	}, nil
}

// Execute exécute une commande à distance et retourne la sortie.
func (c *Client) Execute(cmd string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("client SSH non connecté")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("erreur création session SSH: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		// Retourner aussi la sortie en cas d'erreur (utile pour les messages d'erreur)
		return string(output), fmt.Errorf("commande '%s': %w\nSortie: %s", cmd, err, string(output))
	}

	return string(output), nil
}

// ExecuteSudo exécute une commande avec sudo.
func (c *Client) ExecuteSudo(cmd, password string) (string, error) {
	// Utilise echo pour passer le mot de passe via stdin
	sudoCmd := fmt.Sprintf("echo '%s' | sudo -S %s", escapeShellArg(password), cmd)
	return c.Execute(sudoCmd)
}

// StartStream démarre une commande en streaming (pour tail -f, journalctl -f).
// Retourne un canal de lignes et une fonction pour arrêter le stream.
func (c *Client) StartStream(cmd string) (<-chan string, func(), error) {
	if c.conn == nil {
		return nil, nil, fmt.Errorf("client SSH non connecté")
	}

	session, err := c.conn.NewSession()
	if err != nil {
		return nil, nil, fmt.Errorf("erreur création session: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("erreur pipe stdout: %w", err)
	}

	if err := session.Start(cmd); err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("erreur démarrage commande: %w", err)
	}

	lines := make(chan string, 100)
	done := make(chan struct{})

	// Goroutine de lecture
	go func() {
		defer close(lines)
		defer session.Close()

		buf := make([]byte, 4096)
		var partial string
		for {
			select {
			case <-done:
				return
			default:
				n, err := stdout.Read(buf)
				if n > 0 {
					text := partial + string(buf[:n])
					parts := strings.Split(text, "\n")
					// Envoyer toutes les lignes complètes
					for i := 0; i < len(parts)-1; i++ {
						select {
						case lines <- parts[i]:
						case <-done:
							return
						}
					}
					// Garder la dernière partie (potentiellement incomplète)
					partial = parts[len(parts)-1]
				}
				if err != nil {
					if partial != "" {
						select {
						case lines <- partial:
						case <-done:
						}
					}
					return
				}
			}
		}
	}()

	stopFn := func() {
		close(done)
		session.Signal(ssh.SIGTERM)
	}

	return lines, stopFn, nil
}

// IsConnected vérifie si la connexion SSH est active.
func (c *Client) IsConnected() bool {
	if c.conn == nil {
		return false
	}
	_, _, err := c.conn.SendRequest("keepalive@vizualisor", true, nil)
	return err == nil
}

// Close ferme la connexion SSH.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Addr retourne l'adresse du serveur.
func (c *Client) Addr() string {
	return net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.Port))
}

// buildAuthMethods construit les méthodes d'authentification SSH.
func buildAuthMethods(cfg ConnectConfig) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	switch cfg.AuthMethod {
	case "key":
		key, err := os.ReadFile(cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("erreur lecture clé SSH %s: %w", cfg.KeyPath, err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("erreur parsing clé SSH: %w", err)
		}
		methods = append(methods, ssh.PublicKeys(signer))
	case "password":
		if cfg.Password != "" {
			methods = append(methods, ssh.Password(cfg.Password))
		}
	default:
		if cfg.Password != "" {
			methods = append(methods, ssh.Password(cfg.Password))
		}
	}

	return methods, nil
}

// escapeShellArg échappe un argument pour le shell.
func escapeShellArg(arg string) string {
	return strings.ReplaceAll(arg, "'", "'\"'\"'")
}
