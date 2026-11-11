/**
 * i18n (Internationalization) module for OmniVideoRip.
 * Supports: zh-CN (default), en, ja, zh-TW.
 * Auto-detects system locale via Go backend, falls back to navigator.language.
 */

const I18N_TRANSLATIONS = {
  "zh-CN": {
    // Header
    "brand.tagline": "全网视频极速抓取平台",
    "lang.auto": "🌐 自动 (系统语言)",
    // Mode tabs
    "mode.single": "🔗 单链接下载",
    "mode.batch": "📋 批量多链接排队下载",
    "mode.hint.single": "粘贴单个视频页面地址，自动分析可用分辨率",
    "mode.hint.single.active": "单视频极速分析模式已启用。输入单个链接即可快速提取并自由选择分辨率。",
    "mode.hint.batch.active": "多视频批量排队模式已启用。支持多平台跨站混合输入，每平台自动独立并发提取。",
    // Input
    "input.placeholder": "在此粘贴任意视频地址 (如 YouTube, Bilibili, TikTok, X, MissAV, GGJav...)",
    "input.paste": "📋 粘贴",
    "input.paste.title": "一键粘贴剪贴板",
    "input.detect.auto": "自动识别",
    "btn.analyze": "解析画质",
    // Batch
    "batch.placeholder": "在此粘贴多个视频地址，每行一个链接 (或空格分隔)。例如:\nhttps://www.youtube.com/watch?v=HlLx7oE7q3I\nhttps://missav.live/dm26/ja/jac-082\nhttps://www.bilibili.com/video/BV1xx411c7mD",
    "batch.paste": "📋 粘贴剪贴板多链接",
    "batch.clear": "清空列表",
    "batch.submit": "开始批量下载",
    // Path
    "path.label": "📁 下载保存位置:",
    "path.placeholder": "选择或输入视频保存路径",
    "path.browse": "更改目录 📁",
    // Resolution card
    "res.loading.title": "正在读取视频标题...",
    "res.label": "🎯 请选择下载画质 / 分辨率:",
    "res.close": "✕ 关闭",
    "res.confirm": "开始下载",
    "res.best": "最佳画质 (Best Quality)",
    "res.best.note": "自动拼接最高画质与最佳音轨 (推荐)",
    "res.best.note.short": "自动混流最优视频轨与音频轨",
    "res.original": "原画",
    // Active downloads
    "active.title": "⚡ 正在下载",
    "active.badge": "{n} 个进行中",
    "active.pause": "⏸️ 暂停",
    "active.delete": "🗑️ 删除",
    "active.delete.title": "删除任务并清理缓存",
    "active.msg.default": "正在全力下载切片与数据...",
    // Queue
    "queue.title": "当前队列",
    "queue.badge": "{n} 个任务",
    "queue.waiting": "⏳ 等待中: {n}",
    "queue.failed": "❌ 失败: {n}",
    "queue.done": "✅ 已完成: {n}",
    "queue.clear": "清空队列与缓存 🧹",
    "queue.status.queued": "⏳ 排队等待中",
    "queue.status.paused": "⏸️ 已暂停",
    "queue.status.failed": "❌ 下载失败",
    "queue.status.completed": "✅ 下载完成",
    "queue.resume": "▶️ 继续下载",
    "queue.copy.path": "📋 复制路径",
    "queue.error.prefix": "⚠️ 错误: ",
    "queue.file.prefix": "📂 文件: ",
    "queue.file.saved": "已保存至本地",
    "queue.waiting.msg": "排队等待空闲通道启动...",
    // Update
    "update.ytdlp.title": "🔄 正在更新 yt-dlp 解析引擎...",
    "update.ytdlp.checking": "🔄 正在检测 yt-dlp 版本...",
    "update.ytdlp.btn": "🔄 更新 yt-dlp",
    "update.ytdlp.btn.title": "检查并更新 yt-dlp 解析引擎 (针对各视频网站反爬高频更新)",
    "update.ytdlp.done": "✅ yt-dlp 官方组件更新成功！",
    "update.ytdlp.retry": "🔄 重试更新",
    "update.ffmpeg.title": "🔄 正在更新 FFmpeg 组件...",
    "update.ffmpeg.checking": "🔄 正在检测 FFmpeg 组件...",
    "update.ffmpeg.btn": "🔄 更新 ffmpeg",
    "update.ffmpeg.btn.title": "检查并更新 ffmpeg 组件 (音视频合并库，通常无需更新)",
    "update.ffmpeg.done": "✅ FFmpeg 组件更新成功！",
    "update.ffmpeg.retry": "🔄 重试更新",
    "update.checking.btn": "⏳ 检测中...",
    "update.connecting": "正在连接 GitHub 官方源比对版本...",
    "update.downloading": "正在下载...",
    "update.from.github": "正在从 GitHub 官方源获取最新版本...",
    "update.ffmpeg.detecting": "正在检测 FFmpeg 运行库...",
    "update.fail": "❌ 更新失败: {err}",
    "update.ffmpeg.fail": "❌ FFmpeg 更新失败: {err}",
    "update.fail.timeout": "网络连接超时",
    // Version bar
    "ver.app.label": "主程序",
    "ver.ytdlp.label": "yt-dlp",
    "ver.ffmpeg.label": "ffmpeg",
    "ver.loading": "加载中...",
    "ver.ytdlp.missing": "未安装 (需点击更新)",
    "ver.ffmpeg.missing": "未安装 (需点击更新)",
    // Platforms Ecosystem & Showcase
    "showcase.eyebrow": "1,800+ PLATFORMS POWERED BY YT-DLP & NATIVE ENGINES",
    "showcase.title": "支持全网超千家音视频与社媒平台",
    "showcase.subtitle": "集成官方完整 yt-dlp 1800+ 站点解析生态与自研 16 线程并发流解密引擎，全球主流社媒短视频、海内外长剧集、高码率切片原画一键极速抓取。",
    "showcase.search.btn": "查询全部支持网站名录 (可搜索 1800+ 站点)",
    "showcase.more": "+ 1,800 更多平台",
    "modal.platforms.title": "🔍 支持网站全量查询名录",
    "modal.platforms.subtitle": "收录全球 1,800+ 主流媒体、社交平台与自研解密站点，输入关键字即可实时筛选",
    "modal.search.placeholder": "搜索支持网站或域名（例如：YouTube, B站, TVer, TikTok, Twitter, SupJav...）",
    "modal.count.total": "共 {n}+ 站点",
    "modal.count.found": "找到 {n} 个站点",
    "modal.empty.title": "未找到匹配的预置名称",
    "modal.empty": "未找到匹配的预置名称，但 OmniVideoRip 搭载官方完整 yt-dlp 1,800+ 引擎，您可以直接粘贴视频链接进行智能嗅探与通用解析！",
    "modal.card.fill": "点击填入 {name} 网址",
    "modal.more": "已显示前 {max} 个结果（共 {n} 个），请输入更精确的关键字筛选...",
    "cat.all": "全部 (1800+)",
    "cat.popular": "热门主流",
    "cat.social": "社交短视频",
    "cat.streaming": "流媒体/电视",
    "cat.japan": "日韩专区",
    "cat.audio": "音乐音频",
    "cat.adult": "成人正片解密",
    "platform.generic": "全网通用解析",
    // History
    "history.eyebrow": "RECENT COMPLETED",
    "history.title": "历史记录",
    "history.clear": "清空记录",
    "history.default.filename": "下载视频.mp4",
    // Footer
    "footer.text": "OmniVideoRip Desktop Client © 2026 · 全网视频极速抓取平台",
    // AdSense
    "ad.top": "Google AdSense 广告展示位 (TOP BANNER - 728x90 / 响应式)",
    "ad.bottom": "Google AdSense 广告展示位 (BOTTOM BANNER - 728x90 / 响应式)",
    // Alerts
    "alert.analyze.fail": "解析视频失败: {err}",
    "alert.download.fail": "创建下载任务失败: {err}",
    "alert.batch.empty": "请至少输入一个视频播放页面链接！",
    "alert.batch.fail": "批量提交任务失败: {err}",
    "alert.copy.ok": "文件路径已成功复制到剪贴板！",
    // Dynamic states
    "analyzing": "正在解析画质...",
    "joining.queue": "正在加入下载队列...",
    "download.error": "下载遇到异常",
    "download.starting": "正在建立视频流连接...",
    "download.queued.msg": "排队等待中...",
    "download.starting.multi": "开始极速多线程下载...",
    // Title
    "page.title": "OmniVideoRip | 全网视频极速抓取器",
    "page.meta.description": "OmniVideoRip 全能视频抓取器。支持 YouTube、Bilibili、GGJav、MissAV、TikTok、X (Twitter)、Instagram 等 1800+ 站点的高清音视频本地极速提取与多任务批量排队下载。",
    // Platform name for resolution meta
    "res.meta": "{platform} · 时长: {duration}",
    "res.duration.unknown": "未知",
    "res.title.unknown": "未知视频标题",
    // Video meta mock
    "mock.title": "示例全网视频标题",
    // SupJav Modal & Tutorial
    "supjav.guide.badge": "⚡ F12 抓流指引",
    "supjav.modal.title": "⚡ SupJav 视频下载与 F12 抓流指引",
    "supjav.modal.subtitle": "Cloudflare 5秒盾验证拦截应对指南 · F12 提取 .m3u8 极速直下",
    "supjav.modal.banner": "SupJav 全站受 Cloudflare Turnstile 人机验证保护且播放线路经过动态混淆，常规后台请求会被拦截。只需在浏览器按 F12 抓取 .m3u8 真实视频流即可开启 16 线程满速下载：",
    "supjav.modal.m1.badge": "⭐ 推荐方案：浏览器按 F12 提取 .m3u8 视频流 (100% 成功)",
    "supjav.modal.m1.title": "四步快速获取真实 .m3u8 播放流地址",
    "supjav.modal.m1.s1": "点击下方【1. 在浏览器中打开视频页】，手动勾选一次 Cloudflare “确认您是人类”复选框。",
    "supjav.modal.m1.s2": "在浏览器页面按下键盘 F12（或右键页面选择“检查 / Inspect”），点击顶部【Network (网络)】标签，在过滤框中输入 m3u8。",
    "supjav.modal.m1.s3": "点击网页中的视频播放按钮（若列表为空可按 F5 刷新后再点播放），列表中会出现 .m3u8 条目（如 master.m3u8 / index.m3u8）。",
    "supjav.modal.m1.s4": "右键该 .m3u8 条目 → 选择【复制 (Copy) → 复制链接地址 (Copy URL)】，随后点击下方【2. 读取剪贴板 .m3u8 并下载】即可 16 线程极速直下！",
    "supjav.modal.m1.btn_open": "启动独立小窗口打开",
    "supjav.modal.m1.btn_clip": "2. 读取剪贴板 .m3u8 并下载",
    "supjav.modal.m2.badge": "⚡ 手动粘贴：直接输入 .m3u8 链接或 cURL 命令",
    "supjav.modal.m2.title": "粘贴从 F12 Network 复制的 .m3u8 地址",
    "supjav.modal.m2.desc": "您也可以将从 F12 网络面板复制的 .m3u8 链接（或 Copy as cURL 命令）直接粘贴到下方输入框：",
    "supjav.modal.m2.ph": "在此粘贴 F12 抓取的 https://.../master.m3u8 链接或 cURL...",
    "supjav.modal.m2.btn": "🚀 立即解析下载",
    "supjav.modal.btn_browser": "1. 在浏览器中打开视频页",
    "supjav.modal.btn_close": "我知道了，返回主界面",
    "supjav.clip.success": "✅ 已成功从剪贴板识别到 .m3u8 视频流地址，正在解析并准备下载...",
    "supjav.clip.empty": "⚠️ 剪贴板中未检测到 .m3u8 链接。请在浏览器按 F12 -> Network -> 搜索 m3u8 -> 右键复制 .m3u8 链接地址后再试！",
    "supjav.window.opened": "🌐 已打开视频页！请完成人机验证后按 F12 -> Network -> 筛选 m3u8 -> 播放视频并右键复制 .m3u8 链接地址。",
    // Delete Task Confirmation
    "confirm.delete.title": "🗑️ 删除任务确认",
    "confirm.delete.message": "确定要删除该下载任务吗？",
    "confirm.delete.subtext": "⚠️ 该操作将同时清理未完成的临时分片与下载缓存，释放磁盘空间。",
    "confirm.delete.cancel": "取消",
    "confirm.delete.ok": "确认删除并清理缓存",
    "confirm.delete.prompt": "确定要删除该下载任务吗？该操作将同时清理未完成的临时分片与下载缓存文件。",
  },

  en: {
    "brand.tagline": "Universal Video Extraction Platform",
    "lang.auto": "🌐 Auto (System)",
    "mode.single": "🔗 Single URL",
    "mode.batch": "📋 Batch Multi-URL Queue",
    "mode.hint.single": "Paste a single video page URL to auto-analyze available resolutions",
    "mode.hint.single.active": "Single video analysis mode enabled. Enter one URL to quickly extract and choose resolution.",
    "mode.hint.batch.active": "Batch multi-URL queue mode enabled. Supports cross-platform mixed input with per-platform concurrent extraction.",
    "input.placeholder": "Paste any video URL here (e.g. YouTube, Bilibili, TikTok, X, MissAV, GGJav...)",
    "input.paste": "📋 Paste",
    "input.paste.title": "Paste from clipboard",
    "input.detect.auto": "Auto Detect",
    "btn.analyze": "Analyze",
    "batch.placeholder": "Paste multiple video URLs, one per line (or space-separated). Example:\nhttps://www.youtube.com/watch?v=HlLx7oE7q3I\nhttps://missav.live/dm26/ja/jac-082\nhttps://www.bilibili.com/video/BV1xx411c7mD",
    "batch.paste": "📋 Paste URLs from Clipboard",
    "batch.clear": "Clear List",
    "batch.submit": "Start Batch Download",
    "path.label": "📁 Save to:",
    "path.placeholder": "Select or enter video save path",
    "path.browse": "Browse 📁",
    "res.loading.title": "Loading video title...",
    "res.label": "🎯 Select download quality / resolution:",
    "res.close": "✕ Close",
    "res.confirm": "Start Download",
    "res.best": "Best Quality",
    "res.best.note": "Auto-merge highest video + best audio track (Recommended)",
    "res.best.note.short": "Auto-mux best video and audio tracks",
    "res.original": "Original",
    "active.title": "⚡ Downloading",
    "active.badge": "{n} active",
    "active.pause": "⏸️ Pause",
    "active.delete": "🗑️ Delete",
    "active.delete.title": "Delete task and clear cache",
    "active.msg.default": "Downloading segments and data...",
    "queue.title": "Current Queue",
    "queue.badge": "{n} tasks",
    "queue.waiting": "⏳ Waiting: {n}",
    "queue.failed": "❌ Failed: {n}",
    "queue.done": "✅ Done: {n}",
    "queue.clear": "Clear Queue & Cache 🧹",
    "queue.status.queued": "⏳ Queued",
    "queue.status.paused": "⏸️ Paused",
    "queue.status.failed": "❌ Failed",
    "queue.status.completed": "✅ Completed",
    "queue.resume": "▶️ Resume",
    "queue.copy.path": "📋 Copy Path",
    "queue.error.prefix": "⚠️ Error: ",
    "queue.file.prefix": "📂 File: ",
    "queue.file.saved": "Saved locally",
    "queue.waiting.msg": "Waiting for available slot...",
    "update.ytdlp.title": "🔄 Updating yt-dlp engine...",
    "update.ytdlp.checking": "🔄 Checking yt-dlp version...",
    "update.ytdlp.btn": "🔄 Update yt-dlp",
    "update.ytdlp.btn.title": "Check and update yt-dlp parser engine (frequent anti-scrape updates)",
    "update.ytdlp.done": "✅ yt-dlp updated successfully!",
    "update.ytdlp.retry": "🔄 Retry Update",
    "update.ffmpeg.title": "🔄 Updating FFmpeg...",
    "update.ffmpeg.checking": "🔄 Checking FFmpeg...",
    "update.ffmpeg.btn": "🔄 Update ffmpeg",
    "update.ffmpeg.btn.title": "Check and update ffmpeg component (A/V merge library, rarely needs updating)",
    "update.ffmpeg.done": "✅ FFmpeg updated successfully!",
    "update.ffmpeg.retry": "🔄 Retry Update",
    "update.checking.btn": "⏳ Checking...",
    "update.connecting": "Connecting to GitHub to compare versions...",
    "update.downloading": "Downloading...",
    "update.from.github": "Fetching latest version from GitHub...",
    "update.ffmpeg.detecting": "Detecting FFmpeg runtime...",
    "update.fail": "❌ Update failed: {err}",
    "update.ffmpeg.fail": "❌ FFmpeg update failed: {err}",
    "update.fail.timeout": "Network connection timed out",
    "ver.app.label": "App",
    "ver.ytdlp.label": "yt-dlp",
    "ver.ffmpeg.label": "ffmpeg",
    "ver.loading": "Loading...",
    "ver.ytdlp.missing": "Not installed (click update)",
    "ver.ffmpeg.missing": "Not installed (click update)",
    // Platforms Ecosystem & Showcase
    "showcase.eyebrow": "1,800+ PLATFORMS POWERED BY YT-DLP & NATIVE ENGINES",
    "showcase.title": "Over 1,800 Supported Video & Social Sites",
    "showcase.subtitle": "Powered by the official yt-dlp ecosystem covering 1,800+ sites and our proprietary 16-thread stream decryption engine. Download short videos, full series, and high-bitrate HLS streams at full speed.",
    "showcase.search.btn": "Browse All Supported Sites (Search 1,800+ Sites)",
    "showcase.more": "+ 1,800 More Sites",
    "modal.platforms.title": "🔍 Supported Sites Directory",
    "modal.platforms.subtitle": "Over 1,800 mainstream video sites, social media, and native decryption platforms. Type keywords to filter instantly.",
    "modal.search.placeholder": "Search supported sites or domains (e.g., YouTube, Bilibili, TVer, TikTok, Twitter, SupJav...)",
    "modal.count.total": "{n}+ Sites Total",
    "modal.count.found": "{n} Sites Found",
    "modal.empty.title": "No Matching Preset Found",
    "modal.empty": "No exact presets found, but OmniVideoRip includes the complete 1,800+ yt-dlp extractor engine. Simply paste your video URL to parse and download!",
    "modal.card.fill": "Click to fill {name} URL",
    "modal.more": "Showing first {max} of {n} results. Type a more specific keyword to narrow down...",
    "cat.all": "All (1800+)",
    "cat.popular": "Popular",
    "cat.social": "Shorts & Social",
    "cat.streaming": "Streaming & TV",
    "cat.japan": "Japan & Asia",
    "cat.audio": "Audio & Music",
    "cat.adult": "Adult Decryptors",
    "platform.generic": "Universal Parser",
    "history.eyebrow": "RECENT COMPLETED",
    "history.title": "History",
    "history.clear": "Clear History",
    "history.default.filename": "downloaded_video.mp4",
    "footer.text": "OmniVideoRip Desktop Client © 2026 · Universal Video Extraction Platform",
    "ad.top": "Google AdSense Ad Slot (TOP BANNER - 728x90 / Responsive)",
    "ad.bottom": "Google AdSense Ad Slot (BOTTOM BANNER - 728x90 / Responsive)",
    "alert.analyze.fail": "Failed to analyze video: {err}",
    "alert.download.fail": "Failed to create download task: {err}",
    "alert.batch.empty": "Please enter at least one video URL!",
    "alert.batch.fail": "Failed to submit batch tasks: {err}",
    "alert.copy.ok": "File path copied to clipboard!",
    "analyzing": "Analyzing...",
    "joining.queue": "Adding to download queue...",
    "download.error": "Download encountered an error",
    "download.starting": "Establishing video stream connection...",
    "download.queued.msg": "Queued, waiting...",
    "download.starting.multi": "Starting high-speed multi-thread download...",
    "page.title": "OmniVideoRip | Universal Video Extraction",
    "page.meta.description": "OmniVideoRip universal video extractor. Supports YouTube, Bilibili, GGJav, MissAV, TikTok, X (Twitter), Instagram and 1800+ sites for HD local extraction and batch queue download.",
    "res.meta": "{platform} · Duration: {duration}",
    "res.duration.unknown": "Unknown",
    "res.title.unknown": "Unknown video title",
    "mock.title": "Sample Video Title",
    // SupJav Modal & Tutorial
    "supjav.guide.badge": "⚡ F12 Stream Guide",
    "supjav.modal.title": "⚡ SupJav Download & F12 .m3u8 Extraction Guide",
    "supjav.modal.subtitle": "Cloudflare Turnstile Bypass · Extract .m3u8 via F12 DevTools",
    "supjav.modal.banner": "SupJav uses Cloudflare Turnstile anti-bot protection and obfuscated player links. Simply press F12 in your browser to grab the .m3u8 stream URL for 16-thread full-speed download:",
    "supjav.modal.m1.badge": "⭐ Recommended: Press F12 in Browser to Grab .m3u8 Stream (100% Works)",
    "supjav.modal.m1.title": "4 Easy Steps to Extract the .m3u8 Stream URL",
    "supjav.modal.m1.s1": "Click '1. Open Video in Browser' below and check the Cloudflare 'Verify you are human' box.",
    "supjav.modal.m1.s2": "Press F12 on your keyboard (or right-click page -> Inspect), select the 'Network' tab at the top, and type 'm3u8' in the filter box.",
    "supjav.modal.m1.s3": "Click Play on the video (if no item appears, press F5 to refresh and click Play again) until an .m3u8 request (e.g. master.m3u8 / index.m3u8) shows up.",
    "supjav.modal.m1.s4": "Right-click the .m3u8 item -> Copy -> Copy URL, then click '2. Read Clipboard .m3u8 & Download' below for 16-thread full-speed download!",
    "supjav.modal.m1.btn_open": "Open in Standalone Window",
    "supjav.modal.m1.btn_clip": "2. Read Clipboard .m3u8 & Download",
    "supjav.modal.m2.badge": "⚡ Manual Paste: Paste .m3u8 URL or cURL Command",
    "supjav.modal.m2.title": "Paste .m3u8 Link Copied from F12 Network Tab",
    "supjav.modal.m2.desc": "You can also paste the .m3u8 link (or Copy as cURL command) copied from F12 Network directly below:",
    "supjav.modal.m2.ph": "Paste https://.../master.m3u8 link or cURL here...",
    "supjav.modal.m2.btn": "🚀 Parse & Download",
    "supjav.modal.btn_browser": "1. Open Video in Browser",
    "supjav.modal.btn_close": "Got It, Return",
    "supjav.clip.success": "✅ Detected .m3u8 stream URL from clipboard, starting analysis...",
    "supjav.clip.empty": "⚠️ No .m3u8 link found in clipboard. Please press F12 -> Network -> filter 'm3u8' -> right-click and Copy URL first!",
    "supjav.window.opened": "🌐 Video page opened! Complete verification, press F12 -> Network -> filter 'm3u8', play the video, and copy the .m3u8 URL.",
    // Delete Task Confirmation
    "confirm.delete.title": "🗑️ Confirm Task Deletion",
    "confirm.delete.message": "Are you sure you want to delete this download task?",
    "confirm.delete.subtext": "⚠️ This will also remove any incomplete segment and cache files to free disk space.",
    "confirm.delete.cancel": "Cancel",
    "confirm.delete.ok": "Delete & Clear Cache",
    "confirm.delete.prompt": "Are you sure you want to delete this download task? This will also remove any incomplete segment and cache files.",
  },

  ja: {
    "brand.tagline": "全ネット動画高速ダウンロードプラットフォーム",
    "lang.auto": "🌐 自動 (システム言語)",
    "mode.single": "🔗 単一リンクダウンロード",
    "mode.batch": "📋 一括マルチリンクキュー",
    "mode.hint.single": "動画ページURLを貼り付けて、利用可能な解像度を自動分析",
    "mode.hint.single.active": "単一動画解析モードが有効です。URLを入力して解像度を自由に選択できます。",
    "mode.hint.batch.active": "一括マルチURLキューモードが有効です。クロスプラットフォーム混合入力をサポートし、プラットフォーム別に並行抽出します。",
    "input.placeholder": "任意の動画URLをここに貼り付け (YouTube, Bilibili, TikTok, X, MissAV, GGJav...)",
    "input.paste": "📋 貼付",
    "input.paste.title": "クリップボードから貼り付け",
    "input.detect.auto": "自動検出",
    "btn.analyze": "画質解析",
    "batch.placeholder": "複数の動画URLを貼り付け（1行に1リンク、またはスペース区切り）。例：\nhttps://www.youtube.com/watch?v=HlLx7oE7q3I\nhttps://missav.live/dm26/ja/jac-082\nhttps://www.bilibili.com/video/BV1xx411c7mD",
    "batch.paste": "📋 クリップボードから複数URL貼付",
    "batch.clear": "リストをクリア",
    "batch.submit": "一括ダウンロード開始",
    "path.label": "📁 保存先:",
    "path.placeholder": "動画の保存パスを選択または入力",
    "path.browse": "フォルダ変更 📁",
    "res.loading.title": "動画タイトルを読み込み中...",
    "res.label": "🎯 ダウンロード画質 / 解像度を選択:",
    "res.close": "✕ 閉じる",
    "res.confirm": "ダウンロード開始",
    "res.best": "最高画質 (Best Quality)",
    "res.best.note": "最高画質と最適音声トラックを自動マージ (推奨)",
    "res.best.note.short": "最適な映像・音声トラックを自動ミックス",
    "res.original": "オリジナル",
    "active.title": "⚡ ダウンロード中",
    "active.badge": "{n} 件実行中",
    "active.pause": "⏸️ 一時停止",
    "active.delete": "🗑️ 削除",
    "active.delete.title": "タスクを削除してキャッシュをクリア",
    "active.msg.default": "セグメントとデータをダウンロード中...",
    "queue.title": "現在のキュー",
    "queue.badge": "{n} 件のタスク",
    "queue.waiting": "⏳ 待機中: {n}",
    "queue.failed": "❌ 失敗: {n}",
    "queue.done": "✅ 完了: {n}",
    "queue.clear": "キューとキャッシュをクリア 🧹",
    "queue.status.queued": "⏳ キュー待機中",
    "queue.status.paused": "⏸️ 一時停止中",
    "queue.status.failed": "❌ ダウンロード失敗",
    "queue.status.completed": "✅ ダウンロード完了",
    "queue.resume": "▶️ 再開",
    "queue.copy.path": "📋 パスをコピー",
    "queue.error.prefix": "⚠️ エラー: ",
    "queue.file.prefix": "📂 ファイル: ",
    "queue.file.saved": "ローカルに保存済み",
    "queue.waiting.msg": "空きスロットを待機中...",
    "update.ytdlp.title": "🔄 yt-dlp エンジンを更新中...",
    "update.ytdlp.checking": "🔄 yt-dlp バージョンを確認中...",
    "update.ytdlp.btn": "🔄 yt-dlp 更新",
    "update.ytdlp.btn.title": "yt-dlp パーサーエンジンを確認・更新 (動画サイトのアンチスクレイプ対策に頻繁に更新)",
    "update.ytdlp.done": "✅ yt-dlp の更新に成功しました！",
    "update.ytdlp.retry": "🔄 再試行",
    "update.ffmpeg.title": "🔄 FFmpeg を更新中...",
    "update.ffmpeg.checking": "🔄 FFmpeg を確認中...",
    "update.ffmpeg.btn": "🔄 ffmpeg 更新",
    "update.ffmpeg.btn.title": "ffmpeg コンポーネントを確認・更新 (通常は更新不要)",
    "update.ffmpeg.done": "✅ FFmpeg の更新に成功しました！",
    "update.ffmpeg.retry": "🔄 再試行",
    "update.checking.btn": "⏳ 確認中...",
    "update.connecting": "GitHub 公式ソースに接続してバージョンを比較中...",
    "update.downloading": "ダウンロード中...",
    "update.from.github": "GitHub 公式ソースから最新版を取得中...",
    "update.ffmpeg.detecting": "FFmpeg ランタイムを検出中...",
    "update.fail": "❌ 更新に失敗: {err}",
    "update.ffmpeg.fail": "❌ FFmpeg 更新に失敗: {err}",
    "update.fail.timeout": "ネットワーク接続タイムアウト",
    "ver.app.label": "アプリ",
    "ver.ytdlp.label": "yt-dlp",
    "ver.ffmpeg.label": "ffmpeg",
    "ver.loading": "読込中...",
    "ver.ytdlp.missing": "未インストール (更新をクリック)",
    "ver.ffmpeg.missing": "未インストール (更新をクリック)",
    // Platforms Ecosystem & Showcase
    "showcase.eyebrow": "1,800+ PLATFORMS POWERED BY YT-DLP & NATIVE ENGINES",
    "showcase.title": "世界中の主要動画・SNSプラットフォームに対応",
    "showcase.subtitle": "1,800以上のサイトを網羅する公式 yt-dlp 完全互換エンジンと独自開発の 16スレッド並行 HLS 復号エンジンを搭載。世界中の主要 SNS ショート動画から長編ドラマ、高画質配信までワンクリックで最高速保存。",
    "showcase.search.btn": "全対応サイト一覧・検索 (1,800以上のサイトを検索)",
    "showcase.more": "+ 1,800 以上のサイト",
    "modal.platforms.title": "🔍 対応動画サイト検索ディレクトリ",
    "modal.platforms.subtitle": "世界中の 1,800 以上の主要配信・SNS・独自復号サイトを収録。キーワード入力で瞬時に検索できます。",
    "modal.search.placeholder": "対応サイト名やドメインで検索（例：YouTube, Bilibili, TVer, TikTok, Twitter, SupJav...）",
    "modal.count.total": "全 {n}+ サイト",
    "modal.count.found": "{n} 件のサイト",
    "modal.empty.title": "一致するプリセットが見つかりません",
    "modal.empty": "完全一致するプリセットは見つかりませんでしたが、OmniVideoRip は 1,800 以上の yt-dlp フルエンジンを搭載しています。動画リンクを貼り付けるだけで自動解析・ダウンロードが可能です！",
    "modal.card.fill": "クリックして {name} のURLを入力",
    "modal.more": "最初の {max} 件を表示中（全 {n} 件）。より具体的なキーワードで絞り込んでください...",
    "cat.all": "すべて (1800+)",
    "cat.popular": "人気・主要",
    "cat.social": "ショート・SNS",
    "cat.streaming": "配信・テレビ",
    "cat.japan": "日本・アジア",
    "cat.audio": "音楽・音声",
    "cat.adult": "成人・独自復号",
    "platform.generic": "汎用パーサー",
    "history.eyebrow": "RECENT COMPLETED",
    "history.title": "履歴",
    "history.clear": "履歴をクリア",
    "history.default.filename": "ダウンロード動画.mp4",
    "footer.text": "OmniVideoRip Desktop Client © 2026 · 全ネット動画高速ダウンロード",
    "ad.top": "Google AdSense 広告スロット (TOP BANNER - 728x90 / レスポンシブ)",
    "ad.bottom": "Google AdSense 広告スロット (BOTTOM BANNER - 728x90 / レスポンシブ)",
    "alert.analyze.fail": "動画の解析に失敗: {err}",
    "alert.download.fail": "ダウンロードタスクの作成に失敗: {err}",
    "alert.batch.empty": "少なくとも1つの動画URLを入力してください！",
    "alert.batch.fail": "一括タスクの送信に失敗: {err}",
    "alert.copy.ok": "ファイルパスをクリップボードにコピーしました！",
    "analyzing": "解析中...",
    "joining.queue": "ダウンロードキューに追加中...",
    "download.error": "ダウンロードエラーが発生",
    "download.starting": "動画ストリーム接続を確立中...",
    "download.queued.msg": "キュー待機中...",
    "download.starting.multi": "高速マルチスレッドダウンロードを開始...",
    "page.title": "OmniVideoRip | 全ネット動画高速ダウンロード",
    "page.meta.description": "OmniVideoRip 全機能動画ダウンローダー。YouTube、Bilibili、GGJav、MissAV、TikTok、X (Twitter)、Instagram 等 1800+ サイトのHD動画をローカルに高速抽出・一括キューダウンロード。",
    "res.meta": "{platform} · 再生時間: {duration}",
    "res.duration.unknown": "不明",
    "res.title.unknown": "不明な動画タイトル",
    "mock.title": "サンプル動画タイトル",
    // SupJav Modal & Tutorial
    "supjav.guide.badge": "⚡ F12 抽出ガイド",
    "supjav.modal.title": "⚡ SupJav ダウンロード・F12 .m3u8 抽出ガイド",
    "supjav.modal.subtitle": "Cloudflare Turnstile 対策 · F12 開発者ツールで .m3u8 を取得",
    "supjav.modal.banner": "SupJav は Cloudflare Turnstile 認証と難読化された再生リンクを使用しています。ブラウザで F12 を押して .m3u8 ストリームURLをコピーするだけで、16スレッド最高速ダウンロードが可能です：",
    "supjav.modal.m1.badge": "⭐ 推奨方法: ブラウザで F12 を押して .m3u8 ストリームを取得 (100% 確実)",
    "supjav.modal.m1.title": "4ステップで .m3u8 ストリームURLを取得",
    "supjav.modal.m1.s1": "下の【1. ブラウザで動画ページを開く】をクリックし、Cloudflare の「人間であることを確認します」にチェックを入れます。",
    "supjav.modal.m1.s2": "キーボードの F12 キーを押す（または右クリック → 検証 / Inspect）と開発者ツールが開くので、上部の【Network (ネットワーク)】タブを選び、検索欄に m3u8 と入力します。",
    "supjav.modal.m1.s3": "動画の再生ボタンをクリックします（表示されない場合は F5 で再読み込みして再生）。一覧に .m3u8 項目（master.m3u8 / index.m3u8 等）が表示されます。",
    "supjav.modal.m1.s4": ".m3u8 項目を右クリック → 【コピー (Copy) → リンクのアドレスをコピー (Copy URL)】を選択し、下の【2. クリップボードの .m3u8 を読み取って保存】をクリックします！",
    "supjav.modal.m1.btn_open": "独立小ウィンドウで開く",
    "supjav.modal.m1.btn_clip": "2. クリップボードの .m3u8 を読み取って保存",
    "supjav.modal.m2.badge": "⚡ 手動貼り付け: .m3u8 URL または cURL を直接入力",
    "supjav.modal.m2.title": "F12 Network からコピーした .m3u8 リンクを貼り付け",
    "supjav.modal.m2.desc": "F12 ネットワークパネルからコピーした .m3u8 リンク（または Copy as cURL）を直接下に貼り付けることもできます：",
    "supjav.modal.m2.ph": "https://.../master.m3u8 リンクまたは cURL をここに貼り付け...",
    "supjav.modal.m2.btn": "🚀 解析してダウンロード",
    "supjav.modal.btn_browser": "1. ブラウザで動画ページを開く",
    "supjav.modal.btn_close": "閉じて戻る",
    "supjav.clip.success": "✅ クリップボードから .m3u8 ストリームURLを検出しました。解析を開始します...",
    "supjav.clip.empty": "⚠️ クリップボードに .m3u8 リンクが見つかりません。ブラウザで F12 -> Network -> m3u8 で検索し、URLをコピーしてからお試しください！",
    "supjav.window.opened": "🌐 動画ページを開きました！認証完了後、F12 -> Network -> m3u8 でフィルタし、動画を再生して .m3u8 URLをコピーしてください。",
    // Delete Task Confirmation
    "confirm.delete.title": "🗑️ タスク削除の確認",
    "confirm.delete.message": "このダウンロードタスクを削除してもよろしいですか？",
    "confirm.delete.subtext": "⚠️ 未完了のセグメントおよびダウンロードキャッシュも同時に削除されます。",
    "confirm.delete.cancel": "キャンセル",
    "confirm.delete.ok": "削除してキャッシュをクリア",
    "confirm.delete.prompt": "このダウンロードタスクを削除してもよろしいですか？未完了のセグメントおよびダウンロードキャッシュも同時に削除されます。",
  },

  "zh-TW": {
    "brand.tagline": "全網影片極速擷取平台",
    "lang.auto": "🌐 自動 (系統語言)",
    "mode.single": "🔗 單連結下載",
    "mode.batch": "📋 批次多連結排隊下載",
    "mode.hint.single": "貼上單一影片頁面網址，自動分析可用解析度",
    "mode.hint.single.active": "單一影片極速分析模式已啟用。輸入單一連結即可快速擷取並自由選擇解析度。",
    "mode.hint.batch.active": "多影片批次排隊模式已啟用。支援多平台跨站混合輸入，每平台自動獨立並行擷取。",
    "input.placeholder": "在此貼上任意影片網址 (如 YouTube, Bilibili, TikTok, X, MissAV, GGJav...)",
    "input.paste": "📋 貼上",
    "input.paste.title": "一鍵貼上剪貼簿",
    "input.detect.auto": "自動識別",
    "btn.analyze": "解析畫質",
    "batch.placeholder": "在此貼上多個影片網址，每行一個連結（或空格分隔）。例如：\nhttps://www.youtube.com/watch?v=HlLx7oE7q3I\nhttps://missav.live/dm26/ja/jac-082\nhttps://www.bilibili.com/video/BV1xx411c7mD",
    "batch.paste": "📋 貼上剪貼簿多連結",
    "batch.clear": "清空列表",
    "batch.submit": "開始批次下載",
    "path.label": "📁 下載儲存位置:",
    "path.placeholder": "選擇或輸入影片儲存路徑",
    "path.browse": "變更目錄 📁",
    "res.loading.title": "正在讀取影片標題...",
    "res.label": "🎯 請選擇下載畫質 / 解析度:",
    "res.close": "✕ 關閉",
    "res.confirm": "開始下載",
    "res.best": "最佳畫質 (Best Quality)",
    "res.best.note": "自動拼接最高畫質與最佳音軌 (推薦)",
    "res.best.note.short": "自動混流最優影片軌與音訊軌",
    "res.original": "原畫",
    "active.title": "⚡ 正在下載",
    "active.badge": "{n} 個進行中",
    "active.pause": "⏸️ 暫停",
    "active.delete": "🗑️ 刪除",
    "active.delete.title": "刪除任務並清理快取",
    "active.msg.default": "正在全力下載切片與資料...",
    "queue.title": "目前佇列",
    "queue.badge": "{n} 個任務",
    "queue.waiting": "⏳ 等待中: {n}",
    "queue.failed": "❌ 失敗: {n}",
    "queue.done": "✅ 已完成: {n}",
    "queue.clear": "清空佇列與快取 🧹",
    "queue.status.queued": "⏳ 排隊等待中",
    "queue.status.paused": "⏸️ 已暫停",
    "queue.status.failed": "❌ 下載失敗",
    "queue.status.completed": "✅ 下載完成",
    "queue.resume": "▶️ 繼續下載",
    "queue.copy.path": "📋 複製路徑",
    "queue.error.prefix": "⚠️ 錯誤: ",
    "queue.file.prefix": "📂 檔案: ",
    "queue.file.saved": "已儲存至本機",
    "queue.waiting.msg": "排隊等待空閒通道啟動...",
    "update.ytdlp.title": "🔄 正在更新 yt-dlp 解析引擎...",
    "update.ytdlp.checking": "🔄 正在偵測 yt-dlp 版本...",
    "update.ytdlp.btn": "🔄 更新 yt-dlp",
    "update.ytdlp.btn.title": "檢查並更新 yt-dlp 解析引擎（針對各影片網站反爬蟲高頻更新）",
    "update.ytdlp.done": "✅ yt-dlp 官方元件更新成功！",
    "update.ytdlp.retry": "🔄 重試更新",
    "update.ffmpeg.title": "🔄 正在更新 FFmpeg 元件...",
    "update.ffmpeg.checking": "🔄 正在偵測 FFmpeg 元件...",
    "update.ffmpeg.btn": "🔄 更新 ffmpeg",
    "update.ffmpeg.btn.title": "檢查並更新 ffmpeg 元件（音影片合併庫，通常無需更新）",
    "update.ffmpeg.done": "✅ FFmpeg 元件更新成功！",
    "update.ffmpeg.retry": "🔄 重試更新",
    "update.checking.btn": "⏳ 偵測中...",
    "update.connecting": "正在連線 GitHub 官方源比對版本...",
    "update.downloading": "正在下載...",
    "update.from.github": "正在從 GitHub 官方源取得最新版本...",
    "update.ffmpeg.detecting": "正在偵測 FFmpeg 執行庫...",
    "update.fail": "❌ 更新失敗: {err}",
    "update.ffmpeg.fail": "❌ FFmpeg 更新失敗: {err}",
    "update.fail.timeout": "網路連線逾時",
    "ver.app.label": "主程式",
    "ver.ytdlp.label": "yt-dlp",
    "ver.ffmpeg.label": "ffmpeg",
    "ver.loading": "載入中...",
    "ver.ytdlp.missing": "未安裝 (需點擊更新)",
    "ver.ffmpeg.missing": "未安裝 (需點擊更新)",
    // Platforms Ecosystem & Showcase
    "showcase.eyebrow": "1,800+ PLATFORMS POWERED BY YT-DLP & NATIVE ENGINES",
    "showcase.title": "支援全網超千家影音與社群平台",
    "showcase.subtitle": "整合官方完整 yt-dlp 1800+ 網站解析生態與自研 16 執行緒並行串流解密核心，全球主流社群短影音、海內外長劇集、高碼率切片原畫一鍵極速下載。",
    "showcase.search.btn": "查詢全部支援網站名錄 (可搜尋 1800+ 網站)",
    "showcase.more": "+ 1,800 更多平台",
    "modal.platforms.title": "🔍 支援網站全量查詢名錄",
    "modal.platforms.subtitle": "收錄全球 1,800+ 主流媒體、社群平台與自研解密網站，輸入關鍵字即可即時篩選",
    "modal.search.placeholder": "搜尋支援網站或網域（例如：YouTube, B站, TVer, TikTok, Twitter, SupJav...）",
    "modal.count.total": "共 {n}+ 網站",
    "modal.count.found": "找到 {n} 個網站",
    "modal.empty.title": "未找到相符的預置名稱",
    "modal.empty": "未找到相符的預置名稱，但 OmniVideoRip 搭載官方完整 yt-dlp 1,800+ 引擎，您可以直接貼上影片連結進行智慧嗅探與通用解析！",
    "modal.card.fill": "點擊填入 {name} 網址",
    "modal.more": "已顯示前 {max} 個結果（共 {n} 個），請輸入更精確的關鍵字篩選...",
    "cat.all": "全部 (1800+)",
    "cat.popular": "熱門主流",
    "cat.social": "社群短影音",
    "cat.streaming": "串流媒體/電視",
    "cat.japan": "日韓專區",
    "cat.audio": "音樂音訊",
    "cat.adult": "成人正片解密",
    "platform.generic": "全網通用解析",
    "history.eyebrow": "RECENT COMPLETED",
    "history.title": "歷史紀錄",
    "history.clear": "清空紀錄",
    "history.default.filename": "下載影片.mp4",
    "footer.text": "OmniVideoRip Desktop Client © 2026 · 全網影片極速擷取平台",
    "ad.top": "Google AdSense 廣告展示位 (TOP BANNER - 728x90 / 響應式)",
    "ad.bottom": "Google AdSense 廣告展示位 (BOTTOM BANNER - 728x90 / 響應式)",
    "alert.analyze.fail": "解析影片失敗: {err}",
    "alert.download.fail": "建立下載任務失敗: {err}",
    "alert.batch.empty": "請至少輸入一個影片播放頁面連結！",
    "alert.batch.fail": "批次提交任務失敗: {err}",
    "alert.copy.ok": "檔案路徑已成功複製到剪貼簿！",
    "analyzing": "正在解析畫質...",
    "joining.queue": "正在加入下載佇列...",
    "download.error": "下載遇到異常",
    "download.starting": "正在建立影片串流連線...",
    "download.queued.msg": "排隊等待中...",
    "download.starting.multi": "開始極速多執行緒下載...",
    "page.title": "OmniVideoRip | 全網影片極速擷取器",
    "page.meta.description": "OmniVideoRip 全能影片擷取器。支援 YouTube、Bilibili、GGJav、MissAV、TikTok、X (Twitter)、Instagram 等 1800+ 站點的高清音影片本地極速擷取與多任務批次排隊下載。",
    "res.meta": "{platform} · 時長: {duration}",
    "res.duration.unknown": "未知",
    "res.title.unknown": "未知影片標題",
    "mock.title": "範例全網影片標題",
    // SupJav Modal & Tutorial
    "supjav.guide.badge": "⚡ F12 抓流指引",
    "supjav.modal.title": "⚡ SupJav 影片下載與 F12 抓流指引",
    "supjav.modal.subtitle": "Cloudflare 5秒盾驗證攔截應對指南 · F12 擷取 .m3u8 極速直下",
    "supjav.modal.banner": "SupJav 全站受 Cloudflare Turnstile 人機驗證保護且播放線路經過動態混淆，常規後台請求會被攔截。只需在瀏覽器按 F12 抓取 .m3u8 真實影片串流即可開啟 16 執行緒滿速下載：",
    "supjav.modal.m1.badge": "⭐ 推薦方案：瀏覽器按 F12 擷取 .m3u8 影片串流 (100% 成功)",
    "supjav.modal.m1.title": "四步快速取得真實 .m3u8 播放串流網址",
    "supjav.modal.m1.s1": "點擊下方【1. 在瀏覽器中打開影片頁】，手動勾選一次 Cloudflare「確認您是人類」核取方塊。",
    "supjav.modal.m1.s2": "在瀏覽器頁面按下鍵盤 F12（或右鍵頁面選擇「檢查 / Inspect」），點擊頂部【Network (網路)】分頁，在過濾框中輸入 m3u8。",
    "supjav.modal.m1.s3": "點擊網頁中的影片播放按鈕（若列表為空可按 F5 重新整理後再點播放），列表中會出現 .m3u8 項目（如 master.m3u8 / index.m3u8）。",
    "supjav.modal.m1.s4": "右鍵該 .m3u8 項目 → 選擇【複製 (Copy) → 複製連結網址 (Copy URL)】，隨後點擊下方【2. 讀取剪貼簿 .m3u8 並下載】即可 16 執行緒極速直下！",
    "supjav.modal.m1.btn_open": "啟動獨立小視窗打開",
    "supjav.modal.m1.btn_clip": "2. 讀取剪貼簿 .m3u8 並下載",
    "supjav.modal.m2.badge": "⚡ 手動貼上：直接輸入 .m3u8 連結或 cURL 指令",
    "supjav.modal.m2.title": "貼上從 F12 Network 複製的 .m3u8 網址",
    "supjav.modal.m2.desc": "您也可以將從 F12 網路面板複製的 .m3u8 連結（或 Copy as cURL 指令）直接貼到下方輸入框：",
    "supjav.modal.m2.ph": "在此貼上 F12 抓取的 https://.../master.m3u8 連結或 cURL...",
    "supjav.modal.m2.btn": "🚀 立即解析下載",
    "supjav.modal.btn_browser": "1. 在瀏覽器中打開影片頁",
    "supjav.modal.btn_close": "我知道了，返回主介面",
    "supjav.clip.success": "✅ 已成功從剪貼簿識別到 .m3u8 影片串流網址，正在解析並準備下載...",
    "supjav.clip.empty": "⚠️ 剪貼簿中未偵測到 .m3u8 連結。請在瀏覽器按 F12 -> Network -> 搜尋 m3u8 -> 右鍵複製 .m3u8 連結網址後再試！",
    "supjav.window.opened": "🌐 已打開影片頁！請完成人機驗證後按 F12 -> Network -> 篩選 m3u8 -> 播放影片並右鍵複製 .m3u8 連結網址。",
    // Delete Task Confirmation
    "confirm.delete.title": "🗑️ 刪除任務確認",
    "confirm.delete.message": "確定要刪除該下載任務嗎？",
    "confirm.delete.subtext": "⚠️ 此操作將同時清理未完成的暫存切片與下載快取，釋放磁碟空間。",
    "confirm.delete.cancel": "取消",
    "confirm.delete.ok": "確認刪除並清理快取",
    "confirm.delete.prompt": "確定要刪除該下載任務嗎？此操作將同時清理未完成的暫存切片與下載快取檔案。",
  },
};

