package extractors

import (
	"context"
	"net/url"
	"path"
	"strings"
	"time"
	"videodownloader/internal/sidecars"
)

// Direct HLS links may carry download hints in the URL fragment, which is never sent over the network:
//   https://cdn.example.com/master.m3u8#omni_referer=https%3A%2F%2Fplayer.example%2F&omni_title=SupJav_390226
const (
	fragmentRefererKey = "omni_referer"
	fragmentTitleKey   = "omni_title"
)

// parseDirectM3U8 reports whether rawURL points directly at an HLS playlist and extracts fragment hints.
func parseDirectM3U8(rawURL string) (cleanURL, referer, title string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", "", "", false
	}
	lowerPath := strings.ToLower(u.Path)
	lowerQuery := strings.ToLower(u.RawQuery)
	if !strings.HasSuffix(lowerPath, ".m3u8") && !strings.Contains(lowerPath, ".m3u8/") && !strings.Contains(lowerQuery, ".m3u8") {
		return "", "", "", false
	}
	if hints, err := url.ParseQuery(u.Fragment); err == nil {
		referer = strings.TrimSpace(hints.Get(fragmentRefererKey))
		title = strings.TrimSpace(hints.Get(fragmentTitleKey))
	}
	if referer != "" && validateHTTPURL(referer) != nil {
		referer = ""
	}
	u.Fragment = ""
	u.RawFragment = ""
	if title == "" {
		base := strings.TrimSuffix(path.Base(u.Path), path.Ext(u.Path))
		lowerBase := strings.ToLower(base)
		if base == "" || lowerBase == "master" || lowerBase == "index" || lowerBase == "playlist" || lowerBase == "video" || lowerBase == "manifest" {
			title = "HLS_M3U8_Video"
		} else {
			title = base
		}
	}
	return u.String(), referer, sanitizeFilename(title), true
}

func analyzeDirectM3U8(pageURL, title string) VideoMetaData {
	return VideoMetaData{
		Title:    title,
		Duration: formatDuration(0),
		Platform: "m3u8",
		URL:      pageURL,
		Formats: []VideoFormatOption{
			{FormatID: "best", Resolution: "最佳画质 (Best Quality)", Extension: "mp4", Note: "16线程 HLS 原生并发直下 (推荐)"},
			{FormatID: "1080p", Resolution: "1080P 全高清", Extension: "mp4", Note: "1080P M3U8 流"},
			{FormatID: "720p", Resolution: "720P 高清", Extension: "mp4", Note: "720P M3U8 流"},
			{FormatID: "480p", Resolution: "480P 标清", Extension: "mp4", Note: "480P M3U8 流"},
		},
	}
}

func (r *ExtractorRouter) downloadDirectM3U8(ctx context.Context, playlistURL, referer, title, outDir, ffPath, formatID string, progressCb ProgressCallback) (string, error) {
	client := newSafeHTTPClient(30 * time.Second)
	tryReferer := referer
	if tryReferer == "" {
		tryReferer = "https://supjav.com/"
	}
	out, err := downloadSupJavHLS(ctx, client, playlistURL, tryReferer, outDir, title, ffPath, formatID, progressCb)
	if err == nil || ctx.Err() != nil {
		return out, err
	}
	// Retry once with origin of playlist URL if referer was defaulted and failed
	if referer == "" {
		if u, parseErr := url.Parse(playlistURL); parseErr == nil && u.Scheme != "" && u.Host != "" {
			originRef := u.Scheme + "://" + u.Host + "/"
			if out2, err2 := downloadSupJavHLS(ctx, client, playlistURL, originRef, outDir, title, ffPath, formatID, progressCb); err2 == nil || ctx.Err() != nil {
				return out2, err2
			}
		}
	}
	// Native downloader cannot handle every playlist (e.g. AES-128 encrypted segments); let yt-dlp try.
	ytPath, _ := sidecars.ResolveYTDLPPath(r.ytdlpPath)
	if ytPath == "" {
		return "", err
	}
	return r.runYTDLPWithReferer(ctx, ytPath, ffPath, playlistURL, outDir, formatID, "", tryReferer, progressCb)
}
