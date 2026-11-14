package extractors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "youtube"},
		{"https://youtu.be/dQw4w9WgXcQ", "youtube"},
		{"https://www.bilibili.com/video/BV1xx411c7mD", "bilibili"},
		{"https://b23.tv/BV1xx411c7mD", "bilibili"},
		{"https://missav.live/dm26/ja/jac-082", "missav"},
		{"https://missav.ws/dm26/ja/jac-082", "missav"},
		{"https://ggjav.tv/main/video?id=150804", "ggjav"},
		{"https://rou.video/v/cmt6sx2gg0000s6zm7hpzvx3x", "rouvideo"},
		{"https://www.tiktok.com/@user/video/1234567890", "tiktok"},
		{"https://x.com/username/status/1234567890", "x"},
		{"https://www.instagram.com/reel/C321abc/", "instagram"},
		{"https://example.com/video.mp4", "example.com"},
		{"https://spankbang.com/video123", "spankbang.com"},
	}

	for _, tt := range tests {
		got := detectPlatform(tt.url)
		if got != tt.expected {
			t.Errorf("detectPlatform(%s) = %s; want %s", tt.url, got, tt.expected)
		}
	}
}

func TestDetectPlatformRejectsLookalikeHosts(t *testing.T) {
	tests := []struct {
		url         string
		wrongTarget string
	}{
		{"https://ggjav.tv.attacker.example/video", "ggjav"},
		{"https://missav.live.attacker.example/video", "missav"},
		{"https://notyoutube.com/video", "youtube"},
		{"https://twitter.com.attacker.example/video", "x"},
	}

	for _, tt := range tests {
		if got := detectPlatform(tt.url); got == tt.wrongTarget {
			t.Errorf("detectPlatform(%q) unexpectedly returned %q", tt.url, got)
		}
	}
}

func TestValidateHTTPURLRejectsUnsafeTargets(t *testing.T) {
	tests := []string{
		"file:///C:/Windows/win.ini",
		"ftp://example.com/video",
		"http://localhost/video",
		"http://127.0.0.1/video",
		"http://[::1]/video",
		"http://169.254.169.254/latest/meta-data",
	}

	for _, rawURL := range tests {
		if err := validateHTTPURL(rawURL); err == nil {
			t.Errorf("validateHTTPURL(%q) unexpectedly succeeded", rawURL)
		}
	}
}

func TestDownloadSegmentsFailsWhenAnySegmentFails(t *testing.T) {
	dir := t.TempDir()
	_, err := downloadSegments(context.Background(), []string{"https://example.com/1.ts", "https://example.com/2.ts"}, dir, 2, nil,
		func(_ context.Context, _ string, destPath string) error {
			if strings.HasSuffix(destPath, "00000.ts") {
				return os.WriteFile(destPath, []byte("segment"), 0600)
			}
			return os.ErrNotExist
		})
	if err == nil {
		t.Fatal("downloadSegments unexpectedly succeeded with a missing segment")
	}
}

