package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

// publicIPViaSocks5 interroge un service d'IP publique EN PASSANT par le
// proxy SOCKS5 local. Le résultat prouve que le trafic sort réellement par
// le tunnel ZiVPN (et affiche l'IP de sortie = celle du serveur).
func publicIPViaSocks5(socksAddr string) (string, error) {
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return "", fmt.Errorf("proxy SOCKS5 injoignable: %w", err)
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		},
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	resp, err := client.Get("http://api.ipify.org?format=json")
	if err != nil {
		return "", fmt.Errorf("requête IP via le tunnel impossible: %w", err)
	}
	defer resp.Body.Close()
	var out struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("réponse IP illisible: %w", err)
	}
	if out.IP == "" {
		return "", fmt.Errorf("le service n'a retourné aucune IP")
	}
	return out.IP, nil
}
