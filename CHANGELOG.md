# Changelog

All notable changes to **OmniVideoRip** are documented in this file.

---

## v1.6.5 (2026-11-14)

### Improved
- **Locked Binary Hot-Swap on Windows (`build.ps1`)**: Stages builds to `OmniVideoRip.exe.new` and, when `OmniVideoRip.exe` is locked by a running instance, renames the running binary to `OmniVideoRip.exe.old` (or a timestamped `.old` fallback) so the fresh `OmniVideoRip.exe` lands immediately without interrupting active downloads.
- **Automatic `.old` Backup Cleanup (`app.go`)**: Cleans unlocked `OmniVideoRip.exe*.old` backups automatically on startup and added `appVersion` unification across single/batch task metadata (`ClientVersion`).

---

## v1.6.4 (2026-11-14)

### Changed
- **Unified Supported Site Count to `1,200+`**: Updated all desktop UI labels, i18n translations (`zh-CN`, `en`, `ja`, `zh-TW`), metadata, and platform list descriptions in `frontend/src/` from `1,800+` / `1,850+` to `1,200+` to match the website and actual dataset (`1,227` platforms).

---

## v1.6.3 (2026-11-14)

### Fixed
- **Protected `cookies.txt` Against `yt-dlp` Overwrite (`prepareTempCookieCopy`)**: Cloned imported `cookies.txt` files to an isolated temporary file per extraction/download call so `yt-dlp`'s cookie-jar write-back never corrupts or invalidates the user's original exported cookie file during retries or concurrent batch runs.
- **Updated Cookie Export Guide for Session Longevity**: Added clear instructions in `#youtube-cookie-modal` recommending Incognito/InPrivate export (and closing the incognito window directly) to avoid rapid YouTube cookie rotation.

---

## v1.6.2 (2026-11-14)

### Fixed
- **Clean Download Output Folder (Relocated Resume Cache)**: Moved `.omnivideorip-resume-*` temporary task directories from the user's download folder (`outDir`) into the OS application cache directory (`os.UserCacheDir()/OmniVideoRip/resume`) while keeping cross-volume file moves working seamlessly (`moveFile`) and cleaning completed/empty temp cache folders automatically.

---

## v1.6.1 (2026-11-14)

### Added & Fixed
- **Simplified `cookies.txt` Import & Interactive Step-by-Step Guide Modal**: Removed unreliable Windows direct browser DB reading (`Chrome`/`Edge` lock & App-Bound Encryption issues) in favor of a clear `cookies.txt` workflow, complete with a guided 3-step modal (`Get cookies.txt LOCALLY` extension store link, export instructions, and one-click file import).
- **Batch Quality Selector**: Added a shared quality selector (`Best`, `1080P`, `720P`, `480P`, `Audio MP3`) to the batch download action bar (`SubmitBatchDownloadsWithQuality`).
- **Fixed Queue List Jumping & Header Version Sync**: Fixed `Store.List()` random Go map ordering and JS string-subtraction `NaN` sorting in `renderQueuePanel` so failed/queued tasks stay in deterministic order; synced `.version-badge` and native window title (`OmniVideoRip v1.6.1`).

---

## v1.6.0 (2026-11-14)

### Added & Fixed
- **Account Cookies Authentication (`Browser` & `cookies.txt`)**: Added a compact cookie source selector (`Edge`, `Chrome`, `Firefox`, `Brave`, or local Netscape `cookies.txt`) persisted in `config.json` and applied across both metadata analysis (`Analyze`) and queued downloads.
- **Clear YouTube Bot Check & Error Diagnostics**: Fixed `ExtractorRouter.Analyze` discarding `yt-dlp` stderr behind a generic network error message. Surfaced structured diagnostics (`YOUTUBE_BOT_CHECK`, `COOKIE_DB_LOCKED`, DNS/network errors, and raw upstream `ERROR:` lines) with 4-language remediation guidance.
- **Improved Windows JS Runtime Discovery & Cookie-Compatible Client Routing**: Added standard Windows `node.exe` / `deno.exe` path fallbacks for GUI launches and avoided forcing `web_embedded` when account cookies are configured.

---

## v1.5.3 (2026-11-13)

### Added & Improved
- **F12 & cURL Stream Extraction Guidance (`SupJav`)**: Replaced automated browser-based cookie probing with a clear, reliable `F12` DevTools `.m3u8` / `Copy as cURL` workflow guide across all 4 desktop UI languages (`zh-CN`, `zh-TW`, `en`, `ja`).
- **Website Overhaul (`website/`)**: Rebuilt the official landing site with a dark architectural palette matched to the app icon, independent product positioning, `1,200+` (`1,227` indexed) unified site metrics, OS compatibility metadata (`Windows 10/11 x64`), and multi-language parity.

---

## v1.5.2 (2026-08-26)

### Added
- **In-App Component Update System**: Added one-click component updates in **Settings -> Components** with upstream release manifest verification, SHA-256 checksum validation, and Windows atomic binary replacement (`internal/sidecars`).
- **Task Deletion Cache Cleanup**: Deleting a task in **Queue** or **History** now prompts whether to also purge residual `.tmp` and segment cache files on disk (`DeleteTaskWithOption`).
- **Direct HTTP/cURL Stream Engine**: Added direct downloader routing for pasted `.mp4` / `.m3u8` stream links and `Copy as cURL` commands, preserving `Cookie`, `Referer`, and `User-Agent` headers automatically.
- **SupJav Cloudflare Bypass Guidance**: Detects Turnstile/JS challenges cleanly and guides users to copy the target embed source or direct stream URL.

