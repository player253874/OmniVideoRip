package extractors

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	supjavServerRegex = regexp.MustCompile(`(?i)<(?:div|button|a)[^>]*class="[^"]*btn-server[^"]*"[^>]*(?:data-link|data-src|data-file|data-url)=["']([^"']+)["']`)
	supjavIframeRegex = regexp.MustCompile(`(?i)<iframe[^>]+src=["']([^"']+)["']`)
	supjavEntryTitle  = regexp.MustCompile(`(?i)<h1[^>]*class=["'][^"']*entry-title[^"']*["'][^>]*>([\s\S]*?)</h1>`)
)

func isCloudflareChallenge(statusCode int, body string, header http.Header) bool {
	if statusCode == http.StatusForbidden || statusCode == http.StatusServiceUnavailable {
		if header.Get("Cf-Mitigated") == "challenge" {
			return true
		}
		if strings.Contains(body, "Just a moment...") || strings.Contains(body, "challenge-platform") || strings.Contains(body, "_cf_chl_opt") {
			return true
		}
	}
	return false
}

func extractSupJavEmbeds(html string) []string {
	var candidates []string
	seen := make(map[string]bool)

	addCandidate := func(link string) {
		link = strings.TrimSpace(link)
		if link == "" || seen[link] {
			return
		}
		// If base64 encoded URL
		if strings.HasPrefix(link, "aHR0cDov") || strings.HasPrefix(link, "aHR0cHM6") {
			if decoded, err := base64.StdEncoding.DecodeString(link); err == nil {
				link = string(decoded)
			}
		}
		if strings.HasPrefix(link, "//") {
			link = "https:" + link
		}
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			seen[link] = true
			candidates = append(candidates, link)
		}
	}

	// 1. Check btn-server data attributes
	serverMatches := supjavServerRegex.FindAllStringSubmatch(html, -1)
	for _, m := range serverMatches {
		if len(m) > 1 {
			addCandidate(m[1])
		}
	}

	// 2. Check iframes
	iframeMatches := supjavIframeRegex.FindAllStringSubmatch(html, -1)
	for _, m := range iframeMatches {
		if len(m) > 1 {
			addCandidate(m[1])
		}
	}

	// 3. Check direct m3u8 in scripts
	m3u8Matches := m3u8Regex.FindAllString(html, -1)
	for _, m := range m3u8Matches {
		addCandidate(m)
	}

	return candidates
}

func parseSupJavMetadata(html, pageURL string) (string, string) {
	title := ""
	if m := supjavEntryTitle.FindStringSubmatch(html); len(m) > 1 {
		title = strings.TrimSpace(stripHTMLTags(m[1]))
	}
	if title == "" {
		if m := titlePattern.FindStringSubmatch(html); len(m) > 1 {
			title = strings.TrimSpace(m[1])
		}
	}
	if title == "" {
		title = codeNameFromURL(pageURL, "SUPJAV")
	} else {
		title = strings.TrimSuffix(title, " - SupJav")
		title = strings.TrimSuffix(title, " | SupJav")
		title = strings.TrimSuffix(title, " - 高清在线看")
		title = strings.TrimSpace(title)
	}

	thumb := ""
	if m := ogImgPattern.FindStringSubmatch(html); len(m) > 1 {
		thumb = strings.TrimSpace(m[1])
	}
	return title, thumb
}

func stripHTMLTags(s string) string {
	tagRe := regexp.MustCompile(`<[^>]*>`)
	return tagRe.ReplaceAllString(s, "")
}

func AnalyzeSupJav(ctx context.Context, pageURL string) (VideoMetaData, error) {
	return analyzeSupJavWithClient(ctx, newSafeHTTPClient(25*time.Second), pageURL)
}

