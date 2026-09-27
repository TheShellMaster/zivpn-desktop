#!/bin/bash
set -e

VERSION="v2.2.3"
BASE_URL="https://github.com/apernet/hysteria/releases/download/app%2F${VERSION}"

mkdir -p dist/linux dist/windows dist/macos

echo "==> Téléchargement et Patch des moteurs réseau..."
# Linux
if [ ! -f "dist/linux/hysteria-stock" ]; then
    curl -sL "${BASE_URL}/hysteria-linux-amd64" -o dist/linux/hysteria-stock
    chmod +x dist/linux/hysteria-stock
fi
python3 scripts/patch-engine.py dist/linux/hysteria-stock dist/linux/zivpn-engine
chmod +x dist/linux/zivpn-engine

# Windows
if [ ! -f "dist/windows/hysteria-stock.exe" ]; then
    curl -sL "${BASE_URL}/hysteria-windows-amd64.exe" -o dist/windows/hysteria-stock.exe
fi
if [ ! -f "dist/windows/wintun.dll" ]; then
    curl -sL https://www.wintun.net/builds/wintun-0.14.1.zip -o wintun.zip
    unzip -p wintun.zip wintun/bin/amd64/wintun.dll > dist/windows/wintun.dll
    rm wintun.zip
fi
python3 scripts/patch-engine.py dist/windows/hysteria-stock.exe dist/windows/zivpn-engine.exe

# macOS (AMD64)
if [ ! -f "dist/macos/hysteria-stock" ]; then
    curl -sL "${BASE_URL}/hysteria-darwin-amd64" -o dist/macos/hysteria-stock
    chmod +x dist/macos/hysteria-stock
fi
python3 scripts/patch-engine.py dist/macos/hysteria-stock dist/macos/zivpn-engine
chmod +x dist/macos/zivpn-engine

echo "==> Cross-Compilation de l'interface Fyne avec fyne-cross (Docker)..."
export PATH=$PATH:$(go env GOPATH)/bin

echo "  -> Linux"
fyne-cross linux -arch=amd64 -app-id com.zivpn.desktop -icon assets/icon.png -dir ./cmd/zivpn-desktop
cp fyne-cross/bin/linux-amd64/zivpn-desktop dist/linux/

echo "  -> Windows"
fyne-cross windows -arch=amd64 -app-id com.zivpn.desktop -icon assets/icon.png -dir ./cmd/zivpn-desktop
cp fyne-cross/bin/windows-amd64/zivpn-desktop.exe dist/windows/

echo "  -> macOS"
fyne-cross darwin -arch=amd64 -app-id com.zivpn.desktop -icon assets/icon.png -dir ./cmd/zivpn-desktop
cp fyne-cross/bin/darwin-amd64/zivpn-desktop dist/macos/

echo "==> Packaging..."
cd dist/linux && tar -czvf ../zivpn-desktop-linux-amd64.tar.gz zivpn-desktop zivpn-engine && cd ../..
cd dist/windows && zip ../zivpn-desktop-windows-amd64.zip zivpn-desktop.exe zivpn-engine.exe wintun.dll && cd ../..
cd dist/macos && tar -czvf ../zivpn-desktop-macos-amd64.tar.gz zivpn-desktop zivpn-engine && cd ../..

echo "==> TOUT EST PRÊT DANS LE DOSSIER 'dist' !"
