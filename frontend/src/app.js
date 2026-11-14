const PLATFORM_MAP = {
  youtube: { name: "YouTube", icon: "🔴", color: "#ef4444" },
  bilibili: { name: "Bilibili", icon: "🔵", color: "#3b82f6" },
  missav: { name: "MissAV", icon: "🌸", color: "#ec4899" },
  ggjav: { name: "GGJav", icon: "🟣", color: "#a855f7" },
  supjav: { name: "SupJav", icon: "⚡", color: "#f59e0b" },
  rouvideo: { name: "RouVideo (肉视频)", icon: "🥩", color: "#f472b6" },
  tiktok: { name: "TikTok / 抖音", icon: "🎵", color: "#06b6d4" },
  x: { name: "X (Twitter)", icon: "🖤", color: "#94a3b8" },
  instagram: { name: "Instagram", icon: "📷", color: "#e1306c" },
  tver: { name: "TVer", icon: "🇯🇵", color: "#38bdf8" },
  m3u8: { name: "HLS / M3U8 直链", icon: "⚡", color: "#f59e0b" },
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
  cookieConfig: { mode: "none", browser: "", file: "" },
  tasksMap: new Map(),
  history: JSON.parse(localStorage.getItem("download_history") || "[]"),
  currentMetaData: null,
  selectedFormatId: "best",
};

const taskSnapshots = new TaskSnapshots(state.tasksMap);

async function refreshTaskSnapshot() {
  const api = window.go?.main?.App;
  if (!api?.GetAllTasks) return false;
  const token = taskSnapshots.request();
  const tasks = await api.GetAllTasks();
  const previous = new Map(state.tasksMap);
  if (!taskSnapshots.apply(token, tasks)) return false;
  for (const task of tasks || []) {
    if (task.status === 'completed' && previous.get(task.id)?.status !== 'completed') addHistoryItem(task);
  }
  renderActiveDownloads();
  renderQueuePanel();
  return true;
}

const elements = {
  tabSingleMode: document.getElementById("tab-single-mode"),
  tabBatchMode: document.getElementById("tab-batch-mode"),
  modeHintText: document.getElementById("mode-hint-text"),
  singleUrlBox: document.getElementById("single-url-box"),
  batchUrlBox: document.getElementById("batch-url-box"),
  batchUrlInput: document.getElementById("batch-url-input"),
  batchPasteBtn: document.getElementById("batch-paste-btn"),
  batchClearBtn: document.getElementById("batch-clear-btn"),
  batchQualitySelect: document.getElementById("batch-quality-select"),

  singleOutputDirInput: document.getElementById("single-output-dir-input"),
  singleBrowseBtn: document.getElementById("single-browse-btn"),
  cookieSettingsBar: document.getElementById("cookie-settings-bar"),
  cookieStatusBadge: document.getElementById("cookie-status-badge"),
  cookieHelpBtn: document.getElementById("cookie-help-btn"),
  cookieClearBtn: document.getElementById("cookie-clear-btn"),
  cookieFileBrowseBtn: document.getElementById("cookie-file-browse-btn"),

  // YouTube Cookie Guide Modal
  youtubeCookieModal: document.getElementById("youtube-cookie-modal"),
  btnCloseYtCookieModal: document.getElementById("btn-close-yt-cookie-modal"),
  btnYtOpenExt: document.getElementById("btn-yt-open-ext"),
  btnYtImportCookie: document.getElementById("btn-yt-import-cookie"),
  btnYtCookieDone: document.getElementById("btn-yt-cookie-done"),

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

  // Supported Platforms Modal & Showcase
  btnOpenPlatformsModal: document.getElementById("btn-open-platforms-modal"),
  platformsModal: document.getElementById("platforms-modal"),
  btnClosePlatformsModal: document.getElementById("btn-close-platforms-modal"),
  platformsSearchInput: document.getElementById("platforms-search-input"),
  btnClearPlatformsSearch: document.getElementById("btn-clear-platforms-search"),
  platformsSearchCount: document.getElementById("platforms-search-count"),
  platformsCategoryChips: document.getElementById("platforms-category-chips"),
  modalPlatformsList: document.getElementById("modal-platforms-list"),

  // SupJav Tutorial & Verification Modal
  supjavGuideBtn: document.getElementById("supjav-guide-btn"),
  supjavTutorialModal: document.getElementById("supjav-tutorial-modal"),
  btnCloseSupjavModal: document.getElementById("btn-close-supjav-modal"),
  btnSupjavOpenWindow: document.getElementById("btn-supjav-open-window"),
  btnSupjavReadClipboard: document.getElementById("btn-supjav-read-clipboard"),
  supjavClipboardStatus: document.getElementById("supjav-clipboard-status"),
  supjavDirectStreamInput: document.getElementById("supjav-direct-stream-input"),
  btnSupjavDirectDownload: document.getElementById("btn-supjav-direct-download"),
  btnSupjavOpenBrowser: document.getElementById("btn-supjav-open-browser"),
  btnSupjavDone: document.getElementById("btn-supjav-done"),
};

