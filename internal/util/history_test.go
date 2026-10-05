package util

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestHistoryOperations(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)

	// Initially empty
	entries := LoadHistory()
	if len(entries) != 0 {
		t.Fatalf("expected empty history, got %d", len(entries))
	}

	// Add entry 1
	e1 := HistoryEntry{
		URL:       "https://youtube.com/watch?v=abc12345",
		Title:     "Test Video 1",
		Format:    "best",
		FilePath:  "/tmp/video1.mp4",
		FileSize:  1024,
		CreatedAt: time.Now(),
	}
	if err := AddHistoryEntry(e1); err != nil {
		t.Fatalf("failed to add entry: %v", err)
	}

	// Verify FindHistoryByURL
	found := FindHistoryByURL("https://youtube.com/watch?v=abc12345")
	if found == nil || found.Title != "Test Video 1" {
		t.Fatalf("expected to find entry 1, got %+v", found)
	}

	// Not found
	if notFound := FindHistoryByURL("https://youtube.com/watch?v=notexist"); notFound != nil {
		t.Fatalf("expected nil for non-existent url, got %+v", notFound)
	}

	// Add entry 2
	e2 := HistoryEntry{
		URL:      "https://youtube.com/watch?v=xyz98765",
		Title:    "Test Video 2",
		Format:   "mp3",
		FilePath: "/tmp/video2.mp3",
		FileSize: 2048,
	}
	if err := AddHistoryEntry(e2); err != nil {
		t.Fatalf("failed to add entry 2: %v", err)
	}

	entries = LoadHistory()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Title != "Test Video 2" {
		t.Errorf("expected newest entry first, got %q", entries[0].Title)
	}

	// Re-add entry 1 with updated title (deduplication check)
	e1Updated := e1
	e1Updated.Title = "Test Video 1 Updated"
	if err := AddHistoryEntry(e1Updated); err != nil {
		t.Fatalf("failed to update entry 1: %v", err)
	}
	entries = LoadHistory()
	if len(entries) != 2 {
		t.Fatalf("expected still 2 entries after deduplication, got %d", len(entries))
	}
	if entries[0].Title != "Test Video 1 Updated" {
		t.Errorf("expected updated entry to be at top, got %q", entries[0].Title)
	}

	// Delete index 0
	if err := DeleteHistoryEntry(0); err != nil {
		t.Fatalf("failed to delete entry at 0: %v", err)
	}
	entries = LoadHistory()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry left, got %d", len(entries))
	}
	if entries[0].Title != "Test Video 2" {
		t.Errorf("expected Test Video 2 remaining, got %q", entries[0].Title)
	}
}

func TestHistoryLimit(t *testing.T) {
	tmpDir := t.TempDir()
	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", origConfigHome)

	for i := 0; i < 110; i++ {
		e := HistoryEntry{
			URL:      fmt.Sprintf("https://youtube.com/watch?v=%04d", i),
			Title:    fmt.Sprintf("Video %d", i),
			FilePath: fmt.Sprintf("/tmp/v%d.mp4", i),
		}
		_ = AddHistoryEntry(e)
	}

	entries := LoadHistory()
	if len(entries) != 100 {
		t.Fatalf("expected max 100 entries, got %d", len(entries))
	}
	if entries[0].Title != "Video 109" {
		t.Errorf("expected latest Video 109 at index 0, got %q", entries[0].Title)
	}
}
