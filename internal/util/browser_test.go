package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanInstalledBrowsersRealMachine(t *testing.T) {
	browsers := ScanInstalledBrowsers()
	// On this Linux machine, Zen Browser is installed
	foundZen := false
	for _, b := range browsers {
		if b.ID == "zen" {
			foundZen = true
			if !strings.HasPrefix(b.CookieArg, "firefox:") {
				t.Errorf("expected CookieArg to have 'firefox:' prefix, got: %s", b.CookieArg)
			}
		}
	}
	if !foundZen {
		t.Logf("Zen browser not detected or running in isolated env, found: %d browsers", len(browsers))
	}
}

func TestFindGeckoProfilesMock(t *testing.T) {
	tempDir := t.TempDir()
	profile1 := filepath.Join(tempDir, "default.release")
	_ = os.MkdirAll(profile1, 0755)
	cookieFile := filepath.Join(profile1, "cookies.sqlite")
	_ = os.WriteFile(cookieFile, []byte("sqlite-data"), 0644)

	profiles := findGeckoProfiles(tempDir, "testbrowser", "Test Browser")
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].ID != "testbrowser" {
		t.Errorf("expected ID 'testbrowser', got %s", profiles[0].ID)
	}
	expectedArg := "firefox:" + profile1
	if profiles[0].CookieArg != expectedArg {
		t.Errorf("expected CookieArg %s, got %s", expectedArg, profiles[0].CookieArg)
	}
}

func TestGetCookieCascade(t *testing.T) {
	// Empty preference returns nil (cookies off)
	if cascade := GetCookieCascade(""); len(cascade) != 0 {
		t.Errorf("expected empty cascade for empty preference, got %d", len(cascade))
	}

	// Specific browser not installed still becomes first candidate
	cascade := GetCookieCascade("safari")
	if len(cascade) == 0 || cascade[0].ID != "safari" {
		t.Errorf("expected first candidate to be safari, got: %v", cascade)
	}
}
