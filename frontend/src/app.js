const PLATFORM_MAP = {
  youtube: { name: "YouTube", icon: "🔴", color: "#ef4444" },
  bilibili: { name: "Bilibili", icon: "🔵", color: "#3b82f6" },
  missav: { name: "MissAV", icon: "🌸", color: "#ec4899" },
  ggjav: { name: "GGJav", icon: "🟣", color: "#a855f7" },
  rouvideo: { name: "RouVideo (肉视频)", icon: "🥩", color: "#f472b6" },
  tiktok: { name: "TikTok / 抖音", icon: "🎵", color: "#06b6d4" },
  x: { name: "X (Twitter)", icon: "🖤", color: "#94a3b8" },
  instagram: { name: "Instagram", icon: "📷", color: "#e1306c" },
  generic: { get name() { return window.i18n ? window.i18n.t("platform.generic") : "全网通用解析"; }, icon: "🌐", color: "#10b981" },
};

function getPlatformInfo(key) {
  if (key && PLATFORM_MAP[key]) return PLATFORM_MAP[key];
  if (!key || key === "generic") return PLATFORM_MAP.generic;
  return {
    name: key,
    icon: "🌐",
    color: "#10b981",
  };
}

const state = {
  activeMode: "single",
  pollHandle: null,
  outputDir: localStorage.getItem("custom_output_dir") || "",
  tasksMap: new Map(),
  history: JSON.parse(localStorage.getItem("download_history") || "[]"),
  currentMetaData: null,
  selectedFormatId: "best",
};

const elements = {
  tabSingleMode: document.getElementById("tab-single-mode"),
  tabBatchMode: document.getElementById("tab-batch-mode"),
  modeHintText: document.getElementById("mode-hint-text"),
  singleUrlBox: document.getElementById("single-url-box"),
  batchUrlBox: document.getElementById("batch-url-box"),
  batchUrlInput: document.getElementById("batch-url-input"),
  batchPasteBtn: document.getElementById("batch-paste-btn"),
  batchClearBtn: document.getElementById("batch-clear-btn"),

  singleOutputDirInput: document.getElementById("single-output-dir-input"),
  singleBrowseBtn: document.getElementById("single-browse-btn"),

  downloadForm: document.getElementById("download-form"),
  urlInput: document.getElementById("url-input"),
  pasteBtn: document.getElementById("paste-btn"),
  clearBtn: document.getElementById("clear-btn"),
  analyzeSubmitBtn: document.getElementById("analyze-submit-btn"),
  detectedPlatformBadge: document.getElementById("detected-platform-badge"),
  platformIcon: document.getElementById("platform-icon"),
  platformName: document.getElementById("platform-name"),

  // Resolution Selector Card Elements
  resolutionSelectorCard: document.getElementById("resolution-selector-card"),
  resVideoThumb: document.getElementById("res-video-thumb"),
  resVideoTitle: document.getElementById("res-video-title"),
  resVideoMeta: document.getElementById("res-video-meta"),
  resCloseBtn: document.getElementById("res-close-btn"),
  resChipsContainer: document.getElementById("res-chips-container"),
  resConfirmDownloadBtn: document.getElementById("res-confirm-download-btn"),

  // Active Downloads Section (Above Current Queue)
  activeDownloadsSection: document.getElementById("active-downloads-section"),
  activeDownloadsList: document.getElementById("active-downloads-list"),
  activeRunningBadge: document.getElementById("active-running-badge"),

  // Current Queue Panel (Queued -> Failed -> Completed)
  queuePanel: document.getElementById("queue-panel"),
  queueCountBadge: document.getElementById("queue-count-badge"),
  queueWaitingCount: document.getElementById("queue-waiting-count"),
  queueFailedCount: document.getElementById("queue-failed-count"),
  queueDoneCount: document.getElementById("queue-done-count"),
  clearQueueBtn: document.getElementById("clear-queue-btn"),
  queueTasksList: document.getElementById("queue-tasks-list"),

  historySection: document.getElementById("history-section"),
  historyList: document.getElementById("history-list"),
  clearHistoryBtn: document.getElementById("clear-history-btn"),

  // Component Updater
  btnUpdateYtdlp: document.getElementById("btn-update-ytdlp"),
  btnUpdateFFmpeg: document.getElementById("btn-update-ffmpeg"),
  updateProgressCard: document.getElementById("update-progress-card"),
  updateTitle: document.getElementById("update-title"),
  updatePct: document.getElementById("update-pct"),
  updateBarFill: document.getElementById("update-bar-fill"),
  updateMsg: document.getElementById("update-msg"),
};

let _lastVersionInfo = null;

document.addEventListener("DOMContentLoaded", async () => {
  if (window.i18n && window.i18n.initI18n) {
    await window.i18n.initI18n();
  }
  await initWailsApp();
  bindEvents();
  renderHistory();
});

