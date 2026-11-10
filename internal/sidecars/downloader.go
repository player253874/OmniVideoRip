package sidecars

import (
	"context"
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

// CheckLatestYTDLPVersion queries trusted release metadata.
func CheckLatestYTDLPVersion(ctx context.Context) (string, error) {
	asset, err := latestVerifiedAsset(ctx, newSafeHTTPClient(20*time.Second), ytDlpSpec)
	return asset.Tag, err
}

// CheckLatestFFmpegVersion queries trusted release metadata.
func CheckLatestFFmpegVersion(ctx context.Context) (string, error) {
	asset, err := latestVerifiedAsset(ctx, newSafeHTTPClient(20*time.Second), ffmpegSpec)
	return asset.Tag, err
}

func DownloadOrUpdateYTDLP(ctx context.Context, progressCb ProgressCallback) (string, error) {
	return updateVerifiedComponent(ctx, ytDlpSpec, progressCb)
}

func DownloadOrUpdateFFmpeg(ctx context.Context, progressCb ProgressCallback) (string, error) {
	return updateVerifiedComponent(ctx, ffmpegSpec, progressCb)
}
