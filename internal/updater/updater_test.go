package updater

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		remote  string
		current string
		want    bool
	}{
		{"1.1.0", "1.0.0", true},
		{"1.0.1", "1.0.0", true},
		{"2.0.0", "1.9.9", true},
		{"1.0.0", "1.0.0", false},
		{"0.9.9", "1.0.0", false},
		{"1.0.0", "1.1.0", false},
		{"1.1.1", "1.1.0", true},
		{"1.2.0", "1.1.0", true},
		{"1.2.0", "1.2.0", false},
		{"2.0.0", "1.2.0", true},
		{"2.0.0", "2.0.0", false},
		{"2.0.1", "2.0.0", true},
		{"2.0.1", "2.0.1", false},
		{"2.1.0", "2.0.1", true},
		{"2.1.0", "2.1.0", false},
		{"2.2.0", "2.1.0", true},
		{"2.2.0", "2.2.0", false},
		{"2.2.1", "2.2.0", true},
		{"2.2.1", "2.2.1", false},
		{"3.0.1-hotfix", "3.0.0", true},
		{"3.0.1", "3.0.0", true},
		{"3.0.1-hotfix", "3.0.1", false},
	}

	for _, tt := range tests {
		got := isNewerVersion(tt.remote, tt.current)
		if got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v; want %v", tt.remote, tt.current, got, tt.want)
		}
	}
}

func TestParseChecksumLine(t *testing.T) {
	sums := "e51e69e89489f199530f13d32a4342da3ddb2cf34842f364603a066db05afbf3  hmd-linux-amd64\n" +
		"aabbcc  other-file\n"

	got := parseChecksum(sums, "hmd-linux-amd64")
	want := "e51e69e89489f199530f13d32a4342da3ddb2cf34842f364603a066db05afbf3"
	if got != want {
		t.Errorf("parseChecksum = %q, want %q", got, want)
	}

	if got := parseChecksum(sums, "missing-asset"); got != "" {
		t.Errorf("expected empty checksum for unknown asset, got %q", got)
	}
}

func TestApplyUpdateRejectsChecksumMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("malicious-payload"))
	}))
	defer srv.Close()

	err := ApplyUpdate(srv.URL, "0000000000000000000000000000000000000000000000000000000000000000", 0, nil)
	if err == nil {
		t.Fatalf("expected ApplyUpdate to reject a payload whose checksum does not match")
	}
	if !strings.Contains(err.Error(), "checksum") && !strings.Contains(err.Error(), "toàn vẹn") {
		t.Errorf("expected a checksum error, got: %v", err)
	}
}

func TestApplyUpdateRejectsEmptyExpectedChecksum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("payload"))
	}))
	defer srv.Close()

	if err := ApplyUpdate(srv.URL, "", 0, nil); err == nil {
		t.Fatalf("expected ApplyUpdate to refuse installing without a known checksum")
	}
}
