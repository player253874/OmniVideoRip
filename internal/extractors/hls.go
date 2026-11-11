package extractors

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	maxPageBytes     = 8 << 20
	maxPlaylistBytes = 4 << 20
	maxSegmentBytes  = 512 << 20
)

type segmentDownloader func(context.Context, string, string) error

func validateHTTPURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return fmt.Errorf("invalid absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
	}
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if strings.Contains(host, "%") {
		return fmt.Errorf("scoped IP destinations are not allowed")
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return fmt.Errorf("local network targets are not allowed")
	}
	if ip := net.ParseIP(host); ip != nil && isPrivateIP(ip) {
		return fmt.Errorf("local network targets are not allowed")
	}
	return nil
}

func ValidateHTTPURL(rawURL string) error {
	return validateHTTPURL(rawURL)
}

func isPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

var fallbackDNS = []string{
	"1.1.1.1:53",         // Cloudflare DNS
	"8.8.8.8:53",         // Google DNS
	"9.9.9.9:53",         // Quad9 DNS
	"223.5.5.5:53",       // AliDNS
	"119.29.29.29:53",    // DNSPod
	"114.114.114.114:53", // 114DNS
}

// Known fake/poisoned IP addresses commonly injected by DNS pollution
var poisonedIPs = map[string]bool{
	"108.160.173.207": true,
	"59.24.3.173":     true,
	"78.16.49.15":     true,
	"93.46.8.89":      true,
	"37.61.54.158":    true,
	"203.98.7.65":     true,
	"243.185.187.39":  true,
	"46.82.174.68":    true,
	"77.4.7.92":       true,
	"118.5.49.6":      true,
	"159.106.121.75":  true,
}

var staticHostIPs = map[string][]string{
	"ggjav.com":         {"172.67.137.241", "104.21.86.241", "104.21.87.241"},
	"video-1.ggjav.com": {"148.113.161.25"},
	"video-2.ggjav.com": {"148.113.161.25"},
	"video-3.ggjav.com": {"51.81.208.160"},
	"video-4.ggjav.com": {"51.81.208.160"},
	"video-5.ggjav.com": {"51.81.242.19"},
	"video-6.ggjav.com": {"51.81.242.19"},
	"missav.live":       {"104.26.7.107", "104.26.6.107", "172.67.72.106"},
	"missav.ws":         {"104.26.7.107", "104.26.6.107", "172.67.72.106"},
	"missav.ai":         {"104.26.7.107", "104.26.6.107", "172.67.72.106"},
	"missav.com":        {"104.26.7.107", "104.26.6.107", "172.67.72.106"},
	"surrit.com":        {"104.21.75.195", "172.67.195.143"},
}

type dnsCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
}

var resolvedDNSCache = struct {
	sync.Mutex
	entries map[string]dnsCacheEntry
}{entries: make(map[string]dnsCacheEntry)}

func cachedHostIPs(host string) []net.IP {
	resolvedDNSCache.Lock()
	defer resolvedDNSCache.Unlock()
	entry, ok := resolvedDNSCache.entries[host]
	if !ok || time.Now().After(entry.expiresAt) {
		delete(resolvedDNSCache.entries, host)
		return nil
	}
	return append([]net.IP(nil), entry.ips...)
}

func cacheHostIPs(host string, ips []net.IP) {
	if len(ips) == 0 {
		return
	}
	resolvedDNSCache.Lock()
	resolvedDNSCache.entries[host] = dnsCacheEntry{
		ips:       append([]net.IP(nil), ips...),
		expiresAt: time.Now().Add(5 * time.Minute),
	}
	resolvedDNSCache.Unlock()
}

