# 🔬 Architecture & Ingénierie Inverse du Protocole ZiVPN

Ce document détaille l'analyse technique approfondie du protocole **ZiVPN UDP**, la méthode de rétro-ingénierie appliquée pour identifier ses fondations, et les choix de conception architecturale qui ont permis de créer le premier client de bureau fonctionnel (**ZiVPN Desktop**).

---

## 1. Contexte & Problématique

**ZiVPN** est un outil de tunneling UDP populaire dans plusieurs régions pour contourner la censure et optimiser les connexions mobiles instables. 

Cependant, son créateur (disponible via des scripts d'installation sur le dépôt GitHub `zahidbd2/udp-zivpn`) n'a développé qu'une **application Android**. Les utilisateurs de PC (Windows, Linux, macOS) se retrouvaient jusqu'alors bloqués sans possibilité de se connecter directement à leurs serveurs ZiVPN depuis leur ordinateur.

Ce projet a pour objectif d'apporter cette compatibilité desktop sans altérer le serveur existant.

---

## 2. Rétro-ingénierie du Serveur ZiVPN

Pour concevoir un client desktop compatible, nous avons analysé le serveur ZiVPN fonctionnel installé sur un VPS Linux Debian/Ubuntu.

### 2.1. Analyse des fichiers d'installation
Le script d'installation officiel (`zi.sh`) déploie les éléments suivants :
- Un exécutable serveur dans `/usr/local/bin/zivpn`.
- Une configuration dans `/etc/zivpn/config.json`.
- Un fichier de mots de passe autorisés dans `/etc/zivpn/auth.config`.
- Un certificat TLS auto-signé dans `/etc/zivpn/zivpn.crt` et sa clé privée `/etc/zivpn/zivpn.key`.
- Un service systemd `zivpn.service`.

### 2.2. Analyse du binaire serveur
L'inspection des symboles et des chaînes de caractères du binaire `/usr/local/bin/zivpn` (version 1.5.0, compilé en Go 1.21, `quic-go v0.40`) a révélé :
1. La présence massive de paquets Go provenant de l'organisation `github.com/apernet` :
   - `github.com/apernet/hysteria/core/...`
   - `github.com/apernet/hysteria/extras/obfs` (dont `SalamanderObfuscator`)
   - `github.com/apernet/quic-go`
2. Des constantes de framing QUIC **renommées** par rapport à Hysteria :
   - `Hysteria-UDP` → **`Zivpnudp-UDP`**
   - `Hysteria-Auth` → **`Zivpnudp-Auth`**
   - `Hysteria-CC-RX` → **`Zivpnudp-CC-RX`**
   - `Hysteria-Padding` → **`Zivpnudp-Padding`**
3. Une vérification de signature codée en dur côté client : `config.Signature` doit valoir `"hu``hqb`c"`, sinon le binaire officiel s'arrête avec `FATAL wtf!!!No Idea`.

### 💡 Conclusion technique :
**ZiVPN 1.5.0 est un fork d'Hysteria v2 (génération v2.2.x)**, où :
- Le framing QUIC utilise les marqueurs `Zivpnudp-*` au lieu de `Hysteria-*` : un client Hysteria standard voit ses paquets **ignorés silencieusement** par le serveur.
- L'obfuscation est `salamander` (BLAKE2b-256, sel 8 octets), avec le mot de passe `hu``hqb`c` — extrait du client Android officiel (décompilé : `G.java` + XOR `A2.d.k` clé `ZCQV`), et confirmé par connexion réelle au serveur.
- L'authentification est le mot de passe seul (champ `auth`), sans nom d'utilisateur.
- Le port standard par défaut est `5667/udp`.

---

## 3. Spécification Détaillée du Protocole

Pour qu'un client puisse négocier avec succès une session avec le serveur ZiVPN, il doit satisfaire au contrat suivant :

```text
┌─────────────────────────────────────────────────────────────┐
│                       Couche Application                    │
│     Trafic IP brut (TUN) / Requêtes SOCKS5 (Port 1080)       │
├─────────────────────────────────────────────────────────────┤
│                       Couche Chiffrement                    │
│    TLS 1.3 (QUIC-TLS) avec négociation InsecureSkipVerify    │
├─────────────────────────────────────────────────────────────┤
│                       Couche Transport                      │
│      QUIC (Protocole UDP rapide avec multiplexage)          │
│      Framing "Zivpnudp-*" (Auth / UDP / CC-RX / Padding)     │
├─────────────────────────────────────────────────────────────┤
│                      Couche Obfuscation                     │
│    Salamander (BLAKE2b-256, mot de passe : "hu``hqb`c")       │
└─────────────────────────────────────────────────────────────┘
```

### Paramètres de Négociation :
| Paramètre | Valeur attendue | Rôle |
|---|---|---|
| **Protocole de transport** | `UDP` (QUIC) | Transport à faible latence résistant à la perte de paquets. |
| **Port par défaut** | `5667` | Port d'écoute du serveur. |
| **Framing** | `Zivpnudp-*` | Marqueurs QUIC renommés par le fork ZiVPN (voir §2.2). |
| **Obfuscation (`obfs`)** | `salamander`, mot de passe `"zivpn"` | Masque les en-têtes QUIC contre l'inspection DPI (Deep Packet Inspection). |
| **Authentification (`auth`)** | `<mot de passe>` | Doit correspondre à un mot de passe valide du serveur. |
| **Validation TLS (`tls.insecure`)** | `true` | Le certificat du serveur étant auto-signé, le client doit accepter le certificat sans vérification CA publique. |

---

## 4. Architecture de ZiVPN Desktop

L'application est structurée selon une séparation stricte entre **l'interface graphique** et le **moteur réseau**, connectés par une couche de configuration déclarative.

```text
┌───────────────────────────────────────────────────────────────┐
│                    Interface Graphique (GUI)                  │
│                     (Écrite en Go + Fyne)                     │
│                                                               │
│   [ Formulaire IP / Port / MDP ]   [ Bouton CONNECTER ]       │
│   [ Indicateur d'état dynamique ]  [ Console de logs ]        │
└───────────────────────────────┬───────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────┐
│              Contrôleur Interne (internal/engine)             │
│                                                               │
│   1. Charge / Sauvegarde le profil local (profile.json)       │
│   2. Valide les entrées utilisateur                           │
│   3. Génère dynamiquement le client.json (format Hysteria v2) │
│   4. Supervise le processus du moteur réseau                  │
│   5. Capture et diffuse les logs en temps réel                │
└───────────────────────────────┬───────────────────────────────┘
                                │ Lance en sous-processus
                                ▼
┌───────────────────────────────────────────────────────────────┐
│                 Moteur Réseau (zivpn-engine)                  │
│              (Hysteria v2.2.3 patché protocole ZiVPN)         │
│                                                               │
│   - Négocie le tunnel QUIC obfusqué vers AWS                  │
│   - Crée l'interface TUN système (zivpn-tun / wintun.dll)     │
│   - Ouvre le proxy local SOCKS5 (127.0.0.1:1080)              │
│   - Route l'ensemble du trafic IP de la machine               │
└───────────────────────────────┬───────────────────────────────┘
                                │ UDP chiffré (Port 5667)
                                ▼
┌───────────────────────────────────────────────────────────────┐
│                     Serveur AWS ZiVPN                         │
└───────────────────────────────────────────────────────────────┘
```

---

## 5. Pourquoi avoir choisi l'approche "Moteur découplé" ?

Lors du développement initial, deux stratégies ont été étudiées :

### Option A : Tout compiler dans un seul gros binaire Go
- **Obstacle majeur** : Hysteria v1 s'appuie sur une version spécifique de `quic-go` (`v0.34.x`). Les versions modernes de Go (Go 1.21, 1.22, 1.23+) ont profondément modifié les API cryptographiques internes (`crypto/tls`), rendant `quic-go v0.34` strictement incompilable sans altérer l'ensemble de la chaîne de compilation.
- Forcer une version obsolète de Go aurait fragilisé la maintenance future du projet.

### Option B : L'architecture Moteur Découplé (Retenue)
- L'interface utilisateur et la logique métier sont compilées avec le **Go moderne (1.23+)**, bénéficiant des dernières optimisations et correctifs de sécurité.
- Le moteur réseau `zivpn-engine` est le binaire officiel Hysteria **v2.2.3**, patché de façon reproductible (`scripts/patch-engine.py`) pour parler le protocole ZiVPN (framing `Zivpnudp-*`, obfs salamander).
- **Avantage décisif** : si le protocole ZiVPN évolue demain, il suffira de remplacer le binaire du moteur sans devoir réécrire l'interface utilisateur.

---

## 6. Structure du Code Source

```text
zivpn-desktop/
├── cmd/
│   └── zivpn-desktop/
│       └── main.go           # Point d'entrée : lance la GUI ou le mode CLI
├── internal/
│   ├── config/
│   │   ├── config.go         # Structures de données, validation, I/O profile.json & client.json
│   │   └── config_test.go    # Tests unitaires de validation
│   ├── engine/
│   │   └── engine.go         # Détection du binaire, lancement, capture des logs et arrêt
│   └── ui/
│       └── ui.go             # Interface graphique Fyne (fenêtre, boutons, état, logs)
├── ARCHITECTURE.md           # Ce document d'ingénierie inverse
├── README.md                 # Guide d'utilisation et de présentation
├── go.mod                    # Définition du module Go et dépendances
└── go.sum                    # Sommes de contrôle cryptographiques des paquets
```

---

## 7. Gestion Multiplateforme (Windows vs Linux)

| Élément | Linux | Windows |
|---|---|---|
| **Pilote TUN** | Périphérique natif Linux `/dev/net/tun` (`zivpn-tun`) | Pilote Wintun haute performance via `wintun.dll` |
| **Permissions** | Nécessite `sudo` ou `CAP_NET_ADMIN` pour monter l'interface | Nécessite le lancement en "Administrateur" |
| **Stockage config** | `~/.config/ZiVPN Desktop/` | `%APPDATA%\ZiVPN Desktop\` |
| **Moteur réseau** | `zivpn-engine` | `zivpn-engine.exe` |
