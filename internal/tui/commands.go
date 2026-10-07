package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"hmd/internal/deps"
	"hmd/internal/downloader"
	"hmd/internal/updater"
	"hmd/internal/util"
)

// Internal message types
type msgDepsProgress struct {
	item       string
	downloaded int64
	total      int64
	percent    float64
}

type msgDepsReady struct {
	paths deps.BinaryPaths
	err   error
}

type msgInfoFetched struct {
	info *downloader.MediaInfo
	err  error
}

type msgDownloadProgress struct {
	progress downloader.ProgressUpdate
}

type msgDownloadStatus struct {
	status string
}

type msgDownloadCompleted struct {
	result *downloader.DownloadResult
	err    error
}

type msgUpdateCheckResult struct {
	info *updater.ReleaseInfo
	err  error
}

type msgUpdateProgress struct {
	downloaded int64
	total      int64
	percent    float64
}

type msgUpdateFinished struct {
	err error
}

type updateProgressMsg = tea.Msg

type msgYtDlpUpdateFinished struct {
	path string
	err  error
}

func waitForChannel(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := updater.CheckForUpdate()
		return msgUpdateCheckResult{info: info, err: err}
	}
}

func startSelfUpdateCmd(assetURL string) (tea.Cmd, chan tea.Msg) {
	ch := make(chan tea.Msg, 50)
	go func() {
		err := updater.ApplyUpdate(assetURL, func(dl, total int64, pct float64) {
			select {
			case ch <- msgUpdateProgress{downloaded: dl, total: total, percent: pct}:
			default:
			}
		})
		ch <- msgUpdateFinished{err: err}
		close(ch)
	}()
	return waitForChannel(ch), ch
}

func updateYtDlpCmd() (tea.Cmd, chan updateProgressMsg) {
	ch := make(chan updateProgressMsg, 50)
	go func() {
		newPath, err := deps.UpdateYtDlp(func(name string, dl, total int64, pct float64) {
			select {
			case ch <- msgUpdateProgress{downloaded: dl, total: total, percent: pct}:
			default:
			}
		})
		ch <- msgYtDlpUpdateFinished{path: newPath, err: err}
		close(ch)
	}()
	return waitForChannel(ch), ch
}

func initDepsCmd() tea.Cmd {
	return func() tea.Msg {
		paths, missing := deps.FindBinaries()
		if len(missing) == 0 {
			return msgDepsReady{paths: paths, err: nil}
		}

		// Download missing dependencies
		readyPaths, err := deps.EnsureDependencies(nil)
		return msgDepsReady{paths: readyPaths, err: err}
	}
}

func fetchInfoCmd(ytdlpPath, url, cookiesBrowser string) tea.Cmd {
	return func() tea.Msg {
		info, err := downloader.FetchInfo(ytdlpPath, url, cookiesBrowser)
		return msgInfoFetched{info: info, err: err}
	}
}

func startDownloadCmd(m *Model) (tea.Cmd, chan tea.Msg) {
	ch := make(chan tea.Msg, 50)
	ctx, cancel := context.WithCancel(context.Background())
	m.DownloadCancel = cancel

	preset := downloader.AvailablePresets[0].ID
	if len(m.AvailablePresets) > m.PresetIndex {
		preset = m.AvailablePresets[m.PresetIndex].ID
	} else if len(downloader.AvailablePresets) > m.PresetIndex {
		preset = downloader.AvailablePresets[m.PresetIndex].ID
	}
	rawURL := m.MediaInfo.RawURL
	outDir := m.OutputDir
	if outDir == "" {
		outDir = util.GetDefaultDownloadDir()
	}

	opts := downloader.DownloadOptions{
		YtDlpPath:           m.Paths.YtDlp,
		FFmpegPath:          m.Paths.FFmpeg,
		GalleryDlPath:       m.Paths.GalleryDl,
		URL:                 rawURL,
		Title:               m.MediaInfo.Title,
		OutputDir:           outDir,
		Preset:              preset,
		DownloadPlaylist:    m.DownloadPlaylist,
		TrimRange:           m.TrimRange,
		EnableSubtitles:     m.EnableSubtitles,
		CookiesBrowser:      m.CookiesBrowser,
		ConcurrentFragments: m.ConcurrentFragments,
		ConcurrentTracks:    3,
		MaxRetries:          m.MaxRetries,
		FragmentRetries:     m.FragmentRetries,
		IsSpotify:           m.MediaInfo.IsSpotify,
		IsImage:             m.MediaInfo.IsImage,
		SpotifyTracks:       m.MediaInfo.SpotifyTracks,
		OnProgress: func(pu downloader.ProgressUpdate) {
			select {
			case <-ctx.Done():
				return
			case ch <- msgDownloadProgress{progress: pu}:
			default:
			}
		},
		OnStatus: func(st string) {
			select {
			case <-ctx.Done():
				return
			case ch <- msgDownloadStatus{status: st}:
			default:
			}
		},
	}

	go func() {
		res, err := downloader.ExecuteDownload(ctx, opts)
		select {
		case <-ctx.Done():
		case ch <- msgDownloadCompleted{result: res, err: err}:
		}
		close(ch)
	}()

	return waitForChannel(ch), ch
}
