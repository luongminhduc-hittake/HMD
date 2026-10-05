package updater

import "testing"

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
	}

	for _, tt := range tests {
		got := isNewerVersion(tt.remote, tt.current)
		if got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v; want %v", tt.remote, tt.current, got, tt.want)
		}
	}
}
