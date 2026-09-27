#!/usr/bin/env bash
set -e

if [ "$EUID" -ne 0 ]; then
  echo "Erreur: Ce script doit être exécuté avec les privilèges administrateur (sudo ./uninstall.sh)" >&2
  exit 1
fi

echo "🗑️ Désinstallation de ZiVPN Desktop..."

rm -f /usr/local/bin/zivpn-desktop
rm -f /usr/local/bin/zivpn-engine
rm -f /usr/share/icons/hicolor/512x512/apps/zivpn-desktop.png
rm -f /usr/share/applications/zivpn-desktop.desktop

if command -v update-desktop-database >/dev/null 2>&1; then
  update-desktop-database /usr/share/applications 2>/dev/null || true
fi

echo "✅ ZiVPN Desktop a été complètement désinstallé de votre système."