// i18n helper shorthand
function t(key, params) {
  return window.i18n ? window.i18n.t(key, params) : key;
}

function renderVersionInfo(verInfo) {
  if (!verInfo) return;
  _lastVersionInfo = verInfo;

  const verApp = document.getElementById("ver-app");
  const verYtdlp = document.getElementById("ver-ytdlp");
  const verFfmpeg = document.getElementById("ver-ffmpeg");

  if (verApp && verInfo.appVersion) {
    verApp.textContent = "v" + verInfo.appVersion;
  }

  if (verYtdlp) {
    if (verInfo.ytdlpVersion && verInfo.ytdlpVersion.trim() !== "") {
      verYtdlp.textContent = verInfo.ytdlpVersion;
      verYtdlp.classList.remove("ver-missing");
      verYtdlp.classList.add("ver-ready");
    } else {
      verYtdlp.textContent = t("ver.ytdlp.missing");
      verYtdlp.classList.remove("ver-ready");
      verYtdlp.classList.add("ver-missing");
    }
  }

  if (verFfmpeg) {
    if (verInfo.ffmpegVersion && verInfo.ffmpegVersion.trim() !== "") {
      verFfmpeg.textContent = verInfo.ffmpegVersion;
      verFfmpeg.classList.remove("ver-missing");
      verFfmpeg.classList.add("ver-ready");
    } else {
      verFfmpeg.textContent = t("ver.ffmpeg.missing");
      verFfmpeg.classList.remove("ver-ready");
      verFfmpeg.classList.add("ver-missing");
    }
  }
}

async function refreshVersionInfo() {
  if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetVersionInfo) {
    try {
      const verInfo = await window.go.main.App.GetVersionInfo();
      if (verInfo) {
        renderVersionInfo(verInfo);
      }
    } catch (e) {
      console.warn("Refresh version error:", e);
    }
  }
}

async function initWailsApp() {
  if (window.go && window.go.main && window.go.main.App) {
    try {
      await refreshVersionInfo();

      const defaultDir = await window.go.main.App.GetDefaultDownloadDir();
      if (defaultDir && !localStorage.getItem("custom_output_dir")) {
        state.outputDir = defaultDir;
      }

      const allTasks = await window.go.main.App.GetAllTasks();
      if (allTasks && allTasks.length > 0) {
        allTasks.forEach((t) => state.tasksMap.set(t.id, t));
        renderActiveDownloads();
        renderQueuePanel();
      }
    } catch (e) {
      console.warn("Wails binding error:", e);
    }
  }
  syncOutputDirUI();

  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn("version:ready", (verInfo) => {
      if (verInfo) {
        renderVersionInfo(verInfo);
      }
    });

    window.runtime.EventsOn("task:progress", (task) => {
      if (task && task.id) {
        const prev = state.tasksMap.get(task.id);
        state.tasksMap.set(task.id, task);

        if (task.status === "completed" && (!prev || prev.status !== "completed")) {
          addHistoryItem(task);
        }

        renderActiveDownloads();
        renderQueuePanel();
      }
    });

    window.runtime.EventsOn("task:deleted", (taskId) => {
      state.tasksMap.delete(taskId);
      renderActiveDownloads();
      renderQueuePanel();
    });

    window.runtime.EventsOn("tasks:cleared", () => {
      state.tasksMap.clear();
      renderActiveDownloads();
      renderQueuePanel();
    });

    // yt-dlp Update Events
    window.runtime.EventsOn("ytdlp:update:progress", (data) => {
      if (!data) return;
      if (elements.updateProgressCard) {
        elements.updateProgressCard.classList.remove("hidden");
      }
      if (elements.updateTitle) elements.updateTitle.textContent = t("update.ytdlp.title");
      if (elements.updatePct) elements.updatePct.textContent = `${data.percent}%`;
      if (elements.updateBarFill) elements.updateBarFill.style.width = `${data.percent}%`;
      if (elements.updateMsg) elements.updateMsg.textContent = data.message || t("update.downloading");
    });

    window.runtime.EventsOn("ytdlp:update:done", (data) => {
      if (elements.updatePct) elements.updatePct.textContent = "100%";
      if (elements.updateBarFill) elements.updateBarFill.style.width = "100%";
      if (elements.updateMsg) elements.updateMsg.textContent = data.message || t("update.ytdlp.done");
      if (elements.btnUpdateYtdlp) {
        elements.btnUpdateYtdlp.disabled = false;
        elements.btnUpdateYtdlp.textContent = t("update.ytdlp.btn");
      }
      refreshVersionInfo();
      setTimeout(() => {
        if (elements.updateProgressCard) {
          elements.updateProgressCard.classList.add("hidden");
        }
      }, 3500);
    });

    window.runtime.EventsOn("ytdlp:update:error", (data) => {
      if (elements.updateMsg) {
        elements.updateMsg.textContent = t("update.fail", { err: data.error || t("update.fail.timeout") });
      }
      if (elements.btnUpdateYtdlp) {
        elements.btnUpdateYtdlp.disabled = false;
        elements.btnUpdateYtdlp.textContent = t("update.ytdlp.retry");
      }
    });

    // FFmpeg Update Events
    window.runtime.EventsOn("ffmpeg:update:progress", (data) => {
      if (!data) return;
      if (elements.updateProgressCard) {
        elements.updateProgressCard.classList.remove("hidden");
      }
      if (elements.updateTitle) elements.updateTitle.textContent = t("update.ffmpeg.title");
      if (elements.updatePct) elements.updatePct.textContent = `${data.percent}%`;
      if (elements.updateBarFill) elements.updateBarFill.style.width = `${data.percent}%`;
      if (elements.updateMsg) elements.updateMsg.textContent = data.message || t("update.downloading");
    });

    window.runtime.EventsOn("ffmpeg:update:done", (data) => {
      if (elements.updatePct) elements.updatePct.textContent = "100%";
      if (elements.updateBarFill) elements.updateBarFill.style.width = "100%";
      if (elements.updateMsg) elements.updateMsg.textContent = data.message || t("update.ffmpeg.done");
      if (elements.btnUpdateFFmpeg) {
        elements.btnUpdateFFmpeg.disabled = false;
        elements.btnUpdateFFmpeg.textContent = t("update.ffmpeg.btn");
      }
      refreshVersionInfo();
      setTimeout(() => {
        if (elements.updateProgressCard) {
          elements.updateProgressCard.classList.add("hidden");
        }
      }, 3500);
    });

    window.runtime.EventsOn("ffmpeg:update:error", (data) => {
      if (elements.updateMsg) {
        elements.updateMsg.textContent = t("update.ffmpeg.fail", { err: data.error || t("update.fail.timeout") });
      }
      if (elements.btnUpdateFFmpeg) {
        elements.btnUpdateFFmpeg.disabled = false;
        elements.btnUpdateFFmpeg.textContent = t("update.ffmpeg.retry");
      }
    });
  }
}

