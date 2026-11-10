package extractors

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"videodownloader/internal/sidecars"
)

type ProgressCallback func(percent int, message string)
type taskCacheKey struct{}

type VideoFormatOption struct {
	FormatID    string `json:"formatId"`
	Resolution  string `json:"resolution"`
	Extension   string `json:"ext"`
	FileSizeStr string `json:"fileSizeStr"`
	Fps         int    `json:"fps"`
	Note        string `json:"note"`
}

type VideoMetaData struct {
	Title     string              `json:"title"`
	Thumbnail string              `json:"thumbnail"`
	Duration  string              `json:"duration"`
	Platform  string              `json:"platform"`
	URL       string              `json:"url"`
	Formats   []VideoFormatOption `json:"formats"`
}

type ExtractorRouter struct {
	outputDir  string
	ytdlpPath  string
	ffmpegPath string
}

func NewRouter(outputDir, ytdlpPath, ffmpegPath string) *ExtractorRouter {
	return &ExtractorRouter{
		outputDir:  outputDir,
		ytdlpPath:  ytdlpPath,
		ffmpegPath: ffmpegPath,
	}
}

func (r *ExtractorRouter) Analyze(ctx context.Context, pageURL string) (VideoMetaData, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return VideoMetaData{}, fmt.Errorf("invalid video URL: %w", err)
	}
	platform := detectPlatform(pageURL)

	if platform == "missav" {
		return AnalyzeMissAV(ctx, pageURL)
	}
	if platform == "ggjav" {
		return AnalyzeGGJav(ctx, pageURL)
	}
	if platform == "rouvideo" {
		return AnalyzeRouVideo(ctx, pageURL)
	}
	if platform == "supjav" {
		return AnalyzeSupJav(ctx, pageURL)
	}

	if platform == "x" {
		if meta, err := AnalyzeTwitter(ctx, pageURL); err == nil {
			return meta, nil
		}
	}

	ytPath, _ := sidecars.ResolveYTDLPPath(r.ytdlpPath)
	if ytPath == "" {
		return VideoMetaData{}, fmt.Errorf("全网解析依赖 (yt-dlp) 未就绪")
	}

	args := []string{
		"--ignore-config",
		"--no-playlist",
		"--dump-single-json",
		"--encoding", "utf-8",
	}
	args = append(args, getJSRuntimeArgs()...)
	if u, err := url.Parse(pageURL); err == nil {
		if proxyURL, _ := GetSystemProxyURL(u); proxyURL != nil {
			args = append(args, "--proxy", proxyURL.String())
		}
	}
	args = append(args, pageURL)

	cmd := exec.CommandContext(ctx, ytPath, args...)
	hideCmdWindow(cmd)

	outBytes, err := cmd.Output()
	if err != nil {
		return VideoMetaData{}, fmt.Errorf("解析视频元信息失败: 请检查网络连接或链接有效性")
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(outBytes, &rawData); err != nil {
		return VideoMetaData{}, fmt.Errorf("解析视频 JSON 失败: %w", err)
	}

	title, _ := rawData["title"].(string)
	if title == "" {
		title = "全网视频"
	}
	thumbnail, _ := rawData["thumbnail"].(string)
	durationSec, _ := rawData["duration"].(float64)
	durationStr := formatDuration(int(durationSec))

	formats := []VideoFormatOption{
		{FormatID: "best", Resolution: "最佳画质 (Best Quality)", Extension: "mp4", Note: "自动拼接最高画质与最佳音轨 (推荐)"},
		{FormatID: "1080p", Resolution: "1080P 全高清", Extension: "mp4", Note: "1080p 60fps / 30fps"},
		{FormatID: "720p", Resolution: "720P 高清", Extension: "mp4", Note: "720p 高清画质"},
		{FormatID: "480p", Resolution: "480P 标清", Extension: "mp4", Note: "480p 标清画质"},
		{FormatID: "audio", Resolution: "纯音频提取 (MP3)", Extension: "mp3", Note: "无损分离背景音频/音乐"},
	}

	return VideoMetaData{
		Title:     title,
		Thumbnail: thumbnail,
		Duration:  durationStr,
		Platform:  platform,
		URL:       pageURL,
		Formats:   formats,
	}, nil
}