let _lastVersionInfo = null;

document.addEventListener("DOMContentLoaded", async () => {
  if (window.i18n && window.i18n.initI18n) {
    await window.i18n.initI18n();
  }
  await initWailsApp();
  bindEvents();
  initPlatformsModal();
  initSupJavModal();
  initYouTubeCookieModal();
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
  const verBadge = document.querySelector(".version-badge");
  const verYtdlp = document.getElementById("ver-ytdlp");
  const verFfmpeg = document.getElementById("ver-ffmpeg");

  if (verInfo.appVersion) {
    const formatted = "v" + verInfo.appVersion;
    if (verApp) verApp.textContent = formatted;
    if (verBadge) verBadge.textContent = formatted;
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

      if (window.go.main.App.GetCookieConfig) {
        const cookieCfg = await window.go.main.App.GetCookieConfig();
        syncCookieUI(cookieCfg);
      }

      await refreshTaskSnapshot();
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
        if (!taskSnapshots.progress(task)) return;

        if (task.status === "completed" && (!prev || prev.status !== "completed")) {
          addHistoryItem(task);
        }

        renderActiveDownloads();
        renderQueuePanel();
      }
    });

    window.runtime.EventsOn("task:deleted", (taskId) => {
      taskSnapshots.remove(taskId);
      renderActiveDownloads();
      renderQueuePanel();
    });

    window.runtime.EventsOn("tasks:cleared", () => {
      taskSnapshots.clear();
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
      syncCookieUI(state.cookieConfig);
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
      if (elements.platformsModal && !elements.platformsModal.classList.contains("hidden")) {
        renderPlatformsModalList(elements.platformsSearchInput?.value || "", currentPlatformCategory);
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

  elements.cookieHelpBtn?.addEventListener("click", () => {
    openYouTubeCookieModal();
  });

  elements.cookieClearBtn?.addEventListener("click", async () => {
    await clearCookieFileConfig();
  });

  elements.cookieFileBrowseBtn?.addEventListener("click", async () => {
    await browseAndApplyCookiesFile();
  });

  elements.resCloseBtn?.addEventListener("click", () => {
    elements.resolutionSelectorCard.classList.add("hidden");
  });

  elements.resConfirmDownloadBtn?.addEventListener("click", () => {
    startDownloadAfterResolution();
  });

  elements.clearQueueBtn?.addEventListener("click", async () => {
    if (elements.clearQueueBtn.disabled) return;
    elements.clearQueueBtn.disabled = true;
    stopPolling();
    taskSnapshots.invalidate();
    try {
      const appApi = window.go?.main?.App;
      if (appApi?.ClearAllTasksAndCaches) {
        await appApi.ClearAllTasksAndCaches();
      } else if (appApi?.ClearTasks) {
        await appApi.ClearTasks();
      } else {
        throw new Error("Clear task API is unavailable");
      }
      taskSnapshots.clear();
      renderActiveDownloads();
      renderQueuePanel();
    } catch (err) {
      console.warn("Clear tasks and caches failed:", err);
      window.alert(`清空队列与缓存失败：${err?.message || err}`);
    } finally {
      taskSnapshots.invalidate();
      elements.clearQueueBtn.disabled = false;
      startPolling();
    }
  });

  elements.clearHistoryBtn?.addEventListener("click", () => {
    state.history = [];
    localStorage.removeItem("download_history");
    renderHistory();
  });
}

let currentPlatformCategory = "all";
let searchDebounceTimer = null;

function initPlatformsModal() {
  if (!elements.btnOpenPlatformsModal || !elements.platformsModal) return;

  function openModal(initialSearch = "", initialCategory = "") {
    if (initialCategory) {
      currentPlatformCategory = initialCategory;
      if (elements.platformsCategoryChips) {
        elements.platformsCategoryChips.querySelectorAll(".chip-btn").forEach((btn) => {
          btn.classList.toggle("active", btn.getAttribute("data-cat") === initialCategory);
        });
      }
    }
    if (elements.platformsSearchInput) {
      elements.platformsSearchInput.value = initialSearch;
      if (elements.btnClearPlatformsSearch) {
        elements.btnClearPlatformsSearch.classList.toggle("hidden", !initialSearch);
      }
    }
    elements.platformsModal.classList.remove("hidden");
    renderPlatformsModalList(initialSearch, currentPlatformCategory);
    setTimeout(() => {
      elements.platformsSearchInput?.focus();
    }, 50);
  }

  function closeModal() {
    elements.platformsModal.classList.add("hidden");
  }

  elements.btnOpenPlatformsModal.addEventListener("click", () => {
    openModal("", "all");
  });

  elements.btnClosePlatformsModal?.addEventListener("click", closeModal);

  elements.platformsModal.addEventListener("click", (e) => {
    if (e.target === elements.platformsModal) {
      closeModal();
    }
  });

  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && !elements.platformsModal.classList.contains("hidden")) {
      closeModal();
    }
  });

  // Showcase pills click event
  document.querySelectorAll(".showcase-pill").forEach((pill) => {
    pill.addEventListener("click", () => {
      const search = pill.getAttribute("data-search") || "";
      const cat = pill.getAttribute("data-cat") || "all";
      openModal(search, cat);
    });
  });

  // Category filter chips
  elements.platformsCategoryChips?.addEventListener("click", (e) => {
    const chip = e.target.closest(".chip-btn");
    if (!chip) return;
    const cat = chip.getAttribute("data-cat") || "all";
    currentPlatformCategory = cat;
    elements.platformsCategoryChips.querySelectorAll(".chip-btn").forEach((b) => b.classList.remove("active"));
    chip.classList.add("active");
    renderPlatformsModalList(elements.platformsSearchInput?.value || "", currentPlatformCategory);
  });

  // Search input with debounce
  elements.platformsSearchInput?.addEventListener("input", (e) => {
    const val = e.target.value.trim();
    if (elements.btnClearPlatformsSearch) {
      elements.btnClearPlatformsSearch.classList.toggle("hidden", !val);
    }
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = setTimeout(() => {
      renderPlatformsModalList(val, currentPlatformCategory);
    }, 100);
  });

  // Clear search button
  elements.btnClearPlatformsSearch?.addEventListener("click", () => {
    if (elements.platformsSearchInput) {
      elements.platformsSearchInput.value = "";
      elements.platformsSearchInput.focus();
    }
    elements.btnClearPlatformsSearch.classList.add("hidden");
    renderPlatformsModalList("", currentPlatformCategory);
  });

  // Delegated click on site cards
  elements.modalPlatformsList?.addEventListener("click", (e) => {
    const card = e.target.closest(".modal-site-card");
    if (!card) return;
    const domain = card.getAttribute("data-domain") || "";
    if (domain) {
      const primaryDomain = domain.split(",")[0].trim();
      const sampleUrl = "https://" + primaryDomain;
      elements.urlInput.value = sampleUrl;
      updatePlatformBadge(sampleUrl);
      closeModal();
      elements.urlInput.focus();
      elements.urlInput.select();
    }
  });
}