function syncOutputDirUI() {
  if (elements.singleOutputDirInput) {
    elements.singleOutputDirInput.value = state.outputDir;
  }
}

function bindEvents() {
  const langSelect = document.getElementById("lang-select");
  langSelect?.addEventListener("change", async (e) => {
    const selectedLang = e.target.value;
    if (window.i18n && window.i18n.setLanguage) {
      await window.i18n.setLanguage(selectedLang);
      renderActiveDownloads();
      renderQueuePanel();
      renderHistory();
      if (_lastVersionInfo) {
        renderVersionInfo(_lastVersionInfo);
      } else {
        refreshVersionInfo();
      }
      if (state.currentMetaData) {
        renderResolutionOptions(state.currentMetaData);
      }
      if (elements.modeHintText) {
        if (state.activeMode === "single") {
          elements.modeHintText.textContent = t("mode.hint.single.active");
        } else {
          elements.modeHintText.textContent = t("mode.hint.batch.active");
        }
      }
    }
  });

  elements.btnUpdateYtdlp?.addEventListener("click", async () => {
    if (elements.btnUpdateYtdlp) {
      elements.btnUpdateYtdlp.disabled = true;
      elements.btnUpdateYtdlp.textContent = t("update.checking.btn");
    }
    if (elements.updateProgressCard) {
      elements.updateProgressCard.classList.remove("hidden");
      elements.updateTitle.textContent = t("update.ytdlp.checking");
      elements.updatePct.textContent = "0%";
      elements.updateBarFill.style.width = "0%";
      elements.updateMsg.textContent = t("update.connecting");
    }
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.UpdateYTDLPComponent) {
        await window.go.main.App.UpdateYTDLPComponent();
      }
    } catch (err) {
      if (elements.updateMsg) {
        elements.updateMsg.textContent = t("update.fail", { err });
      }
      if (elements.btnUpdateYtdlp) {
        elements.btnUpdateYtdlp.disabled = false;
        elements.btnUpdateYtdlp.textContent = t("update.ytdlp.btn");
      }
    }
  });

  elements.btnUpdateFFmpeg?.addEventListener("click", async () => {
    if (elements.btnUpdateFFmpeg) {
      elements.btnUpdateFFmpeg.disabled = true;
      elements.btnUpdateFFmpeg.textContent = t("update.checking.btn");
    }
    if (elements.updateProgressCard) {
      elements.updateProgressCard.classList.remove("hidden");
      elements.updateTitle.textContent = t("update.ffmpeg.checking");
      elements.updatePct.textContent = "0%";
      elements.updateBarFill.style.width = "0%";
      elements.updateMsg.textContent = t("update.ffmpeg.detecting");
    }
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.UpdateFFmpegComponent) {
        await window.go.main.App.UpdateFFmpegComponent();
      }
    } catch (err) {
      if (elements.updateMsg) {
        elements.updateMsg.textContent = t("update.fail", { err });
      }
      if (elements.btnUpdateFFmpeg) {
        elements.btnUpdateFFmpeg.disabled = false;
        elements.btnUpdateFFmpeg.textContent = t("update.ffmpeg.btn");
      }
    }
  });

  elements.tabSingleMode?.addEventListener("click", () => {
    state.activeMode = "single";
    elements.tabSingleMode.classList.add("active");
    elements.tabBatchMode.classList.remove("active");
    elements.singleUrlBox.classList.remove("hidden");
    elements.batchUrlBox.classList.add("hidden");
    elements.modeHintText.textContent = t("mode.hint.single.active");
  });

  elements.tabBatchMode?.addEventListener("click", () => {
    state.activeMode = "batch";
    elements.tabBatchMode.classList.add("active");
    elements.tabSingleMode.classList.remove("active");
    elements.batchUrlBox.classList.remove("hidden");
    elements.singleUrlBox.classList.add("hidden");
    elements.modeHintText.textContent = t("mode.hint.batch.active");
  });

  elements.urlInput?.addEventListener("input", (e) => {
    updatePlatformBadge(e.target.value.trim());
  });

  elements.urlInput?.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
      e.preventDefault();
      analyzeSingleVideo();
    }
  });

  elements.pasteBtn?.addEventListener("click", async () => {
    try {
      const text = await navigator.clipboard.readText();
      if (text) {
        elements.urlInput.value = text.trim();
        updatePlatformBadge(text.trim());
        analyzeSingleVideo();
      }
    } catch (err) {
      console.warn("Clipboard access failed:", err);
    }
  });

  elements.clearBtn?.addEventListener("click", () => {
    elements.urlInput.value = "";
    updatePlatformBadge("");
    elements.resolutionSelectorCard.classList.add("hidden");
    elements.urlInput.focus();
  });

  elements.analyzeSubmitBtn?.addEventListener("click", () => {
    analyzeSingleVideo();
  });

  elements.batchPasteBtn?.addEventListener("click", async () => {
    try {
      const text = await navigator.clipboard.readText();
      if (text) {
        elements.batchUrlInput.value = text.trim();
      }
    } catch (err) {
      console.warn("Clipboard access failed:", err);
    }
  });

  elements.batchClearBtn?.addEventListener("click", () => {
    elements.batchUrlInput.value = "";
    elements.batchUrlInput.focus();
  });

  elements.downloadForm?.addEventListener("submit", (e) => {
    e.preventDefault();
    if (state.activeMode === "batch") {
      submitBatchTasks();
    } else {
      analyzeSingleVideo();
    }
  });

  elements.singleBrowseBtn?.addEventListener("click", async () => {
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.SelectDirectory) {
        const selected = await window.go.main.App.SelectDirectory();
        if (selected) {
          state.outputDir = selected;
          localStorage.setItem("custom_output_dir", selected);
          syncOutputDirUI();
          if (window.go.main.App.SetDefaultDownloadDir) {
            await window.go.main.App.SetDefaultDownloadDir(selected);
          }
        }
      }
    } catch (err) {
      console.warn("Folder dialog failed:", err);
    }
  });

  elements.resCloseBtn?.addEventListener("click", () => {
    elements.resolutionSelectorCard.classList.add("hidden");
  });

  elements.resConfirmDownloadBtn?.addEventListener("click", () => {
    startDownloadAfterResolution();
  });

  elements.clearQueueBtn?.addEventListener("click", async () => {
    stopPolling();
    try {
      const appApi = window.go?.main?.App;
      if (appApi?.ClearAllTasksAndCaches) {
        await appApi.ClearAllTasksAndCaches();
      } else if (appApi?.ClearTasks) {
        await appApi.ClearTasks();
      } else {
        throw new Error("Clear task API is unavailable");
      }
      state.tasksMap.clear();
      renderActiveDownloads();
      renderQueuePanel();
    } catch (err) {
      console.warn("Clear tasks and caches failed:", err);
      window.alert(`清空队列与缓存失败：${err?.message || err}`);
    } finally {
      startPolling();
    }
  });

  elements.clearHistoryBtn?.addEventListener("click", () => {
    state.history = [];
    localStorage.removeItem("download_history");
    renderHistory();
  });
}

