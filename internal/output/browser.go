package output

import (
	"os/exec"
	"runtime"
)

func OpenURL(u string) {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", u)
	case "linux":
		c = exec.Command("xdg-open", u)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	}
	if c != nil {
		_ = c.Start()
	}
}