function renderPlatformsModalList(query = "", category = "all") {
  if (!elements.modalPlatformsList) return;

  const dataset = (window.SUPPORTED_PLATFORMS || globalThis.SUPPORTED_PLATFORMS || []);
  const cleanQuery = query.toLowerCase().trim();
  const locale = window.i18n?.getCurrentLocale?.() || "zh-CN";
  const locTag = (tg) => (globalThis.localizePlatformTag ? globalThis.localizePlatformTag(tg, locale) : tg);
  const locName = (n) => (globalThis.localizePlatformName ? globalThis.localizePlatformName(n, locale) : n);

  let filtered = dataset.filter((p) => {
    if (category !== "all" && p.category !== category) return false;
    if (!cleanQuery) return true;
    if (p.name && (p.name.toLowerCase().includes(cleanQuery) || locName(p.name).toLowerCase().includes(cleanQuery))) return true;
    if (p.domain && p.domain.toLowerCase().includes(cleanQuery)) return true;
    if (p.id && p.id.toLowerCase().includes(cleanQuery)) return true;
    if (p.tags && Array.isArray(p.tags)) {
      return p.tags.some((tg) => tg.toLowerCase().includes(cleanQuery) || locTag(tg).toLowerCase().includes(cleanQuery));
    }
    return false;
  });

  // Update count badge
  if (elements.platformsSearchCount) {
    if (!cleanQuery && category === "all") {
      elements.platformsSearchCount.textContent = t("modal.count.total", { n: dataset.length });
    } else {
      elements.platformsSearchCount.textContent = t("modal.count.found", { n: filtered.length });
    }
  }

  if (filtered.length === 0) {
    elements.modalPlatformsList.innerHTML = `
      <div class="modal-empty-state">
        <div class="empty-icon">🔎</div>
        <div class="empty-title">${t("modal.empty.title")}</div>
        <div class="empty-desc">${t("modal.empty")}</div>
      </div>
    `;
    return;
  }

  // Render cards (render first 150 items to keep UI snappy)
  const MAX_RENDER = 150;
  const slice = filtered.slice(0, MAX_RENDER);
  const cardsHtml = slice.map((p) => {
    const name = locName(p.name);
    const tagsHtml = (p.tags && Array.isArray(p.tags))
      ? p.tags.map((tg) => `<span class="site-tag">${escapeHtml(locTag(tg))}</span>`).join("")
      : "";
    return `
      <div class="modal-site-card" data-domain="${escapeHtml(p.domain)}" title="${escapeHtml(t("modal.card.fill", { name }))}">
        <div class="site-card-top">
          <span class="site-card-icon">${p.icon || "🌐"}</span>
          <span class="site-card-name">${escapeHtml(name)}</span>
        </div>
        <div class="site-card-domain">${escapeHtml(p.domain)}</div>
        ${tagsHtml ? `<div class="site-card-tags">${tagsHtml}</div>` : ""}
      </div>
    `;
  }).join("");

  const moreNotice = filtered.length > MAX_RENDER
    ? `<div style="grid-column: 1 / -1; text-align: center; color: var(--text-muted); font-size: 12px; padding: 12px 0;">${t("modal.more", { max: MAX_RENDER, n: filtered.length })}</div>`
    : "";

  elements.modalPlatformsList.innerHTML = cardsHtml + moreNotice;
}

