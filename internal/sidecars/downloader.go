package sidecars

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ProgressCallback func(percent int, message string)

var staticGitHubIPs = map[string][]string{
	"github.com":                           {"20.205.243.166", "140.82.113.3", "140.82.112.4"},
	"api.github.com":                       {"20.205.243.168", "140.82.114.6"},
	"objects.githubusercontent.com":        {"185.199.108.133", "185.199.109.133", "185.199.110.133", "185.199.111.133"},
	"raw.githubusercontent.com":            {"185.199.108.133", "185.199.109.133", "185.199.110.133", "185.199.111.133"},
	"release-assets.githubusercontent.com": {"185.199.108.133", "185.199.109.133", "185.199.110.133", "185.199.111.133"},
}

func newSafeHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			ForceAttemptHTTP2: false,
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				host, port, err := net.SplitHostPort(address)
				if err != nil {
					return nil, err
				}
				if ips, ok := staticGitHubIPs[strings.ToLower(host)]; ok {
					var lastErr error
					for _, ip := range ips {
						dialCtx, dialCancel := context.WithTimeout(ctx, 3*time.Second)
						d := &net.Dialer{Timeout: 3 * time.Second}
						conn, err := d.DialContext(dialCtx, network, net.JoinHostPort(ip, port))
						dialCancel()
						if err == nil {
							return conn, nil
						}
						lastErr = err
					}
					if lastErr != nil {
						// fallback to system dial if static fails
					}
				}
				d := &net.Dialer{Timeout: 5 * time.Second}
				return d.DialContext(ctx, network, address)
			},
		},
		Timeout: timeout,
	}
}

// GetBinDir returns the target binary directory (e.g. ./bin next to executable, or %LOCALAPPDATA%/OmniVideoRip/bin).
func GetBinDir() string {
	// 1. If ./bin exists in current working directory or can be created
	if absBin, err := filepath.Abs("bin"); err == nil {
		if err := os.MkdirAll(absBin, 0755); err == nil {
			return absBin
		}
	}

	// 2. Next to running executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		if !strings.Contains(strings.ToLower(exeDir), "temp") {
			binDir := filepath.Join(exeDir, "bin")
			if err := os.MkdirAll(binDir, 0755); err == nil {
				return binDir
			}
		}
	}

	// 3. Fallback to User AppData
	if userConfig, err := os.UserConfigDir(); err == nil {
		appBin := filepath.Join(userConfig, "OmniVideoRip", "bin")
		_ = os.MkdirAll(appBin, 0755)
		return appBin
	}
	return "bin"
}

// ResolveYTDLPPath searches for yt-dlp.exe in priority order:
// 1. App-relative ./bin/yt-dlp.exe or ./yt-dlp.exe or ../bin/yt-dlp.exe
// 2. Working-dir ./bin/yt-dlp.exe
// 3. AppData bin/yt-dlp.exe
// 4. System PATH
func ResolveYTDLPPath(hintPath string) (string, string) {
	if hintPath != "" && fileExists(hintPath) {
		return hintPath, "custom"
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		// 1. Check exeDir/bin/yt-dlp.exe
		localBin := filepath.Join(exeDir, "bin", "yt-dlp.exe")
		if fileExists(localBin) {
			return localBin, "local_bin"
		}
		// 2. Check same directory: exeDir/yt-dlp.exe
		sameDir := filepath.Join(exeDir, "yt-dlp.exe")
		if fileExists(sameDir) {
			return sameDir, "local_bin"
		}
		// 3. Check parent directory bin: exeDir/../bin/yt-dlp.exe (e.g. if exe is in release/)
		parentBin := filepath.Join(filepath.Dir(exeDir), "bin", "yt-dlp.exe")
		if fileExists(parentBin) {
			return parentBin, "local_bin"
		}
	}

	if fileExists("bin/yt-dlp.exe") {
		abs, _ := filepath.Abs("bin/yt-dlp.exe")
		return abs, "local_bin"
	}
	if fileExists("yt-dlp.exe") {
		abs, _ := filepath.Abs("yt-dlp.exe")
		return abs, "local_bin"
	}

	if userConfig, err := os.UserConfigDir(); err == nil {
		appBin := filepath.Join(userConfig, "OmniVideoRip", "bin", "yt-dlp.exe")
		if fileExists(appBin) {
			return appBin, "appdata"
		}
	}

	if sysPath, err := exec.LookPath("yt-dlp"); err == nil && sysPath != "" {
		return sysPath, "system"
	}
	if sysPath, err := exec.LookPath("yt-dlp.exe"); err == nil && sysPath != "" {
		return sysPath, "system"
	}

	return "", "none"
}

