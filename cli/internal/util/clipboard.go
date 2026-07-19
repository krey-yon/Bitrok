package util

import (
	"os/exec"
	"runtime"
)

// CopyToClipboard copies s to the system clipboard by shelling out to the
// platform tool. Best-effort: returns an error if no tool is available, callers
// print a hint instead of failing. ponytail: no clipboard dep — pbcopy/xclip
// are already on every dev machine; add a Go clipboard lib only if this bites.
func CopyToClipboard(s string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default: // linux/bsd
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			return errNoClipboard
		}
	}
	c, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := c.Write([]byte(s)); err != nil {
		return err
	}
	c.Close()
	return cmd.Wait()
}

var errNoClipboard = &clipboardErr{"no clipboard tool found (install xclip or wl-copy)"}

type clipboardErr struct{ msg string }

func (e *clipboardErr) Error() string { return e.msg }
