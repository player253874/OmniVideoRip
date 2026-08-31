package extractors

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLiveRouVideoAnalyze(t *testing.T) {
	if os.Getenv("OMNIVIDEORIP_LIVE_TESTS") != "1" {
		t.Skip("set OMNIVIDEORIP_LIVE_TESTS=1 to run live site checks")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	meta, err := AnalyzeRouVideo(ctx, "https://rou.video/v/cmt6sx2gg0000s6zm7hpzvx3x")
	if err != nil {
		t.Fatalf("AnalyzeRouVideo failed: %v", err)
	}
	if meta.Platform != "rouvideo" || meta.Title == "" {
		t.Fatalf("unexpected RouVideo metadata: %+v", meta)
	}
	t.Logf("Metadata: Title=%q, Duration=%q", meta.Title, meta.Duration)
}
func TestLiveRouReferenceStream(t *testing.T) {
	if os.Getenv("OMNIVIDEORIP_LIVE_TESTS") != "1" {
		t.Skip("set OMNIVIDEORIP_LIVE_TESTS=1 to run live site checks")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := newSafeHTTPClient(30 * time.Second)
	refURL := "https://ppp.porn/v/yw6bqv/"
	body, finalURL, err := fetchBytesWithFinalURL(ctx, client, refURL, "https://rou.video/", maxPageBytes)
	if err != nil {
		t.Fatalf("fetch reference page failed: %v", err)
	}
	match := rouReferenceStreamRegex.FindSubmatch(body)
	if len(match) < 2 {
		t.Fatal("reference page did not expose an HLS stream")
	}
	streamURL, err := resolveURL(finalURL, string(match[1]))
	if err != nil {
		t.Fatal(err)
	}
	playlist, err := fetchBytes(ctx, client, streamURL, finalURL, maxPlaylistBytes)
	if err != nil {
		t.Fatalf("fetch reference playlist failed: %v", err)
	}
	if !strings.Contains(string(playlist), "#EXTM3U") {
		t.Fatal("reference stream did not return an HLS playlist")
	}
	t.Logf("Reference HLS: %s", streamURL)
}