// collectAllIPs prefers the system resolver so VPN and Tailscale DNS/routing
// remain authoritative. Public resolvers are queried concurrently only when
// the system resolver and static mappings produce no usable address.
func collectAllIPs(ctx context.Context, host string) []net.IP {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) || poisonedIPs[ip.String()] {
			return nil
		}
		return []net.IP{ip}
	}
	if cached := cachedHostIPs(host); len(cached) > 0 {
		return cached
	}

	seen := make(map[string]bool)
	var result []net.IP
	addIP := func(ip net.IP) {
		s := ip.String()
		if !seen[s] && !isPrivateIP(ip) && !poisonedIPs[s] {
			seen[s] = true
			result = append(result, ip)
		}
	}

	if staticIPs, ok := staticHostIPs[host]; ok {
		for _, s := range staticIPs {
			if parsed := net.ParseIP(s); parsed != nil {
				addIP(parsed)
			}
		}
	}

	sysCtx, sysCancel := context.WithTimeout(ctx, 3*time.Second)
	sysIPs, sysErr := net.DefaultResolver.LookupIP(sysCtx, "ip", host)
	sysCancel()
	if sysErr == nil {
		for _, ip := range sysIPs {
			addIP(ip)
		}
	}
	if len(result) > 0 {
		cacheHostIPs(host, result)
		return result
	}

	lookupCtx, lookupCancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer lookupCancel()
	responses := make(chan []net.IP, len(fallbackDNS))
	for _, dnsAddr := range fallbackDNS {
		addr := dnsAddr
		go func() {
			resolver := &net.Resolver{
				PreferGo: true,
				Dial: func(c context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{Timeout: 2 * time.Second}
					return d.DialContext(c, "udp", addr)
				},
			}
			ips, _ := resolver.LookupIP(lookupCtx, "ip", host)
			responses <- ips
		}()
	}
	for range fallbackDNS {
		select {
		case ips := <-responses:
			for _, ip := range ips {
				addIP(ip)
			}
		case <-lookupCtx.Done():
			cacheHostIPs(host, result)
			return result
		}
	}
	cacheHostIPs(host, result)
	return result
}

type proxyAddressKey struct{}

func proxyDialAddress(proxy *url.URL) string {
	port := proxy.Port()
	if port == "" {
		switch proxy.Scheme {
		case "https":
			port = "443"
		case "socks5", "socks5h":
			port = "1080"
		default:
			port = "80"
		}
	}
	return net.JoinHostPort(proxy.Hostname(), port)
}

func newSafeHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			if err := validateHTTPURL(req.URL.String()); err != nil {
				return nil, err
			}
			proxyURL, err := GetSystemProxyURL(req.URL)
			proxyAddress := ""
			if proxyURL != nil && err == nil {
				proxyAddress = proxyDialAddress(proxyURL)
			}
			// Explicit user-configured proxies are trusted to resolve the remote target.
			*req = *req.WithContext(context.WithValue(req.Context(), proxyAddressKey{}, proxyAddress))
			return proxyURL, err
		},
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 25 * time.Second,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialSafeAddress(ctx, network, address, collectAllIPs)
		},
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return validateHTTPURL(req.URL.String())
		},
	}
}

func dialSafeAddress(ctx context.Context, network, address string, lookup func(context.Context, string) []net.IP) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	// If host is loopback or local address (e.g. local proxy 127.0.0.1:7890), dial directly without DNS lookup
	if proxyAddress, _ := ctx.Value(proxyAddressKey{}).(string); proxyAddress != "" && proxyAddress == address {
		d := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
		return d.DialContext(ctx, network, address)
	}
	if err := validateHTTPURL("http://" + address); err != nil {
		return nil, err
	}

	ips := lookup(ctx, host)
	if len(ips) == 0 {
		return nil, fmt.Errorf("lookup %s: no permitted public addresses", host)
	}

	// Try each IP with a per-IP dial timeout
	perIPTimeout := 4 * time.Second
	var lastErr error
	for _, ip := range ips {
		if isPrivateIP(ip) {
			continue
		}
		dialCtx, dialCancel := context.WithTimeout(ctx, perIPTimeout)
		d := &net.Dialer{Timeout: perIPTimeout, KeepAlive: 30 * time.Second}
		conn, err := d.DialContext(dialCtx, network, net.JoinHostPort(ip.String(), port))
		dialCancel()
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("dial %s: all %d addresses failed, last error: %w", host, len(ips), lastErr)
}

func readLimitedBody(resp *http.Response, limit int64) ([]byte, error) {
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}
	limited := io.LimitReader(resp.Body, limit+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds %d byte limit", limit)
	}
	return data, nil
}

func resolveURL(baseURL, reference string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(strings.TrimSpace(reference))
	if err != nil {
		return "", err
	}
	resolved := base.ResolveReference(ref).String()
	if err := validateHTTPURL(resolved); err != nil {
		return "", err
	}
	return resolved, nil
}