// ResolveFFmpegPath searches for ffmpeg.exe in priority order.
func ResolveFFmpegPath(hintPath string) (string, string) {
	if hintPath != "" && fileExists(hintPath) {
		return hintPath, "custom"
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		// 1. Check exeDir/bin/ffmpeg.exe
		localBin := filepath.Join(exeDir, "bin", "ffmpeg.exe")
		if fileExists(localBin) {
			return localBin, "local_bin"
		}
		// 2. Check same directory: exeDir/ffmpeg.exe
		sameDir := filepath.Join(exeDir, "ffmpeg.exe")
		if fileExists(sameDir) {
			return sameDir, "local_bin"
		}
		// 3. Check parent directory bin: exeDir/../bin/ffmpeg.exe
		parentBin := filepath.Join(filepath.Dir(exeDir), "bin", "ffmpeg.exe")
		if fileExists(parentBin) {
			return parentBin, "local_bin"
		}
	}

	if fileExists("bin/ffmpeg.exe") {
		abs, _ := filepath.Abs("bin/ffmpeg.exe")
		return abs, "local_bin"
	}
	if fileExists("ffmpeg.exe") {
		abs, _ := filepath.Abs("ffmpeg.exe")
		return abs, "local_bin"
	}

	if userConfig, err := os.UserConfigDir(); err == nil {
		appBin := filepath.Join(userConfig, "OmniVideoRip", "bin", "ffmpeg.exe")
		if fileExists(appBin) {
			return appBin, "appdata"
		}
	}

	if sysPath, err := exec.LookPath("ffmpeg"); err == nil && sysPath != "" {
		return sysPath, "system"
	}
	if sysPath, err := exec.LookPath("ffmpeg.exe"); err == nil && sysPath != "" {
		return sysPath, "system"
	}

	return "", "none"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Size() > 0
}

// CheckLatestYTDLPVersion queries GitHub API for the latest release tag.
func CheckLatestYTDLPVersion(ctx context.Context) (string, error) {
	client := newSafeHTTPClient(10 * time.Second)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "OmniVideoRip-Updater/1.3.6")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("连接 GitHub 检查更新失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub 返回状态码: %d", resp.StatusCode)
	}

	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.TagName, nil
}

// DownloadOrUpdateYTDLP downloads the latest yt-dlp.exe to the local bin directory with progress reporting.
func DownloadOrUpdateYTDLP(ctx context.Context, progressCb ProgressCallback) (string, error) {
	targetDir := GetBinDir()
	targetPath := filepath.Join(targetDir, "yt-dlp.exe")
	tempPath := targetPath + ".download.tmp"

	if progressCb != nil {
		progressCb(5, "正在连接下载源...")
	}

	downloadSources := []string{
		"https://ghproxy.net/https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe",
		"https://mirror.ghproxy.com/https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe",
		"https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe",
	}

	clients := []*http.Client{
		{Timeout: 10 * time.Minute},
		newSafeHTTPClient(10 * time.Minute),
	}

	var resp *http.Response
	var lastErr error

	for _, cl := range clients {
		for _, src := range downloadSources {
			req, err := http.NewRequestWithContext(ctx, "GET", src, nil)
			if err != nil {
				lastErr = err
				continue
			}
			req.Header.Set("User-Agent", "OmniVideoRip-Updater/1.3.6")

			r, err := cl.Do(req)
			if err == nil && r.StatusCode == http.StatusOK {
				resp = r
				break
			}
			if r != nil {
				r.Body.Close()
			}
			lastErr = err
		}
		if resp != nil {
			break
		}
	}

	if resp == nil {
		return "", fmt.Errorf("所有下载节点连接失败: %v", lastErr)
	}
	defer resp.Body.Close()

	totalBytes := resp.ContentLength
	out, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}

	var written int64
	buf := make([]byte, 64*1024)
	lastReport := time.Now()

	for {
		if ctx.Err() != nil {
			out.Close()
			os.Remove(tempPath)
			return "", ctx.Err()
		}

		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := out.Write(buf[:n]); wErr != nil {
				out.Close()
				os.Remove(tempPath)
				return "", wErr
			}
			written += int64(n)

			if totalBytes > 0 && time.Since(lastReport) > 100*time.Millisecond {
				pct := int(float64(written) / float64(totalBytes) * 90)
				if progressCb != nil {
					mbWritten := float64(written) / (1024 * 1024)
					mbTotal := float64(totalBytes) / (1024 * 1024)
					progressCb(10+pct, fmt.Sprintf("正在下载 yt-dlp: %.1f MB / %.1f MB (%d%%)", mbWritten, mbTotal, pct))
				}
				lastReport = time.Now()
			}
		}

		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			out.Close()
			os.Remove(tempPath)
			return "", rErr
		}
	}

	if err := out.Close(); err != nil {
		os.Remove(tempPath)
		return "", err
	}

	if written < 10*1024*1024 {
		os.Remove(tempPath)
		return "", fmt.Errorf("下载文件体积异常 (仅 %d 字节)，可能下载不完整", written)
	}

	if progressCb != nil {
		progressCb(98, "下载完成，正在安装替换...")
	}

	_ = os.Remove(targetPath)
	if err := os.Rename(tempPath, targetPath); err != nil {
		return "", fmt.Errorf("安装 yt-dlp.exe 失败: %w", err)
	}

	if progressCb != nil {
		progressCb(100, "yt-dlp 已成功更新至最新版本！")
	}

	return targetPath, nil
}

