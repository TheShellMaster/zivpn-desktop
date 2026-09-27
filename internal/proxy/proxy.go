package proxy

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Enable activates the system-wide proxy so all applications and browsers
// automatically route their traffic through the local ZiVPN proxy.
func Enable(port int) error {
	switch runtime.GOOS {
	case "linux":
		enableLinuxGnome(port)
		enableLinuxKDE(port)
	case "windows":
		// Windows Internet Settings registry keys
		_ = exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "1", "/f").Run()
		_ = exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`, "/v", "ProxyServer", "/t", "REG_SZ", "/d", "socks=127.0.0.1:1080", "/f").Run()
	case "darwin":
		// macOS network services
		out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, "*") {
					_ = exec.Command("networksetup", "-setsocksfirewallproxy", line, "127.0.0.1", "1080").Run()
				}
			}
		}
	}
	return nil
}

// Disable restores direct internet access by resetting system proxy settings.
func Disable() error {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("gsettings"); err == nil {
			_ = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
		}
		disableLinuxKDE()
	case "windows":
		_ = exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`, "/v", "ProxyEnable", "/t", "REG_DWORD", "/d", "0", "/f").Run()
	case "darwin":
		out, err := exec.Command("networksetup", "-listallnetworkservices").Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.Contains(line, "*") {
					_ = exec.Command("networksetup", "-setsocksfirewallproxystate", line, "off").Run()
				}
			}
		}
	}
	return nil
}

// enableLinuxGnome configure le proxy via GSettings (GNOME, Ubuntu, Cinnamon…).
func enableLinuxGnome(port int) {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return
	}
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "host", "127.0.0.1").Run()
	_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "port", strconv.Itoa(port)).Run()
}

// enableLinuxKDE configure le proxy via kioslaverc (KDE Plasma).
// Vérifié : kwriteconfig5 écrit ~/.config/kioslaverc utilisé par les
// applications KDE. Chaque appel est best-effort (erreurs ignorées).
func enableLinuxKDE(port int) {
	if _, err := exec.LookPath("kwriteconfig5"); err != nil {
		return
	}
	_ = exec.Command("kwriteconfig5",
		"--file", "kioslaverc",
		"--group", "Proxy Settings",
		"--key", "ProxyType", "1").Run()
	_ = exec.Command("kwriteconfig5",
		"--file", "kioslaverc",
		"--group", "Proxy Settings",
		"--key", "socksProxy",
		"socks://127.0.0.1:"+strconv.Itoa(port)).Run()
	// Demande aux esclaves KIO de relire la configuration.
	_ = exec.Command("dbus-send", "--session",
		"--type=method_call", "--dest=org.kde.KIO.Scheduler",
		"/Scheduler", "org.kde.KIO.Scheduler.reparseSlaveConfiguration",
		"string:").Run()
}

// disableLinuxKDE restaure l'accès direct (ProxyType=0) sous KDE Plasma.
func disableLinuxKDE() {
	if _, err := exec.LookPath("kwriteconfig5"); err != nil {
		return
	}
	_ = exec.Command("kwriteconfig5",
		"--file", "kioslaverc",
		"--group", "Proxy Settings",
		"--key", "ProxyType", "0").Run()
	_ = exec.Command("dbus-send", "--session",
		"--type=method_call", "--dest=org.kde.KIO.Scheduler",
		"/Scheduler", "org.kde.KIO.Scheduler.reparseSlaveConfiguration",
		"string:").Run()
}
