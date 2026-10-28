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

type msgDepsStarted struct {
	ch chan tea.Msg
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

func startSelfUpdateCmd(assetURL, checksum string, size int64) (tea.Cmd, chan tea.Msg) {
	ch := make(chan tea.Msg, 50)
	go func() {
		err := updater.ApplyUpdate(assetURL, checksum, size, func(dl, total int64, pct float64) {
			select {
			case ch <- msgUpdateProgress{downloaded: dl, total: total, percent: pct}:
			default:
			}
		})
		select {
		case ch <- msgUpdateFinished{err: err}:
		default:
		}
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
		select {
		case ch <- msgYtDlpUpdateFinished{path: newPath, err: err}:
		default:
		}
		close(ch)
	}()
	return waitForChannel(ch), ch
}

// initDepsCmd checks for the required binaries and installs the missing ones,
// streaming progress so the "checking system" screen is not frozen while a
// 24 MB gallery-dl or 100 MB ffmpeg download runs.
func initDepsCmd() tea.Cmd {
	return func() tea.Msg {
		ch := make(chan tea.Msg, 64)
		go func() {
			defer close(ch)

			paths, missing := deps.FindBinaries()
			if len(missing) == 0 {
				ch <- msgDepsReady{paths: paths}
				return
			}

			ch <- msgDepsProgress{item: missing[0]}

			readyPaths, err := deps.EnsureDependencies(func(name string, dl, total int64, pct float64) {
				select {
				case ch <- msgDepsProgress{item: name, downloaded: dl, total: total, percent: pct}:
				default: // a dropped progress tick is fine; the next one follows immediately
				}
			})
			ch <- msgDepsReady{paths: readyPaths, err: err}
		}()
		return msgDepsStarted{ch: ch}
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
		defer close(ch)
		res, err := downloader.ExecuteDownload(ctx, opts)
		// Progress callbacks drop messages when the buffer is full, so a non-blocking
		// send could lose the only completion message. Block instead: the reader is
		// always draining, and ctx.Done covers the case where the user quits first.
		select {
		case <-ctx.Done():
		case ch <- msgDownloadCompleted{result: res, err: err}:
		}
	}()
	return waitForChannel(ch), ch
}
