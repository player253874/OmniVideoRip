package extractors

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestUnpackDeanEdwards(t *testing.T) {
	packed := `eval(function(p,a,c,k,e,d){while(c--)if(k[c])p=p.replace(new RegExp('\\b'+c.toString(a)+'\\b','g'),k[c]);return p}('0 1="2://3.4/5.6";',36,7,'var|videoUrl|https|cdn|example|master|m3u8'.split('|')))`

	results := unpackDeanEdwards(packed)
	if len(results) == 0 {
		t.Fatalf("unpackDeanEdwards returned no results")
	}

	expected := `var videoUrl="https://cdn.example/master.m3u8";`
	if results[0] != expected {
		t.Errorf("unpackDeanEdwards = %q; want %q", results[0], expected)
	}
}

func TestUnpackDeanEdwardsBase62(t *testing.T) {
	// Base-62 packed script where index 61 ('Z') is replaced
	// We build 62 elements (0 to 61)
	words := make([]string, 62)
	words[0] = "var"
	words[1] = "result"
	words[61] = "secretValue"
	wordsStr := strings.Join(words, "|")

	packed := `eval(function(p,a,c,k,e,d){while(c--)if(k[c])p=p.replace(new RegExp('\\b'+c.toString(a)+'\\b','g'),k[c]);return p}('0 1="Z";',62,62,'` + wordsStr + `'.split('|')))`

	results := unpackDeanEdwards(packed)
	if len(results) == 0 {
		t.Fatalf("unpackDeanEdwards returned no results for base-62")
	}

	expected := `var result="secretValue";`
	if results[0] != expected {
		t.Errorf("unpackDeanEdwards base-62 = %q; want %q", results[0], expected)
	}
}

func TestParseCandidateEmbedHTML(t *testing.T) {
	// 1. Direct m3u8
	directHTML := `<html><body><video src="https://mirror.com/hls/direct.m3u8"></video></body></html>`
	if got := parseCandidateEmbedHTML("https://mirror.com/e/123", directHTML); got != "https://mirror.com/hls/direct.m3u8" {
		t.Errorf("expected direct m3u8, got %q", got)
	}

	// 2. Dean Edwards packed m3u8
	packedHTML := `<html><body><script>
	eval(function(p,a,c,k,e,d){while(c--)if(k[c])p=p.replace(new RegExp('\\b'+c.toString(a)+'\\b','g'),k[c]);return p}('0={"1":"2://3.4/5.6?7=8"};',36,9,'links|hls|https|cdn|stream|master|m3u8|token|abc'.split('|')))
	</script></body></html>`
	expectedM3U8 := "https://cdn.stream/master.m3u8?token=abc"
	if got := parseCandidateEmbedHTML("https://mirror.com/e/123", packedHTML); got != expectedM3U8 {
		t.Errorf("expected packed m3u8 %q, got %q", expectedM3U8, got)
	}

	// 3. Empty / no m3u8
	emptyHTML := `<html><body>File expired</body></html>`
	if got := parseCandidateEmbedHTML("https://mirror.com/e/123", emptyHTML); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGGJavLive304530Resolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in short mode")
	}

	client := newSafeHTTPClient(25 * time.Second)
	pageURL := "https://ggjav.com/main/video?id=304530"

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	bodyBytes, err := fetchBytes(ctx, client, pageURL, "https://ggjav.com/", maxPageBytes)
	if err != nil {
		t.Skipf("skipping live test due to network unreachable: %v", err)
	}

	playlistURL, referer, err := extractPlayableURLWithContext(ctx, client, bodyBytes)
	if err != nil {
		t.Fatalf("extractPlayableURLWithContext failed on live ggjav 304530: %v", err)
	}

	if !strings.Contains(playlistURL, ".m3u8") {
		t.Fatalf("expected m3u8 URL, got: %s", playlistURL)
	}
	if referer == "" {
		t.Fatalf("expected non-empty referer")
	}

	t.Logf("Live GGJav 304530 resolved M3U8: %s (Referer: %s)", playlistURL, referer)
}