// Current locale state
let _currentLocale = "zh-CN";

/**
 * Resolve locale string to a supported key.
 * e.g. "en-US" -> "en", "ja-JP" -> "ja", "zh-TW" -> "zh-TW", "zh-Hant" -> "zh-TW", "zh-CN" -> "zh-CN"
 */
function resolveLocale(raw) {
  if (!raw) return "zh-CN";
  const lower = raw.toLowerCase().replace("_", "-");

  if (lower.startsWith("zh")) {
    if (lower.includes("tw") || lower.includes("hk") || lower.includes("hant") || lower.includes("mo")) {
      return "zh-TW";
    }
    return "zh-CN";
  }
  if (lower.startsWith("ja")) return "ja";
  if (lower.startsWith("en")) return "en";

  // Fallback: try exact match
  if (I18N_TRANSLATIONS[raw]) return raw;

  return "zh-CN";
}

/**
 * Get translated string by key. Supports {placeholder} substitution.
 * @param {string} key - Translation key
 * @param {Object} [params] - Placeholder values, e.g. { n: 3, err: "timeout" }
 * @returns {string}
 */
function t(key, params) {
  const dict = I18N_TRANSLATIONS[_currentLocale] || I18N_TRANSLATIONS["zh-CN"];
  let text = dict[key];
  if (text === undefined) {
    // Fallback to zh-CN
    text = I18N_TRANSLATIONS["zh-CN"][key];
  }
  if (text === undefined) return key; // Return key itself as ultimate fallback

  if (params) {
    for (const [k, v] of Object.entries(params)) {
      text = text.replace(new RegExp(`\\{${k}\\}`, "g"), String(v));
    }
  }
  return text;
}

