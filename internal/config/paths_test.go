package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePaths(t *testing.T) {
	paths, err := ResolvePaths()
	if err != nil {
		t.Fatalf("ResolvePaths failed: %v", err)
	}

	if paths.DownloadsDir == "" {
		t.Errorf("Expected non-empty DownloadsDir")
	}

	if paths.BinDir == "" {
		t.Errorf("Expected non-empty BinDir")
	}
}

func TestEnsureDirectories(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "test_bin")
	downloadsDir := filepath.Join(tempDir, "test_downloads")

	p := Paths{
		BinDir:       binDir,
		DownloadsDir: downloadsDir,
	}

	err := EnsureDirectories(p)
	if err != nil {
		t.Fatalf("EnsureDirectories failed: %v", err)
	}

	if info, err := os.Stat(binDir); err != nil || !info.IsDir() {
		t.Errorf("Expected binDir to exist as a directory")
	}

	if info, err := os.Stat(downloadsDir); err != nil || !info.IsDir() {
		t.Errorf("Expected downloadsDir to exist as a directory")
	}
}