// resumeDir returns a deterministic working directory for a given video URL,
// so interrupted downloads can be resumed by reusing the same directory.
func resumeDir(outDir, platform, videoKey string) string {
	safe := sanitizeFilename(platform + "_" + videoKey)
	if len(safe) > 80 {
		safe = safe[:80]
	}
	return filepath.Join(outDir, ".omnivideorip-resume-"+safe)
}

// GetTaskResumeDir returns the deterministic resume directory for a task.
func GetTaskResumeDir(outDir, platform, rawURL string) string {
	var key string
	switch strings.ToLower(platform) {
	case "missav":
		key = codeNameFromURL(rawURL, "MISSAV")
	case "ggjav":
		key = codeNameFromURL(rawURL, "GGJAV")
	case "rouvideo":
		key = codeNameFromURL(rawURL, "ROUVIDEO")
	default:
		key = codeNameFromURL(rawURL, "VIDEO")
	}
	digest := sha256.Sum256([]byte(rawURL))
	for len(key) > 32 {
		runes := []rune(key)
		key = string(runes[:len(runes)-1])
	}
	return resumeDir(outDir, platform, key+fmt.Sprintf("-%x", digest[:8]))
}

// CleanupTaskResumeCache removes the deterministic resume directory for a given task.
func CleanupTaskResumeCache(outDir, platform, rawURL string) error {
	if strings.TrimSpace(outDir) == "" || strings.TrimSpace(rawURL) == "" {
		return nil
	}
	dir := GetTaskResumeDir(outDir, platform, rawURL)
	info, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing linked cache directory")
	}
	return os.RemoveAll(dir)
}

// CleanupAllUnfinishedCaches removes all .omnivideorip-resume-* and temporary segment caches in outDir and tempDir.
func CleanupAllUnfinishedCaches(outDir string) {
	// No directory-wide deletion: only CleanupTaskResumeCache can establish ownership.
}

func segmentFileIsUsable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() || info.Size() == 0 {
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	prefix := make([]byte, 512)
	n, readErr := file.Read(prefix)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return false
	}
	trimmed := strings.ToLower(strings.TrimSpace(string(prefix[:n])))
	return !strings.HasPrefix(trimmed, "<!doctype html") &&
		!strings.HasPrefix(trimmed, "<html") &&
		!strings.HasPrefix(trimmed, "<?xml")
}

func downloadSegments(ctx context.Context, segmentURLs []string, workDir string, workers int, progressCb ProgressCallback, download segmentDownloader) ([]string, error) {
	if len(segmentURLs) == 0 {
		return nil, fmt.Errorf("playlist contains no segments")
	}
	// Bind numbered fragments to their exact playlist, including variant and signed URLs.
	manifest, err := json.Marshal(segmentURLs)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(manifest)
	workDir = filepath.Join(workDir, fmt.Sprintf("playlist-%x", digest))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(segmentURLs) {
		workers = len(segmentURLs)
	}

	type item struct {
		index int
		url   string
		path  string
	}
	segmentFiles := make([]string, len(segmentURLs))

	skipped := 0
	var pending []item
	for i, segmentURL := range segmentURLs {
		segmentPath := filepath.Join(workDir, fmt.Sprintf("seg_%05d.ts", i))
		segmentFiles[i] = segmentPath
		if segmentFileIsUsable(segmentPath) {
			skipped++
			continue
		}
		_ = os.Remove(segmentPath)
		_ = os.Remove(segmentPath + ".part")
		pending = append(pending, item{index: i, url: segmentURL, path: segmentPath})
	}

	total := len(segmentURLs)
	if skipped > 0 && progressCb != nil {
		pct := 15 + int((float64(skipped)/float64(total))*75)
		progressCb(pct, fmt.Sprintf("已从断点恢复 %d/%d 个切片，继续下载剩余 %d 个...", skipped, total, len(pending)))
	}
	if len(pending) == 0 {
		return segmentFiles, nil
	}

	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan item)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	completed := skipped

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				if downloadCtx.Err() != nil {
					return
				}
				if err := download(downloadCtx, task.url, task.path); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = fmt.Errorf("segment %d: %w", task.index+1, err)
						cancel()
					}
					mu.Unlock()
					return
				}
				mu.Lock()
				completed++
				current := completed
				mu.Unlock()
				if progressCb != nil && (current%10 == 0 || current == total) {
					pct := 15 + int((float64(current)/float64(total))*75)
					progressCb(pct, fmt.Sprintf("正在下载切片: %d/%d (%d%%)", current, total, pct))
				}
			}
		}()
	}