func TestMergeSegmentsWithoutFFmpegReturnsTSFile(t *testing.T) {
	dir := t.TempDir()
	segments := []string{filepath.Join(dir, "one.ts"), filepath.Join(dir, "two.ts")}
	if err := os.WriteFile(segments[0], []byte("one"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(segments[1], []byte("two"), 0600); err != nil {
		t.Fatal(err)
	}

	outputPath, err := mergeSegments(context.Background(), segments, dir, "video", "")
	if err != nil {
		t.Fatalf("mergeSegments failed: %v", err)
	}
	if filepath.Ext(outputPath) != ".ts" {
		t.Fatalf("mergeSegments output = %q; want a .ts file without ffmpeg", outputPath)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "onetwo" {
		t.Fatalf("merged contents = %q; want onetwo", data)
	}
}

func TestParseYTDLPOutputPath(t *testing.T) {
	path, ok := parseYTDLPOutputPath("__OMNIVIDEORIP_PATH__:C:\\Downloads\\video.mp4")
	if !ok || path != `C:\Downloads\video.mp4` {
		t.Fatalf("parseYTDLPOutputPath returned %q, %v", path, ok)
	}
}

func TestParseBestQualityM3U8(t *testing.T) {
	playlist := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360
640x360/video.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1400000,RESOLUTION=842x480
842x480/video.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2800000,RESOLUTION=1280x720
1280x720/video.m3u8`

	got := parseBestQualityM3U8(playlist)
	if got != "1280x720/video.m3u8" {
		t.Errorf("Expected '1280x720/video.m3u8', got '%s'", got)
	}
}

func TestParseSegmentURLsRejectsUnsupportedPlaylistFeatures(t *testing.T) {
	for _, tag := range []string{
		`#EXT-X-KEY:METHOD=AES-128,URI="key.bin"`,
		`#EXT-X-MAP:URI="init.mp4"`,
		`#EXT-X-BYTERANGE:1000@0`,
	} {
		playlist := "#EXTM3U\n" + tag + "\nsegment.ts\n"
		if _, err := parseSegmentURLs("https://cdn.example.com/video/index.m3u8", playlist); err == nil {
			t.Errorf("parseSegmentURLs unexpectedly accepted %s", tag)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	dirty := `JAC-082: Special/Video?id=123*&foo="bar"<baz>|end`
	cleaned := sanitizeFilename(dirty)

	forbidden := []string{"\\", "/", ":", "*", "?", `"`, "<", ">", "|"}
	for _, f := range forbidden {
		if strings.Contains(cleaned, f) {
			t.Errorf("Sanitization failed, string contains forbidden char '%s': %s", f, cleaned)
		}
	}
}

func TestDecryptRouEv(t *testing.T) {
	d := "sVisn5qbpYuoolhwWJ6qqqapcGVlrGSopGhqbGSur7BlnqKpZZmjqmyprmidnWZmZmapbLCjbZ6msKyuaa5lmaOqbKmuaJ2dZmZmZqlssKNtnqawrK5prmNtaGZln6Sam65koKaddaxzbFybrqZzZ21ubW1vaGpmZlyXq6qec25/qnqvsIRvjWtshqSOkIpmbo+Dd5qDh2hno45upnmZmWiOq3yMeIyjm5lYYliqnqujmIyKiouoolhwWJ6qqqapcGVlrGSopGhqbGSur7BlnqKpZZmjqmyprmidnWZmZmapbLCjbZ6msKyuaa5lqp6ro5ipZaqeq6OYpJefomSsqqp1rHNsXJuupnNnbW5tbWZsZmZmXJerqp5zl6ZpnJqNnKGgqWeuioGXoJV4hG6Jb4dnrYOkqqSdY2+Dfn1pbIJuoW5+i1iz"
	k := 54

	ev, err := decryptRouEv(d, k)
	if err != nil {
		t.Fatalf("decryptRouEv failed: %v", err)
	}
	if !strings.Contains(ev.VideoURL, "cmt6sx2gg0000s6zm7hpzvx3x") {
		t.Fatalf("Expected videoURL to contain video ID, got: %s", ev.VideoURL)
	}
}

func TestResolveRouRelativePlaylistURL(t *testing.T) {
	got, err := resolveURL("https://rou.video/v/video-id", "/api/hls/video-id")
	if err != nil {
		t.Fatalf("resolveURL failed: %v", err)
	}
	if got != "https://rou.video/api/hls/video-id" {
		t.Fatalf("resolved URL = %q", got)
	}
}

func TestRouReferenceStreamPattern(t *testing.T) {
	page := []byte(`<script>var stream = 'https://cdn.example/video/index.m3u8?token=abc&amp;x=1';</script>`)
	match := rouReferenceStreamRegex.FindSubmatch(page)
	if len(match) < 2 {
		t.Fatal("reference stream was not detected")
	}
	if !strings.Contains(string(match[1]), "index.m3u8") {
		t.Fatalf("unexpected stream URL: %q", match[1])
	}
}

func TestSegmentFileIsUsableRejectsHTML(t *testing.T) {
	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "error.ts")
	if err := os.WriteFile(htmlPath, []byte("<!DOCTYPE html><title>Not found</title>"), 0600); err != nil {
		t.Fatal(err)
	}
	if segmentFileIsUsable(htmlPath) {
		t.Fatal("HTML error page was accepted as a media segment")
	}

	mediaPath := filepath.Join(dir, "media.ts")
	media := make([]byte, 188)
	media[0] = 0x47
	if err := os.WriteFile(mediaPath, media, 0600); err != nil {
		t.Fatal(err)
	}
	if !segmentFileIsUsable(mediaPath) {
		t.Fatal("valid media payload was rejected")
	}
}

func TestResumeDirKeysMatchNewExtractors(t *testing.T) {
	for _, platform := range []string{"rouvideo"} {
		dir := GetTaskResumeDir(`C:\Downloads`, platform, "https://example.com/v/video-id")
		if !strings.Contains(strings.ToLower(dir), platform+"_video-id") {
			t.Fatalf("resume dir for %s = %q", platform, dir)
		}
	}
}

func TestGetTaskResumeDirNotInOutputDir(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "UserDownloads")
	resumePath := GetTaskResumeDir(outDir, "youtube", "https://www.youtube.com/watch?v=MJ443eAUlPs")
	rel, err := filepath.Rel(outDir, resumePath)
	if err == nil && !strings.HasPrefix(rel, "..") {
		t.Fatalf("resume dir %q should not be created inside user download folder %q", resumePath, outDir)
	}
}

func TestMoveFileAndLegacyCleanup(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "video.downloading")
	dst := filepath.Join(dstDir, "final.mp4")

	if err := os.WriteFile(src, []byte("mp4-payload"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := moveFile(src, dst); err != nil {
		t.Fatalf("moveFile failed: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source temp file %q still exists after move", src)
	}
	got, err := os.ReadFile(dst)
	if err != nil || string(got) != "mp4-payload" {
		t.Fatalf("destination file mismatch: got %q, err=%v", string(got), err)
	}

	legacy := legacyResumeDir(dstDir, "youtube", "https://www.youtube.com/watch?v=abc")
	if err := os.MkdirAll(legacy, 0755); err != nil {
		t.Fatal(err)
	}
	if err := CleanupTaskResumeCache(dstDir, "youtube", "https://www.youtube.com/watch?v=abc"); err != nil {
		t.Fatalf("CleanupTaskResumeCache failed: %v", err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy resume dir %q was not cleaned up", legacy)
	}
}
