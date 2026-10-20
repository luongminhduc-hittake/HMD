package util

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// DetectedBrowser contains metadata for an installed browser profile found on the host.
type DetectedBrowser struct {
	ID        string    // Unique key: "zen", "chrome", "firefox", "brave", "edge", "opera", "vivaldi", "librewolf", "floorp"
	Name      string    // User-friendly display: "Zen Browser", "Google Chrome", etc.
	CookieArg string    // Argument for yt-dlp/gallery-dl, e.g. "firefox:/path/to/profile" or "chrome"
	LastUsed  time.Time // Latest cookie database modification time
}

// findGeckoProfiles searches a base directory for profiles containing cookies.sqlite.
func findGeckoProfiles(baseDir, id, name string) []DetectedBrowser {
	if baseDir == "" {
		return nil
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	var results []DetectedBrowser
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		profileDir := filepath.Join(baseDir, entry.Name())
		cookiePath := filepath.Join(profileDir, "cookies.sqlite")
		if fi, err := os.Stat(cookiePath); err == nil && !fi.IsDir() {
			results = append(results, DetectedBrowser{
				ID:        id,
				Name:      name,
				CookieArg: "firefox:" + profileDir,
				LastUsed:  fi.ModTime(),
			})
		}
	}
	return results
}

// checkChromiumBrowser checks if a Chromium-based browser exists and gets its latest cookie mod time.
func checkChromiumBrowser(userDir, id, name, cliName string) *DetectedBrowser {
	if userDir == "" {
		return nil
	}
	if _, err := os.Stat(userDir); err != nil {
		return nil
	}

	// Find Default or Profile 1 cookies file
	cookieCandidates := []string{
		filepath.Join(userDir, "Default", "Cookies"),
		filepath.Join(userDir, "Default", "Network", "Cookies"),
		filepath.Join(userDir, "Profile 1", "Cookies"),
		filepath.Join(userDir, "Profile 1", "Network", "Cookies"),
	}

	var latestTime time.Time
	found := false
	for _, cand := range cookieCandidates {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			found = true
			if fi.ModTime().After(latestTime) {
				latestTime = fi.ModTime()
			}
		}
	}

	if !found {
		// Even if cookie file not directly visible, userDir exists
		if fi, err := os.Stat(userDir); err == nil {
			latestTime = fi.ModTime()
			found = true
		}
	}

	if found {
		return &DetectedBrowser{
			ID:        id,
			Name:      name,
			CookieArg: cliName,
			LastUsed:  latestTime,
		}
	}
	return nil
}

