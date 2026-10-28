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

// redirectHosts lists domains whose links are shorteners or share wrappers.
// Matching is by full label ("t.co" or "*.t.co"), never a substring: "microsoft.com"
// contains "t.co" and would otherwise be fetched as if it were a shortener.
var redirectHosts = []string{
	"pin.it",
	"vt.tiktok.com",
	"vm.tiktok.com",
	"v.douyin.com",
	"fb.watch",
	"fb.me",
	"threads.net",
	"b23.tv",
	"t.co",
	"redd.it",
	"spotify.link",
	"bit.ly",
	"tinyurl.com",
	"cutt.ly",
	"s.id",
}

// hostMatches reports whether host is exactly domain or a subdomain of it.
func hostMatches(host, domain string) bool {
	return host == domain || strings.HasSuffix(host, "."+domain)
}

// IsRedirectURL checks if a URL belongs to known shortener or redirect domains.
func IsRedirectURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	path := strings.ToLower(u.Path)

	for _, domain := range redirectHosts {
		if hostMatches(host, domain) {
			// tiktok.com and douyin.com are not shorteners themselves; only their /t/ share paths are.
			if domain == "tiktok.com" || domain == "douyin.com" {
				continue
			}
			return true
		}
	}

	return (hostMatches(host, "tiktok.com") && strings.HasPrefix(path, "/t/")) ||
		(hostMatches(host, "facebook.com") && strings.HasPrefix(path, "/share")) ||
		(hostMatches(host, "instagram.com") && strings.HasPrefix(path, "/share"))
}

// ResolveRedirect follows HTTP redirects for shortened or sharing URLs to get the canonical URL.
// If the URL is not a known redirect URL or resolution fails, rawURL is returned unchanged.
func ResolveRedirect(ctx context.Context, rawURL string) string {
	// 1. Instant local canonicalization for youtu.be (0ms network cost, preserves v= param)
	if u, err := url.Parse(rawURL); err == nil && strings.Contains(strings.ToLower(u.Host), "youtu.be") {
		videoID := strings.Trim(u.Path, "/")
		if videoID != "" && !strings.Contains(videoID, "/") {
			q := u.Query()
			q.Set("v", videoID)
			u.Scheme = "https"
			u.Host = "www.youtube.com"
			u.Path = "/watch"
			u.RawQuery = q.Encode()
			return u.String()
		}
	}

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
