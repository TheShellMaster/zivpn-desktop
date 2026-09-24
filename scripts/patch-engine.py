#!/usr/bin/env python3
"""Patche un binaire Hysteria v2.2.3 officiel pour parler le protocole ZiVPN.

Contexte (vérifié par analyse du binaire /usr/local/bin/zivpn v1.5.0) :
- ZiVPN = Hysteria v2 avec les constantes de framing renommées :
    Hysteria-UDP     -> Zivpnudp-UDP
    Hysteria-Auth    -> Zivpnudp-Auth
    Hysteria-CC-RX   -> Zivpnudp-CC-RX
    Hysteria-Padding -> Zivpnudp-Padding
- Obfuscation : salamander (BLAKE2b), mot de passe "zivpn"
  (le serveur utilise `"obfs": "zivpn"` dans /etc/zivpn/config.json).
- Authentification : mot de passe seul (champ "auth").

Usage :
    python3 scripts/patch-engine.py <binaire-entree> <binaire-sortie>
"""
import re
import sys


def main() -> None:
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <binaire-entree> <binaire-sortie>")
        sys.exit(2)
    src, dst = sys.argv[1], sys.argv[2]

    with open(src, "rb") as f:
        data = bytearray(f.read())

    count = 0
    targets = [b"Hysteria-UDP", b"Hysteria-Auth", b"Hysteria-CC-RX", b"Hysteria-Padding"]
    for target in targets:
        replacement = target.replace(b"Hysteria-", b"Zivpnudp-")
        idx = 0
        while True:
            pos = data.find(target, idx)
            if pos == -1:
                break
            data[pos : pos + len(target)] = replacement
            count += 1
            idx = pos + len(target)
            print(f"Patched {target.decode()} -> {replacement.decode()} at {hex(pos)}")

    # ALPN "hysteria" -> "zivpnudp" (hors chemins de debug Go : app/cmd,
    # github.com, Aperture). Les chemins "m/apernet/hysteria/..." restants
    # sont de simples infos de debug (pclntab) sans effet à l'exécution,
    # mais on les laisse intacts quand le contexte l'indique.
    for m in re.finditer(b"hysteria", data):
        pos = m.start()
        context = data[pos - 10 : pos + 20]
        if b"app/cmd" not in context and b"github.com" not in context and b"Aperture" not in context:
            data[pos : pos + 8] = b"zivpnudp"
            count += 1

    with open(dst, "wb") as f:
        f.write(data)

    print(f"Total patches: {count}")
    if count < 4:
        print("ERREUR: les 4 constantes de framing sont introuvables, patch incomplet.")
        sys.exit(1)


if __name__ == "__main__":
    main()
