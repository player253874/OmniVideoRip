package extractors

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const browserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

var (
	uuidPattern  = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	titlePattern = regexp.MustCompile(`(?i)<title>(.*?)</title>`)
	ogImgPattern = regexp.MustCompile(`(?i)<meta\s+property=["']og:image["']\s+content=["']([^"']+)["']`)
)

func fetchMissAVPageWithFallback(ctx context.Context, client *http.Client, rawURL string) ([]byte, string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, rawURL, err
	}

	domains := []string{"missav.live", "missav.ws", "missav.ai", "missav.com"}
	// Put requested domain first
	reqHost := strings.ToLower(parsed.Host)
	var orderedDomains []string
	orderedDomains = append(orderedDomains, reqHost)
	for _, d := range domains {
		if d != reqHost {
			orderedDomains = append(orderedDomains, d)
		}
	}

	// Also try path variations: some domains use /dm31/ja/CODE, others just /CODE
	pathVariants := []string{parsed.RequestURI()}
	basePath := filepath.Base(strings.TrimSuffix(parsed.Path, "/"))
	if basePath != "" && basePath != "." && basePath != "/" {
		altPath := "/" + basePath
		if altPath != parsed.Path && altPath != parsed.RequestURI() {
			pathVariants = append(pathVariants, altPath)
		}
	}

	var lastErr error
	for _, domain := range orderedDomains {
		for _, pathVar := range pathVariants {
			candidateURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, domain, pathVar)
			body, fetchErr := fetchBytes(ctx, client, candidateURL, candidateURL, maxPageBytes)
			if fetchErr != nil {
				lastErr = fetchErr
				continue
			}
			// Verify the page is real content (not a Cloudflare challenge or empty page)
			if len(body) < 1000 {
				lastErr = fmt.Errorf("%s 返回内容过短 (%d 字节), 可能被拦截", domain, len(body))
				continue
			}
			bodyStr := string(body)
			if !uuidPattern.MatchString(bodyStr) && !strings.Contains(bodyStr, "surrit.com") {
				// Page loaded but has no video UUIDs - likely a Cloudflare challenge or wrong content
				lastErr = fmt.Errorf("%s 返回页面不含视频数据 (可能被 Cloudflare 拦截)", domain)
				continue
			}
			return body, candidateURL, nil
		}
	}

	return nil, rawURL, fmt.Errorf("无法获取 MissAV 视频页面 (已尝试全部镜像域名与路径): %w", lastErr)
}

func AnalyzeMissAV(ctx context.Context, pageURL string) (VideoMetaData, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return VideoMetaData{}, err
	}

	client := newSafeHTTPClient(20 * time.Second)
	bodyBytes, finalURL, err := fetchMissAVPageWithFallback(ctx, client, pageURL)
	if err != nil {
		return VideoMetaData{}, err
	}

	html := string(bodyBytes)
	title := "MissAV 视频"
	if m := titlePattern.FindStringSubmatch(html); len(m) > 1 {
		title = strings.TrimSpace(m[1])
		title = strings.TrimSuffix(title, " - MissAV")
		title = strings.TrimSuffix(title, " | MissAV")
	}

	thumb := ""
	if m := ogImgPattern.FindStringSubmatch(html); len(m) > 1 {
		thumb = m[1]
	}

	formats := []VideoFormatOption{
		{FormatID: "best", Resolution: "最佳画质 (自动匹配最高分辨率)", Extension: "mp4", Note: "原生多线程并发直下 (推荐)"},
	}

	seen := make(map[string]bool)
	for _, uid := range uuidPattern.FindAllString(html, -1) {
		if seen[uid] {
			continue
		}
		seen[uid] = true
		masterURL := fmt.Sprintf("https://surrit.com/%s/playlist.m3u8", uid)
		master, fetchErr := fetchBytes(ctx, client, masterURL, finalURL, maxPlaylistBytes)
		if fetchErr != nil {
			continue
		}
		masterText := string(master)
		if strings.Contains(masterText, "#EXTM3U") {
			if strings.Contains(masterText, "1920x1080") || strings.Contains(masterText, "1080p") {
				formats = append(formats, VideoFormatOption{FormatID: "1080p", Resolution: "1080P 全高清", Extension: "mp4", Note: "1920x1080 原画"})
			}
			if strings.Contains(masterText, "1280x720") || strings.Contains(masterText, "720p") {
				formats = append(formats, VideoFormatOption{FormatID: "720p", Resolution: "720P 高清", Extension: "mp4", Note: "1280x720 高清"})
			}
			if strings.Contains(masterText, "842x480") || strings.Contains(masterText, "480p") {
				formats = append(formats, VideoFormatOption{FormatID: "480p", Resolution: "480P 标清", Extension: "mp4", Note: "842x480 标清"})
			}
			if strings.Contains(masterText, "640x360") || strings.Contains(masterText, "360p") {
				formats = append(formats, VideoFormatOption{FormatID: "360p", Resolution: "360P 流畅", Extension: "mp4", Note: "640x360 流畅"})
			}
			break
		}
	}

	return VideoMetaData{
		Title:     title,
		Thumbnail: thumb,
		Platform:  "missav",
		URL:       pageURL,
		Formats:   formats,
	}, nil
}