function updatePlatformBadge(rawUrl) {
  if (!rawUrl) {
    elements.detectedPlatformBadge.classList.add("hidden");
    return;
  }
  const key = detectPlatformKey(rawUrl);
  const pInfo = getPlatformInfo(key);

  elements.detectedPlatformBadge.classList.remove("hidden");
  elements.platformIcon.textContent = pInfo.icon;
  elements.platformName.textContent = pInfo.name;
  elements.detectedPlatformBadge.style.color = pInfo.color;
}

function detectPlatformKey(rawUrl) {
  try {
    const hostname = new URL(rawUrl).hostname.toLowerCase().replace(/^www\./, "");

    if (hostname.includes("youtu")) return "youtube";
    if (hostname.includes("bilibili") || hostname.includes("b23.tv")) return "bilibili";
    if (hostname.includes("ggjav")) return "ggjav";
    if (hostname.includes("missav")) return "missav";
    if (hostname.includes("rou.video")) return "rouvideo";
    if (hostname.includes("tiktok")) return "tiktok";
    if (hostname.includes("x.com") || hostname.includes("twitter")) return "x";
    if (hostname.includes("instagram") || hostname.includes("instagr.am")) return "instagram";

    return hostname || "generic";
  } catch (error) {
    return "generic";
  }
}

