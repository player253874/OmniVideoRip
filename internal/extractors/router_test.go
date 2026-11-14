package extractors

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatYTDLPError(t *testing.T) {
	botErr := "WARNING: [youtube] No JS runtime\nERROR: [youtube] MJ443eAUlPs: Sign in to confirm you’re not a bot. This helps protect our community."
	gotBot := formatYTDLPError(botErr, errors.New("exit status 1"))
	if !strings.HasPrefix(gotBot, "YOUTUBE_BOT_CHECK:") {
		t.Fatalf("expected YOUTUBE_BOT_CHECK prefix, got %q", gotBot)
	}

	lockedErr := "ERROR: Could not copy Chrome cookie database. See https://github.com/yt-dlp/yt-dlp/issues/7271"
	gotLocked := formatYTDLPError(lockedErr, errors.New("exit status 1"))
	if !strings.HasPrefix(gotLocked, "COOKIE_DB_LOCKED:") {
		t.Fatalf("expected COOKIE_DB_LOCKED prefix, got %q", gotLocked)
	}

	dnsErr := "ERROR: Unable to download webpage: HTTPSConnectionPool(host='www.youtube.com', port=443)"
	gotDNS := formatYTDLPError(dnsErr, errors.New("exit status 1"))
	if !strings.Contains(gotDNS, "网络 / DNS 连接失败") {
		t.Fatalf("expected DNS guidance, got %q", gotDNS)
	}

	privateErr := "WARNING: ignoring warning\nERROR: [youtube] abc: Private video. Sign in if you've been granted access"
	gotPrivate := formatYTDLPError(privateErr, errors.New("exit status 1"))
	if !strings.Contains(gotPrivate, "Private video") || strings.Contains(gotPrivate, "WARNING:") {
		t.Fatalf("expected clean ERROR line without WARNING, got %q", gotPrivate)
	}
}

func TestAppendCookieArgs(t *testing.T) {
	base := []string{"--no-playlist"}

	noneArgs := appendCookieArgs(base, "none", "chrome", "/tmp/c.txt")
	if len(noneArgs) != 1 {
		t.Fatalf("expected unchanged args for mode=none, got %v", noneArgs)
	}

	invalidBrowserArgs := appendCookieArgs(base, "browser", "opera-evil", "")
	if len(invalidBrowserArgs) != 1 {
		t.Fatalf("expected unsupported browser to be ignored, got %v", invalidBrowserArgs)
	}

	edgeArgs := appendCookieArgs(base, "browser", "Edge", "")
	if len(edgeArgs) != 3 || edgeArgs[1] != "--cookies-from-browser" || edgeArgs[2] != "edge" {
		t.Fatalf("unexpected browser cookie args: %v", edgeArgs)
	}

	fileArgs := appendCookieArgs(base, "file", "", `C:\temp\cookies.txt`)
	if len(fileArgs) != 3 || fileArgs[1] != "--cookies" || fileArgs[2] != `C:\temp\cookies.txt` {
		t.Fatalf("unexpected file cookie args: %v", fileArgs)
	}
}

func TestPrepareTempCookieCopyPreservesSource(t *testing.T) {
	src := filepath.Join(t.TempDir(), "cookies.txt")
	original := "# Netscape HTTP Cookie File\n.youtube.com\tTRUE\t/\tTRUE\t0\tSID\tabc\n"
	if err := os.WriteFile(src, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}

	tmpPath, cleanup := prepareTempCookieCopy(src)
	if tmpPath == src || tmpPath == "" {
		t.Fatalf("expected isolated temp copy path, got %q", tmpPath)
	}
	if err := os.WriteFile(tmpPath, []byte("overwritten-by-ytdlp"), 0600); err != nil {
		t.Fatal(err)
	}
	cleanup()

	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatalf("temp cookie copy %q should be removed on cleanup", tmpPath)
	}
	gotSrc, err := os.ReadFile(src)
	if err != nil || string(gotSrc) != original {
		t.Fatalf("original cookies.txt was altered: got %q, err=%v", string(gotSrc), err)
	}
}
