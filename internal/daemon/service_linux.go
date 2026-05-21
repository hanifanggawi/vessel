//go:build linux

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

const unitPath = "/etc/systemd/system/" + serviceName + ".service"

// chownToInvoker hands a root-created path back to the user who invoked
// `sudo vessel`, so the per-user config stays user-writable. SUDO_UID/GID are
// set by sudo; absent them (already running as the real user) it is a no-op.
func chownToInvoker(path string) {
	uidStr := os.Getenv("SUDO_UID")
	gidStr := os.Getenv("SUDO_GID")
	if uidStr == "" || gidStr == "" {
		return
	}
	uid, err1 := strconv.Atoi(uidStr)
	gid, err2 := strconv.Atoi(gidStr)
	if err1 != nil || err2 != nil {
		return
	}
	_ = os.Chown(path, uid, gid)
}

func unitContents(exe, configPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Vessel hosts-file reconciliation daemon
After=network.target

[Service]
Type=simple
ExecStart=%s daemon run
Environment=VESSEL_DOMAINS_CONFIG_PATH=%s
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, exe, configPath)
}

func osInstall(exe, configPath string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("installing the system service requires root: re-run with sudo")
	}

	restoreSELinuxContext(exe)

	if err := os.WriteFile(unitPath, []byte(unitContents(exe, configPath)), 0o644); err != nil {
		return fmt.Errorf("could not write unit file %s: %w", unitPath, err)
	}

	if err := run("systemctl", "daemon-reload"); err != nil {
		return err
	}
	if err := run("systemctl", "enable", "--now", serviceName+".service"); err != nil {
		return err
	}

	fmt.Printf("Installed and started %s.service (config: %s)\n", serviceName, configPath)
	fmt.Printf("Manage it with: systemctl status %s\n", serviceName)
	return nil
}

func osUninstall() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("removing the system service requires root: re-run with sudo")
	}

	// Best-effort: ignore errors so a partially-installed service can still
	// be cleaned up.
	_ = run("systemctl", "disable", "--now", serviceName+".service")

	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not remove unit file %s: %w", unitPath, err)
	}
	_ = run("systemctl", "daemon-reload")

	fmt.Printf("Removed %s.service\n", serviceName)
	return nil
}

// restoreSELinuxContext resets the SELinux label on the vessel binary to the
// system default for its path (typically bin_t under /usr/local/bin). Without
// this, a binary installed from a FUSE-mounted source (ntfs-3g, sshfs, etc.)
// keeps a fusefs_t label, which systemd's init_t domain cannot execute, so the
// unit fails to start at boot with an opaque AVC denial.
//
// Best-effort: no SELinux, no restorecon, or a relabel failure is not fatal.
func restoreSELinuxContext(path string) {
	if _, err := os.Stat("/sys/fs/selinux/enforce"); err != nil {
		return
	}
	restorecon, err := exec.LookPath("restorecon")
	if err != nil {
		return
	}
	cmd := exec.Command(restorecon, "-F", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not restore SELinux context on %s: %v\n", path, err)
	}
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
