package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConnectionEngineUsesPasswordOnly(t *testing.T) {
	connection := Connection{Server: "198.51.100.20", Port: 5667, Password: "admin"}
	engine := connection.Engine()
	if engine.Server != "198.51.100.20:5667" || engine.Auth != "admin" {
		t.Fatalf("configuration ZiVPN inattendue: %+v", engine)
	}
	if engine.Obfs.Type != "salamander" || engine.Obfs.Salamander.Password != SalamanderPassword {
		t.Fatalf("obfuscation ZiVPN inattendue: %+v", engine.Obfs)
	}
	if !engine.TLS.Insecure {
		t.Fatalf("TLS insecure attendu pour le certificat auto-signé: %+v", engine.TLS)
	}
	if engine.Socks5 == nil || engine.Socks5.Listen != "127.0.0.1:1080" {
		t.Fatalf("proxy SOCKS5 local attendu: %+v", engine.Socks5)
	}
	encoded, err := json.Marshal(engine)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" {
		t.Fatal("configuration vide")
	}
}

func TestConnectionRejectsInvalidInput(t *testing.T) {
	cases := []Connection{
		{Server: "", Port: 5667, Password: "admin"},
		{Server: "198.51.100.20", Port: 0, Password: "admin"},
		{Server: "198.51.100.20", Port: 5667, Password: "has space"},
	}
	for _, connection := range cases {
		if err := connection.Validate(); err == nil {
			t.Fatalf("entrée acceptée à tort: %+v", connection)
		}
	}
}

func TestWriteEngineConfigIsPrivate(t *testing.T) {
	dir := t.TempDir()
	path, err := (Connection{Server: "198.51.100.20", Port: 5667, Password: "admin"}).WriteEngineConfig(filepath.Join(dir, "config"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions configuration inattendues: %o", info.Mode().Perm())
	}
}
