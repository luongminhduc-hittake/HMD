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
