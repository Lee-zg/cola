// browser 包根据用户偏好打开书签链接；找不到指定浏览器时回退系统默认浏览器。
package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenURL(rawURL, preferred string) error {
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	if preferred != "" && preferred != "default" {
		if path := findBrowser(preferred); path != "" {
			return exec.Command(path, rawURL).Start()
		}
	}
	return openDefault(rawURL)
}

func openDefault(rawURL string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL).Start()
	case "darwin":
		return exec.Command("open", rawURL).Start()
	default:
		return exec.Command("xdg-open", rawURL).Start()
	}
}
