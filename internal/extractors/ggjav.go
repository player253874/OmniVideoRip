package extractors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	m3u8Regex = regexp.MustCompile(`https?://[^\s"\'<>]+\.m3u8[^\s"\'<>]*`)
	varLRegex = regexp.MustCompile(`var\s+l\s*=\s*"([^"]+)";`)
)

func AnalyzeGGJav(ctx context.Context, pageURL string) (VideoMetaData, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return VideoMetaData{}, err
	}

	client := newSafeHTTPClient(20 * time.Second)
	bodyBytes, err := fetchBytes(ctx, client, pageURL, "https://ggjav.com/", maxPageBytes)
	if err != nil {
		return VideoMetaData{}, fmt.Errorf("连接 GGJav 页面失败: %w", err)
	}

	html := string(bodyBytes)
	title := "GGJav 视频"
	if m := titlePattern.FindStringSubmatch(html); len(m) > 1 {
		title = strings.TrimSpace(m[1])
		title = strings.TrimSuffix(title, " - GGJAV | 日本AV视频在线看")
		title = strings.TrimSuffix(title, " - GGJAV")
	}

	thumb := ""
	if m := ogImgPattern.FindStringSubmatch(html); len(m) > 1 {
		thumb = m[1]
	}

	matches := varLRegex.FindSubmatch(bodyBytes)
	if len(matches) > 1 {
		rawB64 := string(matches[1])
		decodedBytes, err := base64.StdEncoding.DecodeString(rawB64)
		if err == nil && len(decodedBytes) > 0 {
			decrypted := make([]byte, len(decodedBytes))
			for i, b := range decodedBytes {
				decrypted[i] = byte((int(b) - 0x58 + 256) % 256)
			}
			var linkMap map[string][]string
			if err := json.Unmarshal(decrypted, &linkMap); err == nil {
				if list, ok := linkMap["ggjav"]; ok && len(list) > 0 {
					if parsedURL, err := url.Parse(list[0]); err == nil {
						if poster := parsedURL.Query().Get("poster"); poster != "" && thumb == "" {
							thumb = poster
						}
					}
				}
			}
		}
	}

	formats := []VideoFormatOption{
		{FormatID: "best", Resolution: "最佳画质 (原画 HLS / MP4)", Extension: "mp4", Note: "原生多线程并发直下 (推荐)"},
	}

	return VideoMetaData{
		Title:     title,
		Thumbnail: thumb,
		Platform:  "ggjav",
		URL:       pageURL,
		Formats:   formats,
	}, nil
}

