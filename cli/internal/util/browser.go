package util

import (
	"os/exec"
	"runtime"
)

// OpenBrowser opens url in the user's default browser.
func OpenBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	_ = exec.Command(cmd, args...).Start()
}
