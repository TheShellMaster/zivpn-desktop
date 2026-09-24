package engine

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

type Process struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	mu     sync.Mutex
}

func FindBinary() (string, error) {
	name := "zivpn-engine"
	if runtime.GOOS == "windows" {
		name = "zivpn-engine.exe"
	}

	// 1. Dossier de l'exécutable courant
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), name)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, nil
		}
	}

	// 2. Dossier de travail actuel
	if fi, err := os.Stat(name); err == nil && !fi.IsDir() {
		abs, err := filepath.Abs(name)
		if err == nil {
			return abs, nil
		}
		return name, nil
	}

	// 3. Variable d'environnement PATH
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("moteur réseau '%s' introuvable dans le dossier de l'application", name)
}

func Start(ctx context.Context, binary, configPath string, onLog func(string), onExit func(error)) (*Process, error) {
	if binary == "" {
		var err error
		binary, err = FindBinary()
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(ctx)

	// Élévation des privilèges par OS — affiche un dialogue graphique natif :
	// • Linux  → pkexec (PolicyKit) : fenêtre "Authentification requise"
	// • macOS  → osascript          : fenêtre "Entrez votre mot de passe"
	// • Windows → manifest UAC      : fenêtre "Voulez-vous autoriser…" (au lancement)
	var cmd *exec.Cmd
	switch {
	case runtime.GOOS == "linux" && os.Geteuid() != 0:
		if pkexec, err := exec.LookPath("pkexec"); err == nil {
			cmd = exec.CommandContext(ctx, pkexec, binary, "--no-check", "-c", configPath)
		} else if sudoPath, err := exec.LookPath("sudo"); err == nil {
			cmd = exec.CommandContext(ctx, sudoPath, "-n", binary, "--no-check", "-c", configPath)
		} else {
			cmd = exec.CommandContext(ctx, binary, "--no-check", "-c", configPath)
		}
	case runtime.GOOS == "darwin" && os.Geteuid() != 0:
		// osascript demande le mot de passe via une fenêtre macOS native.
		script := fmt.Sprintf(
			`do shell script "%s --no-check -c %s" with administrator privileges`,
			binary, configPath,
		)
		cmd = exec.CommandContext(ctx, "osascript", "-e", script)
	default:
		cmd = exec.CommandContext(ctx, binary, "--no-check", "-c", configPath)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("capture stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("capture stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("démarrage du moteur ZiVPN: %w", err)
	}

	proc := &Process{
		cmd:    cmd,
		cancel: cancel,
	}

	go streamLogs(stdout, onLog)
	go streamLogs(stderr, onLog)

	go func() {
		waitErr := cmd.Wait()
		if onExit != nil {
			onExit(waitErr)
		}
	}()

	return proc, nil
}

func streamLogs(r io.Reader, callback func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if callback != nil {
			callback(line)
		}
	}
}

func (p *Process) Stop() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	return nil
}