func (r *ExtractorRouter) DownloadWithFormat(ctx context.Context, pageURL, customOutDir, formatID string, progressCb ProgressCallback) (string, error) {
	if err := validateHTTPURL(pageURL); err != nil {
		return "", fmt.Errorf("invalid video URL: %w", err)
	}
	outDir := r.outputDir
	if strings.TrimSpace(customOutDir) != "" {
		outDir = strings.TrimSpace(customOutDir)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	platform := detectPlatform(pageURL)
	ctx = context.WithValue(ctx, taskCacheKey{}, GetTaskResumeDir(outDir, platform, pageURL))

	switch platform {
	case "missav":
		ffPath, _ := sidecars.ResolveFFmpegPath(r.ffmpegPath)
		return DownloadMissAVWithFormat(ctx, pageURL, outDir, ffPath, formatID, progressCb)
	case "ggjav":
		ffPath, _ := sidecars.ResolveFFmpegPath(r.ffmpegPath)
		return DownloadGGJav(ctx, pageURL, outDir, ffPath, progressCb)
	case "rouvideo":
		ffPath, _ := sidecars.ResolveFFmpegPath(r.ffmpegPath)
		return DownloadRouVideo(ctx, r, pageURL, outDir, ffPath, formatID, progressCb)
	case "supjav":
		ffPath, _ := sidecars.ResolveFFmpegPath(r.ffmpegPath)
		return DownloadSupJav(ctx, r, pageURL, outDir, ffPath, formatID, progressCb)

	case "x":
		outPath, err := DownloadTwitter(ctx, pageURL, outDir, formatID, progressCb)
		if err == nil {
			return outPath, nil
		}
		return r.downloadWithYTDLPAndFormat(ctx, pageURL, outDir, formatID, progressCb)
	default:
		return r.downloadWithYTDLPAndFormat(ctx, pageURL, outDir, formatID, progressCb)
	}
}

var pctRegex = regexp.MustCompile(`\[download\]\s+([0-9\.]+)%`)

func getJSRuntimeArgs() []string {
	if _, err := exec.LookPath("node"); err == nil {
		return []string{"--js-runtimes", "node"}
	}
	if _, err := exec.LookPath("deno"); err == nil {
		return []string{"--js-runtimes", "deno"}
	}
	return nil
}

func (r *ExtractorRouter) downloadWithYTDLPAndFormat(ctx context.Context, pageURL, outDir, formatID string, progressCb ProgressCallback) (string, error) {
	ytPath, ytSource := sidecars.ResolveYTDLPPath(r.ytdlpPath)
	if ytPath == "" {
		return "", fmt.Errorf("全网解析依赖 (yt-dlp) 未找到")
	}

	ffPath, _ := sidecars.ResolveFFmpegPath(r.ffmpegPath)

	if progressCb != nil {
		sourceTip := "系统组件"
		if ytSource == "local_bin" {
			sourceTip = "同级 ./bin/ 组件"
		}
		progressCb(5, fmt.Sprintf("正在启动解析引擎 (%s)...", sourceTip))
	}

	primaryClient := ""
	if detectPlatform(pageURL) == "youtube" {
		primaryClient = "web_embedded"
	}

	outPath, err := r.runYTDLP(ctx, ytPath, ffPath, pageURL, outDir, formatID, primaryClient, progressCb)
	if err != nil && ctx.Err() == nil && strings.Contains(err.Error(), "403") && detectPlatform(pageURL) == "youtube" {
		if progressCb != nil {
			progressCb(10, "检测到 YouTube 流媒体防爬限制，正在自动切换兼容模式重试...")
		}
		return r.runYTDLP(ctx, ytPath, ffPath, pageURL, outDir, formatID, "android", progressCb)
	}

	return outPath, err
}

func (r *ExtractorRouter) runYTDLP(ctx context.Context, ytPath, ffPath, pageURL, outDir, formatID, clientOverride string, progressCb ProgressCallback) (string, error) {
	return r.runYTDLPWithReferer(ctx, ytPath, ffPath, pageURL, outDir, formatID, clientOverride, "", progressCb)
}

func (r *ExtractorRouter) runYTDLPWithReferer(ctx context.Context, ytPath, ffPath, pageURL, outDir, formatID, clientOverride, referer string, progressCb ProgressCallback) (string, error) {
	formatArg := "bv*+ba[ext=m4a]/bv*+ba/best"
	if clientOverride == "android" {
		formatArg = "b/best"
	} else if formatID == "1080p" {
		formatArg = "bv*[height<=1080]+ba[ext=m4a]/bv*[height<=1080]+ba/best"
	} else if formatID == "720p" {
		formatArg = "bv*[height<=720]+ba[ext=m4a]/bv*[height<=720]+ba/best"
	} else if formatID == "480p" {
		formatArg = "bv*[height<=480]+ba[ext=m4a]/bv*[height<=480]+ba/best"
	} else if formatID == "audio" {
		formatArg = "ba/b"
	} else if formatID == "best" {
		formatArg = "bv*+ba[ext=m4a]/bv*+ba/best"
	}

	args := []string{
		"--ignore-config",
		"--no-playlist",
		"--newline",
		"--progress",
		"--continue",
		"--encoding", "utf-8",
	}
	cacheDir := GetTaskResumeDir(outDir, detectPlatform(pageURL), pageURL)
	if taskCache, ok := ctx.Value(taskCacheKey{}).(string); ok {
		cacheDir = taskCache
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}
	args = append(args, "--paths", "temp:"+cacheDir)

	args = append(args, getJSRuntimeArgs()...)

	if clientOverride != "" {
		args = append(args, "--extractor-args", fmt.Sprintf("youtube:player_client=%s", clientOverride))
	} else {
		args = append(args, "--extractor-args", "generic:impersonate")
	}

	args = append(args,
		"-f", formatArg,
		"-P", outDir,
		"-o", "%(title)s [%(id)s].%(ext)s",
		"--print", "after_move:__OMNIVIDEORIP_PATH__:%(filepath)s",
	)

	if formatID == "audio" {
		args = append(args, "--extract-audio", "--audio-format", "mp3")
	} else if ffPath != "" {
		args = append(args, "--merge-output-format", "mp4", "--postprocessor-args", "Merger:-c:a aac")
	}

	if ffPath != "" {
		args = append(args, "--ffmpeg-location", filepath.Dir(ffPath))
	}
	if referer != "" {
		args = append(args, "--referer", referer)
	}

	if u, err := url.Parse(pageURL); err == nil {
		if proxyURL, _ := GetSystemProxyURL(u); proxyURL != nil {
			args = append(args, "--proxy", proxyURL.String())
		}
	}

	args = append(args, "--", pageURL)

	cmd := exec.CommandContext(ctx, ytPath, args...)
	hideCmdWindow(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("获取全网解析输出管道失败: %w", err)
	}

	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("启动全网解析引擎失败: %w", err)
	}

	var outputPath string
	stderrDone := make(chan []string, 1)
	go func() {
		var errLines []string
		if stderr != nil {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				errLines = append(errLines, scanner.Text())
			}
		}
		stderrDone <- errLines
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if finalPath, ok := parseYTDLPOutputPath(line); ok {
			outputPath = finalPath
			continue
		}

		if strings.Contains(line, "[download]") {
			matches := pctRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, parseErr := strconv.ParseFloat(matches[1], 64); parseErr == nil {
					pct := int(val)
					if progressCb != nil {
						progressCb(pct, strings.TrimSpace(line))
					}
				}
			} else if progressCb != nil {
				progressCb(50, strings.TrimSpace(line))
			}
		}

		if strings.Contains(line, "Destination: ") {
			parts := strings.Split(line, "Destination: ")
			if len(parts) > 1 {
				outputPath = strings.TrimSpace(parts[1])
			}
		}

		if strings.Contains(line, "has already been downloaded") {
			parts := strings.Split(line, "[download] ")
			if len(parts) > 1 {
				alreadyPath := strings.TrimSuffix(parts[1], " has already been downloaded")
				outputPath = strings.TrimSpace(alreadyPath)
			}
		}
	}

	waitErr := cmd.Wait()
	errLines := <-stderrDone
	if scanErr := scanner.Err(); scanErr != nil && waitErr == nil {
		waitErr = scanErr
	}
	if waitErr != nil {
		errMsg := waitErr.Error()
		if len(errLines) > 0 {
			rawErr := strings.Join(errLines, "\n")
			if strings.Contains(rawErr, "Failed to resolve") || strings.Contains(rawErr, "getaddrinfo failed") || strings.Contains(rawErr, "HTTPSConnection") {
				errMsg = "网络 / DNS 连接失败：无法连接视频站点或媒体 CDN。请检查网络、代理或 Tailscale 出口设置！"
			} else {
				errMsg = rawErr
			}
		}
		return outputPath, fmt.Errorf("全网解析下载失败: %s", errMsg)
	}

	if outputPath == "" {
		return "", fmt.Errorf("download completed but yt-dlp did not report the final output path")
	}
	if info, err := os.Stat(outputPath); err != nil || info.IsDir() || info.Size() == 0 {
		return "", fmt.Errorf("download completed but final output file is missing: %s", outputPath)
	}

	return outputPath, nil
}