let currentSupJavUrl = "";

function openSupJavTutorialModal(url = "") {
  currentSupJavUrl = url || elements.urlInput?.value?.trim() || "";
  if (elements.supjavDirectStreamInput) {
    elements.supjavDirectStreamInput.value = "";
  }
  if (elements.supjavClipboardStatus) {
    elements.supjavClipboardStatus.textContent = "";
    elements.supjavClipboardStatus.classList.add("hidden");
  }
  elements.supjavTutorialModal?.classList.remove("hidden");
}

function closeSupJavTutorialModal() {
  elements.supjavTutorialModal?.classList.add("hidden");
}

function decorateSupJavStreamUrl(streamUrl) {
  if (!streamUrl || !streamUrl.toLowerCase().includes(".m3u8")) return streamUrl;
  if (streamUrl.includes("omni_referer=") || streamUrl.includes("omni_title=")) return streamUrl;
  const hints = [];
  if (currentSupJavUrl && currentSupJavUrl.includes("supjav.com")) {
    hints.push(`omni_referer=${encodeURIComponent(currentSupJavUrl)}`);
    const m = currentSupJavUrl.match(/\/([a-zA-Z0-9_-]+)\.html/i);
    if (m && m[1]) {
      hints.push(`omni_title=${encodeURIComponent("SupJav_" + m[1])}`);
    }
  }
  if (hints.length === 0) return streamUrl;
  const sep = streamUrl.includes("#") ? "&" : "#";
  return `${streamUrl}${sep}${hints.join("&")}`;
}

