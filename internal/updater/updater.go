package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	CurrentVersion = "1.2.0"
	GitHubRepo     = "luongminhduc-hittake/ytdownloader"
)

type gitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// ReleaseInfo contains resolved update metadata.
type ReleaseInfo struct {
	TagName     string
	DownloadURL string
	AssetSize   int64
	FileName    string
}

// CheckForUpdate queries GitHub Releases for a newer version.
func CheckForUpdate() (*ReleaseInfo, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ytdl-self-updater")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // No releases yet
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel gitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("không thể đọc dữ liệu phiên bản: %w", err)
	}

	remoteVer := strings.TrimPrefix(strings.TrimSpace(rel.TagName), "v")
	currentVer := strings.TrimPrefix(strings.TrimSpace(CurrentVersion), "v")

	if !isNewerVersion(remoteVer, currentVer) {
		return nil, nil
	}

	// Match asset based on OS
	var targetAssetURL string
	var targetSize int64
	var targetName string

	isWin := runtime.GOOS == "windows"
	for _, asset := range rel.Assets {
		nameLower := strings.ToLower(asset.Name)
		if isWin {
			if strings.HasSuffix(nameLower, ".exe") {
				targetAssetURL = asset.BrowserDownloadURL
				targetSize = asset.Size
				targetName = asset.Name
				break
			}
		} else {
			if strings.Contains(nameLower, "linux") || (!strings.HasSuffix(nameLower, ".exe") && !strings.HasSuffix(nameLower, ".zip")) {
				targetAssetURL = asset.BrowserDownloadURL
				targetSize = asset.Size
				targetName = asset.Name
				break
			}
		}
	}

	if targetAssetURL == "" {
		return nil, nil
	}

	return &ReleaseInfo{
		TagName:     rel.TagName,
		DownloadURL: targetAssetURL,
		AssetSize:   targetSize,
		FileName:    targetName,
	}, nil
}

// isNewerVersion compares semver strings "X.Y.Z".
func isNewerVersion(remote, current string) bool {
	rParts := strings.Split(remote, ".")
	cParts := strings.Split(current, ".")

	for i := 0; i < len(rParts) && i < len(cParts); i++ {
		rNum, err1 := strconv.Atoi(rParts[i])
		cNum, err2 := strconv.Atoi(cParts[i])
		if err1 == nil && err2 == nil {
			if rNum > cNum {
				return true
			}
			if rNum < cNum {
				return false
			}
		} else {
			if rParts[i] > cParts[i] {
				return true
			}
			if rParts[i] < cParts[i] {
				return false
			}
		}
	}
	return len(rParts) > len(cParts)
}

// ApplyUpdate downloads the new binary and atomically replaces the currently running executable.
func ApplyUpdate(downloadURL string, onProgress func(dl, total int64, pct float64)) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("không thể xác định vị trí file chạy: %w", err)
	}
	if realPath, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = realPath
	}

	execDir := filepath.Dir(execPath)
	tmpPath := filepath.Join(execDir, fmt.Sprintf(".ytdl_update_%d.tmp", time.Now().UnixNano()))

	// Download file
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("lỗi kết nối tải bản cập nhật: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("máy chủ trả về mã lỗi: %d", resp.StatusCode)
	}

	out, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("không thể tạo file tạm: %w", err)
	}

	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		nr, rerr := resp.Body.Read(buf)
		if nr > 0 {
			nw, werr := out.Write(buf[0:nr])
			if werr != nil {
				out.Close()
				_ = os.Remove(tmpPath)
				return fmt.Errorf("lỗi ghi file cập nhật: %w", werr)
			}
			downloaded += int64(nw)
			if onProgress != nil && total > 0 {
				pct := (float64(downloaded) / float64(total)) * 100.0
				onProgress(downloaded, total, pct)
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				break
			}
			out.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("lỗi tải dữ liệu cập nhật: %w", rerr)
		}
	}
	out.Close()

	// Safe replacement
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		_ = os.Remove(oldPath) // remove previous old file if exists
		if err := os.Rename(execPath, oldPath); err != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("lỗi hoán đổi file trên Windows: %w", err)
		}
		if err := os.Rename(tmpPath, execPath); err != nil {
			// Try to recover
			_ = os.Rename(oldPath, execPath)
			return fmt.Errorf("lỗi cài đặt file mới: %w", err)
		}
	} else {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("lỗi cấp quyền thực thi: %w", err)
		}
		if err := os.Rename(tmpPath, execPath); err != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("lỗi thay thế file: %w", err)
		}
	}

	return nil
}