// ScanInstalledBrowsers discovers all installed web browsers on the machine,
// ordered by latest cookie usage time.
func ScanInstalledBrowsers() []DetectedBrowser {
	home, _ := os.UserHomeDir()
	var candidates []DetectedBrowser

	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		localAppData := os.Getenv("LOCALAPPDATA")

		// Zen Browser
		candidates = append(candidates, findGeckoProfiles(filepath.Join(appData, "zen", "Profiles"), "zen", "Zen Browser")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(appData, "Zen", "Profiles"), "zen", "Zen Browser")...)

		// Floorp & LibreWolf & Waterfox
		candidates = append(candidates, findGeckoProfiles(filepath.Join(appData, "Floorp", "Profiles"), "floorp", "Floorp")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(appData, "librewolf", "Profiles"), "librewolf", "LibreWolf")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(appData, "Waterfox", "Profiles"), "waterfox", "Waterfox")...)

		// Firefox
		candidates = append(candidates, findGeckoProfiles(filepath.Join(appData, "Mozilla", "Firefox", "Profiles"), "firefox", "Firefox")...)

		// Chromium variants
		if b := checkChromiumBrowser(filepath.Join(localAppData, "Google", "Chrome", "User Data"), "chrome", "Chrome", "chrome"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data"), "brave", "Brave", "brave"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(localAppData, "Microsoft", "Edge", "User Data"), "edge", "Edge", "edge"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(localAppData, "Arc", "User Data"), "arc", "Arc Browser", "arc"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(localAppData, "Vivaldi", "User Data"), "vivaldi", "Vivaldi", "vivaldi"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(appData, "Opera Software", "Opera Stable"), "opera", "Opera", "opera"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(appData, "Opera Software", "Opera GX Stable"), "opera-gx", "Opera GX", "opera"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(localAppData, "CocCoc", "Browser", "User Data"), "coccoc", "Cốc Cốc", "chromium"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(localAppData, "Yandex", "YandexBrowser", "User Data"), "yandex", "Yandex", "yandex"); b != nil {
			candidates = append(candidates, *b)
		}
	} else {
		// Linux & macOS
		configDir, _ := os.UserConfigDir()

		// Zen Browser
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".zen"), "zen", "Zen Browser")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".var", "app", "app.zen_browser.zen", ".zen"), "zen", "Zen Browser (Flatpak)")...)
		if runtime.GOOS == "darwin" {
			candidates = append(candidates, findGeckoProfiles(filepath.Join(home, "Library", "Application Support", "zen", "Profiles"), "zen", "Zen Browser")...)
		}

		// Floorp & LibreWolf & Waterfox
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".floorp"), "floorp", "Floorp")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".var", "app", "one.ablaze.floorp", ".floorp"), "floorp", "Floorp (Flatpak)")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".librewolf"), "librewolf", "LibreWolf")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".var", "app", "io.gitlab.librewolf-community", ".librewolf"), "librewolf", "LibreWolf (Flatpak)")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".waterfox"), "waterfox", "Waterfox")...)

		// Firefox
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".mozilla", "firefox"), "firefox", "Firefox")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, "snap", "firefox", "common", ".mozilla", "firefox"), "firefox", "Firefox (Snap)")...)
		candidates = append(candidates, findGeckoProfiles(filepath.Join(home, ".var", "app", "org.mozilla.firefox", ".mozilla", "firefox"), "firefox", "Firefox (Flatpak)")...)

		// Chromium variants
		if b := checkChromiumBrowser(filepath.Join(configDir, "google-chrome"), "chrome", "Chrome", "chrome"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(home, ".var", "app", "com.google.Chrome", "config", "google-chrome"), "chrome", "Chrome (Flatpak)", "chrome"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(configDir, "BraveSoftware", "Brave-Browser"), "brave", "Brave", "brave"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(home, ".var", "app", "com.brave.Browser", "config", "BraveSoftware", "Brave-Browser"), "brave", "Brave (Flatpak)", "brave"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(home, "snap", "brave", "current", ".config", "BraveSoftware", "Brave-Browser"), "brave", "Brave (Snap)", "brave"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(configDir, "microsoft-edge"), "edge", "Edge", "edge"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(home, ".var", "app", "com.microsoft.Edge", "config", "microsoft-edge"), "edge", "Edge (Flatpak)", "edge"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(configDir, "vivaldi"), "vivaldi", "Vivaldi", "vivaldi"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(configDir, "opera"), "opera", "Opera", "opera"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(configDir, "chromium"), "chromium", "Chromium", "chromium"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(home, "snap", "chromium", "current", ".config", "chromium"), "chromium", "Chromium (Snap)", "chromium"); b != nil {
			candidates = append(candidates, *b)
		}
		if b := checkChromiumBrowser(filepath.Join(configDir, "coccoc"), "coccoc", "Cốc Cốc", "chromium"); b != nil {
			candidates = append(candidates, *b)
		}

		// macOS Specifics
		if runtime.GOOS == "darwin" {
			appSupport := filepath.Join(home, "Library", "Application Support")
			candidates = append(candidates, DetectedBrowser{
				ID:        "safari",
				Name:      "Safari",
				CookieArg: "safari",
			})
			if b := checkChromiumBrowser(filepath.Join(appSupport, "Arc", "User Data"), "arc", "Arc Browser", "arc"); b != nil {
				candidates = append(candidates, *b)
			}
			if b := checkChromiumBrowser(filepath.Join(appSupport, "Google", "Chrome"), "chrome", "Chrome", "chrome"); b != nil {
				candidates = append(candidates, *b)
			}
			if b := checkChromiumBrowser(filepath.Join(appSupport, "BraveSoftware", "Brave-Browser"), "brave", "Brave", "brave"); b != nil {
				candidates = append(candidates, *b)
			}
			if b := checkChromiumBrowser(filepath.Join(appSupport, "Microsoft Edge"), "edge", "Edge", "edge"); b != nil {
				candidates = append(candidates, *b)
			}
		}
	}

	// Check for local cookies.txt in current directory
	if fi, err := os.Stat("cookies.txt"); err == nil && !fi.IsDir() {
		absPath, _ := filepath.Abs("cookies.txt")
		candidates = append(candidates, DetectedBrowser{
			ID:        "file",
			Name:      "cookies.txt (Local File)",
			CookieArg: absPath,
			LastUsed:  fi.ModTime(),
		})
	}

	// De-duplicate by ID, keeping the one with latest LastUsed
	bestByID := make(map[string]DetectedBrowser)
	for _, b := range candidates {
		if existing, ok := bestByID[b.ID]; !ok || b.LastUsed.After(existing.LastUsed) {
			bestByID[b.ID] = b
		}
	}

	var results []DetectedBrowser
	for _, b := range bestByID {
		results = append(results, b)
	}

	// Sort by LastUsed descending (most recently active browser first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].LastUsed.After(results[j].LastUsed)
	})

	return results
}

