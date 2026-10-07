package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"hmd/internal/spotify"
)

// InjectMetadata downloads the cover image from Spotify and embeds metadata + cover art
// into the audio file using FFmpeg atomic remuxing.
func InjectMetadata(ctx context.Context, ffmpegPath, audioPath string, track spotify.TrackInfo) error {
	if ffmpegPath == "" {
		return nil
	}
	if _, err := os.Stat(audioPath); err != nil {
		return err
	}

	var coverPath string
	if track.CoverURL != "" {
		tmpCover, err := downloadTempCover(ctx, track.CoverURL)
		if err == nil {
			coverPath = tmpCover
			defer os.Remove(coverPath)
		}
	}

	ext := filepath.Ext(audioPath)
	tmpAudioPath := filepath.Join(filepath.Dir(audioPath), fmt.Sprintf(".tmp_%d_%s", time.Now().UnixNano(), filepath.Base(audioPath)))
	defer os.Remove(tmpAudioPath)

	var args []string
	args = append(args, "-y", "-i", audioPath)

	hasCover := coverPath != ""
	if hasCover {
		args = append(args, "-i", coverPath)
		args = append(args, "-map", "0:a", "-map", "1:0")
		args = append(args, "-metadata:s:v", "title=Album cover", "-metadata:s:v", "comment=Cover (front)")
	}

	args = append(args, "-c", "copy")

	// If mp3, force ID3v2.3 for maximum compatibility
	if strings.ToLower(ext) == ".mp3" {
		args = append(args, "-id3v2_version", "3")
	}

	if track.Title != "" {
		args = append(args, "-metadata", fmt.Sprintf("title=%s", track.Title))
	}
	if track.Artist != "" {
		args = append(args, "-metadata", fmt.Sprintf("artist=%s", track.Artist))
	}
	if track.Album != "" {
		args = append(args, "-metadata", fmt.Sprintf("album=%s", track.Album))
	}

	args = append(args, tmpAudioPath)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg metadata injection failed: %w (output: %s)", err, string(output))
	}

	// Atomic replacement
	if err := os.Rename(tmpAudioPath, audioPath); err != nil {
		return fmt.Errorf("failed to replace audio file with tagged version: %w", err)
	}

	return nil
}

func downloadTempCover(ctx context.Context, coverURL string) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", coverURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cover HTTP status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "hmd_cover_*.jpg")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, io.LimitReader(resp.Body, 10*1024*1024)); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}
