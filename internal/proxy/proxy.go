package proxy

import (
	"os/exec"
	"runtime"
	"strings"
)

// Enable activates the system-wide proxy so all applications and browsers
// automatically route their traffic through the local ZiVPN proxy.
func Enable(port int) error {
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
		_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "host", "127.0.0.1").Run()
		_ = exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "port", "1080").Run()
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
		_ = exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
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