async function analyzeSingleVideo() {
  const rawUrl = elements.urlInput.value.trim();
  if (!rawUrl) return;

  const btn = elements.analyzeSubmitBtn;
  const originalText = btn.innerHTML;
  btn.disabled = true;
  btn.innerHTML = t("analyzing");

  elements.resolutionSelectorCard.classList.add("hidden");

  try {
    let metaData;
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.AnalyzeVideo) {
      metaData = await window.go.main.App.AnalyzeVideo(rawUrl);
    } else {
      metaData = {
        title: t("mock.title"),
        duration: "03:45",
        platform: detectPlatformKey(rawUrl),
        formats: [
          { formatId: "best", resolution: t("res.best"), ext: "mp4", note: t("res.best.note") },
          { formatId: "1080p", resolution: "1080P", ext: "mp4", note: "1080p" },
          { formatId: "720p", resolution: "720P", ext: "mp4", note: "720p" },
        ],
      };
    }

    state.currentMetaData = metaData;
    renderResolutionOptions(metaData);
  } catch (err) {
    alert(t("alert.analyze.fail", { err: err.message || err }));
  } finally {
    btn.disabled = false;
    btn.innerHTML = originalText;
  }
}

function renderResolutionOptions(metaData) {
  elements.resolutionSelectorCard.classList.remove("hidden");

  elements.resVideoTitle.textContent = metaData.title || t("res.title.unknown");
  elements.resVideoThumb.src = metaData.thumbnail || "";
  elements.resVideoThumb.onerror = () => {
    elements.resVideoThumb.src = "https://images.unsplash.com/photo-1574375927938-d5a98e8ffe85?w=200&auto=format&fit=crop&q=60";
  };

  const pKey = metaData.platform || detectPlatformKey(metaData.url || elements.urlInput.value.trim());
  const pInfo = getPlatformInfo(pKey);
  elements.resVideoMeta.textContent = t("res.meta", { platform: pInfo.name, duration: metaData.duration || t("res.duration.unknown") });

  const formats = metaData.formats && metaData.formats.length > 0
    ? metaData.formats
    : [{ formatId: "best", resolution: t("res.best"), ext: "mp4", note: t("res.best.note.short") }];

  elements.resChipsContainer.innerHTML = formats
    .map(
      (f, idx) => `
      <div class="res-chip ${idx === 0 ? "selected" : ""}" data-id="${f.formatId}">
        <div class="res-chip-title">${escapeHtml(f.resolution || f.formatId)}</div>
        <div class="res-chip-note">${escapeHtml(f.note || (f.fileSizeStr ? f.fileSizeStr : t("res.original")))}</div>
      </div>
    `
    )
    .join("");

  state.selectedFormatId = formats[0].formatId;

  document.querySelectorAll(".res-chip").forEach((chip) => {
    chip.addEventListener("click", () => {
      document.querySelectorAll(".res-chip").forEach((c) => c.classList.remove("selected"));
      chip.classList.add("selected");
      state.selectedFormatId = chip.dataset.id;
    });
  });
}

async function startDownloadAfterResolution() {
  const rawUrl = elements.urlInput.value.trim();
  if (!rawUrl) return;

  const btn = elements.resConfirmDownloadBtn;
  const originalText = btn.innerHTML;
  btn.disabled = true;
  btn.innerHTML = t("joining.queue");

  try {
    let task;
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SubmitDownloadWithQuality) {
      task = await window.go.main.App.SubmitDownloadWithQuality(rawUrl, state.outputDir, state.selectedFormatId);
    } else {
      task = {
        id: `mock_${Date.now()}`,
        url: rawUrl,
        platform: detectPlatformKey(rawUrl),
        status: "running",
        progress: 10,
        message: t("download.starting.multi"),
      };
    }

    state.tasksMap.set(task.id, task);
    elements.resolutionSelectorCard.classList.add("hidden");
    renderActiveDownloads();
    renderQueuePanel();
    startPolling();
  } catch (err) {
    alert(t("alert.download.fail", { err: err.message || err }));
  } finally {
    btn.disabled = false;
    btn.innerHTML = originalText;
  }
}

