package browser

import (
	"os/exec"
	"runtime"
)

// Open は指定された URL をデフォルトブラウザで開く
func Open(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		// 未知の OS では xdg-open を試す
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
