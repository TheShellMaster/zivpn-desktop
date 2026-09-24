package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Connection is the minimal ZiVPN desktop configuration. ZiVPN authenticates
// with a password only; there is no separate username in the server protocol.
type Connection struct {
	Server   string `json:"server"`
	Port     int    `json:"-"`
	Password string `json:"auth_str"`
	Obfs     string `json:"obfs"`
}

type EngineConfig struct {
	Server   string        `json:"server"`
	Up       int           `json:"up_mbps"`
	Down     int           `json:"down_mbps"`
	Protocol string        `json:"protocol,omitempty"`
	Obfs     string        `json:"obfs"`
	AuthStr  string        `json:"auth_str"`
	Insecure bool          `json:"insecure"`
	Tun      *TunConfig    `json:"tun,omitempty"`
	Socks5   *Socks5Config `json:"socks5,omitempty"`
}

type TunConfig struct {
	Name    string `json:"name"`
	Timeout int    `json:"timeout"`
	MTU     int    `json:"mtu"`
}

type Socks5Config struct {
	Listen string `json:"listen"`
}

func (c Connection) Validate() error {
	if strings.TrimSpace(c.Server) == "" {
		return fmt.Errorf("l'adresse du serveur est obligatoire")
	}
	if net.ParseIP(strings.Trim(c.Server, "[]")) == nil && strings.ContainsAny(c.Server, " /\\") {
		return fmt.Errorf("adresse serveur invalide")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("le port doit être compris entre 1 et 65535")
	}
	if strings.TrimSpace(c.Password) == "" || strings.ContainsAny(c.Password, " \t\r\n") {
		return fmt.Errorf("le mot de passe est obligatoire")
	}
	return nil
}

func (c Connection) Engine() EngineConfig {
	return EngineConfig{
		Server:   net.JoinHostPort(c.Server, strconv.Itoa(c.Port)),
		Up:       100,
		Down:     100,
		Obfs:     valueOr(c.Obfs, "zivpn"),
		AuthStr:  c.Password,
		Insecure: true,
		Tun: &TunConfig{
			Name:    "zivpn-tun",
			Timeout: 300,
			MTU:     1500,
		},
		Socks5: &Socks5Config{
			Listen: "127.0.0.1:1080",
		},
	}
}

func (c Connection) SaveProfile(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, "profile.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func LoadProfile(dir string) (Connection, error) {
	path := filepath.Join(dir, "profile.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return Connection{Server: "", Port: 5667, Password: ""}, err
	}
	var c Connection
	if err := json.Unmarshal(data, &c); err != nil {
		return Connection{Server: "", Port: 5667, Password: ""}, err
	}
	if c.Port == 0 {
		c.Port = 5667
	}
	return c, nil
}

func (c Connection) WriteEngineConfig(dir string) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("création du dossier de configuration: %w", err)
	}
	path := filepath.Join(dir, "client.json")
	payload, err := json.MarshalIndent(c.Engine(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("encodage de la configuration: %w", err)
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("écriture de la configuration: %w", err)
	}
	return path, nil
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