async function resolveStreamUrlFromInput(rawText) {
  const trimmed = rawText ? rawText.trim() : "";
  if (!trimmed) return "";

  let links = [];
  if (window.go?.main?.App?.ExtractSupJavLinksFromText) {
    try {
      links = await window.go.main.App.ExtractSupJavLinksFromText(trimmed);
    } catch (e) {
      console.warn("ExtractSupJavLinksFromText error:", e);
    }
  }
  if (!links || links.length === 0) {
    const m3u8Matches = trimmed.match(/https?:\/\/[^\s"'<>]+?\.m3u8[^\s"'<>]*/gi);
    if (m3u8Matches && m3u8Matches.length > 0) {
      links = m3u8Matches;
    } else {
      const embedMatches = trimmed.match(/https?:\/\/[^\s"'<>]*(?:faststream\.to|fststream\.to|tvlogy\.to|streamtape\.com|streamwish\.to|voe\.sx)[^\s"'<>]*/gi);
      if (embedMatches) links = embedMatches;
    }
  }
  if (links && links.length > 0) {
    return decorateSupJavStreamUrl(links[0]);
  }
  if (/^https?:\/\//i.test(trimmed)) {
    return decorateSupJavStreamUrl(trimmed);
  }
  return "";
}

function initSupJavModal() {
  if (!elements.supjavTutorialModal) return;

  elements.supjavGuideBtn?.addEventListener("click", () => {
    openSupJavTutorialModal(elements.urlInput?.value?.trim());
  });

  elements.btnCloseSupjavModal?.addEventListener("click", closeSupJavTutorialModal);
  elements.btnSupjavDone?.addEventListener("click", closeSupJavTutorialModal);

  elements.supjavTutorialModal.addEventListener("click", (e) => {
    if (e.target === elements.supjavTutorialModal) {
      closeSupJavTutorialModal();
    }
  });

  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && !elements.supjavTutorialModal.classList.contains("hidden")) {
      closeSupJavTutorialModal();
    }
  });

  // Action 1: Open verification window
  elements.btnSupjavOpenWindow?.addEventListener("click", async () => {
    const targetUrl = currentSupJavUrl || elements.urlInput?.value?.trim() || "https://supjav.com/";
    try {
      if (window.go?.main?.App?.OpenSupJavVerificationWindow) {
        await window.go.main.App.OpenSupJavVerificationWindow(targetUrl);
      } else {
        window.open(targetUrl, "_blank");
      }
      if (elements.supjavClipboardStatus) {
        elements.supjavClipboardStatus.textContent = t("supjav.window.opened");
        elements.supjavClipboardStatus.classList.remove("hidden");
      }
    } catch (err) {
      console.warn("Open verification window failed:", err);
      window.open(targetUrl, "_blank");
    }
  });

  // Action 2: Read clipboard & download
  elements.btnSupjavReadClipboard?.addEventListener("click", async () => {
    try {
      const clipText = await navigator.clipboard.readText();
      const trimmed = clipText ? clipText.trim() : "";
      if (!trimmed) {
        if (elements.supjavClipboardStatus) {
          elements.supjavClipboardStatus.textContent = t("supjav.clip.empty");
          elements.supjavClipboardStatus.classList.remove("hidden");
        }
        return;
      }

      // Check if it's HTML content
      if (trimmed.includes("<") && trimmed.includes(">") && (trimmed.includes("btn-server") || trimmed.includes("entry-title") || trimmed.includes("faststream") || trimmed.includes("tvlogy"))) {
        if (window.go?.main?.App?.AnalyzeSupJavHTML) {
          try {
            const metaData = await window.go.main.App.AnalyzeSupJavHTML(trimmed, currentSupJavUrl);
            if (metaData && metaData.formats && metaData.formats.length > 0) {
              closeSupJavTutorialModal();
              state.currentMetaData = metaData;
              renderResolutionOptions(metaData);
              return;
            }
          } catch (e) {
            console.warn("HTML parse error:", e);
          }
        }
      }

      const streamUrl = await resolveStreamUrlFromInput(trimmed);
      if (streamUrl && (streamUrl.toLowerCase().includes(".m3u8") || /faststream|fststream|tvlogy|streamtape|streamwish|voe\.sx/i.test(streamUrl))) {
        if (elements.supjavClipboardStatus) {
          elements.supjavClipboardStatus.textContent = t("supjav.clip.success");
          elements.supjavClipboardStatus.classList.remove("hidden");
        }
        setTimeout(() => {
          closeSupJavTutorialModal();
          elements.urlInput.value = streamUrl;
          updatePlatformBadge(streamUrl);
          analyzeSingleVideo();
        }, 400);
      } else {
        if (elements.supjavClipboardStatus) {
          elements.supjavClipboardStatus.textContent = t("supjav.clip.empty");
          elements.supjavClipboardStatus.classList.remove("hidden");
        }
      }
    } catch (err) {
      console.warn("Clipboard access error:", err);
      if (elements.supjavClipboardStatus) {
        elements.supjavClipboardStatus.textContent = `读取剪贴板失败: ${err.message || err}`;
        elements.supjavClipboardStatus.classList.remove("hidden");
      }
    }
  });

  // Action 3: Direct stream download
  elements.btnSupjavDirectDownload?.addEventListener("click", async () => {
    const val = elements.supjavDirectStreamInput?.value?.trim();
    if (!val) return;
    const resolved = (await resolveStreamUrlFromInput(val)) || val;
    closeSupJavTutorialModal();
    elements.urlInput.value = resolved;
    updatePlatformBadge(resolved);
    analyzeSingleVideo();
  });

  // Action 4: Open in default browser
  elements.btnSupjavOpenBrowser?.addEventListener("click", async () => {
    const targetUrl = currentSupJavUrl || elements.urlInput?.value?.trim() || "https://supjav.com/";
    if (window.go?.main?.App?.OpenURLInBrowser) {
      await window.go.main.App.OpenURLInBrowser(targetUrl);
    } else {
      window.open(targetUrl, "_blank");
    }
    if (elements.supjavClipboardStatus) {
      elements.supjavClipboardStatus.textContent = t("supjav.window.opened");
      elements.supjavClipboardStatus.classList.remove("hidden");
    }
  });
}

function updatePlatformBadge(rawUrl) {
  if (!rawUrl) {
    elements.detectedPlatformBadge.classList.add("hidden");
    elements.supjavGuideBtn?.classList.add("hidden");
    return;
  }
  const key = detectPlatformKey(rawUrl);
  const pInfo = getPlatformInfo(key);

  elements.detectedPlatformBadge.classList.remove("hidden");
  elements.platformIcon.textContent = pInfo.icon;
  elements.platformName.textContent = pInfo.name;
  elements.detectedPlatformBadge.style.color = pInfo.color;

  if (key === "supjav") {
    elements.supjavGuideBtn?.classList.remove("hidden");
  } else {
    elements.supjavGuideBtn?.classList.add("hidden");
  }
}

function detectPlatformKey(rawUrl) {
  try {
    const u = new URL(rawUrl);
    const hostname = u.hostname.toLowerCase().replace(/^www\./, "");
    if (u.pathname.toLowerCase().includes(".m3u8") || u.search.toLowerCase().includes(".m3u8")) {
      return "m3u8";
    }

    if (hostname.includes("youtu")) return "youtube";
    if (hostname.includes("bilibili") || hostname.includes("b23.tv")) return "bilibili";
    if (hostname.includes("ggjav")) return "ggjav";
    if (hostname.includes("supjav")) return "supjav";
    if (hostname.includes("missav")) return "missav";
    if (hostname.includes("rou.video")) return "rouvideo";
    if (hostname.includes("tiktok")) return "tiktok";
    if (hostname.includes("tver.jp")) return "tver";
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
    const errMsg = String(err?.message || err);
    if (rawUrl.includes("supjav.com") || errMsg.includes("CLOUDFLARE_CHALLENGE_DETECTED") || errMsg.includes("Cloudflare")) {
      openSupJavTutorialModal(rawUrl);
      return;
    }
    if (errMsg.includes("YOUTUBE_BOT_CHECK") || errMsg.includes("COOKIE_DB_LOCKED")) {
      highlightCookieSettingsBar();
      openYouTubeCookieModal();
      return;
    }
    alert(t("alert.analyze.fail", { err: errMsg }));
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

    taskSnapshots.invalidate();
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

  const batchQuality = elements.batchQualitySelect ? elements.batchQualitySelect.value : "best";

  try {
    let newTasks = [];
    if (window.go && window.go.main && window.go.main.App && window.go.main.App.SubmitBatchDownloadsWithQuality) {
      newTasks = await window.go.main.App.SubmitBatchDownloadsWithQuality(rawText, state.outputDir, batchQuality);
    } else if (window.go && window.go.main && window.go.main.App && window.go.main.App.SubmitBatchDownloads) {
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

    taskSnapshots.invalidate();
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
        if (await refreshTaskSnapshot()) {
          const hasActive = [...state.tasksMap.values()].some(task => ['running', 'downloading', 'merging', 'queued', 'pending'].includes(task.status));
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
    taskSnapshots.invalidate();
    await refreshTaskSnapshot();
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
    taskSnapshots.invalidate();
    await refreshTaskSnapshot();
    renderActiveDownloads();
    renderQueuePanel();
    startPolling();
  } catch (err) {
    console.warn("Resume task error:", err);
  }
};

function confirmDeleteTask(taskId) {
  return new Promise((resolve) => {
    const modal = document.getElementById("confirm-delete-modal");
    if (!modal) {
      const ok = window.confirm(t("confirm.delete.prompt"));
      resolve(ok);
      return;
    }
    const btnConfirm = document.getElementById("btn-confirm-delete");
    const btnCancel = document.getElementById("btn-cancel-delete");
    const btnClose = document.getElementById("btn-close-confirm-modal");

    function cleanup() {
      modal.classList.add("hidden");
      btnConfirm?.removeEventListener("click", onConfirm);
      btnCancel?.removeEventListener("click", onCancel);
      btnClose?.removeEventListener("click", onCancel);
      modal.removeEventListener("click", onBackdrop);
      document.removeEventListener("keydown", onKeyDown);
    }
    function onConfirm() {
      cleanup();
      resolve(true);
    }
    function onCancel() {
      cleanup();
      resolve(false);
    }
    function onBackdrop(e) {
      if (e.target === modal) {
        cleanup();
        resolve(false);
      }
    }
    function onKeyDown(e) {
      if (e.key === "Escape") {
        cleanup();
        resolve(false);
      }
    }

    btnConfirm?.addEventListener("click", onConfirm);
    btnCancel?.addEventListener("click", onCancel);
    btnClose?.addEventListener("click", onCancel);
    modal.addEventListener("click", onBackdrop);
    document.addEventListener("keydown", onKeyDown);
    modal.classList.remove("hidden");
  });
}

window.deleteTaskFromQueue = async function (taskId) {
  try {
    const confirmed = await confirmDeleteTask(taskId);
    if (!confirmed) return;

    if (window.go && window.go.main && window.go.main.App && window.go.main.App.DeleteTask) {
      await window.go.main.App.DeleteTask(taskId);
    }
    taskSnapshots.remove(taskId);
    await refreshTaskSnapshot();
    renderActiveDownloads();
    renderQueuePanel();
    startPolling();
  } catch (err) {
    console.warn("Delete task error:", err);
    window.alert(err?.message || String(err));
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

function parseTaskTimestamp(ts) {
  if (typeof ts === "number" && Number.isFinite(ts)) return ts;
  if (!ts) return 0;
  const parsed = Date.parse(ts);
  return Number.isNaN(parsed) ? 0 : parsed;
}

function compareTasksDeterministic(a, b) {
  const ta = parseTaskTimestamp(a.createdAt);
  const tb = parseTaskTimestamp(b.createdAt);
  if (tb !== ta) return tb - ta;
  return String(a.id || "").localeCompare(String(b.id || ""));
}

function renderActiveDownloads() {
  if (!elements.activeDownloadsSection || !elements.activeDownloadsList) return;

  const runningTasks = Array.from(state.tasksMap.values()).filter(
    (t) => t.status === "running" || t.status === "downloading" || t.status === "merging"
  );
  runningTasks.sort(compareTasksDeterministic);

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

  // Strict sorting: 待下载 (1) -> 失败 (2) -> 已完成 (3), with stable timestamp + ID tie-breaker
  queueTasks.sort((a, b) => {
    const wa = getTaskSortWeight(a.status);
    const wb = getTaskSortWeight(b.status);
    if (wa !== wb) return wa - wb;
    return compareTasksDeterministic(a, b);
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
  if (!elements.historyList || !elements.historySection) return;
  if (!state.history || state.history.length === 0) {
    elements.historySection.classList.add("hidden");
    return;
  }

  elements.historySection.classList.remove("hidden");
  const countBadge = document.getElementById("history-count-badge");
  if (countBadge) {
    countBadge.textContent = state.history.length;
  }

  elements.historyList.innerHTML = state.history
    .slice(0, 50)
    .map(
      (item) => `
      <div class="history-item">
        <div class="history-item-main">
          <div class="history-title-row">
            <span class="history-item-icon">🎬</span>
            <div class="history-title" title="${escapeHtml(item.fileName || item.outputPath || '')}">${escapeHtml(item.fileName)}</div>
            <div class="history-time">${escapeHtml(item.time || '')}</div>
          </div>
          <div class="history-path" title="${escapeHtml(item.outputPath || '')}">${escapeHtml(item.outputPath || '')}</div>
        </div>
        <div class="history-item-actions">
          <button type="button" class="history-action-btn" title="${t("queue.copy.path")}" onclick="window.copyFilePath('${escapePath(item.outputPath || '')}')">📋</button>
        </div>
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

function syncCookieUI(cfg) {
  const normalized = cfg || { mode: "none", browser: "", file: "" };
  state.cookieConfig = normalized;

  const isFile = normalized.mode === "file" && Boolean(normalized.file);
  if (elements.cookieStatusBadge) {
    if (isFile) {
      const baseName = normalized.file.split(/[/\\]/).pop() || normalized.file;
      elements.cookieStatusBadge.textContent = t("cookie.status.loaded", { file: baseName });
      elements.cookieStatusBadge.title = normalized.file;
      elements.cookieStatusBadge.classList.add("cookie-loaded");
    } else {
      elements.cookieStatusBadge.textContent = t("cookie.status.none");
      elements.cookieStatusBadge.title = "";
      elements.cookieStatusBadge.classList.remove("cookie-loaded");
    }
  }

  if (elements.cookieClearBtn) {
    elements.cookieClearBtn.classList.toggle("hidden", !isFile);
  }
}

async function clearCookieFileConfig() {
  const api = window.go?.main?.App;
  if (!api?.SetCookieConfig) return;
  try {
    const saved = await api.SetCookieConfig("none", "", "");
    syncCookieUI(saved);
  } catch (err) {
    alert(t("alert.cookie.fail", { err: err?.message || err }));
  }
}

async function browseAndApplyCookiesFile() {
  const api = window.go?.main?.App;
  if (!api?.SelectCookiesFile || !api?.SetCookieConfig) return false;

  try {
    const filePath = await api.SelectCookiesFile();
    if (!filePath) {
      syncCookieUI(state.cookieConfig);
      return false;
    }
    const saved = await api.SetCookieConfig("file", "", filePath);
    syncCookieUI(saved);
    return true;
  } catch (err) {
    syncCookieUI(state.cookieConfig);
    alert(t("alert.cookie.fail", { err: err?.message || err }));
    return false;
  }
}

function highlightCookieSettingsBar() {
  if (!elements.cookieSettingsBar) return;
  elements.cookieSettingsBar.classList.remove("cookie-highlight");
  void elements.cookieSettingsBar.offsetWidth;
  elements.cookieSettingsBar.classList.add("cookie-highlight");
  elements.cookieFileBrowseBtn?.focus();
}

function openYouTubeCookieModal() {
  if (elements.youtubeCookieModal) {
    elements.youtubeCookieModal.classList.remove("hidden");
  }
}

function closeYouTubeCookieModal() {
  if (elements.youtubeCookieModal) {
    elements.youtubeCookieModal.classList.add("hidden");
  }
}

function initYouTubeCookieModal() {
  const extStoreUrl = "https://chromewebstore.google.com/detail/get-cookiestxt-locally/cclelndahbckbenkjhflpdbgdldlbecc";

  elements.btnCloseYtCookieModal?.addEventListener("click", closeYouTubeCookieModal);
  elements.btnYtCookieDone?.addEventListener("click", closeYouTubeCookieModal);
  elements.youtubeCookieModal?.addEventListener("click", (e) => {
    if (e.target === elements.youtubeCookieModal) {
      closeYouTubeCookieModal();
    }
  });

  elements.btnYtOpenExt?.addEventListener("click", () => {
    if (window.go?.main?.App?.OpenURLInBrowser) {
      window.go.main.App.OpenURLInBrowser(extStoreUrl);
    } else {
      window.open(extStoreUrl, "_blank");
    }
  });

  elements.btnYtImportCookie?.addEventListener("click", async () => {
    const imported = await browseAndApplyCookiesFile();
    if (imported) {
      closeYouTubeCookieModal();
    }
  });
}
