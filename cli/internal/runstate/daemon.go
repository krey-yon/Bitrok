package runstate

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// DetachEnv is set on the child process so it knows it is already daemonized.
const DetachEnv = "BITROK_DETACHED"

// IsDetached reports whether the current process is a background worker.
func IsDetached() bool {
	return os.Getenv(DetachEnv) == "1"
}

// Detach re-execs the current binary without -d/--detach, with stdout/stderr
// redirected to a log file and a new session (setsid). Parent returns the
// child's PID; the child continues as a normal tunnel process.
//
// name is used for the log file path only.
func Detach(name string, argv []string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("resolve executable: %w", err)
	}

	// Drop detach flags so the child does not re-detach.
	childArgs := filterDetachFlags(argv)

	if err := os.MkdirAll(RunDir(), 0700); err != nil {
		return 0, err
	}
	logPath := LogPath(name)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, fmt.Errorf("open log: %w", err)
	}

	cmd := exec.Command(exe, childArgs...)
	cmd.Env = append(os.Environ(), DetachEnv+"=1")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // new session — survives parent exit
	}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return 0, fmt.Errorf("detach: %w", err)
	}
	// Parent must not wait; close our handle to the log (child still has it).
	logFile.Close()
	return cmd.Process.Pid, nil
}

func filterDetachFlags(argv []string) []string {
	out := make([]string, 0, len(argv))
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch a {
		case "-d", "--detach", "--daemon":
			continue
		default:
			// --detach=true style
			if strings.HasPrefix(a, "--detach=") || strings.HasPrefix(a, "--daemon=") {
				continue
			}
			out = append(out, a)
		}
	}
	return out
}

// SelfArgv returns os.Args[1:] suitable for Detach.
func SelfArgv() []string {
	if len(os.Args) < 2 {
		return nil
	}
	return append([]string{}, os.Args[1:]...)
}

// AbsLogPath resolves a log path for display.
func AbsLogPath(name string) string {
	p, err := filepath.Abs(LogPath(name))
	if err != nil {
		return LogPath(name)
	}
	return p
}
