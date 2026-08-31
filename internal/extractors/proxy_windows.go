//go:build windows

package extractors

import (
	"net/url"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// GetSystemProxyURL returns the proxy URL configured in environment variables or Windows system settings.
func GetSystemProxyURL(targetURL *url.URL) (*url.URL, error) {
	if targetURL == nil {
		return nil, nil
	}

	// 1. Explicitly check environment variables (HTTP_PROXY, HTTPS_PROXY, ALL_PROXY)
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

	// 2. Check Windows System Internet Settings in Registry
	return getWindowsRegistryProxy(targetURL)
}

func getWindowsRegistryProxy(targetURL *url.URL) (*url.URL, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		return nil, nil
	}
	defer k.Close()

	enable, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil || enable != 1 {
		return nil, nil
	}

	proxyServer, _, err := k.GetStringValue("ProxyServer")
	if err != nil || strings.TrimSpace(proxyServer) == "" {
		return nil, nil
	}

	// Check bypass / override list
	if override, _, err := k.GetStringValue("ProxyOverride"); err == nil && override != "" {
		if shouldBypassProxy(targetURL.Hostname(), override) {
			return nil, nil
		}
	}

	return parseProxyServerString(proxyServer, targetURL.Scheme)
}

func shouldBypassProxy(host, override string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	// Replace commas or semicolons
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

func parseProxyServerString(proxyServer, scheme string) (*url.URL, error) {
	proxyServer = strings.TrimSpace(proxyServer)
	if proxyServer == "" {
		return nil, nil
	}

	scheme = strings.ToLower(scheme)
	if scheme == "" {
		scheme = "https"
	}

	// Handle format: "http=127.0.0.1:7890;https=127.0.0.1:7890;socks=127.0.0.1:7890"
	if strings.Contains(proxyServer, "=") {
		parts := strings.Split(proxyServer, ";")
		var matchedAddr string
		var fallbackAddr string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				k := strings.ToLower(strings.TrimSpace(kv[0]))
				v := strings.TrimSpace(kv[1])
				if k == scheme {
					matchedAddr = v
					break
				}
				if k == "http" || k == "https" || k == "socks" {
					fallbackAddr = v
				}
			}
		}
		if matchedAddr != "" {
			return normalizeProxyURL(matchedAddr)
		}
		if fallbackAddr != "" {
			return normalizeProxyURL(fallbackAddr)
		}
	}

	return normalizeProxyURL(proxyServer)
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
