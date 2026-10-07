package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ItemType represents the kind of Spotify resource.
type ItemType string

const (
	ItemTrack    ItemType = "track"
	ItemAlbum    ItemType = "album"
	ItemPlaylist ItemType = "playlist"
)

// TrackInfo holds metadata for a single Spotify track.
type TrackInfo struct {
	ID         string
	Title      string
	Artist     string
	Album      string
	DurationMs int
	CoverURL   string
	IsPlayable bool
}

// CollectionInfo holds metadata for a track, album, or playlist.
type CollectionInfo struct {
	Type     ItemType
	ID       string
	Title    string
	Subtitle string
	CoverURL string
	Tracks   []TrackInfo
}

var (
	spotifyURLRegex = regexp.MustCompile(`(?:open\.spotify\.com(?:/[a-zA-Z-]+)?/|spotify:)(track|album|playlist)[/:]([a-zA-Z0-9]+)`)
	nextDataRegex   = regexp.MustCompile(`<script\s+id="__NEXT_DATA__"\s+type="application/json">(.*?)</script>`)
)

// DefaultHTTPClient provides a timeout-configured client with standard browser User-Agent.
var DefaultHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// IsSpotifyURL checks if the URL belongs to Spotify domains.
func IsSpotifyURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	return strings.Contains(host, "spotify.com") || strings.Contains(host, "spotify.link") || strings.HasPrefix(rawURL, "spotify:")
}

// ParseURL extracts the ItemType and Spotify ID from a Spotify URL or URI.
func ParseURL(rawURL string) (ItemType, string, error) {
	matches := spotifyURLRegex.FindStringSubmatch(rawURL)
	if len(matches) < 3 {
		return "", "", fmt.Errorf("không nhận diện được định dạng liên kết Spotify hợp lệ")
	}

	itemType := ItemType(strings.ToLower(matches[1]))
	id := matches[2]

	switch itemType {
	case ItemTrack, ItemAlbum, ItemPlaylist:
		return itemType, id, nil
	default:
		return "", "", fmt.Errorf("loại nội dung Spotify %q hiện chưa được hỗ trợ", itemType)
	}
}

