//go:build !windows

package browser

func findBrowser(string) string {
	return ""
}
