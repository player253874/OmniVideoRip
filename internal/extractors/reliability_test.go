package extractors

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPSConnectProxyRemainsUsable(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("secure")) }))
	defer upstream.Close()
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodConnect {
			http.Error(w, "CONNECT required", 400)
			return
		}
		remote, err := net.Dial("tcp", upstream.Listener.Addr().String())
		if err != nil {
			http.Error(w, "dial", 502)
			return
		}
		c, rw, err := w.(http.Hijacker).Hijack()
		if err != nil {
			remote.Close()
			return
		}
		rw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
		rw.Flush()
		go func() { io.Copy(remote, c); remote.Close() }()
		io.Copy(c, remote)
		c.Close()
	}))
	defer proxy.Close()
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	client := newSafeHTTPClient(3 * time.Second)
	roots := x509.NewCertPool()
	roots.AddCert(upstream.Certificate())
	client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	resp, err := client.Get("https://example.com/video")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	client.CloseIdleConnections()
	if string(data) != "secure" {
		t.Fatal("CONNECT failed")
	}
}

func TestSOCKSProxyRemainsUsable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		c, err := listener.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(3 * time.Second))
		header := make([]byte, 2)
		if _, err := io.ReadFull(c, header); err != nil {
			return
		}
		methods := make([]byte, int(header[1]))
		io.ReadFull(c, methods)
		c.Write([]byte{5, 0})
		request := make([]byte, 4)
		if _, err := io.ReadFull(c, request); err != nil {
			return
		}
		if request[3] != 3 {
			return
		}
		length := make([]byte, 1)
		io.ReadFull(c, length)
		address := make([]byte, int(length[0])+2)
		io.ReadFull(c, address)
		c.Write([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 80})
		req, err := http.ReadRequest(bufio.NewReader(c))
		if err != nil {
			return
		}
		req.Body.Close()
		io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nOK")
	}()
	t.Setenv("HTTP_PROXY", "socks5://"+listener.Addr().String())
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	resp, err := newSafeHTTPClient(3 * time.Second).Get("http://example.com/video")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(data) != "OK" {
		t.Fatal("SOCKS response mismatch")
	}
	<-done
}

func TestDNSRejectionDoesNotFallBackToRawDial(t *testing.T) {
	for _, ips := range [][]net.IP{nil, {net.ParseIP("127.0.0.1")}, {net.ParseIP("10.0.0.1")}, {net.ParseIP("::1")}} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		conn, err := dialSafeAddress(ctx, "tcp", "blocked.example:80", func(context.Context, string) []net.IP { return ips })
		cancel()
		if conn != nil {
			conn.Close()
		}
		if err == nil {
			t.Fatal("unapproved target dialed")
		}
	}
	if ips := collectAllIPs(context.Background(), "2606:4700:4700::1111"); len(ips) != 1 {
		t.Fatal("public IPv6 literal rejected")
	}
}

func TestExplicitLocalProxyStillWorksAndLocalTargetsAreRejected(t *testing.T) {
	var calls atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.Write([]byte("proxied")) }))
	defer proxy.Close()
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	client := newSafeHTTPClient(2 * time.Second)
	resp, err := client.Get("http://public.example/video")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "proxied" {
		t.Fatal("proxy response mismatch")
	}
	for _, target := range []string{"http://127.0.0.1/", "http://[::1]/", "http://169.254.169.254/", "http://[::1%25lo]/", "http://[fe80::1%25eth0]/", "http://[::ffff:127.0.0.1]/"} {
		if r, err := client.Get(target); err == nil {
			r.Body.Close()
			t.Fatalf("local target accepted: %s", target)
		}
	}
	if calls.Load() != 1 {
		t.Fatal("local target leaked to proxy")
	}
}

func TestRedirectToLocalNetworkIsRejected(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "http://127.0.0.1/secret", 302) }))
	defer proxy.Close()
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	resp, err := newSafeHTTPClient(2 * time.Second).Get("http://public.example/video")
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("private redirect accepted")
	}
}

func TestRedirectToScopedIPv6IsRejected(t *testing.T) {
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "http://[::1%25lo]/secret", 302) }))
	defer proxy.Close()
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	resp, err := newSafeHTTPClient(2 * time.Second).Get("http://public.example/video")
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("scoped private redirect accepted")
	}
}

func TestDifferentPlaylistsCannotReuseNumberedFragments(t *testing.T) {
	root := t.TempDir()
	calls := 0
	download := func(ctx context.Context, url, path string) error {
		calls++
		return os.WriteFile(path, []byte(url), 0600)
	}
	a, err := downloadSegments(context.Background(), []string{"https://cdn.example/720/0.ts"}, root, 1, nil, download)
	if err != nil {
		t.Fatal(err)
	}
	b, err := downloadSegments(context.Background(), []string{"https://cdn.example/1080/0.ts"}, root, 1, nil, download)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || a[0] == b[0] {
		t.Fatal("variant fragments reused")
	}
	_, err = downloadSegments(context.Background(), []string{"https://cdn.example/1080/0.ts"}, root, 1, nil, download)
	if err != nil || calls != 2 {
		t.Fatal("same-playlist resume failed")
	}
	body, _ := os.ReadFile(b[0])
	if !strings.Contains(string(body), "1080") {
		t.Fatal("wrong variant served")
	}
}

func TestOnlyOwnedCacheIsRemoved(t *testing.T) {
	dir := t.TempDir()
	url := "https://example.com/video"
	cache := GetTaskResumeDir(dir, "generic", url)
	os.MkdirAll(cache, 0700)
	os.WriteFile(filepath.Join(cache, "seg_0.part"), []byte("ours"), 0600)
	unrelated := filepath.Join(dir, "browser-download.part")
	os.WriteFile(unrelated, []byte("keep"), 0600)
	CleanupAllUnfinishedCaches(dir)
	if err := CleanupTaskResumeCache(dir, "generic", url); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Fatal("owned cache remains")
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatal("unrelated download was removed")
	}
}
