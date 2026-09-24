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
	Password string `json:"password"`
}

// SalamanderPassword is the obfuscation password used by ZiVPN servers.
// Verified by decompiling the official Android client (com.zi.zivpn) :
// G.java builds the client config with "obfs" = A2.d.k(C1.n.e), where
// A2.d.k XOR-decodes with key "ZCQV", yielding exactly this string.
// The official server binary enforces the same value ("wtf!!!No Idea"
// otherwise). It is NOT the "zivpn" string from the server config file.
const SalamanderPassword = "hu``hqb`c"

type EngineConfig struct {
	Server string        `json:"server"`
	Auth   string        `json:"auth"`
	TLS    TLSConfig     `json:"tls"`
	Obfs   ObfsConfig    `json:"obfs"`
	Socks5 *Socks5Config `json:"socks5,omitempty"`
}

type TLSConfig struct {
	SNI      string `json:"sni,omitempty"`
	Insecure bool   `json:"insecure"`
}

type ObfsConfig struct {
	Type       string           `json:"type"`
	Salamander SalamanderConfig `json:"salamander"`
}

type SalamanderConfig struct {
	Password string `json:"password"`
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
		Server: net.JoinHostPort(c.Server, strconv.Itoa(c.Port)),
		Auth:   c.Password,
		TLS: TLSConfig{
			Insecure: true,
		},
		Obfs: ObfsConfig{
			Type: "salamander",
			Salamander: SalamanderConfig{
				Password: SalamanderPassword,
			},
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