/**
 * Apply translations to all elements with data-i18n attribute.
 */
function applyStaticTranslations() {
  document.querySelectorAll("[data-i18n]").forEach((el) => {
    const key = el.getAttribute("data-i18n");
    const translated = t(key);
    if (translated !== key) {
      el.textContent = translated;
    }
  });

  // Handle special attributes
  document.querySelectorAll("[data-i18n-placeholder]").forEach((el) => {
    const key = el.getAttribute("data-i18n-placeholder");
    const translated = t(key);
    if (translated !== key) {
      el.placeholder = translated;
    }
  });

  document.querySelectorAll("[data-i18n-title]").forEach((el) => {
    const key = el.getAttribute("data-i18n-title");
    const translated = t(key);
    if (translated !== key) {
      el.title = translated;
    }
  });

  // Update page title
  document.title = t("page.title");

  // Update html lang attribute
  const langMap = { "zh-CN": "zh-CN", en: "en", ja: "ja", "zh-TW": "zh-TW" };
  document.documentElement.lang = langMap[_currentLocale] || "zh-CN";
}

// Language change listener callbacks
const _changeListeners = [];

/**
 * Initialize i18n. Detects system locale or loads manual preference from localStorage.
 */
async function initI18n() {
  const saved = localStorage.getItem("user_language") || "auto";

  if (saved !== "auto" && I18N_TRANSLATIONS[saved]) {
    _currentLocale = saved;
    console.log(`[i18n] Using user manual language: ${_currentLocale}`);
  } else {
    let rawLocale = null;
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetSystemLocale) {
        rawLocale = await window.go.main.App.GetSystemLocale();
      }
    } catch (e) {
      console.warn("Failed to get system locale from backend:", e);
    }

    if (!rawLocale) {
      rawLocale = navigator.language || navigator.userLanguage || "zh-CN";
    }

    _currentLocale = resolveLocale(rawLocale);
    console.log(`[i18n] Auto-detected system locale: ${rawLocale} → resolved: ${_currentLocale}`);
  }

  applyStaticTranslations();
  syncLangSelectUI();
}