async function submitBatchTasks() {
  const rawText = elements.batchUrlInput.value.trim();
  if (!rawText) {
    alert(t("alert.batch.empty"));
    return;
  }

  try {
    let newTasks = [];
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SubmitBatchDownloads) {
      newTasks = await window.go.main.App.SubmitBatchDownloads(rawText, state.outputDir);
    } else {
      const urls = rawText.split("\n").filter((u) => u.trim());
      newTasks = urls.map((u, i) => ({
        id: `batch_mock_${Date.now()}_${i}`,
        url: u.trim(),
        platform: detectPlatformKey(u),
        status: i === 0 ? "running" : "queued",
        progress: i === 0 ? 5 : 0,
        message: i === 0 ? t("download.starting") : t("download.queued.msg"),
      }));
    }

    newTasks.forEach((task) => state.tasksMap.set(task.id, task));
    elements.batchUrlInput.value = "";
    renderActiveDownloads();
    renderQueuePanel();
    startPolling();
  } catch (err) {
    alert(t("alert.batch.fail", { err: err.message || err }));
  }
}

function startPolling() {
  stopPolling();
  state.pollHandle = setInterval(async () => {
    try {
      if (window.go && window.go.main && window.go.main.App && window.go.main.App.GetAllTasks) {
        const tasksList = await window.go.main.App.GetAllTasks();
        if (tasksList) {
          let hasActive = false;
          tasksList.forEach((t) => {
            const prev = state.tasksMap.get(t.id);
            state.tasksMap.set(t.id, t);
            if (t.status === "completed" && (!prev || prev.status !== "completed")) {
              addHistoryItem(t);
            }
            if (t.status === "running" || t.status === "downloading" || t.status === "merging" || t.status === "queued" || t.status === "pending") {
              hasActive = true;
            }
          });
          renderActiveDownloads();
          renderQueuePanel();
          if (!hasActive) stopPolling();
        }
      }
    } catch (e) {
      console.warn("Global poll error:", e);
    }
  }, 1000);
}

function stopPolling() {
  if (state.pollHandle) {
    clearInterval(state.pollHandle);
    state.pollHandle = null;
  }
}

// Global actions exposed to onclick handlers
window.pauseTaskFromQueue = async function (taskId) {
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.PauseTask) {
      await window.go.main.App.PauseTask(taskId);
    }
    const allTasks = await window.go.main.App.GetAllTasks();
    if (allTasks) {
      state.tasksMap.clear();
      allTasks.forEach((t) => state.tasksMap.set(t.id, t));
    }
    renderActiveDownloads();
    renderQueuePanel();
    startPolling();
  } catch (err) {
    console.warn("Pause task error:", err);
  }
};

window.resumeTaskFromQueue = async function (taskId) {
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.ResumeTask) {
      await window.go.main.App.ResumeTask(taskId);
    }
    const allTasks = await window.go.main.App.GetAllTasks();
    if (allTasks) {
      state.tasksMap.clear();
      allTasks.forEach((t) => state.tasksMap.set(t.id, t));
    }
    renderActiveDownloads();
    renderQueuePanel();
    startPolling();
  } catch (err) {
    console.warn("Resume task error:", err);
  }
};

window.deleteTaskFromQueue = async function (taskId) {
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.DeleteTask) {
      await window.go.main.App.DeleteTask(taskId);
    }
    state.tasksMap.delete(taskId);
    const allTasks = await window.go.main.App.GetAllTasks();
    if (allTasks) {
      state.tasksMap.clear();
      allTasks.forEach((t) => state.tasksMap.set(t.id, t));
    }
    renderActiveDownloads();
    renderQueuePanel();
    startPolling();
  } catch (err) {
    console.warn("Delete task error:", err);
  }
};

window.copyFilePath = async function (pathStr) {
  try {
    await navigator.clipboard.writeText(pathStr);
    alert(t("alert.copy.ok"));
  } catch (e) {
    console.warn("Copy error:", e);
  }
};

