#!/bin/bash
set -e

echo "Building ZiVPN Desktop Native Release..."

mkdir -p dist/linux dist/windows dist/macos
mkdir -p release/linux release/windows release/macos

# 1. Copier les moteurs
echo "Copie des moteurs..."
# Linux (déjà build)
cp dist/linux/zivpn-engine release/linux/zivpn-engine 2>/dev/null || true
# Windows (déjà build)
cp dist/windows/zivpn-engine.exe release/windows/zivpn-engine.exe 2>/dev/null || true
# Mac (déjà build)
cp dist/macos/zivpn-engine release/macos/zivpn-engine 2>/dev/null || true

# 2. Build Linux App
echo "Compilation Fyne Linux..."
pushd cmd/zivpn-desktop
go build -o ../../release/linux/zivpn-desktop .
popd

# 3. Build Windows App (requires mingw-w64)
echo "Compilation Fyne Windows..."
if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    pushd cmd/zivpn-desktop
    CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags -H=windowsgui -o ../../release/windows/zivpn-desktop.exe .
    popd
    # Copier wintun.dll si présent (sinon il faudra le télécharger)
    wget -qO wintun.zip "https://www.wintun.net/builds/wintun-0.14.1.zip"
    unzip -j wintun.zip "wintun/bin/amd64/wintun.dll" -d release/windows/ || true
else
    echo "x86_64-w64-mingw32-gcc manquant. Installez mingw-w64."
fi

# 4. Mac OS
echo "Pour Mac OS, la compilation Fyne doit se faire depuis un Mac ou via fyne-cross/docker."

# 5. Création des archives
echo "Création des archives ZIP/TAR..."
cd release
tar -czvf zivpn-desktop-linux-amd64.tar.gz linux/
zip -r zivpn-desktop-windows-amd64.zip windows/

echo "Terminé! Les applications se trouvent dans le dossier release/"
