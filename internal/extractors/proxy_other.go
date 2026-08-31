//go:build !windows

package extractors

import (
	"net/url"
	"os"
	"strings"
)

// GetSystemProxyURL returns the proxy URL configured in environment variables.
func GetSystemProxyURL(targetURL *url.URL) (*url.URL, error) {
	if targetURL == nil {
		return nil, nil
	}

	var rawProxy string
	if strings.ToLower(targetURL.Scheme) == "https" {
		if v := os.Getenv("HTTPS_PROXY"); v != "" {
			rawProxy = v
		} else if v := os.Getenv("https_proxy"); v != "" {
			rawProxy = v
		}
	}
	if rawProxy == "" {
		if v := os.Getenv("HTTP_PROXY"); v != "" {
			rawProxy = v
		} else if v := os.Getenv("http_proxy"); v != "" {
			rawProxy = v
		}
	}
	if rawProxy == "" {
		if v := os.Getenv("ALL_PROXY"); v != "" {
			rawProxy = v
		} else if v := os.Getenv("all_proxy"); v != "" {
			rawProxy = v
		}
	}

	if rawProxy != "" {
		noProxy := os.Getenv("NO_PROXY")
		if noProxy == "" {
			noProxy = os.Getenv("no_proxy")
		}
		if noProxy != "" && shouldBypassProxy(targetURL.Hostname(), noProxy) {
			return nil, nil
		}
		return normalizeProxyURL(rawProxy)
	}

	return nil, nil
}

func shouldBypassProxy(host, override string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	override = strings.ReplaceAll(override, ",", ";")
	rules := strings.Split(override, ";")
	for _, r := range rules {
		r = strings.ToLower(strings.TrimSpace(r))
		if r == "" {
			continue
		}
		if r == "<local>" {
			if host == "localhost" || host == "127.0.0.1" || !strings.Contains(host, ".") {
				return true
			}
			continue
		}
		if strings.HasPrefix(r, "*.") {
			domain := strings.TrimPrefix(r, "*.")
			if host == domain || strings.HasSuffix(host, "."+domain) {
				return true
			}
		} else if strings.HasPrefix(r, ".") {
			domain := strings.TrimPrefix(r, ".")
			if host == domain || strings.HasSuffix(host, "."+domain) {
				return true
			}
		} else if host == r {
			return true
		}
	}
	return false
}

func normalizeProxyURL(addr string) (*url.URL, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, nil
	}

	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}

	return url.Parse(addr)
}
