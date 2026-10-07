package downloader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hmd/internal/spotify"
	"hmd/internal/util"
)

// FetchInfo retrieves metadata for a given URL without downloading media.
func FetchInfo(ytdlpPath, rawURL string, cookiesBrowser ...string) (*MediaInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Resolve redirect/shortlink to canonical URL
	rawURL = ResolveRedirect(ctx, rawURL)

	// Intercept Spotify URLs to resolve via Spotify native engine
	if spotify.IsSpotifyURL(rawURL) {
		coll, err := spotify.Resolve(ctx, rawURL)
		if err != nil {
			return nil, err
		}
		durationStr := ""
		if len(coll.Tracks) == 1 && coll.Tracks[0].DurationMs > 0 {
			durationSec := coll.Tracks[0].DurationMs / 1000
			durationStr = fmt.Sprintf("%02d:%02d", durationSec/60, durationSec%60)
		} else if len(coll.Tracks) > 1 {
			durationStr = fmt.Sprintf("%d bài hát", len(coll.Tracks))
		}
		isPlaylist := coll.Type == spotify.ItemAlbum || coll.Type == spotify.ItemPlaylist
		return &MediaInfo{
			Title:            coll.Title,
			Duration:         durationStr,
			Uploader:         coll.Subtitle,
			IsPlaylist:       isPlaylist,
			RawURL:           rawURL,
			IsSpotify:        true,
			SpotifyTracks:    coll.Tracks,
			ThumbnailURL:     coll.CoverURL,
			MediaType:        MediaTypeAudio,
			AvailablePresets: AudioPresets,
		}, nil
	}

	// Check if URL indicates a playlist
	isPlaylist := false
	isPurePlaylist := false
	if u, err := url.Parse(rawURL); err == nil {
		q := u.Query()
		if q.Get("list") != "" || strings.Contains(u.Path, "playlist") {
			isPlaylist = true
			if q.Get("v") == "" {
				isPurePlaylist = true
			}
		}
	}

	args := []string{"--no-config"}
	var cookies string
	if len(cookiesBrowser) > 0 {
		cookies = strings.TrimSpace(cookiesBrowser[0])
	}
	if cookies != "" && util.IsAllowedBrowser(cookies) {
		args = append(args, "--cookies-from-browser", cookies)
	}

	if isPurePlaylist {
		args = append(args,
			"--flat-playlist",
			"--no-warnings",
			"--ignore-errors",
			"--playlist-items", "1",
			"--print", "title:%(playlist_title|title)s",
			"--print", "uploader:%(uploader|channel)s",
			"--",
			rawURL,
		)
	} else {
		args = append(args,
			"--simulate",
			"--no-warnings",
			"--ignore-errors",
			"--no-playlist",
			"--print", "title:%(title)s",
			"--print", "duration_string:%(duration_string)s",
			"--print", "uploader:%(uploader)s",
			"--print", "height:%(height)s",
			"--print", "vcodec:%(vcodec)s",
			"--",
			rawURL,
		)
	}

	cmd := execCommandContext(ctx, ytdlpPath, args...)
	prepareCommand(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if strings.Contains(errStr, "database is locked") {
			return nil, fmt.Errorf("Trình duyệt đang mở và khóa file cookies. Cậu chủ vui lòng đóng trình duyệt rồi thử lại.")
		}

		// Fallback to gallery-dl for images, photo galleries, pins, carousels
		if gInfo, gErr := FetchGalleryInfo(ctx, rawURL, cookiesBrowser...); gErr == nil && gInfo != nil {
			return gInfo, nil
		}

		if errStr != "" {
			return nil, fmt.Errorf("%s", errStr)
		}
		return nil, fmt.Errorf("không thể phân tích link: %w", err)
	}

	info := &MediaInfo{
		RawURL:     rawURL,
		IsPlaylist: isPlaylist,
		Title:      "Media Video",
		Duration:   "N/A",
		Uploader:   "Media",
	}
	if isPurePlaylist {
		info.Duration = "Playlist"
		info.Title = "Media Playlist"
	}

	var vcodec string
	var maxHeight int

	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "title:"):
			t := strings.TrimPrefix(line, "title:")
			if t != "NA" && t != "" {
				info.Title = t
			}
		case strings.HasPrefix(line, "duration_string:"):
			d := strings.TrimPrefix(line, "duration_string:")
			if d != "NA" && d != "" {
				info.Duration = d
			}
		case strings.HasPrefix(line, "uploader:"):
			u := strings.TrimPrefix(line, "uploader:")
			if u != "NA" && u != "" {
				info.Uploader = u
			}
		case strings.HasPrefix(line, "height:"):
			hStr := strings.TrimPrefix(line, "height:")
			if h, err := strconv.Atoi(hStr); err == nil {
				maxHeight = h
			}
		case strings.HasPrefix(line, "vcodec:"):
			vcodec = strings.TrimPrefix(line, "vcodec:")
		}
	}

	info.MaxHeight = maxHeight
	platform := DetectPlatform(rawURL)
	if vcodec == "none" || platform.Name == "SoundCloud" || (vcodec == "" && maxHeight == 0) {
		info.MediaType = MediaTypeAudio
	} else {
		info.MediaType = MediaTypeVideo
	}
	info.AvailablePresets = DetermineAvailablePresets(info.MediaType, info.MaxHeight)

	return info, nil
}
