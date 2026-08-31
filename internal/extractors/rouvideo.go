package extractors

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	rouNextDataRegex        = regexp.MustCompile(`<script id="__NEXT_DATA__" type="application/json">([\s\S]*?)</script>`)
	rouReferenceStreamRegex = regexp.MustCompile(`(?i)var\s+stream\s*=\s*['"]([^'"]+\.m3u8[^'"]*)['"]`)
)

type rouNextData struct {
	Props struct {
		PageProps struct {
			Video struct {
				ID            string  `json:"id"`
				Name          string  `json:"name"`
				NameZh        string  `json:"nameZh"`
				Duration      float64 `json:"duration"`
				CoverImageUrl string  `json:"coverImageUrl"`
				Ref           string  `json:"ref"`
			} `json:"video"`
			Ev struct {
				D string `json:"d"`
				K int    `json:"k"`
			} `json:"ev"`
		} `json:"pageProps"`
	} `json:"props"`
}

type rouDecryptedEv struct {
	VideoURL    string `json:"videoUrl"`
	ThumbVTTURL string `json:"thumbVTTUrl"`
}

func parseRouNextData(html []byte) (*rouNextData, error) {
	match := rouNextDataRegex.FindSubmatch(html)
	if len(match) < 2 {
		return nil, fmt.Errorf("未找到 RouVideo 页面核心数据 (__NEXT_DATA__)")
	}
	var data rouNextData
	if err := json.Unmarshal(match[1], &data); err != nil {
		return nil, fmt.Errorf("解析 RouVideo 数据 JSON 失败: %w", err)
	}
	return &data, nil
}

func decryptRouEv(d string, k int) (*rouDecryptedEv, error) {
	if d == "" {
		return nil, fmt.Errorf("RouVideo 视频加密数据为空")
	}
	rawBytes, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		return nil, fmt.Errorf("Base64 解码失败: %w", err)
	}
	decrypted := make([]byte, len(rawBytes))
	for i, b := range rawBytes {
		decrypted[i] = byte((int(b) - k + 256) % 256)
	}
	var ev rouDecryptedEv
	if err := json.Unmarshal(decrypted, &ev); err != nil {
		return nil, fmt.Errorf("解析解密后的播放数据失败: %w", err)
	}
	return &ev, nil
}

// AnalyzeRouVideo extracts metadata and playback URLs for RouVideo.
func AnalyzeRouVideo(ctx context.Context, pageURL string) (VideoMetaData, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return VideoMetaData{}, err
	}

	client := newSafeHTTPClient(25 * time.Second)
	bodyBytes, err := fetchBytes(ctx, client, pageURL, "https://rou.video/", maxPageBytes)
	if err != nil {
		return VideoMetaData{}, fmt.Errorf("连接 RouVideo 页面失败: %w", err)
	}

	nextData, err := parseRouNextData(bodyBytes)
	if err != nil {
		return VideoMetaData{}, err
	}

	title := nextData.Props.PageProps.Video.NameZh
	if title == "" {
		title = nextData.Props.PageProps.Video.Name
	}
	if title == "" {
		title = "RouVideo 视频"
	}

	thumb := nextData.Props.PageProps.Video.CoverImageUrl
	durationSec := int(nextData.Props.PageProps.Video.Duration)
	durationStr := formatDuration(durationSec)

	formats := []VideoFormatOption{
		{FormatID: "best", Resolution: "最佳画质 (原画 HLS / 720P / 1080P)", Extension: "mp4", Note: "Go 原生 16 线程满速切片直下 (推荐)"},
		{FormatID: "audio", Resolution: "纯音频提取 (MP3)", Extension: "mp3", Note: "自动抽取原声无损音频"},
	}

	return VideoMetaData{
		Title:     title,
		Thumbnail: thumb,
		Duration:  durationStr,
		Platform:  "rouvideo",
		URL:       pageURL,
		Formats:   formats,
	}, nil
}

func downloadRouPlaylist(ctx context.Context, playlistURL, pageURL, outDir, title, ffmpegPath string, progressCb ProgressCallback) (string, error) {
	client := newSafeHTTPClient(90 * time.Second)
	playlistBytes, finalPlaylistURL, err := fetchBytesWithFinalURL(ctx, client, playlistURL, pageURL, maxPlaylistBytes)
	if err != nil {
		return "", fmt.Errorf("获取播放清单失败: %w", err)
	}
	segmentURLs, err := parseSegmentURLs(finalPlaylistURL, string(playlistBytes))
	if err != nil {
		return "", err
	}

	workDir := resumeDir(outDir, "rouvideo", codeNameFromURL(pageURL, "ROUVIDEO"))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", err
	}
	if progressCb != nil {
		progressCb(12, fmt.Sprintf("发现 %d 个分片，启动稳定并发下载...", len(segmentURLs)))
	}
	segmentFiles, err := downloadSegments(ctx, segmentURLs, workDir, 8, progressCb, func(ctx context.Context, rawURL, destPath string) error {
		return downloadSegmentWithRetry(ctx, client, rawURL, destPath, pageURL)
	})
	if err != nil {
		return "", fmt.Errorf("下载 RouVideo 切片失败: %w", err)
	}
	if progressCb != nil {
		progressCb(92, "正在合并视频切片并输出 MP4...")
	}
	outputPath, err := mergeSegments(ctx, segmentFiles, outDir, title, ffmpegPath)
	if err != nil {
		return "", err
	}
	_ = os.RemoveAll(workDir)
	return outputPath, nil
}

