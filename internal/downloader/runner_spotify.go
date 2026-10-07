package downloader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"hmd/internal/util"
)

// runSpotifyDownload orchestrates downloading Spotify tracks by matching them on YouTube
// and running standard downloads with ID3/cover tagging.
func runSpotifyDownload(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	// If single track and not downloading as playlist
	if len(opts.SpotifyTracks) == 1 && !opts.DownloadPlaylist {
		track := opts.SpotifyTracks[0]
		singleOpts := opts
		singleOpts.IsSpotify = false
		singleOpts.CustomFilename = fmt.Sprintf("%s - %s", track.Artist, track.Title)
		if opts.OnStatus != nil {
			opts.OnStatus(fmt.Sprintf("Đang tìm & tải: %s - %s", track.Artist, track.Title))
		}

		// Tier 1: ytsearch1:... audio
		singleOpts.URL = fmt.Sprintf("ytsearch1:%s - %s audio", track.Artist, track.Title)
		res, err := runStandardDownload(ctx, singleOpts)
		if err != nil || res == nil || res.FilePath == "" {
			// Tier 2: ytsearch1:...
			singleOpts.URL = fmt.Sprintf("ytsearch1:%s - %s", track.Artist, track.Title)
			res, err = runStandardDownload(ctx, singleOpts)
		}
		if err != nil {
			return nil, fmt.Errorf("không thể tìm hoặc tải bài hát %q: %w", track.Title, err)
		}

		// Inject metadata & cover
		_ = InjectMetadata(ctx, opts.FFmpegPath, res.FilePath, track)
		if fi, err := os.Stat(res.FilePath); err == nil {
			res.FileSize = fi.Size()
			res.FormattedSize = util.FormatBytes(fi.Size())
		}
		res.Title = fmt.Sprintf("%s - %s", track.Artist, track.Title)
		return res, nil
	}

	// Multiple tracks / Album / Playlist (Concurrent Worker Pool)
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = util.GetDefaultDownloadDir()
	}
	if opts.Title != "" {
		outputDir = filepath.Join(outputDir, util.SanitizeFilename(opts.Title))
	}
	_ = os.MkdirAll(outputDir, 0755)

	total := len(opts.SpotifyTracks)
	concurrency := opts.ConcurrentTracks
	if concurrency <= 0 {
		concurrency = 3
	}
	if concurrency > 5 {
		concurrency = 5
	}
	if total < concurrency {
		concurrency = total
	}

	var (
		mu              sync.Mutex
		downloadedCount int
		skippedCount    int
		completedCount  int
		failedTracks    []string
		wg              sync.WaitGroup
	)

	jobs := make(chan int, total)
	for i := 0; i < total; i++ {
		jobs <- i
	}
	close(jobs)

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				if ctx.Err() != nil {
					return
				}
				track := opts.SpotifyTracks[idx]

				if !track.IsPlayable {
					mu.Lock()
					skippedCount++
					completedCount++
					failedTracks = append(failedTracks, fmt.Sprintf("%s - %s (bị khóa bản quyền trên Spotify)", track.Artist, track.Title))
					mu.Unlock()
					continue
				}

				mu.Lock()
				if opts.OnStatus != nil {
					opts.OnStatus(fmt.Sprintf("[%d/%d] Đang tải: %s - %s", completedCount+1, total, track.Artist, track.Title))
				}
				mu.Unlock()

				trackOpts := opts
				trackOpts.IsSpotify = false
				trackOpts.OutputDir = outputDir
				trackOpts.DownloadPlaylist = false
				trackOpts.CustomFilename = fmt.Sprintf("%s - %s", track.Artist, track.Title)
				trackOpts.OnProgress = nil

				// Tier 1: ytsearch1:... audio
				trackOpts.URL = fmt.Sprintf("ytsearch1:%s - %s audio", track.Artist, track.Title)
				singleRes, err := runStandardDownload(ctx, trackOpts)
				if err != nil || singleRes == nil || singleRes.FilePath == "" {
					// Tier 2: ytsearch1:...
					trackOpts.URL = fmt.Sprintf("ytsearch1:%s - %s", track.Artist, track.Title)
					singleRes, err = runStandardDownload(ctx, trackOpts)
				}

				mu.Lock()
				completedCount++
				if err != nil || singleRes == nil || singleRes.FilePath == "" {
					skippedCount++
					failedTracks = append(failedTracks, fmt.Sprintf("%s - %s (không tìm thấy trên YouTube)", track.Artist, track.Title))
				} else {
					downloadedCount++
					_ = InjectMetadata(ctx, opts.FFmpegPath, singleRes.FilePath, track)
				}
				if opts.OnProgress != nil {
					opts.OnProgress(ProgressUpdate{
						Percent:       float64(completedCount) / float64(total) * 100,
						StatusMessage: fmt.Sprintf("[%d/%d] Đã xong %d bài", completedCount, total, downloadedCount),
					})
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return &DownloadResult{
		FilePath:        outputDir,
		FileName:        filepath.Base(outputDir),
		Title:           opts.Title,
		IsPlaylist:      true,
		DownloadedCount: downloadedCount,
		SkippedCount:    skippedCount,
		FailedTracks:    failedTracks,
	}, nil
}
