package downloader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

// FetchInfo retrieves metadata for a given YouTube URL without downloading media.
func FetchInfo(ytdlpPath, rawURL string) (*MediaInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

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

	var args []string
	if isPurePlaylist {
		args = []string{
			"--flat-playlist",
			"--no-warnings",
			"--ignore-errors",
			"--playlist-items", "1",
			"--print", "title:%(playlist_title|title)s",
			"--print", "uploader:%(uploader|channel)s",
			rawURL,
		}
	} else {
		args = []string{
			"--simulate",
			"--no-warnings",
			"--ignore-errors",
			"--no-playlist",
			"--print", "title:%(title)s",
			"--print", "duration_string:%(duration_string)s",
			"--print", "uploader:%(uploader)s",
			rawURL,
		}
	}

	cmd := exec.CommandContext(ctx, ytdlpPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr != "" {
			return nil, fmt.Errorf("%s", errStr)
		}
		return nil, fmt.Errorf("không thể phân tích link: %w", err)
	}

	info := &MediaInfo{
		RawURL:     rawURL,
		IsPlaylist: isPlaylist,
		Title:      "YouTube Video",
		Duration:   "N/A",
		Uploader:   "YouTube",
	}
	if isPurePlaylist {
		info.Duration = "Playlist"
		info.Title = "YouTube Playlist"
	}

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
		}
	}

	return info, nil
}
