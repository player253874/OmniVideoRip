package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"videodownloader/internal/config"
	"videodownloader/internal/executor"
	"videodownloader/internal/extractors"
	"videodownloader/internal/sidecars"
	"videodownloader/internal/tasks"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppConfig struct {
	DownloadDir string `json:"downloadDir"`
}

type VersionInfo struct {
	AppVersion    string `json:"appVersion"`
	YTDLPVersion  string `json:"ytdlpVersion"`
	FFmpegVersion string `json:"ffmpegVersion"`
}

type App struct {
	taskMu       sync.Mutex
	ctx          context.Context
	logger       *log.Logger
	paths        config.Paths
	taskStore    *tasks.Store
	taskExecutor *executor.Executor
	ytdlpPath    string
	ffmpegPath   string

	verMu      sync.RWMutex
	cachedInfo *VersionInfo
}

func NewApp() *App {
	return &App{}
}

func (a *App) LoadConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	configPath := filepath.Join(configDir, "OmniVideoRip", "config.json")
	data, err := os.ReadFile(configPath)
	if err == nil {
		var cfg AppConfig
		if err := json.Unmarshal(data, &cfg); err == nil && cfg.DownloadDir != "" {
			a.paths.DownloadsDir = cfg.DownloadDir
		}
	}
}

func (a *App) SaveConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	appConfigDir := filepath.Join(configDir, "OmniVideoRip")
	_ = os.MkdirAll(appConfigDir, 0755)
	configPath := filepath.Join(appConfigDir, "config.json")
	cfg := AppConfig{DownloadDir: a.paths.DownloadsDir}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(configPath, data, 0644)
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.logger = log.New(os.Stdout, "[OmniVideoRip] ", log.LstdFlags)

	paths, err := config.ResolvePaths()
	if err != nil {
		a.logger.Printf("Failed to resolve paths: %v", err)
		return
	}
	_ = config.EnsureDirectories(paths)
	a.paths = paths

	a.LoadConfig()

	a.ytdlpPath, _ = sidecars.ResolveYTDLPPath("")
	a.ffmpegPath, _ = sidecars.ResolveFFmpegPath("")

	tasksFile := ""
	if configDir, err := os.UserConfigDir(); err == nil {
		appConfigDir := filepath.Join(configDir, "OmniVideoRip")
		_ = os.MkdirAll(appConfigDir, 0755)
		tasksFile = filepath.Join(appConfigDir, "tasks.json")
	}

	a.taskStore = tasks.NewStore(tasksFile)
	a.taskExecutor = executor.New(a.logger, a.taskStore, a.paths.DownloadsDir, a.ytdlpPath, a.ffmpegPath)
	a.taskExecutor.SetWailsContext(ctx)

	go a.warmupVersionInfo()
	go a.taskExecutor.Run(ctx)
}

func (a *App) UpdateYTDLPComponent() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application is not ready")
	}

	runtime.EventsEmit(a.ctx, "ytdlp:update:progress", map[string]interface{}{
		"percent": 5,
		"message": "正在检查官方最新发布版本...",
	})

	targetPath, err := sidecars.DownloadOrUpdateYTDLP(a.ctx, func(pct int, msg string) {
		runtime.EventsEmit(a.ctx, "ytdlp:update:progress", map[string]interface{}{
			"percent": pct,
			"message": msg,
		})
	})
	if err != nil {
		runtime.EventsEmit(a.ctx, "ytdlp:update:error", map[string]interface{}{
			"error": err.Error(),
		})
		return "", err
	}

	a.ytdlpPath = targetPath
	if a.taskExecutor != nil {
		a.taskExecutor.SetSidecarPaths(a.ytdlpPath, a.ffmpegPath)
	}

	newInfo := a.warmupVersionInfo()
	runtime.EventsEmit(a.ctx, "ytdlp:update:done", map[string]interface{}{
		"path":         targetPath,
		"ytdlpVersion": newInfo.YTDLPVersion,
		"message":      fmt.Sprintf("yt-dlp 已成功更新至最新版本 (%s)！", newInfo.YTDLPVersion),
	})
	return targetPath, nil
}

