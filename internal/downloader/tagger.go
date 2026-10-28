package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
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

	tmpAudioPath := filepath.Join(filepath.Dir(audioPath), fmt.Sprintf(".tmp_%d_%s", time.Now().UnixNano(), filepath.Base(audioPath)))
	defer os.Remove(tmpAudioPath)

	args := buildFFmpegMetadataArgs(audioPath, coverPath, track)
	args = append(args, tmpAudioPath)

	cmd := execCommandContext(ctx, ffmpegPath, args...)
	prepareCommand(cmd)
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

// containerSupportsCoverArt reports whether the container's muxer accepts an
// attached_pic video stream. wav and opus/ogg reject video streams entirely.
func containerSupportsCoverArt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".mp3", ".m4a", ".mp4", ".m4b", ".flac", ".aac":
		return true
	default:
		return false
	}
}

// buildFFmpegMetadataArgs builds the ffmpeg argument list that remuxes audioPath
// (plus an optional cover image) into a temp file, keeping the audio stream as-is.
func buildFFmpegMetadataArgs(audioPath, coverPath string, track spotify.TrackInfo) []string {
	ext := filepath.Ext(audioPath)

	if coverPath != "" && !containerSupportsCoverArt(ext) {
		coverPath = ""
	}

	args := []string{"-y", "-i", audioPath}

	if coverPath != "" {
		args = append(args, "-i", coverPath)
		args = append(args, "-map", "0:a", "-map", "1:0")
		// Container rules: only mp3/flac accept a copied mjpeg cover. Everything else
		// (m4a/mp4/opus/ogg) needs the image re-encoded as mjpeg and flagged attached_pic,
		// otherwise ffmpeg fails with "codec not currently supported in container".
		args = append(args, "-c:v", "mjpeg", "-disposition:v", "attached_pic")
		args = append(args, "-metadata:s:v", "title=Album cover", "-metadata:s:v", "comment=Cover (front)")
	}

	args = append(args, "-c:a", "copy")

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

	return args
}

func downloadTempCover(ctx context.Context, coverURL string) (string, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", coverURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
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
	tmpFileName := tmpFile.Name()
	defer func() {
		if err := tmpFile.Close(); err != nil {
			_ = os.Remove(tmpFileName)
		}
	}()

	if _, err := io.Copy(tmpFile, io.LimitReader(resp.Body, 10*1024*1024)); err != nil {
		_ = os.Remove(tmpFileName)
		return "", err
	}

	return tmpFileName, nil
}