func DownloadGGJav(ctx context.Context, pageURL, outDir, ffmpegPath string, progressCb ProgressCallback) (string, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return "", err
	}
	if progressCb != nil {
		progressCb(5, "正在连接 GGJav 页面并解析视频地址...")
	}

	client := newSafeHTTPClient(30 * time.Second)
	bodyBytes, err := fetchBytes(ctx, client, pageURL, "https://ggjav.com/", maxPageBytes)
	if err != nil {
		return "", fmt.Errorf("连接 GGJav 页面失败: %w", err)
	}

	html := string(bodyBytes)
	title := ""
	if m := titlePattern.FindStringSubmatch(html); len(m) > 1 {
		title = strings.TrimSpace(m[1])
		title = strings.TrimSuffix(title, " - GGJAV | 日本AV视频在线看")
		title = strings.TrimSuffix(title, " - GGJAV")
	}

	playlistURL := extractPlayableURLFromGGJavPage(bodyBytes)
	if playlistURL == "" {
		return "", fmt.Errorf("未在页面中找到有效的播放地址")
	}
	if err := validateHTTPURL(playlistURL); err != nil {
		return "", fmt.Errorf("不安全的播放地址: %w", err)
	}

	if progressCb != nil {
		progressCb(15, "正在解析 HLS 播放列表...")
	}
	playlistBytes, err := fetchBytes(ctx, client, playlistURL, "https://ggjav.com/", maxPlaylistBytes)
	if err != nil {
		return "", fmt.Errorf("获取 GGJav 播放列表失败: %w", err)
	}
	playlistText := string(playlistBytes)
	if subPath := parseBestQualityM3U8(playlistText); subPath != "" {
		subURL, resolveErr := resolveURL(playlistURL, subPath)
		if resolveErr != nil {
			return "", resolveErr
		}
		playlistBytes, err = fetchBytes(ctx, client, subURL, "https://ggjav.com/", maxPlaylistBytes)
		if err != nil {
			return "", fmt.Errorf("获取 GGJav 子播放列表失败: %w", err)
		}
		playlistURL = subURL
		playlistText = string(playlistBytes)
	}
	segmentURLs, err := parseSegmentURLs(playlistURL, playlistText)
	if err != nil {
		return "", err
	}

	workDir := GetTaskResumeDir(outDir, "ggjav", pageURL)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", err
	}
	segmentFiles, err := downloadSegments(ctx, segmentURLs, workDir, 8, progressCb, func(ctx context.Context, rawURL, destPath string) error {
		return downloadSegmentWithRetry(ctx, client, rawURL, destPath, "https://ggjav.com/")
	})
	if err != nil {
		return "", fmt.Errorf("下载 GGJav 切片失败: %w", err)
	}

	if progressCb != nil {
		progressCb(92, "正在合并视频切片...")
	}
	baseName := title
	if strings.TrimSpace(baseName) == "" {
		baseName = codeNameFromURL(pageURL, "GGJAV")
	}
	outputPath, err := mergeSegments(ctx, segmentFiles, outDir, baseName, ffmpegPath)
	if err != nil {
		return "", err
	}
	// Only clean up working directory after successful merge
	_ = os.RemoveAll(workDir)
	if progressCb != nil {
		progressCb(100, "GGJav 视频下载并合并完成！")
	}
	return outputPath, nil
}

func extractPlayableURLFromGGJavPage(bodyBytes []byte) string {
	matches := varLRegex.FindSubmatch(bodyBytes)
	if len(matches) > 1 {
		rawB64 := string(matches[1])
		decodedBytes, err := base64.StdEncoding.DecodeString(rawB64)
		if err == nil && len(decodedBytes) > 0 {
			decrypted := make([]byte, len(decodedBytes))
			for i, b := range decodedBytes {
				decrypted[i] = byte((int(b) - 0x58 + 256) % 256)
			}
			var linkMap map[string][]string
			if err := json.Unmarshal(decrypted, &linkMap); err == nil {
				// Check 'ggjav' key first
				if list, ok := linkMap["ggjav"]; ok && len(list) > 0 {
					for _, embedURL := range list {
						if parsedURL, err := url.Parse(embedURL); err == nil {
							uParam := parsedURL.Query().Get("u")
							if uParam != "" {
								if uDec, err := base64.StdEncoding.DecodeString(uParam); err == nil && len(uDec) > 0 {
									urlStr := string(uDec)
									if strings.HasSuffix(urlStr, ".mp4") {
										return urlStr + "/index.m3u8"
									}
									return urlStr
								}
							}
						}
						return embedURL
					}
				}
				// Check for any direct m3u8 or mp4 URLs in any key
				for _, list := range linkMap {
					for _, item := range list {
						if strings.Contains(item, ".m3u8") {
							return item
						}
						if strings.HasSuffix(item, ".mp4") {
							return item + "/index.m3u8"
						}
					}
				}
			}
		}
	}

	// Fallback to regex for direct m3u8
	m3u8Matches := m3u8Regex.FindAllString(string(bodyBytes), -1)
	if len(m3u8Matches) > 0 {
		return m3u8Matches[0]
	}

	return ""
}

func sanitizeFilename(name string) string {
	invalid := regexp.MustCompile(`[\\/:*?"<>|\x00-\x1f]`)
	cleaned := strings.TrimSpace(invalid.ReplaceAllString(name, "_"))
	cleaned = strings.TrimRight(cleaned, ". ")
	if len(cleaned) > 120 {
		cleaned = cleaned[:120]
	}
	return cleaned
}
