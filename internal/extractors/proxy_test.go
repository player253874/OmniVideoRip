package extractors

import (
	"net/url"
	"os"
	"testing"
)

func TestShouldBypassProxy(t *testing.T) {
	tests := []struct {
		host     string
		override string
		expected bool
	}{
		{"localhost", "<local>", true},
		{"127.0.0.1", "<local>", true},
		{"internal-service", "<local>", true},
		{"example.com", "<local>", false},
		{"sub.example.com", "*.example.com", true},
		{"example.com", "*.example.com", true},
		{"other.org", "*.example.com", false},
		{"direct.org", "direct.org;*.test.com", true},
		{"sub.test.com", "direct.org;*.test.com", true},
	}

	for _, tt := range tests {
		got := shouldBypassProxy(tt.host, tt.override)
		if got != tt.expected {
			t.Errorf("shouldBypassProxy(%q, %q) = %v; want %v", tt.host, tt.override, got, tt.expected)
		}
	}
}

func TestParseProxyServerString(t *testing.T) {
	// Single host:port
	u, err := parseProxyServerString("127.0.0.1:7890", "https")
	if err != nil || u == nil || u.Host != "127.0.0.1:7890" {
		t.Fatalf("Failed to parse 127.0.0.1:7890: %v, %v", u, err)
	}

	// Multi-protocol string
	u2, err := parseProxyServerString("http=127.0.0.1:8080;https=127.0.0.1:8443;socks=127.0.0.1:1080", "https")
	if err != nil || u2 == nil || u2.Host != "127.0.0.1:8443" {
		t.Fatalf("Failed to parse multi-protocol https: %v, %v", u2, err)
	}

	// SOCKS fallback
	u3, err := parseProxyServerString("socks=127.0.0.1:1080", "https")
	if err != nil || u3 == nil || u3.Host != "127.0.0.1:1080" {
		t.Fatalf("Failed to parse socks fallback: %v, %v", u3, err)
	}
}

func TestGetSystemProxyURLEnv(t *testing.T) {
	_ = os.Setenv("HTTPS_PROXY", "http://127.0.0.1:9999")
	defer os.Unsetenv("HTTPS_PROXY")

	target, _ := url.Parse("https://rou.video/v/test")
	proxyURL, err := GetSystemProxyURL(target)
	if err != nil {
		t.Fatalf("GetSystemProxyURL failed: %v", err)
	}
	if proxyURL == nil || proxyURL.Host != "127.0.0.1:9999" {
		t.Fatalf("Expected proxy 127.0.0.1:9999, got: %v", proxyURL)
	}
}