func downloadRouReference(ctx context.Context, router *ExtractorRouter, refURL, pageURL, outDir, ffmpegPath, formatID string, progressCb ProgressCallback) (string, error) {
	if err := validateHTTPURL(refURL); err != nil {
		return "", fmt.Errorf("RouVideo 备用来源无效: %w", err)
	}
	if progressCb != nil {
		progressCb(15, "主 CDN 当前网络不可用，正在切换原始来源线路...")
	}
	client := newSafeHTTPClient(60 * time.Second)
	body, finalRefURL, err := fetchBytesWithFinalURL(ctx, client, refURL, pageURL, maxPageBytes)
	if err != nil {
		return "", fmt.Errorf("连接 RouVideo 原始来源失败: %w", err)
	}
	match := rouReferenceStreamRegex.FindSubmatch(body)
	if len(match) < 2 {
		return "", fmt.Errorf("原始来源页面未提供可下载的 HLS 流")
	}
	streamURL := html.UnescapeString(strings.TrimSpace(string(match[1])))
	streamURL, err = resolveURL(finalRefURL, streamURL)
	if err != nil {
		return "", fmt.Errorf("原始来源 HLS 地址无效: %w", err)
	}
	return router.runYTDLPWithReferer(ctx, router.ytdlpPath, ffmpegPath, streamURL, outDir, formatID, "", finalRefURL, progressCb)
}

// DownloadRouVideo downloads and merges the RouVideo HLS stream.
func DownloadRouVideo(ctx context.Context, router *ExtractorRouter, pageURL, outDir, ffmpegPath, formatID string, progressCb ProgressCallback) (string, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return "", err
	}
	if progressCb != nil {
		progressCb(5, "正在连接 RouVideo 并解密流媒体播放索引...")
	}

	client := newSafeHTTPClient(60 * time.Second)
	bodyBytes, err := fetchBytes(ctx, client, pageURL, "https://rou.video/", maxPageBytes)
	if err != nil {
		return "", fmt.Errorf("连接 RouVideo 页面失败: %w", err)
	}

	nextData, err := parseRouNextData(bodyBytes)
	if err != nil {
		return "", err
	}

	evData, err := decryptRouEv(nextData.Props.PageProps.Ev.D, nextData.Props.PageProps.Ev.K)
	if err != nil {
		return "", err
	}
	if evData.VideoURL == "" {
		return "", fmt.Errorf("未找到有效视频流地址")
	}

	title := nextData.Props.PageProps.Video.NameZh
	if title == "" {
		title = nextData.Props.PageProps.Video.Name
	}
	if title == "" {
		title = nextData.Props.PageProps.Video.ID
	}
	if title == "" {
		title = codeNameFromURL(pageURL, "ROUVIDEO")
	}

	if progressCb != nil {
		progressCb(10, "正在解析 HLS 切片清单 (Playlist)...")
	}

	playlistURL, resolveErr := resolveURL(pageURL, evData.VideoURL)
	var primaryErr error
	if resolveErr == nil {
		outputPath, err := downloadRouPlaylist(ctx, playlistURL, pageURL, outDir, title, ffmpegPath, progressCb)
		if err == nil {
			if progressCb != nil {
				progressCb(100, "RouVideo 视频下载并合并完成！")
			}
			return outputPath, nil
		}
		primaryErr = err
	} else {
		primaryErr = fmt.Errorf("RouVideo HLS 地址无效: %w", resolveErr)
	}

	if refURL := strings.TrimSpace(nextData.Props.PageProps.Video.Ref); refURL != "" {
		outputPath, fallbackErr := downloadRouReference(ctx, router, refURL, pageURL, outDir, ffmpegPath, formatID, progressCb)
		if fallbackErr == nil {
			if progressCb != nil {
				progressCb(100, "RouVideo 已通过原始来源线路下载完成！")
			}
			return outputPath, nil
		}
		return "", fmt.Errorf("RouVideo 主线路失败: %v; 备用线路失败: %w", primaryErr, fallbackErr)
	}
	return "", primaryErr
}
