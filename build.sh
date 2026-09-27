#!/bin/bash
set -e

echo "=> Installation des dépendances graphiques pour Linux (Fyne)..."
sudo apt-get update
sudo apt-get install -y libgl1-mesa-dev xorg-dev libxxf86vm-dev gcc

echo "=> Patch du moteur ZiVPN (Hysteria v2.2.3)..."
if [ ! -f "hysteria-stock" ]; then
    curl -sL "https://github.com/apernet/hysteria/releases/download/app%2Fv2.2.3/hysteria-linux-amd64" -o hysteria-stock
    chmod +x hysteria-stock
fi
python3 scripts/patch-engine.py hysteria-stock zivpn-engine
chmod +x zivpn-engine

echo "=> Build de l'interface ZiVPN Desktop..."
go mod tidy
go build -o zivpn-desktop ./cmd/zivpn-desktop
chmod +x zivpn-desktop

echo "=> Terminé ! Vous pouvez lancer ./zivpn-desktop en root (sudo ./zivpn-desktop) pour activer le TUN."
