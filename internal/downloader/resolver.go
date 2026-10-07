package downloader

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"hmd/internal/util"
)

var defaultResolverClient = &http.Client{
	Transport: util.SharedTransport,
	Timeout:   5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return http.ErrUseLastResponse
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		return nil
	},
}

var pinIDRegex = regexp.MustCompile(`/pin/(?:[\w-]+--)?(\d+)`)

// IsRedirectURL checks if a URL belongs to known shortener or redirect domains.
func IsRedirectURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Host)
	path := strings.ToLower(u.Path)
	return strings.Contains(host, "pin.it") ||
		strings.Contains(host, "vt.tiktok.com") ||
		strings.Contains(host, "vm.tiktok.com") ||
		(strings.Contains(host, "tiktok.com") && strings.HasPrefix(path, "/t/")) ||
		strings.Contains(host, "fb.watch") ||
		strings.Contains(host, "fb.me") ||
		(strings.Contains(host, "facebook.com") && strings.HasPrefix(path, "/share")) ||
		strings.Contains(host, "b23.tv") ||
		strings.Contains(host, "t.co") ||
		strings.Contains(host, "redd.it") ||
		strings.Contains(host, "youtu.be") ||
		strings.Contains(host, "spotify.link")
}

// ResolveRedirect follows HTTP redirects for shortened or sharing URLs to get the canonical URL.
// If the URL is not a known redirect URL or resolution fails, rawURL is returned unchanged.
func ResolveRedirect(ctx context.Context, rawURL string) string {
	if !IsRedirectURL(rawURL) {
		return rawURL
	}

	resolveCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(resolveCtx, http.MethodGet, rawURL, nil)
	if err != nil {
		return rawURL
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := defaultResolverClient.Do(req)
	if err != nil {
		return rawURL
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
	}()

	finalURL := resp.Request.URL.String()
	if finalURL == "" {
		return rawURL
	}

	// Canonicalize Pinterest URLs by keeping clean pin ID URL
	if strings.Contains(finalURL, "pinterest.com/pin/") {
		if m := pinIDRegex.FindStringSubmatch(finalURL); len(m) > 1 {
			return "https://www.pinterest.com/pin/" + m[1] + "/"
		}
	}

	return finalURL
}