func parseYTDLPOutputPath(line string) (string, bool) {
	const prefix = "__OMNIVIDEORIP_PATH__:"
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return "", false
	}
	path := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	return path, path != ""
}

func (r *ExtractorRouter) Download(ctx context.Context, pageURL, customOutDir string, progressCb ProgressCallback) (string, error) {
	return r.DownloadWithFormat(ctx, pageURL, customOutDir, "best", progressCb)
}

func formatDuration(sec int) string {
	if sec <= 0 {
		return "未知时长"
	}
	m := sec / 60
	s := sec % 60
	h := m / 60
	m = m % 60
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func detectPlatform(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "generic"
	}
	host := strings.ToLower(parsed.Hostname())

	switch {
	case matchesDomain(host, "missav.ws"), matchesDomain(host, "missav.live"), matchesDomain(host, "missav.ai"), matchesDomain(host, "missav.com"):
		return "missav"
	case matchesDomain(host, "ggjav.com"), matchesDomain(host, "ggjav.tv"):
		return "ggjav"
	case matchesDomain(host, "rou.video"):
		return "rouvideo"
	case matchesDomain(host, "supjav.com"):
		return "supjav"

	case matchesDomain(host, "youtube.com"), matchesDomain(host, "youtu.be"):
		return "youtube"
	case matchesDomain(host, "bilibili.com"), matchesDomain(host, "b23.tv"):
		return "bilibili"
	case matchesDomain(host, "tiktok.com"):
		return "tiktok"
	case matchesDomain(host, "x.com"), matchesDomain(host, "twitter.com"):
		return "x"
	case matchesDomain(host, "instagram.com"), matchesDomain(host, "instagr.am"):
		return "instagram"
	default:
		if host != "" {
			return strings.TrimPrefix(host, "www.")
		}
		return "generic"
	}
}

func matchesDomain(host, domain string) bool {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	domain = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
	return host == domain || strings.HasSuffix(host, "."+domain)
}
