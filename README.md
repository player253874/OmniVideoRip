<div align="center">

<img src="website/assets/logo.png" alt="OmniVideoRip Logo" width="96" height="96" />

# OmniVideoRip

**Universal Desktop Video Extractor & High-Speed HLS Downloader for Windows**

[![Version](https://img.shields.io/badge/version-v1.6.5-ff3e32?style=flat-square)](CHANGELOG.md)
[![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11%20(x64)-0078d4?style=flat-square)](https://github.com/player253874/OmniVideoRip/releases/latest)
[![Stack](https://img.shields.io/badge/stack-Go%201.22%20%2B%20Wails-00add8?style=flat-square)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-2ea44f?style=flat-square)](LICENSE)
[![Website](https://img.shields.io/badge/website-omnivideorip.netlify.app-0c0e17?style=flat-square)](https://omnivideorip.netlify.app/)

[Official Website](https://omnivideorip.netlify.app/) · [Latest Releases](https://github.com/player253874/OmniVideoRip/releases/latest) · [Changelog](CHANGELOG.md) · [Report a Bug](.github/ISSUE_TEMPLATE/bug_report.md)

</div>

---

## Overview

**OmniVideoRip** is a lightweight Windows desktop video extractor built with **Go + Wails (WebView2)**. Designed with an architectural dark UI and zero feature paywalls, it pairs a universal parsing engine covering **1,200+ streaming and social platforms** (`1,227` indexed sites) with a custom **16-thread Go concurrent HLS engine** for fast, reliable offline media capture.

### Key Highlights

- **1,200+ Supported Platforms (`1,227` Indexed Sites)** — Built-in searchable platform directory with categorized filters and quick extraction guides across YouTube, Bilibili, X/Twitter, TikTok, Twitch, Vimeo, Niconico, SupJav, and more.
- **16-Thread Go Concurrent HLS Engine** — Native AES-128 decryption, adaptive retry backoff, and seamless `.ts` segment merging into standard `.mp4` files with hardware-friendly performance.
- **F12 `.m3u8` & `Copy as cURL` Stream Assistant** — Built-in extraction workflow for Cloudflare-protected or dynamic player pages (such as SupJav): capture the live `.m3u8` request from Chrome/Edge DevTools (`F12`), or paste a `Copy as cURL` command to preserve `Cookie`, `Referer`, and `User-Agent` headers automatically.
- **Zero-Config Proxy & Parallel DNS Fallback** — Automatically inherits Windows system proxy settings, detects common local proxy ports (`Clash`, `v2ray`, `SOCKS5`), and falls back across encrypted/public DNS resolvers (`DoH`, `8.8.8.8`, `1.1.1.1`) when local resolution fails.
- **Persistent Queue & Safe Task Cleanup** — Download queues persist across restarts; pausing and resuming works reliably, and deleting a task lets you choose whether to also remove partial temporary files on disk.
- **SHA-256 Verified In-App Component Updates** — Built-in one-click engine updater fetches official release manifests and verifies SHA-256 checksums before applying updates.
- **Multi-Language Desktop UI** — Supports Simplified Chinese (`zh-CN`), Traditional Chinese (`zh-TW`), English (`en`), and Japanese (`ja`) with instant switching and auto-detection.

---

## Download & Installation

Grab the latest Windows x64 build from **[GitHub Releases](https://github.com/player253874/OmniVideoRip/releases/latest)** or the **[Official Website](https://omnivideorip.netlify.app/)**:

| Package | Description |
| :--- | :--- |
| **Full Bundle (`.zip`)** | Complete ready-to-run archive containing `OmniVideoRip.exe` and bundled `bin/` media components. Extract and launch directly. |
| **Standalone Portable (`.exe`)** | Single lightweight executable; missing runtime components can be installed cleanly via the in-app **Settings -> Components** updater. |

> **System Requirements:** Windows 10 or Windows 11 (64-bit) with Microsoft Edge WebView2 Runtime (pre-installed on modern Windows).

---

## Quick Usage

1. **Standard URL Extraction**
   - Paste any supported video page URL, click **Parse**, choose your preferred resolution/format, and click **Download**.
2. **Stream Sniffing for Protected Pages (`F12` / `cURL`)**
   - For Cloudflare-gated pages or players with rotating tokens: open DevTools (`F12`) -> **Network**, filter by `m3u8`, then right-click the request and choose **Copy -> Copy as cURL (bash)** (or **Copy link address**) and paste it straight into OmniVideoRip.
3. **Supported Platforms Directory**
   - Click the **1,200+ Sites** button in the header to search all `1,227` indexed platforms, filter by region/stream protocol, and view site-specific tips.

---

## Build from Source

### Prerequisites

- **Go** `1.22` or higher
- **Node.js** & **npm** (used by `frontend/build.js`)
- **Windows x64** with WebView2 Runtime

### Build & Run

```powershell
# Run the full build pipeline (resolves/verifies bin/ components, builds frontend, embeds resources, compiles OmniVideoRip.exe)
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

> **Note on `build.ps1`:** The build script checks `bin/yt-dlp.exe` and `bin/ffmpeg.exe` (downloading and verifying their SHA-256 checksums automatically if missing), runs `npm run build` in `frontend/`, generates `resource.syso` via `goversioninfo`, and compiles `OmniVideoRip.exe` with `-tags "desktop,production"`.

### Verification & Tests

```powershell
# Run Go backend tests
go test ./... -count=1

# Run static analysis
go vet ./...
```

---

## Repository Structure

```text
OmniVideoRip/
├── app.go                      # Wails backend bindings, settings, updater & platform handlers
├── main.go                     # Desktop entry point and window configuration
├── build.ps1                   # Windows release build and binary verification script
├── wails.json / versioninfo.json
├── frontend/                   # Desktop WebView2 UI (vanilla HTML/CSS/JS + 4-language i18n)
│   └── src/
├── internal/
│   ├── downloader/             # Concurrent HLS segment engine, AES-128 decryptor & task orchestrator
│   ├── extractors/             # Dedicated extractors (X/Twitter, SupJav) & HTML helpers
│   ├── netutil/                # Proxy detection, HTTP clients & DNS fallback resolver
│   ├── sidecars/               # SHA-256 verified binary component updater
│   └── tasks/                  # Persistent JSON task queue & progress snapshots
└── website/                    # Official multi-language product landing site (Netlify)
    ├── index.html
    ├── styles.css
    ├── main.js
    └── i18n-website.js
```

---

## Disclaimer & License

- **Legal Notice:** OmniVideoRip is intended strictly for personal archiving, offline study, and fair-use workflows involving media you have permission to save. Users are solely responsible for complying with the Terms of Service and copyright laws applicable to target websites.
- **License:** Released under the [MIT License](LICENSE).