func (a *App) UpdateFFmpegComponent() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application is not ready")
	}

	info := a.GetVersionInfo()
	ffPath, _ := sidecars.ResolveFFmpegPath(a.ffmpegPath)
	current := false
	if ffPath != "" {
		var err error
		current, err = sidecars.IsFFmpegCurrent(a.ctx, ffPath)
		if err != nil {
			return "", fmt.Errorf("check official FFmpeg release: %w", err)
		}
	}
	if current {
		runtime.EventsEmit(a.ctx, "ffmpeg:update:done", map[string]interface{}{
			"path":          ffPath,
			"ffmpegVersion": info.FFmpegVersion,
			"message":       fmt.Sprintf("当前 FFmpeg 组件已就绪 (%s)，无需重复下载！", info.FFmpegVersion),
			"alreadyLatest": true,
		})
		return ffPath, nil
	}

	targetPath, err := sidecars.DownloadOrUpdateFFmpeg(a.ctx, func(pct int, msg string) {
		runtime.EventsEmit(a.ctx, "ffmpeg:update:progress", map[string]interface{}{
			"percent": pct,
			"message": msg,
		})
	})
	if err != nil {
		runtime.EventsEmit(a.ctx, "ffmpeg:update:error", map[string]interface{}{
			"error": err.Error(),
		})
		return "", err
	}

	a.ffmpegPath = targetPath
	if a.taskExecutor != nil {
		a.taskExecutor.SetSidecarPaths(a.ytdlpPath, a.ffmpegPath)
	}

	newInfo := a.warmupVersionInfo()
	runtime.EventsEmit(a.ctx, "ffmpeg:update:done", map[string]interface{}{
		"path":          targetPath,
		"ffmpegVersion": newInfo.FFmpegVersion,
		"message":       fmt.Sprintf("FFmpeg 已成功安装/更新至最新版本 (%s)！", newInfo.FFmpegVersion),
	})
	return targetPath, nil
}

func (a *App) CheckLatestYTDLP() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application is not ready")
	}
	return sidecars.CheckLatestYTDLPVersion(a.ctx)
}

func (a *App) warmupVersionInfo() VersionInfo {
	info := VersionInfo{AppVersion: "1.3.9"}

	ytPath, _ := sidecars.ResolveYTDLPPath(a.ytdlpPath)
	if ytPath != "" {
		a.ytdlpPath = ytPath
		info.YTDLPVersion = "已就绪"
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		cmd := exec.CommandContext(ctx, ytPath, "--version")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		if out, err := cmd.Output(); err == nil {
			v := strings.TrimSpace(string(out))
			if v != "" {
				info.YTDLPVersion = v
			}
		}
		cancel()
	}

	ffPath, _ := sidecars.ResolveFFmpegPath(a.ffmpegPath)
	if ffPath != "" {
		a.ffmpegPath = ffPath
		info.FFmpegVersion = "已就绪"
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		cmd := exec.CommandContext(ctx, ffPath, "-version")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
		if out, err := cmd.Output(); err == nil {
			v := formatShortFFmpegVersion(string(out))
			if v != "" {
				info.FFmpegVersion = v
			}
		}
		cancel()
	}

	if a.taskExecutor != nil {
		a.taskExecutor.SetSidecarPaths(a.ytdlpPath, a.ffmpegPath)
	}

	a.verMu.Lock()
	a.cachedInfo = &info
	a.verMu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "version:ready", info)
	}

	return info
}

func (a *App) GetVersionInfo() VersionInfo {
	a.verMu.RLock()
	if a.cachedInfo != nil {
		info := *a.cachedInfo
		a.verMu.RUnlock()
		return info
	}
	a.verMu.RUnlock()

	info := VersionInfo{AppVersion: "1.3.9"}
	ytPath, _ := sidecars.ResolveYTDLPPath(a.ytdlpPath)
	if ytPath != "" {
		info.YTDLPVersion = "已就绪"
	}
	ffPath, _ := sidecars.ResolveFFmpegPath(a.ffmpegPath)
	if ffPath != "" {
		info.FFmpegVersion = "已就绪"
	}

	go a.warmupVersionInfo()

	return info
}

// GetSystemLocale returns the system UI locale (e.g. "zh-CN", "en-US", "ja-JP").
func (a *App) GetSystemLocale() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetUserDefaultLocaleName")

	buf := make([]uint16, 85)
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return "zh-CN"
	}
	return syscall.UTF16ToString(buf)
}

