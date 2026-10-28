package updater

import (
	"crypto/sha256"
	"encoding/hex"
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

	"hmd/internal/util"
)

const (
	CurrentVersion = "3.2.0"
	GitHubRepo     = "luongminhduc-hittake/HMD"
)

type gitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
		Digest             string `json:"digest"`
	} `json:"assets"`
}

// ReleaseInfo contains resolved update metadata.
type ReleaseInfo struct {
	TagName     string
	DownloadURL string
	AssetSize   int64
	FileName    string
	// Checksum is the GitHub-published asset digest, e.g. "sha256:abc...". Empty when
	// the release predates asset digests, in which case installing is refused.
	Checksum string
}

// CheckForUpdate queries GitHub Releases for a newer version.
func CheckForUpdate() (*ReleaseInfo, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "hmd-self-updater")
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
	var targetDigest string

	isWin := runtime.GOOS == "windows"
	for _, asset := range rel.Assets {
		nameLower := strings.ToLower(asset.Name)
		if isWin {
			if strings.HasSuffix(nameLower, ".exe") {
				targetAssetURL = asset.BrowserDownloadURL
				targetSize = asset.Size
				targetName = asset.Name
				targetDigest = asset.Digest
				break
			}
		} else {
			if strings.Contains(nameLower, "linux") || (!strings.HasSuffix(nameLower, ".exe") && !strings.HasSuffix(nameLower, ".zip")) {
				targetAssetURL = asset.BrowserDownloadURL
				targetSize = asset.Size
				targetName = asset.Name
				targetDigest = asset.Digest
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
		Checksum:    targetDigest,
	}, nil
}

func parseSemver(v string) (major, minor, patch int) {
	clean := strings.TrimPrefix(v, "v")
	parts := strings.Split(clean, ".")
	if len(parts) > 0 {
		major, _ = strconv.Atoi(strings.Split(parts[0], "-")[0])
	}
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(strings.Split(parts[1], "-")[0])
	}
	if len(parts) > 2 {
		patch, _ = strconv.Atoi(strings.Split(parts[2], "-")[0])
	}
	return
}

// isNewerVersion compares semver strings "X.Y.Z".
func isNewerVersion(remote, current string) bool {
	r1, r2, r3 := parseSemver(remote)
	c1, c2, c3 := parseSemver(current)
	if r1 != c1 {
		return r1 > c1
	}
	if r2 != c2 {
		return r2 > c2
	}
	return r3 > c3
}

// ApplyUpdate downloads the new binary, verifies it against the checksum published
// with the release, and only then atomically replaces the running executable.
func ApplyUpdate(downloadURL, expectedChecksum string, expectedSize int64, onProgress func(dl, total int64, pct float64)) error {
	if strings.TrimSpace(expectedChecksum) == "" {
		return fmt.Errorf("không có checksum cho bản cập nhật, từ chối cài đặt để bảo đảm toàn vẹn")
	}
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("không thể xác định vị trí file chạy: %w", err)
	}
	if realPath, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = realPath
	}

	execDir := filepath.Dir(execPath)
	out, err := os.CreateTemp(execDir, ".hmd_update_*.tmp")
	if err != nil {
		return fmt.Errorf("không thể tạo file tạm: %w", err)
	}
	tmpPath := out.Name()

	// Download file with shared transport
	client := &http.Client{
		Transport: util.SharedTransport,
	}
	resp, err := client.Get(downloadURL)
	if err != nil {
		out.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("lỗi kết nối tải bản cập nhật: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("máy chủ trả về mã lỗi: %d", resp.StatusCode)
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

	// Never install bytes we cannot vouch for: check the published digest and size
	// before touching the running executable.
	sum, hashErr := fileSHA256(tmpPath)
	if hashErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("không thể kiểm tra toàn vẹn bản cập nhật: %w", hashErr)
	}
	if want := normalizeDigest(expectedChecksum); want != "" && !strings.EqualFold(sum, want) {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("checksum bản cập nhật không khớp: mong đợi %s, nhận %s", want, sum)
	}
	if expectedSize > 0 {
		if info, statErr := os.Stat(tmpPath); statErr == nil && info.Size() != expectedSize {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("kích thước bản cập nhật không khớp: mong đợi %d byte, nhận %d byte", expectedSize, info.Size())
		}
	}

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

// fileSHA256 returns the lowercase hex sha256 digest of path.
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// normalizeDigest accepts "sha256:<hex>" or bare "<hex>" and returns the hex part.
func normalizeDigest(d string) string {
	d = strings.TrimSpace(d)
	if i := strings.Index(d, ":"); i >= 0 {
		d = d[i+1:]
	}
	return strings.ToLower(strings.TrimSpace(d))
}

// parseChecksum returns the hex digest listed for assetName in a "hash  name" manifest.
func parseChecksum(manifest, assetName string) string {
	for _, line := range strings.Split(manifest, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == assetName {
			return strings.ToLower(fields[0])
		}
	}
	return ""
}