// GetCookieCascade returns the list of browsers to try in priority order.
func GetCookieCascade(preference string) []DetectedBrowser {
	pref := strings.ToLower(strings.TrimSpace(preference))
	if pref == "" {
		return nil
	}

	all := ScanInstalledBrowsers()
	if pref == "auto" {
		return all
	}

	// If user picked a specific browser, put it first, then add the rest as fallbacks
	var first *DetectedBrowser
	var rest []DetectedBrowser
	for _, b := range all {
		if strings.EqualFold(b.ID, pref) || strings.EqualFold(b.CookieArg, pref) {
			match := b
			first = &match
		} else {
			rest = append(rest, b)
		}
	}

	var cascade []DetectedBrowser
	if first != nil {
		cascade = append(cascade, *first)
	} else {
		// Specific browser chosen but not auto-detected; still add as first choice
		cascade = append(cascade, DetectedBrowser{
			ID:        pref,
			Name:      strings.ToUpper(pref),
			CookieArg: pref,
		})
	}
	cascade = append(cascade, rest...)
	return cascade
}

// CloneLockedGeckoProfile copies cookies.sqlite and WAL files from an active profile
// into a temporary directory to bypass SQLite lock without closing the browser.
func CloneLockedGeckoProfile(cookieArg string) (string, func()) {
	if !strings.HasPrefix(cookieArg, "firefox:") {
		return cookieArg, func() {}
	}
	srcDir := strings.TrimPrefix(cookieArg, "firefox:")
	srcDB := filepath.Join(srcDir, "cookies.sqlite")
	if _, err := os.Stat(srcDB); err != nil {
		return cookieArg, func() {}
	}

	tempDir, err := os.MkdirTemp("", "hmd_cookie_snap_*")
	if err != nil {
		return cookieArg, func() {}
	}
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	// Copy cookies.sqlite, cookies.sqlite-wal, cookies.sqlite-shm
	for _, ext := range []string{"", "-wal", "-shm"} {
		fName := "cookies.sqlite" + ext
		srcFile := filepath.Join(srcDir, fName)
		if data, err := os.ReadFile(srcFile); err == nil {
			_ = os.WriteFile(filepath.Join(tempDir, fName), data, 0600)
		}
	}

	return "firefox:" + tempDir, cleanup
}