function renderActiveDownloads() {
  if (!elements.activeDownloadsSection || !elements.activeDownloadsList) return;

  const runningTasks = Array.from(state.tasksMap.values()).filter(
    (t) => t.status === "running" || t.status === "downloading" || t.status === "merging"
  );

  if (runningTasks.length === 0) {
    elements.activeDownloadsSection.classList.add("hidden");
    elements.activeDownloadsList.innerHTML = "";
    return;
  }

  elements.activeDownloadsSection.classList.remove("hidden");
  if (elements.activeRunningBadge) {
    elements.activeRunningBadge.textContent = t("active.badge", { n: runningTasks.length });
  }

  const runningIds = new Set(runningTasks.map((t) => t.id));

  // 1. Remove cards that are no longer running
  Array.from(elements.activeDownloadsList.children).forEach((child) => {
    const cid = child.id.replace("card-", "");
    if (!runningIds.has(cid)) {
      child.remove();
    }
  });

  // 2. Update existing cards or insert new ones (preserves DOM & stable hover state)
  runningTasks.forEach((task) => {
    const cardId = `card-${task.id}`;
    let card = document.getElementById(cardId);
    const pKey = task.platform || detectPlatformKey(task.url || "");
    const pInfo = getPlatformInfo(pKey);
    const progress = typeof task.progress === "number" ? task.progress : 0;
    const msg = task.message || t("active.msg.default");

    if (card) {
      const fillEl = card.querySelector(".progress-fill-large");
      if (fillEl) fillEl.style.width = `${progress}%`;
      const pctEl = card.querySelector(".pct-text");
      if (pctEl) pctEl.textContent = `${progress}%`;
      const msgEl = card.querySelector(".progress-msg-text");
      if (msgEl) msgEl.textContent = msg;
    } else {
      const div = document.createElement("div");
      div.className = "active-task-card";
      div.id = cardId;
      div.innerHTML = `
        <div class="active-task-header">
          <div class="active-task-title-group">
            <span class="active-task-icon">${pInfo.icon}</span>
            <div>
              <div class="active-task-title">${escapeHtml(task.url || "")}</div>
              <div style="font-size: 11px; color: var(--text-dim); font-family: var(--font-mono);">
                ${pInfo.name.toUpperCase()} · ID: ${task.id}
              </div>
            </div>
          </div>
          <div class="active-task-actions">
            <button type="button" class="button button-small button-action-pause" onclick="window.pauseTaskFromQueue('${task.id}')">${t("active.pause")}</button>
            <button type="button" class="button button-small button-action-delete" title="${t("active.delete.title")}" onclick="window.deleteTaskFromQueue('${task.id}')">${t("active.delete")}</button>
          </div>
        </div>

        <div class="task-progress-box" style="margin-top: 4px;">
          <div class="progress-track-large">
            <div class="progress-fill-large" style="width: ${progress}%;"></div>
          </div>
          <div class="progress-info-row">
            <span class="progress-msg-text" style="font-family: var(--font-mono); font-size: 12px; color: var(--text-muted);">${escapeHtml(msg)}</span>
            <span class="pct-text">${progress}%</span>
          </div>
        </div>
      `;
      elements.activeDownloadsList.appendChild(div);
    }
  });
}

function getTaskSortWeight(status) {
  // 1. 待下载 / 暂停排在最前 (TOP)
  if (status === "queued" || status === "pending" || status === "paused") return 1;
  // 2. 下载错误排在中间 (MIDDLE)
  if (status === "failed") return 2;
  // 3. 已经完成排在最后 (BOTTOM)
  if (status === "completed") return 3;
  return 4;
}