func formatShortFFmpegVersion(output string) string {
	firstLine := strings.SplitN(output, "\n", 2)[0]
	fields := strings.Fields(firstLine)
	for i, f := range fields {
		if strings.EqualFold(f, "version") && i+1 < len(fields) {
			rawVer := strings.Trim(fields[i+1], ",:; ")
			// If it's a date-based git build like 2024-05-02-git-...
			if len(rawVer) >= 10 && rawVer[4] == '-' && rawVer[7] == '-' {
				return rawVer[:10]
			}
			// If it's nightly build like N-92722-gf22fcd4483
			if strings.HasPrefix(rawVer, "N-") {
				parts := strings.Split(rawVer, "-")
				if len(parts) >= 2 {
					return parts[0] + "-" + parts[1]
				}
			}
			if len(rawVer) > 16 {
				return rawVer[:16]
			}
			return rawVer
		}
	}
	if len(firstLine) > 20 {
		return firstLine[:20]
	}
	return strings.TrimSpace(firstLine)
}

func (a *App) AnalyzeVideo(rawURL string) (extractors.VideoMetaData, error) {
	if a.ctx == nil {
		return extractors.VideoMetaData{}, fmt.Errorf("application is not ready")
	}
	router := extractors.NewRouter(a.paths.DownloadsDir, a.ytdlpPath, a.ffmpegPath)
	return router.Analyze(a.ctx, rawURL)
}

func (a *App) SelectDirectory() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "选择 OmniVideoRip 默认保存目录",
		DefaultDirectory: a.paths.DownloadsDir,
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

func (a *App) GetDefaultDownloadDir() string {
	return a.paths.DownloadsDir
}

func (a *App) SetDefaultDownloadDir(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	a.paths.DownloadsDir = path
	a.SaveConfig()
	return nil
}

func (a *App) SubmitDownload(rawURL, outputDir string) (tasks.Task, error) {
	return a.SubmitDownloadWithQuality(rawURL, outputDir, "best")
}

func (a *App) SubmitDownloadWithQuality(rawURL, outputDir, formatID string) (tasks.Task, error) {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()
	if a.taskStore == nil || a.taskExecutor == nil {
		return tasks.Task{}, fmt.Errorf("application is not ready")
	}
	outDir := outputDir
	if outDir == "" {
		outDir = a.paths.DownloadsDir
	}

	req := tasks.CreateTaskRequest{
		URL:           strings.TrimSpace(rawURL),
		OutputDir:     outDir,
		Source:        "OmniVideoRip_Desktop",
		FormatID:      formatID,
		ClientVersion: "1.3.9",
	}

	task, err := a.taskStore.Create(req)
	if err != nil {
		return tasks.Task{}, err
	}

	// Attach formatID to executor
	if err := a.taskExecutor.EnqueueWithFormat(task.ID, formatID); err != nil {
		failedTask, _ := a.taskStore.Fail(task.ID, "queue_full", err.Error())
		return failedTask, err
	}

	currentTask, _ := a.taskStore.Get(task.ID)
	return currentTask, nil
}

func (a *App) SubmitBatchDownloads(rawURLsText, outputDir string) ([]tasks.Task, error) {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()
	if a.taskStore == nil || a.taskExecutor == nil {
		return nil, fmt.Errorf("application is not ready")
	}
	outDir := outputDir
	if outDir == "" {
		outDir = a.paths.DownloadsDir
	}

	lines := strings.Split(rawURLsText, "\n")
	var createdTasks []tasks.Task

	for _, line := range lines {
		parts := strings.Fields(line)
		for _, part := range parts {
			cleanURL := strings.TrimSpace(part)
			if cleanURL == "" || (!strings.HasPrefix(cleanURL, "http://") && !strings.HasPrefix(cleanURL, "https://")) {
				continue
			}

			req := tasks.CreateTaskRequest{
				URL:           cleanURL,
				OutputDir:     outDir,
				Source:        "OmniVideoRip_Batch",
				ClientVersion: "1.3.9",
			}

			task, err := a.taskStore.Create(req)
			if err != nil {
				continue
			}

			if err := a.taskExecutor.Enqueue(task.ID); err == nil {
				currentTask, _ := a.taskStore.Get(task.ID)
				createdTasks = append(createdTasks, currentTask)
			} else {
				failedTask, _ := a.taskStore.Fail(task.ID, "queue_full", err.Error())
				createdTasks = append(createdTasks, failedTask)
			}
		}
	}

	if len(createdTasks) == 0 {
		return nil, fmt.Errorf("未检测到有效的视频下载地址")
	}

	return createdTasks, nil
}

