package extractors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	m3u8Regex     = regexp.MustCompile(`https?://[^\s"\'<>]+\.m3u8[^\s"\'<>]*`)
	varLRegex     = regexp.MustCompile(`var\s+l\s*=\s*"([^"]+)";`)
	packerPattern = regexp.MustCompile(`eval\(function\(p,a,c,k,e,d\)[\s\S]*?return\s+p}\(([\s\S]*?)\.split\('\|'\)`)
	wordPattern   = regexp.MustCompile(`\b[0-9a-zA-Z]+\b`)
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

	playlistURL, referer, err := extractPlayableURLWithContext(ctx, client, bodyBytes)
	if err != nil || playlistURL == "" {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("未在页面中找到有效的播放地址")
	}
	if err := validateHTTPURL(playlistURL); err != nil {
		return "", fmt.Errorf("不安全的播放地址: %w", err)
	}
	if referer == "" {
		referer = "https://ggjav.com/"
	}

	if progressCb != nil {
		progressCb(15, "正在解析 HLS 播放列表...")
	}
	playlistBytes, err := fetchBytes(ctx, client, playlistURL, referer, maxPlaylistBytes)
	if err != nil {
		return "", fmt.Errorf("获取 GGJav 播放列表失败: %w", err)
	}
	playlistText := string(playlistBytes)
	if subPath := parseBestQualityM3U8(playlistText); subPath != "" {
		subURL, resolveErr := resolveURL(playlistURL, subPath)
		if resolveErr != nil {
			return "", resolveErr
		}
		playlistBytes, err = fetchBytes(ctx, client, subURL, referer, maxPlaylistBytes)
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
		return downloadSegmentWithRetry(ctx, client, rawURL, destPath, referer)
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
	urlStr, _, _ := extractPlayableURLWithContext(context.Background(), nil, bodyBytes)
	return urlStr
}

func extractPlayableURLWithContext(ctx context.Context, client *http.Client, bodyBytes []byte) (string, string, error) {
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
				// 1. Check primary 'ggjav' key first
				if list, ok := linkMap["ggjav"]; ok && len(list) > 0 {
					for _, embedURL := range list {
						if parsedURL, err := url.Parse(embedURL); err == nil {
							uParam := parsedURL.Query().Get("u")
							if uParam != "" {
								if uDec, err := base64.StdEncoding.DecodeString(uParam); err == nil && len(uDec) > 0 {
									urlStr := string(uDec)
									if strings.HasSuffix(urlStr, ".mp4") {
										return urlStr + "/index.m3u8", "https://ggjav.com/", nil
									}
									return urlStr, "https://ggjav.com/", nil
								}
							}
						}
						return embedURL, "https://ggjav.com/", nil
					}
				}

				// 2. Check for any direct m3u8 or mp4 URLs in any key
				for _, list := range linkMap {
					for _, item := range list {
						if strings.Contains(item, ".m3u8") {
							return item, "https://ggjav.com/", nil
						}
						if strings.HasSuffix(item, ".mp4") {
							return item + "/index.m3u8", "https://ggjav.com/", nil
						}
					}
				}

				// 3. Fallback: inspect candidate mirror providers (e.g. mmsi01, streamhide, etc.)
				if client != nil {
					var candidates []string
					if list, ok := linkMap["mmsi01"]; ok {
						candidates = append(candidates, list...)
					}
					for k, list := range linkMap {
						if k != "mmsi01" && k != "ggjav" {
							candidates = append(candidates, list...)
						}
					}

					for _, cand := range candidates {
						if err := validateHTTPURL(cand); err != nil {
							continue
						}
						cBytes, err := fetchBytes(ctx, client, cand, "https://ggjav.com/", maxPageBytes)
						if err != nil {
							continue
						}
						cHTML := string(cBytes)
						candReferer := "https://ggjav.com/"
						if p, err := url.Parse(cand); err == nil && p.Scheme != "" && p.Host != "" {
							candReferer = p.Scheme + "://" + p.Host + "/"
						}

						// Check candidate HTML (direct or Dean Edwards unpacked)
						if m := parseCandidateEmbedHTML(cand, cHTML); m != "" {
							return m, candReferer, nil
						}
					}
				}
			}
		}
	}

	// 4. Fallback to regex for direct m3u8 in raw body
	m3u8Matches := m3u8Regex.FindAllString(string(bodyBytes), -1)
	if len(m3u8Matches) > 0 {
		return m3u8Matches[0], "https://ggjav.com/", nil
	}

	return "", "", fmt.Errorf("未在页面中找到有效的播放地址")
}

func parseCandidateEmbedHTML(candURL, candHTML string) string {
	// 1. Direct m3u8 in candidate HTML
	if m := m3u8Regex.FindString(candHTML); m != "" {
		return m
	}

	// 2. Unpack Dean Edwards packed script
	unpackedList := unpackDeanEdwards(candHTML)
	for _, unp := range unpackedList {
		if m := m3u8Regex.FindString(unp); m != "" {
			return m
		}
	}
	return ""
}

func unpackDeanEdwards(script string) []string {
	matches := packerPattern.FindAllStringSubmatch(script, -1)
	var results []string
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		innerArgs := strings.TrimSpace(m[1])
		lastComma := strings.LastIndex(innerArgs, ",")
		if lastComma == -1 {
			continue
		}
		wordsStr := strings.Trim(strings.TrimSpace(innerArgs[lastComma+1:]), `"'`)
		words := strings.Split(wordsStr, "|")

		prefix := strings.TrimSpace(innerArgs[:lastComma])
		lastComma2 := strings.LastIndex(prefix, ",")
		if lastComma2 == -1 {
			continue
		}
		count, _ := strconv.Atoi(strings.TrimSpace(prefix[lastComma2+1:]))

		prefix2 := strings.TrimSpace(prefix[:lastComma2])
		lastComma3 := strings.LastIndex(prefix2, ",")
		if lastComma3 == -1 {
			continue
		}
		radix, _ := strconv.Atoi(strings.TrimSpace(prefix2[lastComma3+1:]))

		payload := strings.Trim(strings.TrimSpace(prefix2[:lastComma3]), `"'`)
		payload = strings.ReplaceAll(payload, `\'`, `'`)
		payload = strings.ReplaceAll(payload, `\"`, `"`)
		payload = strings.ReplaceAll(payload, `\\`, `\`)

		if radix <= 0 || radix > 62 || count <= 0 || len(words) == 0 {
			continue
		}

		decodeBaseN := func(token string) int {
			if radix <= 36 {
				val, err := strconv.ParseInt(token, radix, 64)
				if err != nil {
					return -1
				}
				return int(val)
			}
			val := 0
			for _, ch := range token {
				var digit int
				switch {
				case ch >= '0' && ch <= '9':
					digit = int(ch - '0')
				case ch >= 'a' && ch <= 'z':
					digit = int(ch - 'a' + 10)
				case ch >= 'A' && ch <= 'Z':
					digit = int(ch - 'A' + 36)
				default:
					return -1
				}
				val = val*radix + digit
			}
			return val
		}

		unpacked := wordPattern.ReplaceAllStringFunc(payload, func(token string) string {
			idx := decodeBaseN(token)
			if idx >= 0 && idx < len(words) && words[idx] != "" {
				return words[idx]
			}
			return token
		})
		results = append(results, unpacked)
	}
	return results
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