// Resolve fetches metadata for a Spotify track, album, or playlist.
func Resolve(ctx context.Context, rawURL string) (*CollectionInfo, error) {
	effectiveURL := rawURL

	// Handle spotify.link redirects if needed
	if strings.Contains(rawURL, "spotify.link") {
		req, err := http.NewRequestWithContext(ctx, "HEAD", rawURL, nil)
		if err == nil {
			req.Header.Set("User-Agent", defaultUserAgent)
			clientNoRedirect := &http.Client{
				Timeout: 10 * time.Second,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			resp, err := clientNoRedirect.Do(req)
			if err == nil {
				loc := resp.Header.Get("Location")
				_ = resp.Body.Close()
				if loc != "" {
					effectiveURL = loc
				}
			}
		}
	}

	itemType, id, err := ParseURL(effectiveURL)
	if err != nil {
		return nil, err
	}

	embedURL := fmt.Sprintf("https://open.spotify.com/embed/%s/%s", itemType, id)
	req, err := http.NewRequestWithContext(ctx, "GET", embedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("lỗi khởi tạo yêu cầu Spotify: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,vi;q=0.8")

	resp, err := DefaultHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối máy chủ Spotify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("nội dung Spotify không tồn tại hoặc đã bị xóa (404)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("máy chủ Spotify phản hồi mã lỗi HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("lỗi đọc dữ liệu Spotify: %w", err)
	}

	coll, err := parseEmbedHTML(string(body), itemType, id)
	if err == nil && coll != nil && len(coll.Tracks) > 0 {
		return coll, nil
	}

	// Fallback to oEmbed for single tracks if HTML embed parsing failed
	if itemType == ItemTrack {
		oembedColl, oembedErr := fetchOEmbed(ctx, effectiveURL)
		if oembedErr == nil && oembedColl != nil {
			return oembedColl, nil
		}
	}

	if err != nil {
		return nil, err
	}
	return nil, errors.New("không thể phân tích thông tin bài hát từ Spotify")
}

type embedPayload struct {
	Props struct {
		PageProps struct {
			State struct {
				Data struct {
					Entity struct {
						Type       string `json:"type"`
						Name       string `json:"name"`
						Title      string `json:"title"`
						Subtitle   string `json:"subtitle"`
						Duration   int    `json:"duration"`
						IsPlayable *bool  `json:"isPlayable"`
						CoverArt   struct {
							Sources []struct {
								URL string `json:"url"`
							} `json:"sources"`
						} `json:"coverArt"`
						TrackList []struct {
							URI        string `json:"uri"`
							Title      string `json:"title"`
							Subtitle   string `json:"subtitle"`
							Duration   int    `json:"duration"`
							IsPlayable *bool  `json:"isPlayable"`
						} `json:"trackList"`
					} `json:"entity"`
				} `json:"data"`
			} `json:"state"`
		} `json:"pageProps"`
	} `json:"props"`
}

func parseEmbedHTML(htmlContent string, itemType ItemType, id string) (*CollectionInfo, error) {
	matches := nextDataRegex.FindStringSubmatch(htmlContent)
	var jsonStr string
	if len(matches) >= 2 {
		jsonStr = matches[1]
	} else {
		// Try fallback regex if script tag had attributes in different order
		startIdx := strings.Index(htmlContent, `{"props":`)
		if startIdx != -1 {
			endIdx := strings.Index(htmlContent[startIdx:], `</script>`)
			if endIdx != -1 {
				jsonStr = htmlContent[startIdx : startIdx+endIdx]
			}
		}
	}

	if jsonStr == "" {
		return nil, errors.New("không tìm thấy khối dữ liệu __NEXT_DATA__")
	}

	var payload embedPayload
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		return nil, fmt.Errorf("lỗi giải mã JSON dữ liệu Spotify: %w", err)
	}

	entity := payload.Props.PageProps.State.Data.Entity
	title := entity.Title
	if title == "" {
		title = entity.Name
	}
	if title == "" {
		title = "Spotify Media"
	}

	coverURL := ""
	if len(entity.CoverArt.Sources) > 0 {
		coverURL = entity.CoverArt.Sources[len(entity.CoverArt.Sources)-1].URL
		if coverURL == "" && len(entity.CoverArt.Sources) > 0 {
			coverURL = entity.CoverArt.Sources[0].URL
		}
	}

	coll := &CollectionInfo{
		Type:     itemType,
		ID:       id,
		Title:    title,
		Subtitle: entity.Subtitle,
		CoverURL: coverURL,
	}

	switch itemType {
	case ItemTrack:
		isPlayable := true
		if entity.IsPlayable != nil {
			isPlayable = *entity.IsPlayable
		}
		coll.Tracks = []TrackInfo{
			{
				ID:         id,
				Title:      title,
				Artist:     entity.Subtitle,
				Album:      title,
				DurationMs: entity.Duration,
				CoverURL:   coverURL,
				IsPlayable: isPlayable,
			},
		}

	case ItemAlbum, ItemPlaylist:
		for _, t := range entity.TrackList {
			trackID := strings.TrimPrefix(t.URI, "spotify:track:")
			artist := t.Subtitle
			if artist == "" {
				artist = entity.Subtitle
			}
			albumName := coll.Title
			isPlayable := true
			if t.IsPlayable != nil {
				isPlayable = *t.IsPlayable
			}

			coll.Tracks = append(coll.Tracks, TrackInfo{
				ID:         trackID,
				Title:      t.Title,
				Artist:     artist,
				Album:      albumName,
				DurationMs: t.Duration,
				CoverURL:   coverURL,
				IsPlayable: isPlayable,
			})
		}
	}

	return coll, nil
}

type oembedResponse struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
	AuthorName   string `json:"author_name"`
}

func fetchOEmbed(ctx context.Context, rawURL string) (*CollectionInfo, error) {
	oembedURL := fmt.Sprintf("https://open.spotify.com/oembed?url=%s", url.QueryEscape(rawURL))
	req, err := http.NewRequestWithContext(ctx, "GET", oembedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := DefaultHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oEmbed HTTP %d", resp.StatusCode)
	}

	var data oembedResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &CollectionInfo{
		Type:     ItemTrack,
		Title:    data.Title,
		Subtitle: data.AuthorName,
		CoverURL: data.ThumbnailURL,
		Tracks: []TrackInfo{
			{
				Title:      data.Title,
				Artist:     data.AuthorName,
				CoverURL:   data.ThumbnailURL,
				IsPlayable: true,
			},
		},
	}, nil
}
