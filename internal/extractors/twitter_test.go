package extractors

import (
	"context"
	"testing"
	"time"
)

func TestParseTwitterStatus(t *testing.T) {
	tests := []struct {
		url      string
		wantUser string
		wantID   string
		wantOK   bool
	}{
		{
			url:      "https://x.com/rachelc00k/status/1858188535611740220/video/1",
			wantUser: "rachelc00k",
			wantID:   "1858188535611740220",
			wantOK:   true,
		},
		{
			url:      "https://twitter.com/fridaysub/status/2053386027499503918",
			wantUser: "fridaysub",
			wantID:   "2053386027499503918",
			wantOK:   true,
		},
		{
			url:      "https://example.com/not-twitter",
			wantUser: "",
			wantID:   "",
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		u, id, ok := parseTwitterStatus(tt.url)
		if ok != tt.wantOK || u != tt.wantUser || id != tt.wantID {
			t.Errorf("parseTwitterStatus(%s) = (%s, %s, %v), want (%s, %s, %v)",
				tt.url, u, id, ok, tt.wantUser, tt.wantID, tt.wantOK)
		}
	}
}

func TestAnalyzeTwitterLive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	meta, err := AnalyzeTwitter(ctx, "https://x.com/rachelc00k/status/1858188535611740220/video/1")
	if err != nil {
		t.Logf("Live AnalyzeTwitter skipped or returned err: %v", err)
		return
	}

	if meta.Title == "" {
		t.Errorf("Expected non-empty title")
	}
	if len(meta.Formats) == 0 {
		t.Errorf("Expected formats")
	}
}