### Fixed & Improved
- **Unified Site Count Copy**: Removed inconsistent marketing figures and synchronized platform statistics around the actual `1,227`-entry directory.
- **Download Speed & Resume Reliability**: Tuned 16-thread Go HLS downloader concurrency, added automatic retry backoff, and fixed filename sanitization edge cases.
- **Persistent Queue Recovery**: Unfinished tasks are restored automatically in a paused state across restarts instead of disappearing.

---

## v1.5.1 (2026-08-24)

### Changed
- **Refactored Architecture**: Split `app.go` and `frontend/src/main.js` into focused domain packages (`internal/netutil`, `internal/extractors`, `internal/downloader`, `internal/tasks`, and modular frontend scripts).
- **Task State Snapshots**: Switched task sync updates to structured state snapshots, eliminating brittle progress parsing and duplicated status text.
- **Consistent 4-Language I18n**: Normalized all UI labels and status badges across `zh-CN`, `zh-TW`, `en`, and `ja` (`frontend/src/i18n.js`).

### Fixed
- **SupJav Embed Extraction**: Handled multi-mirror `data-link` extraction and fallback retry behavior across embed providers.
- **X (Twitter) Direct CDN Downloads**: Improved direct CDN stream selection, resolution labeling, and file write buffering.

---

## v1.5.0 (2026-08-20)

### Added
- **1,200+ Supported Sites Directory Modal**: Added an interactive, searchable modal indexing `1,227` supported sites (`frontend/src/platforms-data.js`) with category filtering.
- **Multi-Language Platform Notes**: Added localized usage notes for popular platforms.

---

## v1.4.1 (2026-08-18)

### Fixed
- **Proxy Inheritance for Subprocesses**: Fixed proxy environment variables (`HTTP_PROXY` / `HTTPS_PROXY`) not always propagating to spawned media workers.
- **UTF-8 Windows Path Handling**: Fixed occasional path encoding glitches when saving downloads into folders containing non-ASCII characters.

---

## v1.4.0 (2026-08-15)

### Added
- **X (Twitter) Native Go Extractor**: Built a dedicated Go extractor (`internal/extractors/twitter.go`) resolving tweet media metadata and direct CDN MP4 streams with zero external latency.
- **Parallel DNS Fallback**: Added custom Go `net.Dialer` fallback resolving via encrypted DoH (`dns.google` / `cloudflare-dns.com`) and public UDP DNS (`8.8.8.8`, `1.1.1.1`) when local system DNS fails.

---

## v1.3.9 (2026-08-12)

### Fixed
- **Dark Theme Contrast & Scrollbars**: Improved contrast for secondary labels and custom scrollbars on Windows 10/11 WebView2.
- **Audio/Video Merge Progress**: Fixed progress bar stalling around `90%` during final container muxing.

---

## v1.3.8 (2026-08-10)

### Added
- **Multi-Language Support (i18n)**: Introduced full UI localization for Simplified Chinese, Traditional Chinese, English, and Japanese with persisted preferences.

---

## v1.3.7 (2026-08-09)

### Fixed
- **Windows Console Window Flash**: Applied `CREATE_NO_WINDOW` (`HideWindow: true`) across all spawned background tools so command windows never flash on screen.

---

## v1.3.5 (2026-08-08)

### Added
- **SupJav Native HLS Extractor**: Implemented native Go scraping + concurrent `.ts` segment downloader with AES-128 key decryption (`internal/extractors/supjav.go`).

---

## v1.3.4 (2026-08-06)

### Fixed
- **Bilibili High-Bitrate Audio+Video Merging**: Fixed cases where 1080P+ DASH audio and video tracks remained separated instead of muxing cleanly into `.mp4`.

---

## v1.3.3 (2026-08-05)

### Added
- **Smart Local Proxy Detection**: Automatically inspects Windows Registry Internet Settings plus common local proxy ports (`Clash` 7890, `v2ray` 10808/10809, `SOCKS5` 1080).

---

## v1.3.2 (2026-08-04)

### Fixed
- **Duplicate Task Guard**: Prevented queueing the same URL multiple times simultaneously.

---

## v1.3.1 (2026-08-03)

### Improved
- **Format Picker Sorting**: Sorted extracted streams cleanly from highest to lowest resolution with codec and estimated file size labels.

---

## v1.3.0 (2026-08-01)

### Added
- **Download Queue & Task History Views**: Added dedicated sidebar views for active transfers, paused items, and completed downloads (`Show in Folder` shortcut included).

---

## v1.2.2 (2026-07-28)

### Fixed
- **Filename Sanitization**: Stripped reserved Windows filename characters (`< > : " / \ | ? *`) cleanly before writing output files.

---

## v1.2.1 (2026-07-25)

### Improved
- **Custom Output Directory**: Added native Windows folder picker dialog and persistent download path saving.

---

## v1.2.0 (2026-07-20)

### Added
- **Real-Time Speed & ETA Telemetry**: Added live throughput (`MB/s`), percentage progress bars, and remaining-time telemetry in the UI.

---

## v1.1.2 (2026-07-15)

### Fixed
- **WebView2 Initialization Stability**: Resolved occasional blank-window startup races on slower Windows hosts.

---

## v1.0.0 (2026-07-10)

### Added
- **Initial Release**: Launched **OmniVideoRip** Windows desktop client built on Go + Wails (`v2`) with universal media parsing, thumbnail/duration preview, and one-click MP4 saving.