sendLoop:
	for _, pendingItem := range pending {
		select {
		case jobs <- pendingItem:
		case <-downloadCtx.Done():
			break sendLoop
		}
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for i, segmentPath := range segmentFiles {
		if !segmentFileIsUsable(segmentPath) {
			return nil, fmt.Errorf("segment %d is missing, empty, or not media data", i+1)
		}
	}
	return segmentFiles, nil
}

func mergeSegments(ctx context.Context, segmentFiles []string, outDir, baseName, ffmpegPath string) (string, error) {
	combined, err := os.CreateTemp(outDir, ".omnivideorip-*.ts")
	if err != nil {
		return "", err
	}
	combinedPath := combined.Name()
	defer os.Remove(combinedPath)
	for i, segmentPath := range segmentFiles {
		segment, err := os.Open(segmentPath)
		if err != nil {
			combined.Close()
			return "", fmt.Errorf("open segment %d: %w", i+1, err)
		}
		_, copyErr := io.Copy(combined, segment)
		closeErr := segment.Close()
		if copyErr != nil || closeErr != nil {
			combined.Close()
			return "", errors.Join(copyErr, closeErr)
		}
	}
	if err := combined.Close(); err != nil {
		return "", err
	}

	baseName = sanitizeFilename(baseName)
	if baseName == "" {
		baseName = fmt.Sprintf("VIDEO_%d", time.Now().Unix())
	}
	if ffmpegPath == "" {
		outputPath := filepath.Join(outDir, baseName+".ts")
		_ = os.Remove(outputPath)
		if err := os.Rename(combinedPath, outputPath); err != nil {
			return "", err
		}
		return outputPath, nil
	}

	outputPath := filepath.Join(outDir, baseName+".mp4")
	tempOutput := outputPath + ".partial.mp4"
	_ = os.Remove(tempOutput)
	cmd := exec.CommandContext(ctx, ffmpegPath, "-y", "-i", combinedPath, "-c", "copy", "-movflags", "+faststart", tempOutput)
	hideCmdWindow(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = os.Remove(tempOutput)
		return "", fmt.Errorf("ffmpeg merge failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	_ = os.Remove(outputPath)
	if err := os.Rename(tempOutput, outputPath); err != nil {
		_ = os.Remove(tempOutput)
		return "", err
	}
	return outputPath, nil
}

func downloadSegmentWithRetry(ctx context.Context, client *http.Client, urlStr, destPath, referer string) error {
	if err := validateHTTPURL(urlStr); err != nil {
		return err
	}
	partPath := destPath + ".part"
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		_ = os.Remove(partPath)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8")
		req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
		req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		if referer != "" {
			req.Header.Set("Referer", referer)
		}
		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				contentType := strings.ToLower(resp.Header.Get("Content-Type"))
				out, createErr := os.Create(partPath)
				if createErr == nil {
					written, copyErr := io.Copy(out, io.LimitReader(resp.Body, maxSegmentBytes+1))
					closeErr := out.Close()
					resp.Body.Close()
					if written > maxSegmentBytes {
						copyErr = fmt.Errorf("segment exceeds size limit")
					}
					if copyErr == nil && closeErr == nil && written > 0 {
						if strings.Contains(contentType, "text/html") || !segmentFileIsUsable(partPath) {
							lastErr = fmt.Errorf("segment response is HTML or invalid media data")
						} else {
							_ = os.Remove(destPath)
							if renameErr := os.Rename(partPath, destPath); renameErr == nil {
								return nil
							} else {
								lastErr = renameErr
							}
						}
					} else {
						lastErr = errors.Join(copyErr, closeErr)
					}
				} else {
					resp.Body.Close()
					lastErr = createErr
				}
			} else {
				resp.Body.Close()
				lastErr = fmt.Errorf("unexpected HTTP status %s", resp.Status)
			}
		} else {
			lastErr = err
		}
		_ = os.Remove(partPath)
		backoff := time.Duration(attempt+1) * 300 * time.Millisecond
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	_ = os.Remove(partPath)
	return fmt.Errorf("download segment after 5 attempts: %w", lastErr)
}