function renderQueuePanel() {
  if (!elements.queuePanel || !elements.queueTasksList) return;

  const allTasks = Array.from(state.tasksMap.values());
  // Non-running tasks for the queue panel
  const queueTasks = allTasks.filter(
    (t) => t.status !== "running" && t.status !== "downloading" && t.status !== "merging"
  );

  if (queueTasks.length === 0) {
    elements.queuePanel.classList.add("hidden");
    return;
  }

  elements.queuePanel.classList.remove("hidden");

  let waiting = 0;
  let failed = 0;
  let done = 0;

  queueTasks.forEach((t) => {
    if (t.status === "completed") done++;
    else if (t.status === "failed") failed++;
    else waiting++;
  });

  if (elements.queueCountBadge) elements.queueCountBadge.textContent = t("queue.badge", { n: queueTasks.length });
  if (elements.queueWaitingCount) elements.queueWaitingCount.textContent = t("queue.waiting", { n: waiting });
  if (elements.queueFailedCount) elements.queueFailedCount.textContent = t("queue.failed", { n: failed });
  if (elements.queueDoneCount) elements.queueDoneCount.textContent = t("queue.done", { n: done });

  // Strict sorting: 待下载 (1) -> 失败 (2) -> 已完成 (3)
  queueTasks.sort((a, b) => {
    const wa = getTaskSortWeight(a.status);
    const wb = getTaskSortWeight(b.status);
    if (wa !== wb) return wa - wb;
    return (b.createdAt || 0) - (a.createdAt || 0);
  });

  elements.queueTasksList.innerHTML = queueTasks
    .slice(0, 50)
    .map((task) => {
      const pKey = task.platform || detectPlatformKey(task.url || "");
      const pInfo = getPlatformInfo(pKey);
      const progress = typeof task.progress === "number" ? task.progress : 0;

      const isCompleted = task.status === "completed";
      const isFailed = task.status === "failed";
      const isPaused = task.status === "paused";
      const isQueued = task.status === "queued" || task.status === "pending";

      let statusBadgeHtml = "";
      if (isQueued) {
        statusBadgeHtml = `<span class="badge" style="background: rgba(99, 102, 241, 0.15); color: #818cf8;">${t("queue.status.queued")}</span>`;
      } else if (isPaused) {
        statusBadgeHtml = `<span class="badge" style="background: rgba(245, 158, 11, 0.15); color: #fbbf24;">${t("queue.status.paused")}</span>`;
      } else if (isFailed) {
        statusBadgeHtml = `<span class="badge" style="background: rgba(239, 68, 68, 0.15); color: #f87171;">${t("queue.status.failed")}</span>`;
      } else if (isCompleted) {
        statusBadgeHtml = `<span class="badge" style="background: rgba(16, 185, 129, 0.15); color: #34d399;">${t("queue.status.completed")}</span>`;
      }

      return `
        <div class="queue-item" style="${isFailed ? "border-color: rgba(239,68,68,0.35);" : isCompleted ? "border-color: rgba(16,185,129,0.25);" : ""}">
          <div class="queue-item-top">
            <div class="queue-item-url">${pInfo.icon} ${escapeHtml(task.url)}</div>
            <div class="queue-item-controls">
              ${statusBadgeHtml}
              ${isPaused || isFailed ? `<button type="button" class="queue-action-btn" style="color: #34d399;" onclick="window.resumeTaskFromQueue('${task.id}')">${t("queue.resume")}</button>` : ""}
              ${isCompleted && task.outputPath ? `<button type="button" class="queue-action-btn" onclick="window.copyFilePath('${escapePath(task.outputPath)}')">${t("queue.copy.path")}</button>` : ""}
              <button type="button" class="queue-action-btn queue-action-btn-danger" title="${t("active.delete.title")}" onclick="window.deleteTaskFromQueue('${task.id}')">🗑️</button>
            </div>
          </div>
          ${
            isFailed
              ? `<div style="font-size: 12px; color: #f87171; margin-top: 4px; font-family: var(--font-mono);">${t("queue.error.prefix")}${escapeHtml(task.message || t("download.error"))}</div>`
              : isCompleted
              ? `<div style="font-size: 11px; color: #34d399; margin-top: 4px; font-family: var(--font-mono);">${t("queue.file.prefix")}${escapeHtml(task.outputPath || t("queue.file.saved"))}</div>`
              : `<div style="font-size: 11px; color: ${isPaused ? "#fbbf24" : "var(--text-dim)"}; margin-top: 4px; font-family: var(--font-mono);">${escapeHtml(task.message || t("queue.waiting.msg"))}</div>`
          }
        </div>
      `;
    })
    .join("");
}

function escapePath(str) {
  if (!str) return "";
  return str.replace(/\\/g, "\\\\").replace(/'/g, "\\'");
}

function addHistoryItem(task) {
  const fullPath = task.outputPath || `${state.outputDir}\\${task.id}.mp4`;
  const fileName = extractFileName(fullPath);

  const item = {
    id: task.id,
    url: task.url,
    platform: task.platform,
    fileName: fileName,
    outputPath: fullPath,
    time: new Date().toLocaleTimeString(),
  };

  state.history.unshift(item);
  if (state.history.length > 30) state.history.pop();
  localStorage.setItem("download_history", JSON.stringify(state.history));
  renderHistory();
}

function extractFileName(fullPath) {
  if (!fullPath) return t("history.default.filename");
  const parts = fullPath.split(/[/\\]/);
  return parts[parts.length - 1] || fullPath;
}

function renderHistory() {
  if (!elements.historyList) return;
  if (state.history.length === 0) {
    elements.historySection.classList.add("hidden");
    return;
  }

  elements.historySection.classList.remove("hidden");
  elements.historyList.innerHTML = state.history
    .map(
      (item) => `
      <div class="history-item">
        <div class="history-title-row">
          <div class="history-title">${escapeHtml(item.fileName)}</div>
          <div class="history-time">${escapeHtml(item.time)}</div>
        </div>
        <div class="history-path">${escapeHtml(item.outputPath)}</div>
      </div>
    `
    )
    .join("");
}

function escapeHtml(str) {
  if (!str) return "";
  return String(str)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}