// DownloadOrUpdateFFmpeg downloads the essentials ffmpeg build and extracts ffmpeg.exe.
func DownloadOrUpdateFFmpeg(ctx context.Context, progressCb ProgressCallback) (string, error) {
	targetDir := GetBinDir()
	targetPath := filepath.Join(targetDir, "ffmpeg.exe")
	zipPath := filepath.Join(targetDir, "ffmpeg_download.tmp.zip")

	if progressCb != nil {
		progressCb(5, "正在连接 FFmpeg 官方镜像构建源...")
	}

	downloadSources := []string{
		"https://ghproxy.net/https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
		"https://mirror.ghproxy.com/https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
		"https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
	}

	clients := []*http.Client{
		{Timeout: 15 * time.Minute},
		newSafeHTTPClient(15 * time.Minute),
	}

	var resp *http.Response
	var lastErr error

	for _, cl := range clients {
		for _, src := range downloadSources {
			req, err := http.NewRequestWithContext(ctx, "GET", src, nil)
			if err != nil {
				lastErr = err
				continue
			}
			req.Header.Set("User-Agent", "OmniVideoRip-Updater/1.3.6")

			r, err := cl.Do(req)
			if err == nil && r.StatusCode == http.StatusOK {
				resp = r
				break
			}
			if r != nil {
				r.Body.Close()
			}
			lastErr = err
		}
		if resp != nil {
			break
		}
	}

	if resp == nil {
		return "", fmt.Errorf("FFmpeg 下载源连接失败: %v", lastErr)
	}
	defer resp.Body.Close()

	totalBytes := resp.ContentLength
	out, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}

	var written int64
	buf := make([]byte, 64*1024)
	lastReport := time.Now()

	for {
		if ctx.Err() != nil {
			out.Close()
			os.Remove(zipPath)
			return "", ctx.Err()
		}

		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := out.Write(buf[:n]); wErr != nil {
				out.Close()
				os.Remove(zipPath)
				return "", wErr
			}
			written += int64(n)

			if totalBytes > 0 && time.Since(lastReport) > 100*time.Millisecond {
				pct := int(float64(written) / float64(totalBytes) * 85)
				if progressCb != nil {
					mbWritten := float64(written) / (1024 * 1024)
					mbTotal := float64(totalBytes) / (1024 * 1024)
					progressCb(5+pct, fmt.Sprintf("正在下载 FFmpeg: %.1f MB / %.1f MB (%d%%)", mbWritten, mbTotal, pct))
				}
				lastReport = time.Now()
			}
		}

		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			out.Close()
			os.Remove(zipPath)
			return "", rErr
		}
	}
	out.Close()

	if progressCb != nil {
		progressCb(92, "正在解压并安装 ffmpeg.exe...")
	}

	if err := extractFFmpegExe(zipPath, targetPath); err != nil {
		os.Remove(zipPath)
		return "", fmt.Errorf("解压 ffmpeg 失败: %w", err)
	}
	os.Remove(zipPath)

	if progressCb != nil {
		progressCb(100, "FFmpeg 已成功安装/更新至最新版本！")
	}

	return targetPath, nil
}

func extractFFmpegExe(zipPath, targetExePath string) error {
	r, err := os.Open(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	fi, err := r.Stat()
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(r, fi.Size())
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if strings.EqualFold(filepath.Base(f.Name), "ffmpeg.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			tempTarget := targetExePath + ".tmp"
			out, err := os.Create(tempTarget)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				os.Remove(tempTarget)
				return err
			}
			out.Close()

			_ = os.Remove(targetExePath)
			return os.Rename(tempTarget, targetExePath)
		}
	}
	return fmt.Errorf("未在压缩包中找到 ffmpeg.exe")
}
