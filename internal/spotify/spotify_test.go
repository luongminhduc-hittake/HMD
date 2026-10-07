package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsSpotifyURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT", true},
		{"https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT?si=abc", true},
		{"https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3", true},
		{"https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M", true},
		{"spotify:track:4cOdK2wGLETKBW3PvgPWqT", true},
		{"https://spotify.link/xyz123", true},
		{"https://youtube.com/watch?v=123", false},
		{"https://example.com/song.mp3", false},
	}

	for _, tt := range tests {
		got := IsSpotifyURL(tt.url)
		if got != tt.expected {
			t.Errorf("IsSpotifyURL(%q) = %v; want %v", tt.url, got, tt.expected)
		}
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		url          string
		expectedType ItemType
		expectedID   string
		shouldError  bool
	}{
		{"https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT", ItemTrack, "4cOdK2wGLETKBW3PvgPWqT", false},
		{"https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT?si=123", ItemTrack, "4cOdK2wGLETKBW3PvgPWqT", false},
		{"https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3", ItemAlbum, "1DFixLWuPkv3KT3TnV35m3", false},
		{"https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M", ItemPlaylist, "37i9dQZF1DXcBWIGoYBM5M", false},
		{"spotify:track:4cOdK2wGLETKBW3PvgPWqT", ItemTrack, "4cOdK2wGLETKBW3PvgPWqT", false},
		{"https://open.spotify.com/show/12345", "", "", true}, // Podcasts not supported as music tracks
		{"https://not-spotify.com/track/123", "", "", true},
	}

	for _, tt := range tests {
		itemType, id, err := ParseURL(tt.url)
		if (err != nil) != tt.shouldError {
			t.Errorf("ParseURL(%q) error = %v; want error %v", tt.url, err, tt.shouldError)
			continue
		}
		if !tt.shouldError {
			if itemType != tt.expectedType || id != tt.expectedID {
				t.Errorf("ParseURL(%q) = (%v, %v); want (%v, %v)", tt.url, itemType, id, tt.expectedType, tt.expectedID)
			}
		}
	}
}

func TestParseEmbedHTML_Track(t *testing.T) {
	mockHTML := `<!DOCTYPE html><html><head>
<script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{"type":"track","name":"Never Gonna Give You Up","title":"Never Gonna Give You Up","subtitle":"Rick Astley","duration":213000,"isPlayable":true,"coverArt":{"sources":[{"url":"https://example.com/cover.jpg"}]},"trackList":[]}}}}}}</script>
</head><body></body></html>`

	coll, err := parseEmbedHTML(mockHTML, ItemTrack, "4cOdK2wGLETKBW3PvgPWqT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if coll.Title != "Never Gonna Give You Up" {
		t.Errorf("expected Title %q, got %q", "Never Gonna Give You Up", coll.Title)
	}
	if coll.Subtitle != "Rick Astley" {
		t.Errorf("expected Subtitle %q, got %q", "Rick Astley", coll.Subtitle)
	}
	if coll.CoverURL != "https://example.com/cover.jpg" {
		t.Errorf("expected CoverURL %q, got %q", "https://example.com/cover.jpg", coll.CoverURL)
	}
	if len(coll.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(coll.Tracks))
	}
	if coll.Tracks[0].DurationMs != 213000 {
		t.Errorf("expected duration 213000, got %d", coll.Tracks[0].DurationMs)
	}
}

func TestParseEmbedHTML_AlbumAndPlaylist(t *testing.T) {
	isPlayableFalse := false
	mockHTML := `<!DOCTYPE html><html><head>
<script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{"type":"album","name":"Test Album","title":"Test Album","subtitle":"Great Artist","duration":0,"coverArt":{"sources":[{"url":"https://example.com/album.jpg"}]},"trackList":[{"uri":"spotify:track:t1","title":"Song One","subtitle":"Great Artist","duration":180000,"isPlayable":true},{"uri":"spotify:track:t2","title":"Song Two (Blocked)","subtitle":"Great Artist","duration":200000,"isPlayable":false}]}}}}}}</script>
</head><body></body></html>`

	_ = isPlayableFalse
	coll, err := parseEmbedHTML(mockHTML, ItemAlbum, "album123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if coll.Title != "Test Album" {
		t.Errorf("expected album title 'Test Album', got %q", coll.Title)
	}
	if len(coll.Tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(coll.Tracks))
	}
	if coll.Tracks[0].Title != "Song One" || !coll.Tracks[0].IsPlayable {
		t.Errorf("expected track 0 to be playable 'Song One', got %+v", coll.Tracks[0])
	}
	if coll.Tracks[1].Title != "Song Two (Blocked)" || coll.Tracks[1].IsPlayable {
		t.Errorf("expected track 1 to be unplayable 'Song Two (Blocked)', got %+v", coll.Tracks[1])
	}
}

func TestResolveWithMockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html><html><head>
<script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"state":{"data":{"entity":{"type":"track","name":"Live Track","title":"Live Track","subtitle":"Live Artist","duration":120000,"isPlayable":true,"coverArt":{"sources":[{"url":"https://example.com/live.jpg"}]},"trackList":[]}}}}}}</script>
</head><body></body></html>`
		w.Write([]byte(html))
	}))
	defer ts.Close()

	// Override DefaultHTTPClient Transport to route embed requests to test server
	origClient := DefaultHTTPClient
	DefaultHTTPClient = &http.Client{
		Transport: &testRoundTripper{targetURL: ts.URL},
	}
	defer func() { DefaultHTTPClient = origClient }()

	coll, err := Resolve(context.Background(), "https://open.spotify.com/track/1234567890123456789012")
	if err != nil {
		t.Fatalf("Resolve returned unexpected error: %v", err)
	}
	if coll.Title != "Live Track" {
		t.Errorf("expected Title %q, got %q", "Live Track", coll.Title)
	}
	if coll.Tracks[0].Artist != "Live Artist" {
		t.Errorf("expected Artist %q, got %q", "Live Artist", coll.Tracks[0].Artist)
	}
}

type testRoundTripper struct {
	targetURL string
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	testReq, err := http.NewRequestWithContext(req.Context(), req.Method, rt.targetURL, req.Body)
	if err != nil {
		return nil, err
	}
	testReq.Header = req.Header
	return http.DefaultTransport.RoundTrip(testReq)
}
