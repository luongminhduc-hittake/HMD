package downloader

import (
	"testing"
)

func TestAvailablePresets(t *testing.T) {
	if len(AvailablePresets) != 9 {
		t.Fatalf("expected 9 available presets, got %d", len(AvailablePresets))
	}

	expectedIDs := map[FormatPreset]bool{
		PresetBestVideo: true,
		Preset1080p:     true,
		Preset720p:      true,
		PresetAudioMP3:  true,
		PresetAudioM4A:  true,
		PresetAudioWAV:  true,
		PresetAudioFLAC: true,
		PresetAudioOPUS: true,
		PresetThumbnail: true,
	}

	for _, p := range AvailablePresets {
		if !expectedIDs[p.ID] {
			t.Errorf("unexpected preset ID: %s", p.ID)
		}
		if p.Title == "" {
			t.Errorf("preset %s has empty title", p.ID)
		}
	}
}

func TestDetermineAvailablePresets(t *testing.T) {
	// 1. Image
	imgPresets := DetermineAvailablePresets(MediaTypeImage, 0)
	if len(imgPresets) != 1 || imgPresets[0].ID != PresetImageOriginal {
		t.Errorf("expected 1 image preset (%s), got %+v", PresetImageOriginal, imgPresets)
	}

	// 2. Audio
	audioPresets := DetermineAvailablePresets(MediaTypeAudio, 0)
	if len(audioPresets) != 6 {
		t.Errorf("expected 6 audio presets, got %d", len(audioPresets))
	}
	for _, p := range audioPresets {
		if p.ID == PresetBestVideo || p.ID == Preset1080p || p.ID == Preset720p {
			t.Errorf("audio presets must not contain video option %s", p.ID)
		}
	}

	// 3. Video 1080p
	v1080 := DetermineAvailablePresets(MediaTypeVideo, 1080)
	has1080 := false
	has720 := false
	for _, p := range v1080 {
		if p.ID == Preset1080p {
			has1080 = true
		}
		if p.ID == Preset720p {
			has720 = true
		}
	}
	if !has1080 || !has720 {
		t.Errorf("1080p video should have both 1080p and 720p presets")
	}

	// 4. Video 720p (MaxHeight 720)
	v720 := DetermineAvailablePresets(MediaTypeVideo, 720)
	for _, p := range v720 {
		if p.ID == Preset1080p {
			t.Errorf("720p video must NOT contain 1080p preset")
		}
	}

	// 5. Video 480p (MaxHeight 480)
	v480 := DetermineAvailablePresets(MediaTypeVideo, 480)
	for _, p := range v480 {
		if p.ID == Preset1080p || p.ID == Preset720p {
			t.Errorf("480p video must NOT contain 1080p or 720p preset, found %s", p.ID)
		}
	}
}
