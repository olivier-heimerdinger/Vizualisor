package ssh

import (
	"fmt"
	"sync"
	"time"

	"github.com/olivier-heimerdinger/Vizualisor/config"
)

// Pool gère un ensemble de connexions SSH vers différents serveurs.
type Pool struct {
	mu      sync.RWMutex
	clients map[string]*Client // clé = nom du serveur
	creds   *config.CredentialResolver
}

// NewPool crée un nouveau pool de connexions SSH.
func NewPool(creds *config.CredentialResolver) *Pool {
	return &Pool{
		clients: make(map[string]*Client),
		creds:   creds,
	}
}

// Connect établit une connexion SSH vers un serveur.
func (p *Pool) Connect(server config.ServerConfig) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Vérifier si une connexion existe déjà
	if client, ok := p.clients[server.Name]; ok {
		if client.IsConnected() {
			return client, nil
		}
		// Connexion morte, la nettoyer
		client.Close()
		delete(p.clients, server.Name)
	}

	// Résoudre les credentials
	username := p.creds.ResolveUsername(server)
	password := p.creds.ResolvePassword(server)

	cfg := ConnectConfig{
		Host:       server.Host,
		Port:       server.Port,
		Username:   username,
		Password:   password,
		AuthMethod: server.AuthMethod,
		KeyPath:    server.KeyPath,
	}

	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	p.clients[server.Name] = client
	return client, nil
}

// Get retourne le client SSH pour un serveur donné.
func (p *Pool) Get(serverName string) (*Client, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	client, ok := p.clients[serverName]
	return client, ok
}

// Disconnect ferme la connexion d'un serveur.
func (p *Pool) Disconnect(serverName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if client, ok := p.clients[serverName]; ok {
		err := client.Close()
		delete(p.clients, serverName)
		return err
	}
	return nil
}

// DisconnectAll ferme toutes les connexions.
func (p *Pool) DisconnectAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, client := range p.clients {
		client.Close()
		delete(p.clients, name)
	}
}

// Status retourne le statut de connexion de tous les serveurs.
func (p *Pool) Status() map[string]bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := make(map[string]bool)
	for name, client := range p.clients {
		status[name] = client.IsConnected()
	}
	return status
}

// KeepAlive vérifie et maintient les connexions actives.
func (p *Pool) KeepAlive(servers []config.ServerConfig, interval time.Duration) chan struct{} {
	stop := make(chan struct{})
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				p.mu.RLock()
				for name, client := range p.clients {
					if !client.IsConnected() {
						// Tenter une reconnexion
						for _, srv := range servers {
							if srv.Name == name {
								go func(s config.ServerConfig) {
									_, err := p.Connect(s)
									if err != nil {
										fmt.Printf("Reconnexion échouée pour %s: %v\n", s.Name, err)
									}
								}(srv)
								break
							}
						}
					}
				}
				p.mu.RUnlock()
			}
		}
	}()

	return stop
}
