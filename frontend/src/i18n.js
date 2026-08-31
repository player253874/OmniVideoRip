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
    // Platforms
    "platforms.eyebrow": "SUPPORTED PLATFORMS DIRECTORY",
    "platforms.title": "全网主要支持平台名录",
    "platform.generic": "全网通用解析",
    "platform.instagram.desc": "支持 Instagram Reels 短视频、帖子视频与快拍动态高清原画嗅探提取。",
    "platform.instagram.tag1": "Reels 短视频",
    "platform.instagram.tag2": "高清无水印",
    "platform.youtube.desc": "支持 4K/8K 极速解析、自由选择 1080P/720P/480P 画质与 MP3 无损音频提取。",
    "platform.youtube.tag1": "4K / 8K",
    "platform.youtube.tag2": "画质自选",
    "platform.youtube.tag3": "音频提取",
    "platform.bilibili.desc": "支持番剧、UP 主投稿视频 1080P/4K 提取，多 P 分卷批量抓取。",
    "platform.bilibili.tag1": "1080P/4K",
    "platform.bilibili.tag2": "中文文件名原生支持",
    "platform.missav.desc": "Go 原生 16 线程 Surrit HLS 并发解密引擎，全长高清正片 5~10 倍极速直下。",
    "platform.missav.tag1": "16 线程并发",
    "platform.missav.tag2": "全长正片",
    "platform.ggjav.desc": "Go 原生 16 线程 HLS 切片解密算法，自动合并播放器分段。",
    "platform.ggjav.tag1": "16 线程并发",
    "platform.ggjav.tag2": "原画 MP4",
    "platform.rouvideo.desc": "Go 原生 16 线程自动解密 HLS 索引，全长高清 720P/1080P 极速直下并自动合并。",
    "platform.rouvideo.tag1": "16 线程解密直下",
    "platform.rouvideo.tag2": "原画 MP4",
    "platform.tiktok.desc": "支持无水印 1080P 短视频提取，背景 BGM 原声音频自动分离。",
    "platform.tiktok.tag1": "无水印提取",
    "platform.x.desc": "支持推文内多视频抓取、高清画质自动匹配与推文动图 GIF 转 MP4。",
    "platform.x.tag1": "多视频解析",
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
    "platforms.eyebrow": "SUPPORTED PLATFORMS DIRECTORY",
    "platforms.title": "Supported Platforms",
    "platform.generic": "Universal Parser",
    "platform.instagram.desc": "Supports Instagram Reels, post videos and Stories in HD original quality.",
    "platform.instagram.tag1": "Reels",
    "platform.instagram.tag2": "HD No Watermark",
    "platform.youtube.desc": "Supports 4K/8K fast analysis, freely choose 1080P/720P/480P quality and lossless MP3 audio extraction.",
    "platform.youtube.tag1": "4K / 8K",
    "platform.youtube.tag2": "Quality Select",
    "platform.youtube.tag3": "Audio Extract",
    "platform.bilibili.desc": "Supports anime, UP's videos in 1080P/4K, multi-part batch extraction.",
    "platform.bilibili.tag1": "1080P/4K",
    "platform.bilibili.tag2": "CJK Filename Support",
    "platform.missav.desc": "Native Go 16-thread Surrit HLS concurrent decryption engine, 5-10x faster full-length HD downloads.",
    "platform.missav.tag1": "16 Threads",
    "platform.missav.tag2": "Full Length",
    "platform.ggjav.desc": "Native Go 16-thread HLS segment decryption, auto-merge player segments.",
    "platform.ggjav.tag1": "16 Threads",
    "platform.ggjav.tag2": "Original MP4",
    "platform.rouvideo.desc": "Native Go 16-thread automatic HLS decryption, download 720P/1080P original video fast.",
    "platform.rouvideo.tag1": "16-Thread Fast Rip",
    "platform.rouvideo.tag2": "Original MP4",
    "platform.tiktok.desc": "Supports watermark-free 1080P short video extraction with background BGM audio separation.",
    "platform.tiktok.tag1": "No Watermark",
    "platform.x.desc": "Supports multi-video extraction from tweets, auto HD quality matching and GIF-to-MP4.",
    "platform.x.tag1": "Multi-Video",
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
    "platforms.eyebrow": "SUPPORTED PLATFORMS DIRECTORY",
    "platforms.title": "対応プラットフォーム一覧",
    "platform.generic": "汎用パーサー",
    "platform.instagram.desc": "Instagram Reels ショート動画、投稿動画、ストーリーをHDオリジナル画質で抽出。",
    "platform.instagram.tag1": "Reels ショート",
    "platform.instagram.tag2": "HD 透かしなし",
    "platform.youtube.desc": "4K/8K 高速解析、1080P/720P/480P 画質自由選択、MP3 ロスレス音声抽出対応。",
    "platform.youtube.tag1": "4K / 8K",
    "platform.youtube.tag2": "画質選択",
    "platform.youtube.tag3": "音声抽出",
    "platform.bilibili.desc": "番組、UP主投稿の 1080P/4K 抽出、マルチパート一括ダウンロード対応。",
    "platform.bilibili.tag1": "1080P/4K",
    "platform.bilibili.tag2": "CJKファイル名対応",
    "platform.missav.desc": "Go ネイティブ 16スレッド Surrit HLS 並行復号エンジン、フルHD本編を5〜10倍速でダウンロード。",
    "platform.missav.tag1": "16スレッド並行",
    "platform.missav.tag2": "フル本編",
    "platform.ggjav.desc": "Go ネイティブ 16スレッド HLS セグメント復号アルゴリズム、プレーヤー分割を自動結合。",
    "platform.ggjav.tag1": "16スレッド並行",
    "platform.ggjav.tag2": "オリジナル MP4",
    "platform.rouvideo.desc": "Go ネイティブ 16スレッド自動 HLS 復号、フルHD 720P/1080P を高速直下＆自動結合。",
    "platform.rouvideo.tag1": "16スレッド復号直下",
    "platform.rouvideo.tag2": "オリジナル MP4",
    "platform.tiktok.desc": "透かしなし1080Pショート動画抽出、BGM原音オーディオ自動分離対応。",
    "platform.tiktok.tag1": "透かしなし",
    "platform.x.desc": "ツイート内マルチ動画取得、HD画質自動マッチ、GIF→MP4変換対応。",
    "platform.x.tag1": "マルチ動画解析",
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
    "platforms.eyebrow": "SUPPORTED PLATFORMS DIRECTORY",
    "platforms.title": "全網主要支援平台名錄",
    "platform.generic": "全網通用解析",
    "platform.instagram.desc": "支援 Instagram Reels 短影片、貼文影片與限時動態高清原畫嗅探擷取。",
    "platform.instagram.tag1": "Reels 短影片",
    "platform.instagram.tag2": "高清無浮水印",
    "platform.youtube.desc": "支援 4K/8K 極速解析、自由選擇 1080P/720P/480P 畫質與 MP3 無損音訊擷取。",
    "platform.youtube.tag1": "4K / 8K",
    "platform.youtube.tag2": "畫質自選",
    "platform.youtube.tag3": "音訊擷取",
    "platform.bilibili.desc": "支援番劇、UP 主投稿影片 1080P/4K 擷取，多 P 分卷批次抓取。",
    "platform.bilibili.tag1": "1080P/4K",
    "platform.bilibili.tag2": "中文檔名原生支援",
    "platform.missav.desc": "Go 原生 16 執行緒 Surrit HLS 並行解密引擎，全長高清正片 5~10 倍極速直下。",
    "platform.missav.tag1": "16 執行緒並行",
    "platform.missav.tag2": "全長正片",
    "platform.ggjav.desc": "Go 原生 16 執行緒 HLS 切片解密演算法，自動合併播放器分段。",
    "platform.ggjav.tag1": "16 執行緒並行",
    "platform.ggjav.tag2": "原畫 MP4",
    "platform.rouvideo.desc": "Go 原生 16 執行緒自動解密 HLS 索引，全長高清 720P/1080P 極速直下並自動合併。",
    "platform.rouvideo.tag1": "16 執行緒解密直下",
    "platform.rouvideo.tag2": "原畫 MP4",
    "platform.tiktok.desc": "支援無浮水印 1080P 短影片擷取，背景 BGM 原聲音訊自動分離。",
    "platform.tiktok.tag1": "無浮水印擷取",
    "platform.x.desc": "支援推文內多影片抓取、高清畫質自動匹配與推文動圖 GIF 轉 MP4。",
    "platform.x.tag1": "多影片解析",
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
