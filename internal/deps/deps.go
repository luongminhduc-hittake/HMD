package deps

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"hmd/internal/util"
)

// Paths to binary tools.
type BinaryPaths struct {
	YtDlp     string
	FFmpeg    string
	GalleryDl string
}

// ProgressCallback reports download progress for a dependency.
type ProgressCallback func(name string, downloaded, total int64, percent float64)

// downloadBinary performs a checksum-verified download; replaced in tests.
var downloadBinary = downloadVerified

// GetBinaryNames returns the executable names depending on OS.
func GetBinaryNames() (string, string) {
	if runtime.GOOS == "windows" {
		return "yt-dlp.exe", "ffmpeg.exe"
	}
	return "yt-dlp", "ffmpeg"
}

// GetGalleryDlAssetName returns the release asset file name for gallery-dl.
// It differs from the installed name: Linux/macOS ship "gallery-dl.bin", Windows ships
// "gallery-dl.exe", and both are installed as "gallery-dl[.exe]".
func GetGalleryDlAssetName() string {
	if runtime.GOOS == "windows" {
		return "gallery-dl.exe"
	}
	return "gallery-dl.bin"
}

// GetGalleryDlBinaryName returns the executable name for gallery-dl depending on OS.
func GetGalleryDlBinaryName() string {
	if runtime.GOOS == "windows" {
		return "gallery-dl.exe"
	}
	return "gallery-dl"
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

// GalleryDlVersion is the fallback gallery-dl release, used when the release API
// cannot be reached. Normal installs resolve the newest tag at download time, because
// gallery-dl's site extractors break as sites change.
const GalleryDlVersion = "v1.32.18"

// Codeberg has no /releases/latest/download path, so the newest tag is read from the API.
func galleryDlAPIURL() string {
	return "https://codeberg.org/api/v1/repos/mikf/gallery-dl/releases?limit=1"
}

// galleryDlReleaseBase returns the download root for a given tag.
func galleryDlReleaseBase(tag string) string {
	return "https://codeberg.org/mikf/gallery-dl/releases/download/" + tag
}

// parseLatestTag extracts tag_name from the Codeberg releases JSON array.
func parseLatestTag(body string) (string, error) {
	var releases []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal([]byte(body), &releases); err != nil {
		return "", fmt.Errorf("không đọc được danh sách bản phát hành: %w", err)
	}
	if len(releases) == 0 || releases[0].TagName == "" {
		return "", fmt.Errorf("danh sách bản phát hành rỗng")
	}
	return releases[0].TagName, nil
}

// resolveGalleryDlTag returns the newest gallery-dl tag, falling back to GalleryDlVersion.
func resolveGalleryDlTag() (string, error) {
	body, err := fetchText(galleryDlAPIURL())
	if err != nil {
		return GalleryDlVersion, nil
	}
	tag, err := parseLatestTag(body)
	if err != nil {
		return GalleryDlVersion, nil
	}
	return tag, nil
}

// resolveGalleryDlChecksum returns the published sha256 for assetName at tag.
// An empty result means the caller must refuse to install.
func resolveGalleryDlChecksum(tag, assetName string) string {
	sums, err := fetchText(galleryDlReleaseBase(tag) + "/SHA256SUMS")
	if err != nil {
		return ""
	}
	if sum := parseChecksum(sums, assetName); sum != "" {
		return sum
	}
	// Last resort: the pinned table, valid only for the pinned tag.
	if tag == GalleryDlVersion {
		return galleryDlSHA256[assetName]
	}
	return ""
}

// galleryDlSHA256 mirrors the SHA256SUMS of the pinned fallback release.
var galleryDlSHA256 = map[string]string{
	"gallery-dl.bin": "94afae4d9144fa661d12f8689311164b595de65b30845c657dcc353a14d562a7",
	"gallery-dl.exe": "605ea88b32f77cfc1f88e4d1689294ac0826c0b6a72da4ca935a27a782b63df3",
}

// installGalleryDl downloads gallery-dl at the newest available tag into destPath.
func installGalleryDl(destPath string, cb ProgressCallback) error {
	tag, _ := resolveGalleryDlTag()
	asset := GetGalleryDlAssetName()

	sum := resolveGalleryDlChecksum(tag, asset)
	if sum == "" {
		return fmt.Errorf("không lấy được checksum cho gallery-dl %s, từ chối cài đặt", tag)
	}

	if err := downloadBinary(galleryDlDownloadURLFor(tag), destPath, "gallery-dl", cb, sum); err != nil {
		return err
	}
	return os.Chmod(destPath, 0755)
}

// GetGalleryDlDownloadURL returns the download URL for gallery-dl standalone binary.
// GetGalleryDlDownloadURL returns the download URL for gallery-dl standalone binary.
func GetGalleryDlDownloadURL() string {
	return galleryDlDownloadURLFor(GalleryDlVersion)
}

// galleryDlDownloadURLFor returns the asset URL for a specific tag.
func galleryDlDownloadURLFor(tag string) string {
	if runtime.GOOS == "windows" {
		return galleryDlReleaseBase(tag) + "/gallery-dl.exe"
	}
	return galleryDlReleaseBase(tag) + "/gallery-dl.bin"
}

// GetYtDlpChecksumsURL returns the SHA2-256SUMS file published with the same yt-dlp release.
func GetYtDlpChecksumsURL() string {
	return "https://github.com/yt-dlp/yt-dlp/releases/latest/download/SHA2-256SUMS"
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

// fetchText downloads a small text file (checksum manifests, release metadata).
// It is a variable so tests can stub the network.
var fetchText = func(url string) (string, error) {
	client := &http.Client{Transport: util.SharedTransport}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mã lỗi HTTP: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ytDlpAssetName returns the release asset file name for the current OS.
func ytDlpAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return "yt-dlp.exe"
	case "darwin":
		return "yt-dlp_macos"
	default:
		return "yt-dlp"
	}
}

// FindBinaries locates yt-dlp, ffmpeg, and gallery-dl either in PATH or in the app's bin directory.
func FindBinaries() (BinaryPaths, []string) {
	var paths BinaryPaths
	var missing []string

	ytName, ffName := GetBinaryNames()
	gName := GetGalleryDlBinaryName()
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

	// 3. Find gallery-dl: check binDir candidate first before exec.LookPath
	if binDir != "" {
		candidate := filepath.Join(binDir, gName)
		if fileExists(candidate) {
			paths.GalleryDl = candidate
		}
	}
	if paths.GalleryDl == "" {
		if p, err := exec.LookPath(gName); err == nil {
			paths.GalleryDl = p
		}
	}
	if paths.GalleryDl == "" {
		missing = append(missing, "gallery-dl")
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
	sum, sumErr := ytDlpChecksum()
	if sumErr != nil {
		return "", fmt.Errorf("lỗi khi tải danh sách checksum yt-dlp: %w", sumErr)
	}
	if err := downloadBinary(downloadURL, tmpPath, "yt-dlp", cb, sum); err != nil {
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
	gName := GetGalleryDlBinaryName()

	for _, item := range missing {
		switch item {
		case "yt-dlp":
			downloadURL := GetYtDlpDownloadURL()
			destPath := filepath.Join(binDir, ytName)
			sum, sumErr := ytDlpChecksum()
			if sumErr != nil {
				return paths, fmt.Errorf("lỗi khi tải danh sách checksum yt-dlp: %w", sumErr)
			}
			if err := downloadBinary(downloadURL, destPath, "yt-dlp", cb, sum); err != nil {
				return paths, fmt.Errorf("lỗi khi tải yt-dlp: %w", err)
			}
			_ = os.Chmod(destPath, 0755)
			paths.YtDlp = destPath

		case "gallery-dl":
			destPath := filepath.Join(binDir, gName)
			if err := installGalleryDl(destPath, cb); err != nil {
				return paths, fmt.Errorf("lỗi khi tải gallery-dl: %w", err)
			}
			paths.GalleryDl = destPath

		case "ffmpeg":
			if runtime.GOOS == "windows" {
				const zipName = "ffmpeg-master-latest-win64-gpl.zip"
				zipURL := "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/" + zipName
				tempZip := filepath.Join(binDir, "ffmpeg_temp.zip")
				sum, sumErr := ffmpegBuildChecksum(zipName)
				if sumErr != nil {
					return paths, fmt.Errorf("lỗi khi tải danh sách checksum ffmpeg: %w", sumErr)
				}
				if err := downloadBinary(zipURL, tempZip, "ffmpeg", cb, sum); err != nil {
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

// verifyFileChecksum confirms that path hashes to expectedSHA256 (hex).
// An empty expected value is a hard error: callers must know the hash they expect.
func verifyFileChecksum(path, expectedSHA256 string) error {
	if expectedSHA256 == "" {
		return fmt.Errorf("không có checksum để kiểm tra toàn vẹn tệp %q", filepath.Base(path))
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, strings.TrimSpace(expectedSHA256)) {
		return fmt.Errorf("checksum không khớp cho %q: mong đợi %s, nhận %s", filepath.Base(path), expectedSHA256, got)
	}
	return nil
}

// downloadVerified downloads url to destPath and refuses to install it unless the
// bytes hash to expectedSHA256.
func downloadVerified(url, destPath, itemName string, cb ProgressCallback, expectedSHA256 string) error {
	tmpFile := destPath + ".tmp"
	if err := downloadToFile(url, tmpFile, itemName, cb); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	if err := verifyFileChecksum(tmpFile, expectedSHA256); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	if err := os.Rename(tmpFile, destPath); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	return nil
}

func downloadToFile(url, filePath, itemName string, cb ProgressCallback) error {
	client := &http.Client{
		Transport: util.SharedTransport,
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

func extractFFmpegFromZip(zipPath, targetDir string) error {
	cleanTargetDir := filepath.Clean(targetDir)
	if err := os.MkdirAll(cleanTargetDir, 0755); err != nil {
		return err
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	found := false
	for _, f := range r.File {
		cleanName := filepath.Base(f.Name)
		if strings.EqualFold(cleanName, "ffmpeg.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			cleanDest := filepath.Clean(filepath.Join(targetDir, cleanName))
			if !strings.HasPrefix(cleanDest, cleanTargetDir+string(filepath.Separator)) && cleanDest != cleanTargetDir {
				return fmt.Errorf("đường dẫn không an toàn")
			}

			outFile, err := os.OpenFile(cleanDest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, copyErr := io.Copy(outFile, rc)
			if copyErr != nil {
				return copyErr
			}

			found = true
		}
	}

	if !found {
		return fmt.Errorf("không tìm thấy tệp ffmpeg.exe trong file zip")
	}

	return nil
}

// ytDlpChecksum fetches the SHA2-256SUMS published with the latest yt-dlp release
// and returns the digest for this platform's asset. yt-dlp tracks releases/latest,
// so the manifest is the only up-to-date source of truth.
func ytDlpChecksum() (string, error) {
	manifest, err := fetchText(GetYtDlpChecksumsURL())
	if err != nil {
		return "", err
	}
	asset := ytDlpAssetName()
	sum := parseChecksum(manifest, asset)
	if sum == "" {
		return "", fmt.Errorf("không tìm thấy checksum cho %q trong SHA2-256SUMS", asset)
	}
	return sum, nil
}

// ffmpegBuildChecksum fetches the checksums.sha256 manifest published alongside the
// latest yt-dlp FFmpeg-Builds archive and returns the digest for zipName.
func ffmpegBuildChecksum(zipName string) (string, error) {
	const manifestURL = "https://github.com/yt-dlp/FFmpeg-Builds/releases/download/latest/checksums.sha256"
	manifest, err := fetchText(manifestURL)
	if err != nil {
		return "", err
	}
	sum := parseChecksum(manifest, zipName)
	if sum == "" {
		return "", fmt.Errorf("không tìm thấy checksum cho %q trong checksums.sha256", zipName)
	}
	return sum, nil
}
