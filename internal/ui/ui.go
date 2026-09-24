package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/TheShellMaster/zivpn-desktop/internal/config"
	"github.com/TheShellMaster/zivpn-desktop/internal/engine"
)

type AppUI struct {
	app     fyne.App
	window  fyne.Window
	configDir string

	entryServer   *widget.Entry
	entryPort     *widget.Entry
	entryPassword *widget.Entry

	lblStatus  *widget.Label
	btnAction  *widget.Button
	logView    *widget.Entry
	logScroll  *container.Scroll

	proc   *engine.Process
	mu     sync.Mutex
	status string
}

func Run() {
	a := app.NewWithID("com.zivpn.desktop")
	w := a.NewWindow("ZiVPN Desktop")
	w.Resize(fyne.NewSize(380, 520))
	w.SetFixedSize(false)

	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	appDir := filepath.Join(configDir, "ZiVPN Desktop")

	ui := &AppUI{
		app:       a,
		window:    w,
		configDir: appDir,
		status:    "disconnected",
	}

	ui.build()
	ui.loadProfile()

	w.SetOnClosed(func() {
		ui.disconnect()
	})

	CheckForUpdates(w, true)

	w.ShowAndRun()
}

func (u *AppUI) build() {
	// En-tête
	lblTitle := widget.NewLabelWithStyle("ZiVPN Desktop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	lblSub := widget.NewLabelWithStyle("Client VPN UDP", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// Formulaire
	u.entryServer = widget.NewEntry()
	u.entryServer.SetPlaceHolder("ex: 198.51.100.1")

	u.entryPort = widget.NewEntry()
	u.entryPort.SetText("5667")

	u.entryPassword = widget.NewPasswordEntry()
	u.entryPassword.SetPlaceHolder("Mot de passe")

	form := container.NewVBox(
		widget.NewLabel("Adresse IP du serveur :"),
		u.entryServer,
		widget.NewLabel("Port UDP :"),
		u.entryPort,
		widget.NewLabel("Mot de passe :"),
		u.entryPassword,
	)

	// Statut
	u.lblStatus = widget.NewLabelWithStyle("⚪ Déconnecté", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Bouton de connexion
	u.btnAction = widget.NewButtonWithIcon("CONNECTER", theme.MediaPlayIcon(), func() {
		u.toggleConnection()
	})
	u.btnAction.Importance = widget.HighImportance

	// Console de logs rétractable
	u.logView = widget.NewMultiLineEntry()
	u.logView.Disable()
	u.logScroll = container.NewScroll(u.logView)
	u.logScroll.SetMinSize(fyne.NewSize(340, 130))

	logsAccordion := widget.NewAccordion(
		widget.NewAccordionItem("Afficher les journaux d'activité", u.logScroll),
	)

	btnCheckUpdate := widget.NewButtonWithIcon("Vérifier les mises à jour", theme.ViewRefreshIcon(), func() {
		CheckForUpdates(u.window, false)
	})

	footer := container.NewHBox(
		widget.NewLabel("Version "+CurrentVersion),
		btnCheckUpdate,
	)

	content := container.NewVBox(
		lblTitle,
		lblSub,
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		u.lblStatus,
		u.btnAction,
		widget.NewSeparator(),
		logsAccordion,
		widget.NewSeparator(),
		footer,
	)

	u.window.SetContent(container.NewPadded(content))
}

func (u *AppUI) loadProfile() {
	prof, err := config.LoadProfile(u.configDir)
	if err == nil {
		if prof.Server != "" {
			u.entryServer.SetText(prof.Server)
		}
		if prof.Port > 0 {
			u.entryPort.SetText(strconv.Itoa(prof.Port))
		}
		if prof.Password != "" {
			u.entryPassword.SetText(prof.Password)
		}
	}
}

func (u *AppUI) toggleConnection() {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.status == "connected" || u.status == "connecting" {
		u.disconnectLocked()
	} else {
		u.connectLocked()
	}
}

func (u *AppUI) connectLocked() {
	server := strings.TrimSpace(u.entryServer.Text)
	portStr := strings.TrimSpace(u.entryPort.Text)
	pass := strings.TrimSpace(u.entryPassword.Text)

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		dialog.ShowError(fmt.Errorf("le port doit être un nombre valide entre 1 et 65535"), u.window)
		return
	}

	conn := config.Connection{
		Server:   server,
		Port:     port,
		Password: pass,
	}

	if err := conn.Validate(); err != nil {
		dialog.ShowError(err, u.window)
		return
	}

	// Sauvegarde du profil
	_ = conn.SaveProfile(u.configDir)

	// Génération de client.json
	configPath, err := conn.WriteEngineConfig(u.configDir)
	if err != nil {
		dialog.ShowError(fmt.Errorf("erreur de configuration: %w", err), u.window)
		return
	}

	u.setStatus("connecting", "🟡 Connexion en cours...")
	u.logView.SetText("")
	u.appendLog(fmt.Sprintf("Initialisation du tunnel vers %s:%d...\n", server, port))

	ctx := context.Background()
	proc, err := engine.Start(ctx, "", configPath, func(line string) {
		u.appendLog(line + "\n")
		if strings.Contains(line, "TUN up and running") || strings.Contains(line, "SOCKS5 server up and running") {
			u.setStatus("connected", fmt.Sprintf("🟢 Connecté à %s:%d", server, port))
		}
	}, func(exitErr error) {
		u.mu.Lock()
		defer u.mu.Unlock()
		if u.status == "connected" || u.status == "connecting" {
			if exitErr != nil {
				u.setStatus("error", "🔴 Connexion interrompue")
				u.appendLog(fmt.Sprintf("Arrêt du moteur: %v\n", exitErr))
			} else {
				u.setStatus("disconnected", "⚪ Déconnecté")
			}
			u.proc = nil
		}
	})

	if err != nil {
		u.setStatus("error", "🔴 Échec du démarrage")
		dialog.ShowError(fmt.Errorf("impossible de démarrer le moteur réseau: %w", err), u.window)
		return
	}

	u.proc = proc
	// Par défaut, si le moteur tourne sans erreur immédiate, on marque comme connecté
	u.setStatus("connected", fmt.Sprintf("🟢 Connecté à %s:%d", server, port))
}

func (u *AppUI) disconnect() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.disconnectLocked()
}

func (u *AppUI) disconnectLocked() {
	if u.proc != nil {
		_ = u.proc.Stop()
		u.proc = nil
	}
	u.setStatus("disconnected", "⚪ Déconnecté")
	u.appendLog("Déconnexion effectuée.\n")
}

func (u *AppUI) setStatus(state, label string) {
	u.status = state
	u.lblStatus.SetText(label)

	switch state {
	case "connected":
		u.btnAction.SetText("DÉCONNECTER")
		u.btnAction.SetIcon(theme.MediaStopIcon())
		u.btnAction.Importance = widget.DangerImportance
		u.entryServer.Disable()
		u.entryPort.Disable()
		u.entryPassword.Disable()
	case "connecting":
		u.btnAction.SetText("ANNULER")
		u.btnAction.SetIcon(theme.CancelIcon())
		u.btnAction.Importance = widget.WarningImportance
	default:
		u.btnAction.SetText("CONNECTER")
		u.btnAction.SetIcon(theme.MediaPlayIcon())
		u.btnAction.Importance = widget.HighImportance
		u.entryServer.Enable()
		u.entryPort.Enable()
		u.entryPassword.Enable()
	}
	u.btnAction.Refresh()
}

func (u *AppUI) appendLog(text string) {
	u.logView.SetText(u.logView.Text + text)
	u.logScroll.ScrollToBottom()
}
