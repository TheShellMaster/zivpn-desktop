# 🛡️ ZiVPN Desktop

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows-blue?style=flat)](https://github.com/TheShellMaster/zivpn-desktop/releases)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

Client desktop moderne, léger et graphique pour le protocole **ZiVPN UDP**.

Jusqu'à présent, ZiVPN n'était disponible que sur **Android** et le script serveur d'origine provient du dépôt [zahidbd2/udp-zivpn](https://github.com/zahidbd2/udp-zivpn). Ce projet fournit l'application de bureau complète (Windows & Linux) avec une interface graphique épurée qui reproduit fidèlement la simplicité de l'application mobile.

---

## 📸 Aperçu de l'interface

```text
┌──────────────────────────────────────────────┐
│                ZiVPN Desktop                 │
│              Client VPN UDP                  │
├──────────────────────────────────────────────┤
│                                              │
│  Adresse IP du serveur :                     │
│  [ 198.51.100.1                           ] │
│                                              │
│  Port UDP :                                  │
│  [ 5667                                    ] │
│                                              │
│  Mot de passe :                              │
│  [ ••••••••••••••••                        ] │
│                                              │
├──────────────────────────────────────────────┤
│                ⚪ Déconnecté                  │
│                                              │
│             [   CONNECTER   ]                │
├──────────────────────────────────────────────┤
│  ▶ Afficher les journaux d'activité          │
└──────────────────────────────────────────────┘
```

---

## ✨ Fonctionnalités

- **Expérience ultra-simple** : Pas de configuration complexe. Entrez simplement l'IP, le port et le mot de passe, puis cliquez sur **CONNECTER**.
- **Sauvegarde automatique** : Retient automatiquement vos identifiants pour vos futures sessions.
- **Vrai tunnel VPN (TUN)** : Redirige l'intégralité du trafic réseau du système à travers le tunnel UDP chiffré et obfusqué.
- **Port local SOCKS5 intégré** : Ouvre simultanément un proxy SOCKS5 local sur `127.0.0.1:1080` pour les applications compatibles ou les tests.
- **Journaux d'activité en direct** : Console rétractable intégrée pour suivre l'état de la connexion et diagnostiquer les erreurs.
- **Mode Ligne de Commande (CLI)** : Possibilité d'exécuter le client sans interface graphique (idéal pour les scripts ou les serveurs distants).

---

## 🚀 Téléchargement & Utilisation

Rendez-vous sur la page des [Releases](https://github.com/TheShellMaster/zivpn-desktop/releases) pour télécharger la dernière version.

### 🐧 Sur Linux

1. Téléchargez l'archive `zivpn-desktop-linux-amd64.tar.gz` et extrayez-la.
2. Assurez-vous que les permissions d'exécution sont définies :
   ```bash
   chmod +x zivpn-desktop zivpn-engine
   ```
3. Lancez l'application :
   ```bash
   sudo ./zivpn-desktop
   ```
   *(Note : `sudo` est nécessaire sous Linux pour permettre au système de créer l'interface réseau virtuelle `zivpn-tun`).*

### 🪟 Sur Windows

1. Téléchargez `zivpn-desktop-windows-amd64.zip` et extrayez-le dans un dossier.
2. Assurez-vous que `zivpn-engine.exe` et `wintun.dll` se trouvent dans le même dossier que `zivpn-desktop.exe`.
3. Faites un clic droit sur `zivpn-desktop.exe` et choisissez **Exécuter en tant qu'administrateur**.

---

## 💻 Mode Ligne de Commande (Optionnel)

Si vous souhaitez exécuter le client sans ouvrir l'interface graphique (ex: script d'automatisation, terminal pur) :

```bash
sudo ./zivpn-desktop -server 198.51.100.1 -port 5667 -password "votre_mot_de_passe"
```

Arguments disponibles :
| Drapeau | Description | Valeur par défaut |
|---|---|---|
| `-server` | Adresse IP ou nom d'hôte du serveur ZiVPN | *(Vide -> Lance la GUI)* |
| `-port` | Port d'écoute UDP du serveur | `5667` |
| `-password` | Mot de passe d'authentification | *(Obligatoire en CLI)* |
| `-engine` | Chemin vers un binaire moteur personnalisé | `./zivpn-engine` |

---

## 🏗️ Architecture & Rétro-ingénierie

Pour comprendre comment le protocole ZiVPN a été rétro-conçu, comment l'obfuscation QUIC fonctionne sous le capot, et comment l'application orchestre le moteur réseau, consultez la documentation dédiée :

👉 **[Consulter ARCHITECTURE.md](./ARCHITECTURE.md)**

---

## 🛠️ Compilation depuis les sources

### Prérequis
- **Go 1.23+**
- Compilateur C (**GCC**) et bibliothèques graphiques X11/OpenGL (`libgl1-mesa-dev`, `xorg-dev`).

### Étapes de compilation

```bash
# 1. Cloner le projet
git clone https://github.com/TheShellMaster/zivpn-desktop.git
cd zivpn-desktop

# 2. Télécharger les dépendances
go mod download

# 3. Compiler l'interface utilisateur
go build -o zivpn-desktop ./cmd/zivpn-desktop

# 4. Placer le moteur réseau zivpn-engine à côté
# (Téléchargeable depuis les releases Hysteria v1.3.5)
```

---

## 📄 Licence

Ce projet est distribué sous la licence **GNU General Public License v3 (GPLv3)**.
Le moteur réseau sous-jacent est basé sur le projet open-source [Hysteria v1](https://github.com/apernet/hysteria).
