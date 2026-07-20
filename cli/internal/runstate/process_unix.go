//go:build !windows

package runstate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func configureDetachedCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func processAlive(proc *os.Process) bool {
	return proc.Signal(syscall.Signal(0)) == nil
}

func requestProcessStop(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

func forceProcessStop(proc *os.Process) error {
	return proc.Signal(syscall.SIGKILL)
}

func processMatches(pid int, expectedExecutable string) bool {
	expected := filepath.Base(expectedExecutable)
	if expected == "" || expected == "." {
		return false
	}
	if target, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "exe")); err == nil {
		return filepath.Base(target) == expected
	}
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=").Output()
	if err != nil {
		return false
	}
	return filepath.Base(strings.TrimSpace(string(output))) == expected
}
