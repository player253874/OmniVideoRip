package sidecars

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type updateRoundTrip func(*http.Request) (*http.Response, error)

func (f updateRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func digestOf(b []byte) string                                              { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func updateResponse(body []byte) *http.Response {
	return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), Header: make(http.Header)}
}

func TestMetadataRequiresOfficialDigestAndURL(t *testing.T) {
	for _, tc := range []struct {
		name, digest, url string
		valid             bool
	}{
		{"valid", "sha256:" + strings.Repeat("a", 64), "https://github.com/yt-dlp/yt-dlp/releases/download/v1/yt-dlp.exe", true},
		{"missing", "", "https://github.com/yt-dlp/yt-dlp/releases/download/v1/yt-dlp.exe", false},
		{"invalid", "sha256:nope", "https://github.com/yt-dlp/yt-dlp/releases/download/v1/yt-dlp.exe", false},
		{"lookalike", "sha256:" + strings.Repeat("a", 64), "https://github.com.attacker.example/yt-dlp/yt-dlp/releases/download/v1/yt-dlp.exe", false},
		{"http", "sha256:" + strings.Repeat("a", 64), "http://github.com/yt-dlp/yt-dlp/releases/download/v1/yt-dlp.exe", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{"tag_name": "v1", "assets": []any{map[string]any{"name": "yt-dlp.exe", "digest": tc.digest, "browser_download_url": tc.url, "size": 16}}})
			client := &http.Client{Transport: updateRoundTrip(func(r *http.Request) (*http.Response, error) { return updateResponse(body), nil })}
			_, err := latestVerifiedAsset(context.Background(), client, ytDlpSpec)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
}

func TestMetadataRedirectCannotDelegateTrustToMirror(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: updateRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		resp := updateResponse(nil)
		resp.StatusCode = 302
		resp.Header.Set("Location", "https://mirror.example/metadata")
		return resp, nil
	})}
	if _, err := latestVerifiedAsset(context.Background(), client, ytDlpSpec); err == nil {
		t.Fatal("untrusted metadata redirect accepted")
	}
	if calls != 1 {
		t.Fatal("followed untrusted metadata redirect")
	}
}

func TestBadPayloadNeverReplacesInstalledBinary(t *testing.T) {
	good := []byte("MZ verified executable")
	for _, payload := range [][]byte{[]byte("MZ hostile executable"), good[:5], append(append([]byte{}, good...), 0)} {
		dir := t.TempDir()
		target := filepath.Join(dir, "yt-dlp.exe")
		os.WriteFile(target, []byte("MZ original"), 0600)
		_, err := saveVerifiedResponse(updateResponse(payload), verifiedAsset{Digest: digestOf(good), Size: int64(len(good))}, dir, nil)
		if err == nil {
			t.Fatal("unverified payload accepted")
		}
		original, _ := os.ReadFile(target)
		if string(original) != "MZ original" {
			t.Fatal("old binary changed")
		}
		files, _ := os.ReadDir(dir)
		if len(files) != 1 {
			t.Fatal("failed temporary download leaked")
		}
	}
}

func TestVerifiedMirrorFallbackAndInstallation(t *testing.T) {
	good := []byte("MZ verified executable")
	calls := 0
	client := &http.Client{Transport: updateRoundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return updateResponse([]byte("tampered")), nil
		}
		return updateResponse(good), nil
	})}
	dir := t.TempDir()
	p, err := downloadVerifiedAsset(context.Background(), client, verifiedAsset{URL: "https://github.com/yt-dlp/yt-dlp/releases/download/v1/yt-dlp.exe", Digest: digestOf(good), Size: int64(len(good))}, dir, nil)
	if err != nil || calls != 2 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	if _, err := executableDigest(p); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "yt-dlp.exe")
	os.WriteFile(target, []byte("MZ original"), 0600)
	if err := replaceComponentFile(p, target); err != nil {
		t.Fatal(err)
	}
	actual, _ := os.ReadFile(target)
	if !bytes.Equal(actual, good) {
		t.Fatal("verified content changed")
	}
	if err := replaceComponentFile(filepath.Join(dir, "missing"), target); err == nil {
		t.Fatal("missing candidate accepted")
	}
	after, _ := os.ReadFile(target)
	if !bytes.Equal(after, good) {
		t.Fatal("failed replace lost old executable")
	}
}

func TestFFmpegArchiveAndInstalledDigest(t *testing.T) {
	dir := t.TempDir()
	for _, goodHeader := range []bool{false, true} {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		entry, _ := zw.Create("bin/ffmpeg.exe")
		if goodHeader {
			entry.Write([]byte("MZ executable"))
		} else {
			entry.Write([]byte("<html>error</html>"))
		}
		zw.Close()
		archive := filepath.Join(dir, "asset.zip")
		os.WriteFile(archive, buf.Bytes(), 0600)
		exe, err := unpackVerifiedExecutable(archive, dir, "ffmpeg.exe")
		if (err == nil) != goodHeader {
			t.Fatalf("goodHeader=%v err=%v", goodHeader, err)
		}
		if goodHeader {
			hash, _ := executableDigest(exe)
			marker, _ := json.Marshal(installedComponent{AssetDigest: "digest1", ExecutableDigest: hash, Tag: "latest"})
			os.WriteFile(exe+".release.json", marker, 0600)
			if !componentIsCurrent(exe, verifiedAsset{Digest: "digest1"}) {
				t.Fatal("current component rejected")
			}
			if componentIsCurrent(exe, verifiedAsset{Digest: "digest2"}) {
				t.Fatal("rolling latest change not detected")
			}
			os.WriteFile(exe, []byte("MZ modified"), 0600)
			if componentIsCurrent(exe, verifiedAsset{Digest: "digest1"}) {
				t.Fatal("modified executable trusted")
			}
		}
	}
	broken := filepath.Join(dir, "broken.zip")
	os.WriteFile(broken, []byte("PK truncated"), 0600)
	if _, err := unpackVerifiedExecutable(broken, dir, "ffmpeg.exe"); err == nil {
		t.Fatal("truncated ZIP accepted")
	}
}