func DownloadMissAV(ctx context.Context, pageURL, outDir, ffmpegPath string, progressCb ProgressCallback) (string, error) {
	return DownloadMissAVWithFormat(ctx, pageURL, outDir, ffmpegPath, "best", progressCb)
}

func DownloadMissAVWithFormat(ctx context.Context, pageURL, outDir, ffmpegPath, formatID string, progressCb ProgressCallback) (string, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return "", err
	}
	if progressCb != nil {
		progressCb(5, "正在连接页面并解析 HLS 播放链...")
	}

	client := newSafeHTTPClient(30 * time.Second)
	bodyBytes, finalURL, err := fetchMissAVPageWithFallback(ctx, client, pageURL)
	if err != nil {
		return "", err
	}

	html := string(bodyBytes)
	title := ""
	if m := titlePattern.FindStringSubmatch(html); len(m) > 1 {
		title = strings.TrimSpace(m[1])
		title = strings.TrimSuffix(title, " - MissAV")
		title = strings.TrimSuffix(title, " | MissAV")
	}

	seen := make(map[string]bool)
	var uniqueUUIDs []string
	for _, id := range uuidPattern.FindAllString(html, -1) {
		if !seen[id] {
			seen[id] = true
			uniqueUUIDs = append(uniqueUUIDs, id)
		}
	}
	if len(uniqueUUIDs) == 0 {
		return "", fmt.Errorf("未在页面中找到有效的 HLS 视频切片 ID")
	}

	var mediaPlaylistURL string
	for _, id := range uniqueUUIDs {
		masterURL := fmt.Sprintf("https://surrit.com/%s/playlist.m3u8", id)
		master, fetchErr := fetchBytes(ctx, client, masterURL, finalURL, maxPlaylistBytes)
		if fetchErr != nil || !strings.Contains(string(master), "#EXTM3U") {
			continue
		}
		qualityPath := parseQualityM3U8(string(master), formatID)
		if qualityPath == "" {
			continue
		}
		mediaPlaylistURL, err = resolveURL(masterURL, qualityPath)
		if err == nil {
			break
		}
	}
	if mediaPlaylistURL == "" {
		return "", fmt.Errorf("无法解析 Surrit HLS 播放切片")
	}

	if progressCb != nil {
		progressCb(15, "正在获取视频切片索引...")
	}
	playlistBytes, err := fetchBytes(ctx, client, mediaPlaylistURL, finalURL, maxPlaylistBytes)
	if err != nil {
		return "", fmt.Errorf("获取切片清单失败: %w", err)
	}
	segmentURLs, err := parseSegmentURLs(mediaPlaylistURL, string(playlistBytes))
	if err != nil {
		return "", err
	}

	workDir := GetTaskResumeDir(outDir, "missav", pageURL)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", err
	}
	segmentFiles, err := downloadSegments(ctx, segmentURLs, workDir, 8, progressCb, func(ctx context.Context, rawURL, destPath string) error {
		return downloadSegmentWithRetry(ctx, client, rawURL, destPath, finalURL)
	})
	if err != nil {
		return "", fmt.Errorf("下载 MissAV 切片失败: %w", err)
	}

	if progressCb != nil {
		progressCb(92, "正在合并视频切片...")
	}
	baseName := title
	if strings.TrimSpace(baseName) == "" {
		baseName = codeNameFromURL(pageURL, "MISSAV")
	}
	outputPath, err := mergeSegments(ctx, segmentFiles, outDir, baseName, ffmpegPath)
	if err != nil {
		return "", err
	}
	// Only clean up working directory after successful merge
	_ = os.RemoveAll(workDir)
	if progressCb != nil {
		progressCb(100, "MissAV 视频下载并合并完成！")
	}
	return outputPath, nil
}