func (a *App) GetAllTasks() []tasks.Task {
	if a.taskStore == nil {
		return []tasks.Task{}
	}
	return a.taskStore.List()
}

func (a *App) GetTaskStatus(taskId string) (tasks.Task, error) {
	if a.taskStore == nil {
		return tasks.Task{}, fmt.Errorf("application is not ready")
	}
	task, ok := a.taskStore.Get(taskId)
	if !ok {
		return tasks.Task{}, fmt.Errorf("task not found")
	}
	return task, nil
}

func (a *App) PauseTask(taskId string) (tasks.Task, error) {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()
	if a.taskStore == nil || a.taskExecutor == nil {
		return tasks.Task{}, fmt.Errorf("application is not ready")
	}
	if err := a.taskExecutor.PauseTask(taskId); err != nil {
		return tasks.Task{}, err
	}
	task, _ := a.taskStore.Get(taskId)
	return task, nil
}

func (a *App) ResumeTask(taskId string) (tasks.Task, error) {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()
	if a.taskStore == nil || a.taskExecutor == nil {
		return tasks.Task{}, fmt.Errorf("application is not ready")
	}
	if err := a.taskExecutor.ResumeTask(taskId); err != nil {
		return tasks.Task{}, err
	}
	task, _ := a.taskStore.Get(taskId)
	return task, nil
}

func (a *App) DeleteTask(taskId string) (bool, error) {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()
	if a.taskStore == nil {
		return false, fmt.Errorf("application is not ready")
	}
	task, ok := a.taskStore.Get(taskId)
	if !ok {
		return false, nil
	}

	// 1. Cancel if active
	if a.taskExecutor != nil {
		a.taskExecutor.CancelAndWait(taskId)
	}
	deleted, err := a.taskStore.DeleteWithError(taskId)
	if err != nil {
		return false, fmt.Errorf("delete persisted task: %w", err)
	}
	if !deleted {
		return false, nil
	}

	// 2. If task was not successfully completed, clean up its uncompleted resume cache
	if task.Status != tasks.StatusCompleted {
		outDir := task.OutputDir
		if outDir == "" {
			outDir = a.paths.DownloadsDir
		}
		if err := extractors.CleanupTaskResumeCache(outDir, task.Platform, task.URL); err != nil {
			return true, err
		}
	}

	// 3. Remove from store
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "task:deleted", taskId)
	}
	return true, nil
}

func (a *App) ClearAllTasksAndCaches() error {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()
	if a.taskStore == nil {
		return fmt.Errorf("application is not ready")
	}

	action := func() error { return a.clearStoppedTasks() }
	if a.taskExecutor != nil {
		return a.taskExecutor.WithStoppedWorkers(action)
	}
	return action()
}

func (a *App) clearStoppedTasks() error {
	allTasks := a.taskStore.List()

	// 2. For uncompleted tasks, clean up individual resume caches
	for _, t := range allTasks {
		if t.Status != tasks.StatusCompleted {
			outDir := t.OutputDir
			if outDir == "" {
				outDir = a.paths.DownloadsDir
			}
			if err := extractors.CleanupTaskResumeCache(outDir, t.Platform, t.URL); err != nil {
				return err
			}
		}
	}

	// 4. Persist the empty store before telling the frontend the clear succeeded.
	if err := a.taskStore.Clear(); err != nil {
		return fmt.Errorf("clear persisted tasks: %w", err)
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "tasks:cleared", true)
	}
	return nil
}

// ClearTasks keeps older frontend bindings compatible with the durable clear operation.
func (a *App) ClearTasks() error {
	return a.ClearAllTasksAndCaches()
}

func (a *App) OnShutdown(ctx context.Context) {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()
	if a.taskExecutor != nil {
		a.taskExecutor.CancelAllAndWait()
	}
	if a.taskStore != nil {
		if err := a.taskStore.Flush(); err != nil && a.logger != nil {
			a.logger.Printf("Save task progress: %v", err)
		}
	}
}
