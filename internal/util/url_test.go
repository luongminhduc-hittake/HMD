package util

import (
	"testing"
)

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "standard url",
			input: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:  "double quotes",
			input: "\"https://www.youtube.com/watch?v=dQw4w9WgXcQ\"",
			want:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:  "single quotes",
			input: "'https://www.youtube.com/watch?v=dQw4w9WgXcQ'",
			want:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:  "invisible BOM and zero width spaces",
			input: "\uFEFFhttps://www.youtube.com/watch?v=dQw4w9WgXcQ\u200B",
			want:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:  "missing https scheme",
			input: "www.youtube.com/watch?v=dQw4w9WgXcQ",
			want:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name:  "spotify track uri",
			input: "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
			want:  "https://open.spotify.com/track/6rqhFgbbKwnb9MLmUQDhG6",
		},
		{
			name:  "spotify playlist uri",
			input: "spotify:playlist:37i9dQZF1DXcBWIGoYBM5M",
			want:  "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:    "empty string",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid text without domain",
			input:   "not a valid url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SanitizeURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
