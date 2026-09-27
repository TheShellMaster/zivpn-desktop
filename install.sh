#!/usr/bin/env bash
set -e

if [ "$EUID" -ne 0 ]; then
  echo "Erreur: Ce script doit être exécuté avec les privilèges administrateur (sudo ./install.sh)" >&2
  exit 1
fi

echo "📦 Installation de ZiVPN Desktop sur le système..."

# 1. Copier les exécutables dans /usr/local/bin
install -m 755 zivpn-desktop /usr/local/bin/zivpn-desktop
install -m 755 zivpn-engine /usr/local/bin/zivpn-engine

# 2. Configurer les capacités réseau pour le tunnel sans requérir sudo à chaque lancement
if command -v setcap >/dev/null 2>&1; then
  echo "⚙️ Configuration des capacités réseau Linux (cap_net_admin)..."
  setcap cap_net_admin,cap_net_bind_service=+ep /usr/local/bin/zivpn-engine || true
fi

# 3. Installer l'icône
mkdir -p /usr/share/icons/hicolor/512x512/apps
install -m 644 assets/icon.png /usr/share/icons/hicolor/512x512/apps/zivpn-desktop.png

# 4. Installer le raccourci du menu d'applications
cat << 'EOF' > /usr/share/applications/zivpn-desktop.desktop
[Desktop Entry]
Name=ZiVPN Desktop
Comment=Client VPN UDP pour ZiVPN
Exec=zivpn-desktop
Icon=zivpn-desktop
Terminal=false
Type=Application
Categories=Network;
Keywords=vpn;zivpn;udp;tunnel;
StartupWMClass=com.zivpn.desktop
EOF

# 5. Mettre à jour les caches système
if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database /usr/share/applications 2>/dev/null || true
fi
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  gtk-update-icon-cache -f -t /usr/share/icons/hicolor 2>/dev/null || true
fi

echo "✅ Installation terminée avec succès !"
echo "👉 Vous pouvez maintenant lancer 'ZiVPN Desktop' directement depuis le menu de vos applications."
