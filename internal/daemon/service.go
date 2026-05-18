package daemon

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hanifanggawi/vessel/internal/config"
)

// serviceName is the unit/task name registered with the OS service manager.
const serviceName = "vessel"

// resolveExe returns the absolute path of the running vessel binary, which
// the service definition invokes at boot.
func resolveExe() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not resolve executable path: %w", err)
	}
	return filepath.EvalSymlinks(exe)
}

// ensureConfig guarantees the per-user config file exists before the service
// is enabled, so the boot-time daemon never starts against a missing file.
//
// The path is config.DomainsConfigPath, which Init() has already resolved for
// the *invoking* user (honoring SUDO_USER). Because the service runs as root
// while the user edits config as themselves, the file must stay owned by the
// invoking user — chownToInvoker handles that on Unix (no-op on Windows).
func ensureConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	if err := os.WriteFile(path, []byte("[rules]\n"), 0o644); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}
	chownToInvoker(dir)
	chownToInvoker(path)
	return nil
}

// Install registers vessel with the OS service manager so the reconcile
// daemon runs at startup. The resolved config path is baked into the service
// definition, guaranteeing the root daemon and `sudo vessel` / `vessel`
// always read the same per-user config file.
func Install() error {
	exe, err := resolveExe()
	if err != nil {
		return err
	}

	configPath := config.DomainsConfigPath
	if configPath == "" {
		return fmt.Errorf("config path is not resolved; cannot install service")
	}
	if err := ensureConfig(configPath); err != nil {
		return err
	}

	return osInstall(exe, configPath)
}

// Uninstall stops and removes the OS service definition. It is best-effort:
// a missing service is not an error.
func Uninstall() error {
	return osUninstall()
}
