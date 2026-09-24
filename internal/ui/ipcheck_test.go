package ui

import (
	"os"
	"strings"
	"testing"
)

func TestPublicIPViaSocks5NoProxy(t *testing.T) {
	// Aucun SOCKS5 n'écoute ici : la fonction doit retourner une erreur,
	// jamais une IP vide ni un crash.
	if _, err := publicIPViaSocks5("127.0.0.1:1"); err == nil {
		t.Fatal("aucune erreur alors que le proxy est injoignable")
	}
}

func TestPublicIPViaSocks5Live(t *testing.T) {
	// Test d'intégration manuel : démarre un tunnel ZiVPN local vers
	// 127.0.0.1:1080 puis :
	//   ZIVPN_TEST_SOCKS=127.0.0.1:1080 go test ./internal/ui/ -run Live -v
	addr := os.Getenv("ZIVPN_TEST_SOCKS")
	if addr == "" {
		t.Skip("ZIVPN_TEST_SOCKS non défini, test ignoré")
	}
	ip, err := publicIPViaSocks5(addr)
	if err != nil {
		t.Fatalf("tunnel injoignable: %v", err)
	}
	if !strings.Contains(ip, ".") && !strings.Contains(ip, ":") {
		t.Fatalf("IP invalide retournée: %q", ip)
	}
	t.Logf("IP de sortie via le tunnel: %s", ip)
}
