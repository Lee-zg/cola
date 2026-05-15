package browser

import (
	"os"
	"path/filepath"
)

func findBrowser(name string) string {
	local := os.Getenv("LOCALAPPDATA")
	programFiles := os.Getenv("PROGRAMFILES")
	programFilesX86 := os.Getenv("PROGRAMFILES(X86)")
	candidates := map[string][]string{
		"chrome": {
			filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(local, "Google", "Chrome", "Application", "chrome.exe"),
		},
		"edge": {
			filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(programFilesX86, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(local, "Microsoft", "Edge", "Application", "msedge.exe"),
		},
		"firefox": {
			filepath.Join(programFiles, "Mozilla Firefox", "firefox.exe"),
			filepath.Join(programFilesX86, "Mozilla Firefox", "firefox.exe"),
			filepath.Join(local, "Mozilla Firefox", "firefox.exe"),
		},
	}
	for _, path := range candidates[name] {
		if fileExists(path) {
			return path
		}
	}
	return ""
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}