/**
 * Manually switch language.
 * @param {string} lang - 'auto', 'zh-CN', 'zh-TW', 'en', 'ja'
 */
async function setLanguage(lang) {
  localStorage.setItem("user_language", lang);

  if (lang === "auto") {
    let rawLocale = null;
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetSystemLocale) {
        rawLocale = await window.go.main.App.GetSystemLocale();
      }
    } catch (e) {
      console.warn("Failed to get system locale:", e);
    }
    if (!rawLocale) {
      rawLocale = navigator.language || navigator.userLanguage || "zh-CN";
    }
    _currentLocale = resolveLocale(rawLocale);
  } else if (I18N_TRANSLATIONS[lang]) {
    _currentLocale = lang;
  }

  applyStaticTranslations();
  syncLangSelectUI();

  // Notify listeners
  _changeListeners.forEach((fn) => {
    try {
      fn(_currentLocale);
    } catch (err) {
      console.warn("Error in lang change listener:", err);
    }
  });
}

function syncLangSelectUI() {
  const select = document.getElementById("lang-select");
  if (select) {
    const saved = localStorage.getItem("user_language") || "auto";
    select.value = saved;
  }
}

function onLanguageChange(fn) {
  if (typeof fn === "function") {
    _changeListeners.push(fn);
  }
}

/**
 * Get current active locale key ('zh-CN', 'zh-TW', 'en', 'ja').
 */
function getCurrentLocale() {
  return _currentLocale;
}

/**
 * Get current setting ('auto', 'zh-CN', 'zh-TW', 'en', 'ja').
 */
function getLanguageSetting() {
  return localStorage.getItem("user_language") || "auto";
}

// Export for use in app.js
window.i18n = {
  t,
  initI18n,
  setLanguage,
  getCurrentLocale,
  getLanguageSetting,
  applyStaticTranslations,
  onLanguageChange,
};
