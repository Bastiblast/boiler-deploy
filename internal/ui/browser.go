package ui

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// OpenBrowser opens the specified URL in the default browser
func OpenBrowser(url string) error {
	log.Printf("[BROWSER] Attempting to open URL: %s", url)
	log.Printf("[BROWSER] Detected OS: %s", runtime.GOOS)
	
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		// Try xdg-open first (most common)
		if _, err := exec.LookPath("xdg-open"); err == nil {
			log.Printf("[BROWSER] Using xdg-open")
			cmd = exec.Command("xdg-open", url)
		} else if _, err := exec.LookPath("gnome-open"); err == nil {
			log.Printf("[BROWSER] Using gnome-open")
			cmd = exec.Command("gnome-open", url)
		} else if _, err := exec.LookPath("wslview"); err == nil {
			log.Printf("[BROWSER] Using wslview (WSL)")
			cmd = exec.Command("wslview", url)
		} else {
			log.Printf("[BROWSER] ERROR: No browser opener found")
			return fmt.Errorf("no browser opener found (xdg-open, gnome-open, wslview)")
		}
	case "darwin":
		log.Printf("[BROWSER] Using open (macOS)")
		cmd = exec.Command("open", url)
	case "windows":
		log.Printf("[BROWSER] Using rundll32 (Windows)")
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		log.Printf("[BROWSER] ERROR: Unsupported platform: %s", runtime.GOOS)
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	err := cmd.Start()
	if err != nil {
		log.Printf("[BROWSER] ERROR: Failed to start command: %v", err)
		return err
	}
	
	log.Printf("[BROWSER] Successfully started browser command")
	return nil
}
