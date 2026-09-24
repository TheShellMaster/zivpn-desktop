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
L'inspection des symboles et des chaînes de caractères du binaire `/usr/local/bin/zivpn` a révélé :
1. La présence massive de paquets Go provenant de l'organisation `github.com/apernet` :
   - `github.com/apernet/hysteria/core/cs`
   - `github.com/apernet/hysteria/app/cmd`
   - `github.com/apernet/quic-go`
2. Les messages de logs d'erreurs et les formats de handshake sont identiques à ceux de **Hysteria version 1.3.x**.

### 💡 Conclusion technique :
**ZiVPN est un renommage (rebranding) d'Hysteria v1**, où :
- L'algorithme d'obfuscation par défaut a été configuré avec le mot-clé statique `"zivpn"`.
- L'authentification a été restreinte au mode mot de passe seul (`auth_str`), sans nom d'utilisateur.
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
├─────────────────────────────────────────────────────────────┤
│                      Couche Obfuscation                     │
│    XOR Packet Obfuscator (Clé statique : "zivpn")            │
└─────────────────────────────────────────────────────────────┘
```

### Paramètres de Négociation :
| Paramètre | Valeur attendue | Rôle |
|---|---|---|
| **Protocole de transport** | `UDP` (QUIC) | Transport à faible latence résistant à la perte de paquets. |
| **Port par défaut** | `5667` | Port d'écoute du serveur. |
| **Obfuscation (`obfs`)** | `"zivpn"` | Masque les en-têtes QUIC contre l'inspection DPI (Deep Packet Inspection). |
| **Authentification (`auth_str`)** | `<mot de passe>` | Doit correspondre à une ligne valide du fichier serveur `auth.config`. |
| **Bande passante (`up_mbps` / `down_mbps`)** | Valeur entière (ex: `100`) | Hysteria v1 exige ces valeurs pour initialiser son algorithme de contrôle de congestion brutale. |
| **Validation TLS (`insecure`)** | `true` | Le certificat du serveur étant auto-signé, le client doit accepter le certificat sans vérification CA publique. |

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
│   3. Génère dynamiquement le client.json (format Hysteria v1) │
│   4. Supervise le processus du moteur réseau                  │
│   5. Capture et diffuse les logs en temps réel                │
└───────────────────────────────┬───────────────────────────────┘
                                │ Lance en sous-processus
                                ▼
┌───────────────────────────────────────────────────────────────┐
│                 Moteur Réseau (zivpn-engine)                  │
│                    (Binaire Hysteria v1)                      │
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
- Le moteur réseau `zivpn-engine` utilise le binaire stable officiel v1.3.5, garanti sans bug de compatibilité cryptographique.
- **Avantage décisif** : Si le protocole ZiVPN évolue demain (par exemple vers Hysteria v2), il suffira de remplacer le binaire du moteur sans devoir réécrire l'interface utilisateur.

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
