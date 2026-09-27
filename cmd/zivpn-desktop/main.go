package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheShellMaster/zivpn-desktop/internal/config"
	"github.com/TheShellMaster/zivpn-desktop/internal/engine"
	"github.com/TheShellMaster/zivpn-desktop/internal/ui"
)

func main() {
	server := flag.String("server", "", "IP ou host ZiVPN (laisser vide pour lancer l'interface graphique)")
	port := flag.Int("port", 5667, "port UDP ZiVPN")
	password := flag.String("password", "", "mot de passe ZiVPN")
	binary := flag.String("engine", "", "chemin vers zivpn-engine")
	flag.Parse()

	// Si aucun serveur n'est passé en ligne de commande, on ouvre l'interface graphique
	if *server == "" {
		ui.Run()
		return
	}

	// Mode CLI de secours si un serveur est explicitement spécifié en argument
	connection := config.Connection{Server: *server, Port: *port, Password: *password}
	stateDir, err := os.UserConfigDir()
	if err != nil {
		stateDir = "."
	}
	appDir := stateDir + string(os.PathSeparator) + "ZiVPN Desktop"
	configPath, err := connection.WriteEngineConfig(appDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	process, err := engine.Start(ctx, *binary, configPath, func(line string) {
		fmt.Println(line)
	}, func(err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Moteur arrêté avec erreur: %v\n", err)
		}
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer process.Stop()

	fmt.Printf("ZiVPN Desktop connecté à %s:%d\n", connection.Server, connection.Port)
	<-ctx.Done()
}
