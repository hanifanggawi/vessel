//go:build windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const taskName = serviceName

// chownToInvoker is a no-op on Windows: there is no sudo, UAC elevation keeps
// the same user profile, so the per-user config path stays consistent.
func chownToInvoker(string) {}

// launcherPath is a stable .cmd shim under ProgramData. The scheduled task
// runs as SYSTEM, whose %AppData% differs from the user's, so the resolved
// per-user config path is baked into this shim rather than left to the
// SYSTEM account's defaults. Pointing schtasks at a file also avoids fragile
// nested-quote escaping in /TR.
func launcherPath() string {
	base := os.Getenv("ProgramData")
	if base == "" {
		base = os.Getenv("SystemDrive") + `\ProgramData`
	}
	return filepath.Join(base, "vessel", "vessel-daemon.cmd")
}

func writeLauncher(exe, configPath string) (string, error) {
	path := launcherPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("could not create launcher directory: %w", err)
	}
	contents := fmt.Sprintf("@echo off\r\nset \"VESSEL_DOMAINS_CONFIG_PATH=%s\"\r\n\"%s\" daemon run\r\n", configPath, exe)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return "", fmt.Errorf("could not write launcher %s: %w", path, err)
	}
	return path, nil
}

func osInstall(exe, configPath string) error {
	launcher, err := writeLauncher(exe, configPath)
	if err != nil {
		return err
	}

	args := []string{
		"/Create",
		"/TN", taskName,
		"/SC", "ONSTART",
		"/RU", "SYSTEM",
		"/RL", "HIGHEST",
		"/F",
		"/TR", launcher,
	}
	if err := run("schtasks", args...); err != nil {
		return fmt.Errorf("%w (run this in an Administrator prompt)", err)
	}

	fmt.Printf("Installed scheduled task %q (config: %s)\n", taskName, configPath)
	fmt.Printf("Manage it with: schtasks /Query /TN %s\n", taskName)
	return nil
}

func osUninstall() error {
	if err := run("schtasks", "/Delete", "/TN", taskName, "/F"); err != nil {
		return fmt.Errorf("%w (run this in an Administrator prompt)", err)
	}
	if err := os.Remove(launcherPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove launcher %s: %w", launcherPath(), err)
	}
	fmt.Printf("Removed scheduled task %q\n", taskName)
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v failed: %w", name, args, err)
	}
	return nil
}
