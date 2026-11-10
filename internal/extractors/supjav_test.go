package extractors

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDetectPlatformSupJav(t *testing.T) {
	cases := []struct {
		url      string
		expected string
	}{
		{"https://supjav.com/zh/390226.html", "supjav"},
		{"https://supjav.com/en/12345.html", "supjav"},
		{"https://www.supjav.com/390226.html", "supjav"},
	}
	for _, tc := range cases {
		if got := detectPlatform(tc.url); got != tc.expected {
			t.Errorf("detectPlatform(%q) = %q, want %q", tc.url, got, tc.expected)
		}
	}
}

func TestExtractSupJavEmbeds(t *testing.T) {
	html := `
	<div class="video-container">
		<div class="btn-server" data-link="https://faststream.in/embed-123">FST</div>
		<div class="btn-server" data-src="aHR0cHM6Ly9zdHJlYW10YXBlLmNvbS9lLzQ1Njc4">ST</div>
		<div class="btn-server" data-file="https://tvlogy.to/e/98765">TV</div>
		<iframe src="https://voe.sx/e/voe123"></iframe>
		<script>var stream = "https://cdn.example.com/hls/master.m3u8";</script>
	</div>
	`
	embeds := extractSupJavEmbeds(html)
	if len(embeds) < 5 {
		t.Fatalf("expected at least 5 embeds, got %d: %v", len(embeds), embeds)
	}

	hasFst := false
	hasSt := false
	hasTv := false
	hasVoe := false
	hasM3u8 := false

	for _, e := range embeds {
		if strings.Contains(e, "faststream.in") {
			hasFst = true
		}
		if strings.Contains(e, "streamtape.com") {
			hasSt = true
		}
		if strings.Contains(e, "tvlogy.to") {
			hasTv = true
		}
		if strings.Contains(e, "voe.sx") {
			hasVoe = true
		}
		if strings.Contains(e, "master.m3u8") {
			hasM3u8 = true
		}
	}

	if !hasFst || !hasSt || !hasTv || !hasVoe || !hasM3u8 {
		t.Errorf("missing expected embed candidates, got: %v", embeds)
	}
}

func TestParseSupJavMetadata(t *testing.T) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>IPX-999 超绝美少女 - SupJav</title>
		<meta property="og:image" content="https://img.supjav.com/covers/ipx999.jpg" />
	</head>
	<body>
		<h1 class="entry-title">IPX-999 超绝美少女</h1>
	</body>
	</html>
	`
	title, thumb := parseSupJavMetadata(html, "https://supjav.com/zh/390226.html")
	if title != "IPX-999 超绝美少女" {
		t.Errorf("title mismatch: got %q", title)
	}
	if thumb != "https://img.supjav.com/covers/ipx999.jpg" {
		t.Errorf("thumb mismatch: got %q", thumb)
	}
}

func TestIsCloudflareChallenge(t *testing.T) {
	header := make(http.Header)
	header.Set("Cf-Mitigated", "challenge")
	if !isCloudflareChallenge(http.StatusForbidden, "<html>Just a moment...</html>", header) {
		t.Errorf("expected cloudflare challenge detected")
	}

	header2 := make(http.Header)
	if !isCloudflareChallenge(http.StatusForbidden, "<html><title>Just a moment...</title></html>", header2) {
		t.Errorf("expected cloudflare challenge detected by title")
	}

	if isCloudflareChallenge(http.StatusOK, "<html>normal</html>", header2) {
		t.Errorf("normal page falsely marked as cloudflare challenge")
	}
}

type supjavRoundTrip func(*http.Request) (*http.Response, error)

func (f supjavRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAnalyzeSupJavCloudflareError(t *testing.T) {
	client := &http.Client{
		Transport: supjavRoundTrip(func(r *http.Request) (*http.Response, error) {
			header := make(http.Header)
			header.Set("Cf-Mitigated", "challenge")
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader("<!DOCTYPE html><html><head><title>Just a moment...</title></head><body>Enable JavaScript and cookies to continue</body></html>")),
			}, nil
		}),
	}

	_, err := analyzeSupJavWithClient(context.Background(), client, "https://supjav.com/zh/390226.html")
	if err == nil {
		t.Fatal("expected error for Cloudflare challenge page, got nil")
	}
	if !strings.Contains(err.Error(), "Cloudflare Turnstile") {
		t.Errorf("expected Cloudflare Turnstile error message, got: %v", err)
	}
}

func TestAnalyzeSupJavSuccess(t *testing.T) {
	client := &http.Client{
		Transport: supjavRoundTrip(func(r *http.Request) (*http.Response, error) {
			html := `<!DOCTYPE html>
			<html>
			<head>
				<title>SSIS-123 测试视频 - SupJav</title>
				<meta property="og:image" content="https://img.example.com/cover.jpg" />
			</head>
			<body>
				<h1 class="entry-title">SSIS-123 测试视频</h1>
				<div class="btn-server" data-link="https://faststream.in/embed-abc">FST</div>
			</body>
			</html>`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(html)),
			}, nil
		}),
	}

	meta, err := analyzeSupJavWithClient(context.Background(), client, "https://supjav.com/zh/390226.html")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Title != "SSIS-123 测试视频" {
		t.Errorf("title mismatch: %q", meta.Title)
	}
	if meta.Thumbnail != "https://img.example.com/cover.jpg" {
		t.Errorf("thumb mismatch: %q", meta.Thumbnail)
	}
	if meta.Platform != "supjav" {
		t.Errorf("platform mismatch: %q", meta.Platform)
	}
}