func analyzeSupJavWithClient(ctx context.Context, client *http.Client, pageURL string) (VideoMetaData, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return VideoMetaData{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return VideoMetaData{}, err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Referer", "https://supjav.com/")

	resp, err := client.Do(req)
	if err != nil {
		return VideoMetaData{}, fmt.Errorf("连接 SupJav 页面失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxPageBytes+1))
	if err != nil {
		return VideoMetaData{}, fmt.Errorf("读取 SupJav 页面失败: %w", err)
	}

	bodyStr := string(bodyBytes)
	if isCloudflareChallenge(resp.StatusCode, bodyStr, resp.Header) {
		return VideoMetaData{}, fmt.Errorf("SupJav 站点当前开启了 Cloudflare Turnstile 人机验证拦截 (HTTP %d)。\n建议：请在浏览器中打开该视频，直接复制播放器下方的分流源（如 FastStream / StreamTape / TVLogy / StreamWish 等）链接下载，或在设置中提供有效 Cookie", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return VideoMetaData{}, fmt.Errorf("SupJav 响应异常: HTTP %d", resp.StatusCode)
	}

	title, thumb := parseSupJavMetadata(bodyStr, pageURL)
	embeds := extractSupJavEmbeds(bodyStr)
	if len(embeds) == 0 {
		return VideoMetaData{}, fmt.Errorf("未在 SupJav 页面中发现可用的播放源，请检查视频是否已下架或需要登录")
	}

	formats := []VideoFormatOption{
		{FormatID: "best", Resolution: "最佳画质 (Best Quality)", Extension: "mp4", Note: "原生并发直下或分流转码 (推荐)"},
		{FormatID: "1080p", Resolution: "1080P 全高清", Extension: "mp4", Note: "1080P 极速分流"},
		{FormatID: "720p", Resolution: "720P 高清", Extension: "mp4", Note: "720P 高清画质"},
		{FormatID: "480p", Resolution: "480P 标清", Extension: "mp4", Note: "480P 标清画质"},
	}

	return VideoMetaData{
		Title:     title,
		Thumbnail: thumb,
		Platform:  "supjav",
		URL:       pageURL,
		Formats:   formats,
	}, nil
}

func DownloadSupJav(ctx context.Context, router *ExtractorRouter, pageURL, outDir, ffmpegPath, formatID string, progressCb ProgressCallback) (string, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return "", err
	}
	if progressCb != nil {
		progressCb(5, "正在连接 SupJav 页面并解析分流源...")
	}

	client := newSafeHTTPClient(30 * time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", "https://supjav.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("连接 SupJav 页面失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxPageBytes+1))
	if err != nil {
		return "", fmt.Errorf("读取 SupJav 页面失败: %w", err)
	}

	bodyStr := string(bodyBytes)
	if isCloudflareChallenge(resp.StatusCode, bodyStr, resp.Header) {
		return "", fmt.Errorf("SupJav 站点当前开启了 Cloudflare Turnstile 人机验证拦截 (HTTP %d)。\n建议：请在浏览器中打开该视频，直接复制播放器下方的分流源（如 FastStream / StreamTape / TVLogy / StreamWish 等）链接下载，或在设置中提供有效 Cookie", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SupJav 响应异常: HTTP %d", resp.StatusCode)
	}

	title, _ := parseSupJavMetadata(bodyStr, pageURL)
	embeds := extractSupJavEmbeds(bodyStr)
	if len(embeds) == 0 {
		return "", fmt.Errorf("未在 SupJav 页面中找到有效的分流源")
	}

	var lastErr error
	for idx, embedURL := range embeds {
		if progressCb != nil {
			progressCb(10+idx*5, fmt.Sprintf("尝试解析分流源 [%d/%d]...", idx+1, len(embeds)))
		}

		// 1. If direct m3u8
		if strings.Contains(embedURL, ".m3u8") {
			out, err := downloadSupJavHLS(ctx, client, embedURL, pageURL, outDir, title, ffmpegPath, formatID, progressCb)
			if err == nil && out != "" {
				return out, nil
			}
			lastErr = err
			continue
		}

		// 2. If FastStream or TVLogy embed, try extracting HLS manifest
		if strings.Contains(embedURL, "tvlogy") || strings.Contains(embedURL, "faststream") || strings.Contains(embedURL, "fststream") {
			m3u8URL, ref, err := extractPlayableURLWithContext(ctx, client, []byte(`<iframe src="`+embedURL+`"></iframe>`))
			if err == nil && m3u8URL != "" {
				out, err := downloadSupJavHLS(ctx, client, m3u8URL, ref, outDir, title, ffmpegPath, formatID, progressCb)
				if err == nil && out != "" {
					return out, nil
				}
				lastErr = err
				continue
			}
		}

		// 3. Fallback to router yt-dlp handling for supported hosters (StreamTape, StreamWish, VOE, DoodStream, etc.)
		out, err := router.downloadWithYTDLPAndFormat(ctx, embedURL, outDir, formatID, progressCb)
		if err == nil && out != "" {
			return out, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("SupJav 所有分流源下载均未成功: %v", lastErr)
}

func downloadSupJavHLS(ctx context.Context, client *http.Client, playlistURL, referer, outDir, title, ffmpegPath, formatID string, progressCb ProgressCallback) (string, error) {
	if progressCb != nil {
		progressCb(20, "正在获取 HLS 播放列表...")
	}

	playlistBytes, err := fetchBytes(ctx, client, playlistURL, referer, maxPlaylistBytes)
	if err != nil {
		return "", fmt.Errorf("获取播放列表失败: %w", err)
	}

	playlistText := string(playlistBytes)
	if subPath := parseQualityM3U8(playlistText, formatID); subPath != "" {
		subURL, resolveErr := resolveURL(playlistURL, subPath)
		if resolveErr != nil {
			return "", resolveErr
		}
		playlistBytes, err = fetchBytes(ctx, client, subURL, referer, maxPlaylistBytes)
		if err != nil {
			return "", fmt.Errorf("获取子播放列表失败: %w", err)
		}
		playlistURL = subURL
		playlistText = string(playlistBytes)
	}

	segmentURLs, err := parseSegmentURLs(playlistURL, playlistText)
	if err != nil {
		return "", err
	}

	workDir := GetTaskResumeDir(outDir, "supjav", playlistURL)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", err
	}

	segmentFiles, err := downloadSegments(ctx, segmentURLs, workDir, 16, progressCb, func(ctx context.Context, rawURL, destPath string) error {
		return downloadSegmentWithRetry(ctx, client, rawURL, destPath, referer)
	})
	if err != nil {
		return "", fmt.Errorf("下载 SupJav 切片失败: %w", err)
	}

	if progressCb != nil {
		progressCb(92, "正在合并视频切片为完整 MP4...")
	}
	baseName := title
	if strings.TrimSpace(baseName) == "" {
		baseName = "SupJav_Video"
	}
	outputPath, err := mergeSegments(ctx, segmentFiles, outDir, baseName, ffmpegPath)
	if err != nil {
		return "", err
	}
	_ = os.RemoveAll(workDir)
	if progressCb != nil {
		progressCb(100, "SupJav 视频下载并合并完成！")
	}
	return outputPath, nil
}
