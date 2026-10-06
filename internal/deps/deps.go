package deps

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"hmd/internal/util"
)

// Paths to binary tools.
type BinaryPaths struct {
	YtDlp  string
	FFmpeg string
}

// ProgressCallback reports download progress for a dependency.
type ProgressCallback func(name string, downloaded, total int64, percent float64)

// GetBinaryNames returns the executable names depending on OS.
func GetBinaryNames() (string, string) {
	if runtime.GOOS == "windows" {
		return "yt-dlp.exe", "ffmpeg.exe"
	}
	return "yt-dlp", "ffmpeg"
}

// GetYtDlpDownloadURL returns the download URL for yt-dlp according to the runtime OS.
func GetYtDlpDownloadURL() string {
	switch runtime.GOOS {
	case "windows":
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	case "darwin":
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	default:
		return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
	}
}

// FindBinaries locates yt-dlp and ffmpeg either in PATH or in the app's bin directory.
func FindBinaries() (BinaryPaths, []string) {
	var paths BinaryPaths
	var missing []string

	ytName, ffName := GetBinaryNames()
	binDir, _ := util.GetBinDir()

	// 1. Find yt-dlp: check binDir candidate first before exec.LookPath
	if binDir != "" {
		candidate := filepath.Join(binDir, ytName)
		if fileExists(candidate) {
			paths.YtDlp = candidate
		}
	}
	if paths.YtDlp == "" {
		if p, err := exec.LookPath(ytName); err == nil {
			paths.YtDlp = p
		}
	}
	if paths.YtDlp == "" {
		missing = append(missing, "yt-dlp")
	}

	// 2. Find ffmpeg: check binDir candidate first before exec.LookPath
	if binDir != "" {
		candidate := filepath.Join(binDir, ffName)
		if fileExists(candidate) {
			paths.FFmpeg = candidate
		}
	}
	if paths.FFmpeg == "" {
		if p, err := exec.LookPath(ffName); err == nil {
			paths.FFmpeg = p
		}
	}
	if paths.FFmpeg == "" {
		missing = append(missing, "ffmpeg")
	}

	return paths, missing
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// UpdateYtDlp downloads the latest yt-dlp binary to binDir, atomically swaps it into place, and returns the path.
func UpdateYtDlp(cb ProgressCallback) (string, error) {
	binDir, err := util.GetBinDir()
	if err != nil {
		return "", fmt.Errorf("không thể tạo thư mục lưu công cụ: %w", err)
	}

	ytName, _ := GetBinaryNames()
	destPath := filepath.Join(binDir, ytName)
	tmpPath := filepath.Join(binDir, ytName+".tmp")

	downloadURL := GetYtDlpDownloadURL()
	if err := downloadToFile(downloadURL, tmpPath, "yt-dlp", cb); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("lỗi khi tải yt-dlp: %w", err)
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("lỗi cấp quyền thực thi cho yt-dlp: %w", err)
	}

	if runtime.GOOS == "windows" {
		oldPath := destPath + ".old"
		_ = os.Remove(oldPath)
		if fileExists(destPath) {
			if err := os.Rename(destPath, oldPath); err != nil {
				_ = os.Remove(tmpPath)
				return "", fmt.Errorf("lỗi hoán đổi file yt-dlp trên Windows: %w", err)
			}
		}
		if err := os.Rename(tmpPath, destPath); err != nil {
			if fileExists(oldPath) {
				_ = os.Rename(oldPath, destPath)
			}
			_ = os.Remove(tmpPath)
			return "", fmt.Errorf("lỗi cài đặt file yt-dlp mới: %w", err)
		}
	} else {
		if err := os.Rename(tmpPath, destPath); err != nil {
			_ = os.Remove(tmpPath)
			return "", fmt.Errorf("lỗi thay thế file yt-dlp: %w", err)
		}
	}

	_ = os.Chmod(destPath, 0755)
	return destPath, nil
}

// EnsureDependencies checks and downloads any missing tools (yt-dlp, ffmpeg).
func EnsureDependencies(cb ProgressCallback) (BinaryPaths, error) {
	paths, missing := FindBinaries()
	if len(missing) == 0 {
		return paths, nil
	}

	binDir, err := util.GetBinDir()
	if err != nil {
		return paths, fmt.Errorf("không thể tạo thư mục lưu công cụ: %w", err)
	}

	ytName, ffName := GetBinaryNames()

	for _, item := range missing {
		switch item {
		case "yt-dlp":
			downloadURL := GetYtDlpDownloadURL()
			destPath := filepath.Join(binDir, ytName)
			if err := downloadFile(downloadURL, destPath, "yt-dlp", cb); err != nil {
				return paths, fmt.Errorf("lỗi khi tải yt-dlp: %w", err)
			}
			_ = os.Chmod(destPath, 0755)
			paths.YtDlp = destPath

		case "ffmpeg":
			if runtime.GOOS == "windows" {
				zipURL := "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
				tempZip := filepath.Join(binDir, "ffmpeg_temp.zip")
				if err := downloadFile(zipURL, tempZip, "ffmpeg", cb); err != nil {
					return paths, fmt.Errorf("lỗi khi tải ffmpeg: %w", err)
				}
				defer os.Remove(tempZip)

				// Extract ffmpeg.exe and ffprobe.exe
				if err := extractFFmpegFromZip(tempZip, binDir); err != nil {
					return paths, fmt.Errorf("lỗi khi giải nén ffmpeg: %w", err)
				}
				paths.FFmpeg = filepath.Join(binDir, ffName)
			} else {
				return paths, fmt.Errorf("hệ thống chưa cài đặt ffmpeg. Vui lòng cài đặt ffmpeg qua package manager của hệ điều hành")
			}
		}
	}

	return paths, nil
}

func downloadToFile(url, filePath, itemName string, cb ProgressCallback) error {
	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) hittakeMD")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mã lỗi HTTP: %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	out, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	var downloaded int64
	buf := make([]byte, 64*1024)

	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := out.Write(buf[0:nr])
			if nw < 0 || nr != nw {
				return ew
			}
			downloaded += int64(nw)
			if cb != nil {
				var pct float64
				if totalSize > 0 {
					pct = float64(downloaded) / float64(totalSize) * 100
				}
				cb(itemName, downloaded, totalSize, pct)
			}
		}
		if er != nil {
			if er == io.EOF {
				break
			}
			return er
		}
	}

	return nil
}

func downloadFile(url, destPath, itemName string, cb ProgressCallback) error {
	tmpFile := destPath + ".tmp"
	if err := downloadToFile(url, tmpFile, itemName, cb); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	if err := os.Rename(tmpFile, destPath); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	return nil
}

func extractFFmpegFromZip(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		cleanName := filepath.Base(f.Name)
		if strings.EqualFold(cleanName, "ffmpeg.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}

			destFile := filepath.Join(targetDir, cleanName)
			outFile, err := os.OpenFile(destFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				rc.Close()
				return err
			}

			_, copyErr := io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()

			if copyErr != nil {
				return copyErr
			}
		}
	}
	return nil
}
