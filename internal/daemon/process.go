package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func stateDir() (string, error) {
	base, err := os.UserCacheDir() // ~/.cache on Linux, %AppData%\Local on Windows
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "vessel")
	return dir, os.MkdirAll(dir, 0o700)
}

func pidFilePath() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vessel.pid"), nil
}

func logFilePath() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vessel.log"), nil
}

func writePID(pid int) error {
	path, err := pidFilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o600)
}

func readPID() (int, error) {
	path, err := pidFilePath()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func removePID() {
	path, _ := pidFilePath()
	os.Remove(path)
}

func Start() error {
	if pid, err := readPID(); err == nil {
		if alive, _ := isRunning(pid); alive {
			return fmt.Errorf("daemon is already running (PID %d)", pid)
		}
		removePID() // stale PID file
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %w", err)
	}

	logPath, err := logFilePath()
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}
	defer logFile.Close()

	cmd := exec.Command(exe, "daemon", "run")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	detachProcess(cmd) // OS-specific (see below)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	if err := writePID(cmd.Process.Pid); err != nil {
		return fmt.Errorf("daemon started (PID %d) but could not write PID file: %w", cmd.Process.Pid, err)
	}

	fmt.Printf("daemon started (PID %d), logging to %s\n", cmd.Process.Pid, logPath)
	return nil
}

func Stop() error {
	pid, err := readPID()
	if err != nil {
		return fmt.Errorf("daemon is not running (no PID file)")
	}

	alive, err := isRunning(pid)
	if err != nil || !alive {
		removePID()
		return fmt.Errorf("daemon is not running (stale PID %d)", pid)
	}

	if err := killProcess(pid); err != nil { // OS-specific
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	removePID()
	fmt.Printf("daemon stopped (PID %d)\n", pid)
	return nil
}

func Restart() error {
	pid, err := readPID()
	if err != nil {
		// Nothing running, just start fresh
		return Start()
	}

	if err := Stop(); err != nil {
		return fmt.Errorf("restart: stop failed: %w", err)
	}

	if err := waitForExit(pid, 5*time.Second); err != nil {
		return fmt.Errorf("restart: old daemon did not exit: %w", err)
	}

	return Start()
}

func waitForExit(pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if alive, _ := isRunning(pid); !alive {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for PID %d to exit", pid)
}

func Status() (pid int, running bool, err error) {
	pid, err = readPID()
	if err != nil {
		return 0, false, nil // no PID file = not running, not an error
	}
	running, err = isRunning(pid)
	return pid, running, err
}
