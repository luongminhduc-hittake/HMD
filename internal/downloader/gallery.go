package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"hmd/internal/deps"
	"hmd/internal/util"
)

// FetchGalleryInfo queries gallery-dl for image or photo gallery metadata.
func FetchGalleryInfo(ctx context.Context, rawURL string, cookiesBrowser ...string) (*MediaInfo, error) {
	galleryDlPath := ""
	paths, _ := deps.FindBinaries()
	if paths.GalleryDl != "" {
		galleryDlPath = paths.GalleryDl
	}
	if galleryDlPath == "" {
		if p, err := exec.LookPath("gallery-dl"); err == nil {
			galleryDlPath = p
		}
	}
	if galleryDlPath == "" {
		return nil, fmt.Errorf("không tìm thấy công cụ gallery-dl")
	}

	args := []string{"-j"}
	var cookies string
	if len(cookiesBrowser) > 0 {
		cookies = strings.TrimSpace(cookiesBrowser[0])
	}
	if cookies != "" && util.IsAllowedBrowser(cookies) {
		args = append(args, "--cookies-from-browser", cookies)
	}
	args = append(args, "--", rawURL)

	gCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := execCommandContext(gCtx, galleryDlPath, args...)
	prepareCommand(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr != "" {
			return nil, fmt.Errorf("gallery-dl: %s", errStr)
		}
		return nil, err
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil, fmt.Errorf("gallery-dl: không tìm thấy dữ liệu ảnh")
	}

	var rawData []json.RawMessage
	if err := json.Unmarshal([]byte(output), &rawData); err != nil {
		return nil, fmt.Errorf("gallery-dl: lỗi đọc dữ liệu JSON: %w", err)
	}

	title := ""
	uploader := ""
	thumbnailURL := ""
	imageCount := 0

	for _, elem := range rawData {
		var subList []json.RawMessage
		if err := json.Unmarshal(elem, &subList); err == nil && len(subList) >= 2 {
			var meta map[string]interface{}
			if err := json.Unmarshal(subList[1], &meta); err == nil {
				if t, ok := meta["title"].(string); ok && t != "" && title == "" {
					title = strings.TrimSpace(t)
				}
				if desc, ok := meta["description"].(string); ok && desc != "" && title == "" {
					title = strings.TrimSpace(desc)
				}
				if u, ok := meta["author"].(string); ok && u != "" && uploader == "" {
					uploader = strings.TrimSpace(u)
				}
				if pinner, ok := meta["pinner"].(map[string]interface{}); ok && uploader == "" {
					if fn, ok := pinner["full_name"].(string); ok && fn != "" {
						uploader = strings.TrimSpace(fn)
					}
				}
				if u, ok := meta["uploader"].(string); ok && u != "" && uploader == "" {
					uploader = strings.TrimSpace(u)
				}
				if user, ok := meta["user"].(map[string]interface{}); ok && uploader == "" {
					if name, ok := user["name"].(string); ok && name != "" {
						uploader = strings.TrimSpace(name)
					}
				}
				if imgURL, ok := meta["url"].(string); ok && imgURL != "" {
					imageCount++
					if thumbnailURL == "" {
						thumbnailURL = imgURL
					}
				}
			}
		}
	}

	if title == "" {
		p := DetectPlatform(rawURL)
		title = fmt.Sprintf("Ảnh %s", p.Name)
	}
	if uploader == "" {
		p := DetectPlatform(rawURL)
		uploader = p.Name
	}

	duration := "Hình ảnh"
	if imageCount > 1 {
		duration = fmt.Sprintf("Album %d ảnh", imageCount)
	}

	return &MediaInfo{
		Title:            title,
		Duration:         duration,
		Uploader:         uploader,
		RawURL:           rawURL,
		IsImage:          true,
		ThumbnailURL:     thumbnailURL,
		MediaType:        MediaTypeImage,
		AvailablePresets: ImagePresets,
	}, nil
}

// executeGalleryDlDownload handles downloading images/galleries using gallery-dl.
func executeGalleryDlDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	if opts.OutputDir == "" {
		opts.OutputDir = util.GetDefaultDownloadDir()
	}
	_ = os.MkdirAll(opts.OutputDir, 0755)

	galleryDlPath := opts.GalleryDlPath
	if galleryDlPath == "" {
		paths, _ := deps.FindBinaries()
		galleryDlPath = paths.GalleryDl
	}
	if galleryDlPath == "" {
		if p, err := exec.LookPath("gallery-dl"); err == nil {
			galleryDlPath = p
		}
	}
	if galleryDlPath == "" {
		return nil, fmt.Errorf("không tìm thấy gallery-dl. Vui lòng cài đặt gallery-dl")
	}

	args := []string{
		"-D", opts.OutputDir,
	}

	if opts.CookiesBrowser != "" && util.IsAllowedBrowser(opts.CookiesBrowser) {
		args = append(args, "--cookies-from-browser", opts.CookiesBrowser)
	}
	args = append(args, "--", opts.URL)

	if opts.OnStatus != nil {
		opts.OnStatus("Đang tải ảnh bằng gallery-dl...")
	}

	parser := NewGalleryParser(opts)
	err := streamCommand(ctx, galleryDlPath, args, parser.FeedLine)
	if err != nil && len(parser.downloadedFiles) == 0 {
		return nil, fmt.Errorf("lỗi tải bằng gallery-dl: %w", err)
	}

	return parser.Result()
}
