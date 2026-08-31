package sidecars

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetBinDir(t *testing.T) {
	dir := GetBinDir()
	if dir == "" {
		t.Errorf("GetBinDir returned empty string")
	}
}

func TestResolveYTDLPPathWithHint(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "yt-dlp.exe")
	if err := os.WriteFile(customPath, []byte("fake-ytdlp"), 0755); err != nil {
		t.Fatal(err)
	}

	p, src := ResolveYTDLPPath(customPath)
	if p != customPath || src != "custom" {
		t.Errorf("ResolveYTDLPPath(%q) = (%q, %q); want (%q, %q)", customPath, p, src, customPath, "custom")
	}
}

func TestResolveFFmpegPathWithHint(t *testing.T) {
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "ffmpeg.exe")
	if err := os.WriteFile(customPath, []byte("fake-ffmpeg"), 0755); err != nil {
		t.Fatal(err)
	}

	p, src := ResolveFFmpegPath(customPath)
	if p != customPath || src != "custom" {
		t.Errorf("ResolveFFmpegPath(%q) = (%q, %q); want (%q, %q)", customPath, p, src, customPath, "custom")
	}
}
