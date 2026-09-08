package sidecars

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxUpdateBytes int64 = 512 << 20

type componentSpec struct{ Repo, Asset, Executable string }

var ytDlpSpec = componentSpec{"yt-dlp/yt-dlp", "yt-dlp.exe", "yt-dlp.exe"}
var ffmpegSpec = componentSpec{"BtbN/FFmpeg-Builds", "ffmpeg-master-latest-win64-gpl.zip", "ffmpeg.exe"}
var componentUpdateMu sync.Mutex

type verifiedAsset struct {
	Tag, URL, Digest string
	Size             int64
}
type installedComponent struct {
	AssetDigest      string `json:"assetDigest"`
	ExecutableDigest string `json:"executableDigest"`
	Tag              string `json:"tag"`
}

func latestVerifiedAsset(ctx context.Context, client *http.Client, spec componentSpec) (verifiedAsset, error) {
	guarded := *client
	guarded.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || req.URL.Scheme != "https" || req.URL.Host != "api.github.com" {
			return fmt.Errorf("untrusted release metadata redirect")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+spec.Repo+"/releases/latest", nil)
	if err != nil {
		return verifiedAsset{}, err
	}
	req.Header.Set("User-Agent", "OmniVideoRip-Updater/1.3.9")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := guarded.Do(req)
	if err != nil {
		return verifiedAsset{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return verifiedAsset{}, fmt.Errorf("official metadata HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	if err != nil {
		return verifiedAsset{}, err
	}
	if len(data) > 4<<20 {
		return verifiedAsset{}, fmt.Errorf("release metadata too large")
	}
	var release struct {
		Tag    string `json:"tag_name"`
		Assets []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			Digest string `json:"digest"`
			Size   int64  `json:"size"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(data, &release); err != nil {
		return verifiedAsset{}, err
	}
	for _, asset := range release.Assets {
		if asset.Name != spec.Asset {
			continue
		}
		hash := strings.TrimPrefix(asset.Digest, "sha256:")
		decoded, err := hex.DecodeString(hash)
		if err != nil || len(decoded) != sha256.Size || !strings.HasPrefix(asset.Digest, "sha256:") {
			return verifiedAsset{}, fmt.Errorf("official SHA-256 missing for %s; refusing installation", spec.Asset)
		}
		u, err := url.Parse(asset.URL)
		if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || !strings.HasPrefix(u.Path, "/"+spec.Repo+"/releases/download/") {
			return verifiedAsset{}, fmt.Errorf("unexpected official asset URL")
		}
		if asset.Size <= 0 || asset.Size > maxUpdateBytes {
			return verifiedAsset{}, fmt.Errorf("invalid asset size")
		}
		return verifiedAsset{release.Tag, asset.URL, strings.ToLower(hash), asset.Size}, nil
	}
	return verifiedAsset{}, fmt.Errorf("official release has no %s", spec.Asset)
}

func executableDigest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	header := make([]byte, 2)
	if _, err := io.ReadFull(f, header); err != nil || string(header) != "MZ" {
		return "", fmt.Errorf("invalid Windows executable")
	}
	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, maxUpdateBytes+1))
	if err != nil {
		return "", err
	}
	if n > maxUpdateBytes {
		return "", fmt.Errorf("executable too large")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func componentIsCurrent(path string, asset verifiedAsset) bool {
	data, err := os.ReadFile(path + ".release.json")
	if err != nil {
		return false
	}
	var installed installedComponent
	if json.Unmarshal(data, &installed) != nil || installed.AssetDigest != asset.Digest {
		return false
	}
	actual, err := executableDigest(path)
	return err == nil && actual == installed.ExecutableDigest
}

func IsFFmpegCurrent(ctx context.Context, path string) (bool, error) {
	asset, err := latestVerifiedAsset(ctx, newSafeHTTPClient(20*time.Second), ffmpegSpec)
	if err != nil {
		return false, err
	}
	return componentIsCurrent(path, asset), nil
}

func downloadVerifiedAsset(ctx context.Context, client *http.Client, asset verifiedAsset, dir string, progress ProgressCallback) (string, error) {
	// Mirrors may transport bytes, but only the official metadata authenticates them.
	sources := []string{asset.URL, "https://ghproxy.net/" + asset.URL, "https://mirror.ghproxy.com/" + asset.URL}
	var lastErr error
	for _, source := range sources {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "OmniVideoRip-Updater/1.3.9")
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		path, err := saveVerifiedResponse(resp, asset, dir, progress)
		resp.Body.Close()
		if err == nil {
			return path, nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("verified download failed: %w", lastErr)
}

func saveVerifiedResponse(resp *http.Response, asset verifiedAsset, dir string, progress ProgressCallback) (path string, err error) {
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("asset HTTP %d", resp.StatusCode)
	}
	f, err := os.CreateTemp(dir, ".component-*.download")
	if err != nil {
		return "", err
	}
	name := f.Name()
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(name)
		}
	}()
	h := sha256.New()
	if progress != nil {
		progress(15, "正在下载并校验组件...")
	}
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, asset.Size+1))
	if err != nil {
		return "", err
	}
	if n != asset.Size || hex.EncodeToString(h.Sum(nil)) != asset.Digest {
		return "", fmt.Errorf("asset SHA-256/size verification failed")
	}
	if err = f.Sync(); err != nil {
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	return name, nil
}

func unpackVerifiedExecutable(archive, dir, executable string) (path string, err error) {
	zr, err := zip.OpenReader(archive)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	for _, entry := range zr.File {
		if filepath.Base(entry.Name) != executable || entry.FileInfo().IsDir() {
			continue
		}
		if entry.UncompressedSize64 > uint64(maxUpdateBytes) {
			return "", fmt.Errorf("executable archive member too large")
		}
		rc, err := entry.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		f, err := os.CreateTemp(dir, ".component-*.exe")
		if err != nil {
			return "", err
		}
		name := f.Name()
		defer func() {
			f.Close()
			if err != nil {
				os.Remove(name)
			}
		}()
		n, err := io.Copy(f, io.LimitReader(rc, maxUpdateBytes+1))
		if err != nil {
			return "", err
		}
		if n <= 0 || n > maxUpdateBytes {
			return "", fmt.Errorf("invalid extracted executable size")
		}
		if err = f.Sync(); err != nil {
			return "", err
		}
		if err = f.Close(); err != nil {
			return "", err
		}
		if _, err = executableDigest(name); err != nil {
			return "", err
		}
		return name, nil
	}
	return "", fmt.Errorf("archive has no %s", executable)
}

func updateVerifiedComponent(ctx context.Context, spec componentSpec, progress ProgressCallback) (string, error) {
	componentUpdateMu.Lock()
	defer componentUpdateMu.Unlock()
	dir := GetBinDir()
	target := filepath.Join(dir, spec.Executable)
	client := newSafeHTTPClient(10 * time.Minute)
	metadataClient := newSafeHTTPClient(20 * time.Second)
	asset, err := latestVerifiedAsset(ctx, metadataClient, spec)
	if err != nil {
		return "", err
	}
	if componentIsCurrent(target, asset) {
		if progress != nil {
			progress(100, "组件已是最新版本")
		}
		return target, nil
	}
	downloaded, err := downloadVerifiedAsset(ctx, client, asset, dir, progress)
	if err != nil {
		return "", err
	}
	defer os.Remove(downloaded)
	candidate := downloaded
	if strings.HasSuffix(spec.Asset, ".zip") {
		candidate, err = unpackVerifiedExecutable(downloaded, dir, spec.Executable)
		if err != nil {
			return "", err
		}
		defer os.Remove(candidate)
	}
	digest, err := executableDigest(candidate)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err = replaceComponentFile(candidate, target); err != nil {
		return "", fmt.Errorf("install failed (previous component retained): %w", err)
	}
	marker, _ := json.Marshal(installedComponent{asset.Digest, digest, asset.Tag})
	f, err := os.CreateTemp(dir, ".component-*.json")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(marker); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = replaceComponentFile(f.Name(), target+".release.json"); err != nil {
		return "", err
	}
	if progress != nil {
		progress(100, "校验通过，组件更新完成")
	}
	return target, nil
}
