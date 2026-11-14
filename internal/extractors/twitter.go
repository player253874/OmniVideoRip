package extractors

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var twitterStatusRegex = regexp.MustCompile(`(?i)(?:twitter\.com|x\.com)/([^/]+)/status/(\d+)`)

type fxTwitterResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Tweet   struct {
		Text   string `json:"text"`
		Author struct {
			Name       string `json:"name"`
			ScreenName string `json:"screen_name"`
		} `json:"author"`
		Media struct {
			Videos []struct {
				URL          string  `json:"url"`
				ThumbnailURL string  `json:"thumbnail_url"`
				Duration     float64 `json:"duration"`
				Width        int     `json:"width"`
				Height       int     `json:"height"`
				Variants     []struct {
					URL         string `json:"url"`
					Bitrate     int    `json:"bitrate"`
					ContentType string `json:"content_type"`
				} `json:"variants"`
			} `json:"videos"`
		} `json:"media"`
	} `json:"tweet"`
}

type vxTwitterResponse struct {
	Text          string   `json:"text"`
	UserName      string   `json:"user_name"`
	UserScreen    string   `json:"user_screen_name"`
	MediaURLs     []string `json:"mediaURLs"`
	MediaExtended []struct {
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnail_url"`
		DurationMs   int    `json:"duration_millis"`
		Size         struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"size"`
	} `json:"media_extended"`
}

func parseTwitterStatus(rawURL string) (user string, tweetID string, ok bool) {
	matches := twitterStatusRegex.FindStringSubmatch(rawURL)
	if len(matches) >= 3 {
		return matches[1], matches[2], true
	}
	return "", "", false
}

func AnalyzeTwitter(ctx context.Context, pageURL string) (VideoMetaData, error) {
	user, tweetID, ok := parseTwitterStatus(pageURL)
	if !ok {
		return VideoMetaData{}, fmt.Errorf("无法从链接中解析 X (Twitter) 推文 ID: %s", pageURL)
	}

	metaClient := &http.Client{
		Timeout: 12 * time.Second,
	}

	// 1. Try fxtwitter API (returns detailed resolution variants)
	fxURL := fmt.Sprintf("https://api.fxtwitter.com/%s/status/%s", user, tweetID)
	req, _ := http.NewRequestWithContext(ctx, "GET", fxURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	if resp, err := metaClient.Do(req); err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var fx fxTwitterResponse
		if jsonErr := json.NewDecoder(resp.Body).Decode(&fx); jsonErr == nil && len(fx.Tweet.Media.Videos) > 0 {
			v := fx.Tweet.Media.Videos[0]
			title := fx.Tweet.Text
			if title == "" {
				title = fmt.Sprintf("X_%s_%s", fx.Tweet.Author.ScreenName, tweetID)
			} else {
				if len(title) > 60 {
					title = title[:60] + "..."
				}
				title = fmt.Sprintf("%s - %s", fx.Tweet.Author.Name, title)
			}

			durSec := int(v.Duration)
			durStr := fmt.Sprintf("%02d:%02d", durSec/60, durSec%60)

			// Collect MP4 variants
			type variant struct {
				url        string
				resolution string
				formatId   string
				bitrate    int
			}
			var variants []variant

			for _, vr := range v.Variants {
				if strings.Contains(vr.ContentType, "mp4") || strings.HasSuffix(strings.Split(vr.URL, "?")[0], ".mp4") {
					resStr := "高清"
					fmtId := "best"
					if vr.Bitrate >= 2000000 || v.Height >= 1080 || v.Width >= 1080 {
						resStr = "1080P 全高清"
						fmtId = "1080p"
					} else if vr.Bitrate >= 800000 || v.Height >= 720 || v.Width >= 720 {
						resStr = "720P 高清"
						fmtId = "720p"
					} else if vr.Bitrate >= 400000 || v.Height >= 480 || v.Width >= 480 {
						resStr = "480P 标清"
						fmtId = "480p"
					} else {
						resStr = "360P 流畅"
						fmtId = "360p"
					}
					variants = append(variants, variant{
						url:        vr.URL,
						resolution: resStr,
						formatId:   fmtId,
						bitrate:    vr.Bitrate,
					})
				}
			}

			// Sort by bitrate descending
			sort.Slice(variants, func(i, j int) bool {
				return variants[i].bitrate > variants[j].bitrate
			})

			formats := []VideoFormatOption{
				{FormatID: "best", Resolution: "最佳画质 (Best Quality)", Extension: "mp4", Note: "直连推特 CDN 原画 MP4 (推荐)"},
			}

			seenFmt := make(map[string]bool)
			for _, vr := range variants {
				if !seenFmt[vr.formatId] {
					seenFmt[vr.formatId] = true
					formats = append(formats, VideoFormatOption{
						FormatID:   vr.formatId,
						Resolution: vr.resolution,
						Extension:  "mp4",
						Note:       fmt.Sprintf("%d kbps 极速直链", vr.bitrate/1000),
					})
				}
			}

			return VideoMetaData{
				Title:     sanitizeFilename(title),
				Thumbnail: v.ThumbnailURL,
				Duration:  durStr,
				Platform:  "x",
				URL:       pageURL,
				Formats:   formats,
			}, nil
		}
	}

	// 2. Fallback to vxtwitter API
	vxURL := fmt.Sprintf("https://api.vxtwitter.com/%s/status/%s", user, tweetID)
	reqVx, _ := http.NewRequestWithContext(ctx, "GET", vxURL, nil)
	reqVx.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	if resp, err := metaClient.Do(reqVx); err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var vx vxTwitterResponse
		if jsonErr := json.NewDecoder(resp.Body).Decode(&vx); jsonErr == nil && len(vx.MediaURLs) > 0 {
			title := vx.Text
			if title == "" {
				title = fmt.Sprintf("X_%s_%s", vx.UserScreen, tweetID)
			} else {
				if len(title) > 60 {
					title = title[:60] + "..."
				}
				title = fmt.Sprintf("%s - %s", vx.UserName, title)
			}

			thumb := ""
			durStr := "00:30"
			if len(vx.MediaExtended) > 0 {
				thumb = vx.MediaExtended[0].ThumbnailURL
				durSec := vx.MediaExtended[0].DurationMs / 1000
				if durSec > 0 {
					durStr = fmt.Sprintf("%02d:%02d", durSec/60, durSec%60)
				}
			}

			formats := []VideoFormatOption{
				{FormatID: "best", Resolution: "最佳画质 (Best Quality)", Extension: "mp4", Note: "直连推特 CDN 原画 MP4 (推荐)"},
				{FormatID: "720p", Resolution: "720P 高清", Extension: "mp4", Note: "标准原画"},
			}

			return VideoMetaData{
				Title:     sanitizeFilename(title),
				Thumbnail: thumb,
				Duration:  durStr,
				Platform:  "x",
				URL:       pageURL,
				Formats:   formats,
			}, nil
		}
	}

	return VideoMetaData{}, fmt.Errorf("未能从推文中提取到视频流")
}

func DownloadTwitter(ctx context.Context, pageURL, outDir, formatID string, progressCb ProgressCallback) (string, error) {
	user, tweetID, ok := parseTwitterStatus(pageURL)
	if !ok {
		return "", fmt.Errorf("无法从链接中解析 X (Twitter) 推文 ID")
	}

	if progressCb != nil {
		progressCb(5, "正在获取 X (Twitter) 媒体下载直链...")
	}

	meta, err := AnalyzeTwitter(ctx, pageURL)
	if err != nil {
		return "", err
	}

	metaClient := &http.Client{Timeout: 15 * time.Second}
	var directVideoURL string

	fxURL := fmt.Sprintf("https://api.fxtwitter.com/%s/status/%s", user, tweetID)
	req, _ := http.NewRequestWithContext(ctx, "GET", fxURL, nil)
	if resp, err := metaClient.Do(req); err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var fx fxTwitterResponse
		if jsonErr := json.NewDecoder(resp.Body).Decode(&fx); jsonErr == nil && len(fx.Tweet.Media.Videos) > 0 {
			v := fx.Tweet.Media.Videos[0]
			directVideoURL = v.URL
			if len(v.Variants) > 0 {
				var bestVariant string
				var bestBitrate int
				for _, vr := range v.Variants {
					if (strings.Contains(vr.ContentType, "mp4") || strings.HasSuffix(strings.Split(vr.URL, "?")[0], ".mp4")) && vr.Bitrate > bestBitrate {
						bestBitrate = vr.Bitrate
						bestVariant = vr.URL
					}
				}
				if bestVariant != "" {
					directVideoURL = bestVariant
				}
			}
		}
	}

	if directVideoURL == "" {
		vxURL := fmt.Sprintf("https://api.vxtwitter.com/%s/status/%s", user, tweetID)
		reqVx, _ := http.NewRequestWithContext(ctx, "GET", vxURL, nil)
		if resp, err := metaClient.Do(reqVx); err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var vx vxTwitterResponse
			if jsonErr := json.NewDecoder(resp.Body).Decode(&vx); jsonErr == nil && len(vx.MediaURLs) > 0 {
				directVideoURL = vx.MediaURLs[0]
			}
		}
	}

	if directVideoURL == "" {
		return "", fmt.Errorf("未能获取到有效的推特视频直链")
	}

	if progressCb != nil {
		progressCb(10, "已获取推特 CDN 高速直链，正在建立流传输连接...")
	}

	// Output file path
	fileName := fmt.Sprintf("%s.mp4", sanitizeFilename(meta.Title))
	if len(fileName) < 5 {
		fileName = fmt.Sprintf("X_%s_%s.mp4", user, tweetID)
	}
	outPath := filepath.Join(outDir, fileName)
	workDir := GetTaskResumeDir(outDir, "x", pageURL)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", err
	}
	tempPath := filepath.Join(workDir, "video.downloading")

	// Dedicated streaming client with NO overall request deadline (Timeout: 0)
	streamClient := &http.Client{
		Timeout: 0, // IMPORTANT: No overall timeout so large video streams never get aborted mid-flight
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   15 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Get existing downloaded size for resume support
	var startBytes int64 = 0
	if fi, err := os.Stat(tempPath); err == nil {
		startBytes = fi.Size()
	}

	var totalBytes int64 = 0
	var outFile *os.File

	if startBytes > 0 {
		outFile, err = os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		outFile, err = os.Create(tempPath)
	}
	if err != nil {
		return "", fmt.Errorf("创建本地输出文件失败: %w", err)
	}
	defer outFile.Close()

	dlReq, err := http.NewRequestWithContext(ctx, "GET", directVideoURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建下载请求失败: %w", err)
	}
	dlReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	if startBytes > 0 {
		dlReq.Header.Set("Range", fmt.Sprintf("bytes=%d-", startBytes))
	}

	dlResp, err := streamClient.Do(dlReq)
	if err != nil {
		return "", fmt.Errorf("连接推特视频流失败: %w", err)
	}
	defer dlResp.Body.Close()

	if dlResp.StatusCode == http.StatusPartialContent {
		totalBytes = startBytes + dlResp.ContentLength
	} else if dlResp.StatusCode == http.StatusOK {
		totalBytes = dlResp.ContentLength
		// Server didn't respect Range or fresh start, reset file pointer
		if startBytes > 0 {
			_ = outFile.Truncate(0)
			_, _ = outFile.Seek(0, 0)
			startBytes = 0
		}
	} else {
		return "", fmt.Errorf("推特 CDN 返回 HTTP %d", dlResp.StatusCode)
	}

	downloaded := startBytes
	buf := make([]byte, 64*1024)
	lastReport := time.Now()

	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		n, rErr := dlResp.Body.Read(buf)
		if n > 0 {
			if _, wErr := outFile.Write(buf[:n]); wErr != nil {
				return "", fmt.Errorf("写入文件数据失败: %w", wErr)
			}
			downloaded += int64(n)

			if time.Since(lastReport) > 300*time.Millisecond && totalBytes > 0 {
				lastReport = time.Now()
				pct := int(float64(downloaded)/float64(totalBytes)*90) + 10
				if pct > 99 {
					pct = 99
				}
				mbDownloaded := float64(downloaded) / (1024 * 1024)
				mbTotal := float64(totalBytes) / (1024 * 1024)
				if progressCb != nil {
					progressCb(pct, fmt.Sprintf("[download] 正在极速下载: %.2fMB / %.2fMB (%d%%)", mbDownloaded, mbTotal, pct))
				}
			}
		}

		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return "", fmt.Errorf("下载流中断: %w", rErr)
		}
	}

	_ = outFile.Close()
	// Move temp file to final destination (supports cross-volume cache -> download folder)
	if err := moveFile(tempPath, outPath); err != nil {
		return "", fmt.Errorf("移动完成文件失败: %w", err)
	}
	_ = os.RemoveAll(workDir)

	if progressCb != nil {
		progressCb(100, "推特视频下载完成！")
	}

	return outPath, nil
}