func fetchBytes(ctx context.Context, client *http.Client, rawURL, referer string, limit int64) ([]byte, error) {
	data, _, err := fetchBytesWithFinalURL(ctx, client, rawURL, referer, limit)
	return data, err
}

func fetchBytesWithFinalURL(ctx context.Context, client *http.Client, rawURL, referer string, limit int64) ([]byte, string, error) {
	if err := validateHTTPURL(rawURL); err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7,ja;q=0.6")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	finalURL := rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	data, err := readLimitedBody(resp, limit)
	return data, finalURL, err
}

func parseSegmentURLs(playlistURL, playlistText string) ([]string, error) {
	for _, unsupportedTag := range []string{"#EXT-X-KEY:", "#EXT-X-MAP:", "#EXT-X-BYTERANGE:"} {
		if strings.Contains(playlistText, unsupportedTag) {
			return nil, fmt.Errorf("unsupported HLS playlist feature: %s", strings.TrimSuffix(unsupportedTag, ":"))
		}
	}
	var segmentURLs []string
	scanner := bufio.NewScanner(strings.NewReader(playlistText))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		resolved, err := resolveURL(playlistURL, line)
		if err != nil {
			return nil, fmt.Errorf("invalid segment URL: %w", err)
		}
		segmentURLs = append(segmentURLs, resolved)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(segmentURLs) == 0 {
		return nil, fmt.Errorf("播放列表中未找到视频切片")
	}
	return segmentURLs, nil
}

func codeNameFromURL(rawURL, fallbackPrefix string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil {
		if id := parsed.Query().Get("id"); id != "" {
			if clean := sanitizeFilename(id); clean != "" {
				return clean
			}
		}
		if base := sanitizeFilename(strings.ToUpper(filepath.Base(strings.TrimSuffix(parsed.Path, "/")))); base != "" && base != "." {
			return base
		}
	}
	return fmt.Sprintf("%s_%d", fallbackPrefix, time.Now().Unix())
}

func parseQualityM3U8(playlistText, formatID string) string {
	lines := strings.Split(playlistText, "\n")
	var candidates []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, ".m3u8") {
			candidates = append(candidates, trimmed)
		}
	}
	if len(candidates) == 0 {
		return ""
	}

	if formatID == "1080p" {
		for _, c := range candidates {
			if strings.Contains(c, "1080") {
				return c
			}
		}
	} else if formatID == "720p" {
		for _, c := range candidates {
			if strings.Contains(c, "720") || strings.Contains(c, "1280x720") {
				return c
			}
		}
	} else if formatID == "480p" {
		for _, c := range candidates {
			if strings.Contains(c, "480") || strings.Contains(c, "842x480") {
				return c
			}
		}
	} else if formatID == "360p" {
		for _, c := range candidates {
			if strings.Contains(c, "360") || strings.Contains(c, "640x360") {
				return c
			}
		}
	}

	// Default / best quality: pick candidate with highest resolution indicator or last item
	for _, c := range candidates {
		if strings.Contains(c, "1080") {
			return c
		}
	}
	for _, c := range candidates {
		if strings.Contains(c, "720") {
			return c
		}
	}
	return candidates[len(candidates)-1]
}

func parseBestQualityM3U8(playlistText string) string {
	return parseQualityM3U8(playlistText, "best")
}
