# ZiVPN Desktop 🚀

ZiVPN Desktop est un client VPN avec interface graphique (GUI) multiplateforme permettant de se connecter à un serveur AWS spécifique. L'application redirige **l'intégralité du trafic de l'ordinateur** de manière transparente grâce à une interface réseau virtuelle (TUN), reproduisant ainsi parfaitement le comportement de l'application Android ZiVPN d'origine sur un environnement de bureau.

![ZiVPN Desktop](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-blue) ![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8) ![Fyne](https://img.shields.io/badge/GUI-Fyne-green)

---

## ✨ Fonctionnalités

- **Interface utilisateur simple et sombre** basée sur Fyne, imitant l'esthétique du client mobile.
- **Routage réseau global (TUN natif)** : Contrairement aux proxys SOCKS5 classiques, ZiVPN Desktop encapsule tout le trafic système via l'interface TUN intégrée.
- **Moteur basé sur Hysteria v2** : Le moteur réseau interne utilise une version patchée et optimisée de [Hysteria v2.12.3](https://github.com/apernet/hysteria) pour la prise en charge native du mode `auto_route` et du protocole QUIC (`h3`).
- **Obfuscation Salamander** : Connexion furtive utilisant le chiffrement XOR natif de ZiVPN (bypass d'inspection profonde des paquets).
- **Élévation de privilèges automatique** : L'application demande automatiquement les droits Administrateur (`pkexec` sur Linux, `osascript` sur Mac, UAC sur Windows) indispensables pour configurer les routes réseau du système.

---

## 📥 Téléchargement et Installation

Allez dans l'onglet **[Releases](../../releases/latest)** pour télécharger la dernière version pour votre système d'exploitation.

### 🐧 Linux
1. Téléchargez l'archive `zivpn-desktop-linux-amd64.tar.gz`.
2. Extrayez l'archive dans un dossier :
   ```bash
   tar -xzf zivpn-desktop-linux-amd64.tar.gz
   ```
3. Exécutez l'application (assurez-vous d'avoir `polkit` installé pour la fenêtre de mot de passe sudo) :
   ```bash
   cd linux
   ./zivpn-desktop
   ```

### 🪟 Windows
1. Téléchargez l'archive `zivpn-desktop-windows-amd64.zip`.
2. Décompressez le fichier `.zip`. Assurez-vous que le fichier `wintun.dll` se trouve **dans le même dossier** que `zivpn-desktop.exe` et `zivpn-engine.exe`.
3. Lancez `zivpn-desktop.exe`. L'application demandera les droits administrateur (UAC) lors de la connexion pour créer l'adaptateur virtuel Wintun.

### 🍏 macOS
> *Remarque : Le client macOS nécessite généralement d'être compilé directement sur la machine Apple cible ou téléchargé depuis les releases si disponible.*
1. Placez `zivpn-desktop` et `zivpn-engine` dans le même dossier.
2. Lancez l'application. Elle utilisera `osascript` pour vous demander le mot de passe administrateur lors de l'établissement du tunnel.

---

## 🎮 Comment l'utiliser ?

1. Au lancement de l'application, l'interface vous affichera deux champs principaux.
2. **Payload / IP** : L'adresse IP de votre serveur AWS `184.73.52.96` (port `5667` implicite).
3. **Mot de passe** : Le mot de passe de connexion à votre serveur AWS (ex: `admin` ou `Baba`). 
   *Note: Ne confondez pas ce mot de passe avec le mot de passe d'obfuscation système, ce dernier est géré automatiquement par l'application en arrière-plan.*
4. Cliquez sur le bouton **Connecter**.
5. Une invite système vous demandera votre mot de passe administrateur (requis pour modifier les routes de l'OS).
6. Le statut passera à **Connecté**. Tout le trafic de votre ordinateur transite désormais par votre serveur AWS !
7. Cliquez sur **Déconnecter** pour retrouver votre connexion locale classique.

---

## 🛠️ Compilation depuis les sources (Développeurs)

Si vous souhaitez modifier le code ou compiler vous-même les binaires.

### Prérequis
- **Go 1.21+** installé.
- **Compilateur C** : `gcc` pour Linux, `mingw-w64` pour la cross-compilation Windows.
- **Dépendances graphiques Fyne** (sur Linux) :
  ```bash
  sudo apt-get install libgl1-mesa-dev xorg-dev libxxf86vm-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libegl1-mesa-dev
  ```

### Instructions de compilation automatisées

Un script complet est inclus pour générer les releases Linux et Windows nativement depuis un environnement Linux.

```bash
# Clonez le dépôt
git clone https://github.com/TheShellMaster/zivpn-desktop.git
cd zivpn-desktop

# Lancez le script de build
./build-native.sh
```

Le script va automatiquement :
1. Copier les moteurs réseaux `zivpn-engine` précompilés dans le dossier `release/`.
2. Compiler l'interface graphique Linux native.
3. Utiliser `x86_64-w64-mingw32-gcc` (via CGO) pour compiler l'interface Windows (`.exe`) sans console graphique (`-H=windowsgui`).
4. Télécharger dynamiquement `wintun.dll` pour le support TUN sur Windows.
5. Créer les archives prêtes à l'emploi `.tar.gz` et `.zip` dans `release/`.

---

## 🧠 Détails Techniques et Architecture

### Patch du moteur Hysteria
Le serveur ZiVPN AWS cible est basé sur une version altérée d'Hysteria. Pour que le client de bureau puisse communiquer avec, le moteur fourni (dossier `zivpn-engine`) a été recompilé à partir des sources de Hysteria v2 avec les modifications suivantes :
- L'ALPN a été vérifié et correspond au standard QUIC HTTP/3 (`h3`).
- Les entêtes de négociation HTTP/3 ont été patchées : `Hysteria-` est remplacé par `Zivpnudp-`.
- Le domaine SNI de la requête interne a été ajusté sur `zivpnudp`.

### Obfuscation (Salamander)
Une rétro-ingénierie du client Android a permis d'isoler le mot de passe d'obfuscation. Bien que l'interface de l'APK indique "zivpn", la clé transmise au serveur est une version altérée par une opération binaire (XOR), aboutissant à la clé `hu``hqb`c`. L'application de bureau injecte dynamiquement cette chaîne exacte pour que le handshake QUIC réussisse.

---

*Développé avec passion pour reproduire l'expérience Android en Desktop.* 🚀
