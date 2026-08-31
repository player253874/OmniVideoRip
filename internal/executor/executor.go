package executor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"videodownloader/internal/extractors"
	"videodownloader/internal/tasks"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type QueuedTaskItem struct {
	TaskID   string
	FormatID string
}

type Executor struct {
	logger     *log.Logger
	store      *tasks.Store
	outputDir  string
	ytdlpPath  string
	ffmpegPath string
	wailsCtx   context.Context
	ctxMu      sync.Mutex

	mu               sync.Mutex
	pendingQueue     []QueuedTaskItem
	taskFormat       map[string]string
	activeByPlatform map[string]string             // platform -> taskID
	activeCancels    map[string]context.CancelFunc // taskID -> cancelFunc
	scheduleSignal   chan struct{}
	workerWG         sync.WaitGroup
	clearing         bool
}

func New(logger *log.Logger, store *tasks.Store, outputDir, ytdlpPath, ffmpegPath string) *Executor {
	return &Executor{
		logger:           logger,
		store:            store,
		outputDir:        outputDir,
		ytdlpPath:        ytdlpPath,
		ffmpegPath:       ffmpegPath,
		pendingQueue:     make([]QueuedTaskItem, 0),
		taskFormat:       make(map[string]string),
		activeByPlatform: make(map[string]string),
		activeCancels:    make(map[string]context.CancelFunc),
		scheduleSignal:   make(chan struct{}, 64),
	}
}

func (e *Executor) SetWailsContext(ctx context.Context) {
	e.ctxMu.Lock()
	defer e.ctxMu.Unlock()
	e.wailsCtx = ctx
}

func (e *Executor) SetSidecarPaths(ytdlpPath, ffmpegPath string) {
	e.ctxMu.Lock()
	defer e.ctxMu.Unlock()
	e.ytdlpPath = ytdlpPath
	e.ffmpegPath = ffmpegPath
}

func (e *Executor) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			e.mu.Lock()
			for _, cancel := range e.activeCancels {
				if cancel != nil {
					cancel()
				}
			}
			e.mu.Unlock()
			return
		case <-e.scheduleSignal:
			e.schedule(ctx)
		}
	}
}

func (e *Executor) notifySchedule() {
	select {
	case e.scheduleSignal <- struct{}{}:
	default:
	}
}

func (e *Executor) schedule(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.clearing || len(e.pendingQueue) == 0 {
		return
	}

	remaining := make([]QueuedTaskItem, 0, len(e.pendingQueue))

	for _, item := range e.pendingQueue {
		task, ok := e.store.Get(item.TaskID)
		if !ok || task.Status == tasks.StatusPaused || task.Status == tasks.StatusCompleted || task.Status == tasks.StatusFailed {
			continue
		}

		platform := task.Platform
		if platform == "" {
			platform = "generic"
		}

		// Check if this platform currently has an active task running
		if runningID, busy := e.activeByPlatform[platform]; busy && runningID != "" {
			remaining = append(remaining, item)
			continue
		}

		// Allocate platform slot and launch worker
		taskCtx, taskCancel := context.WithCancel(ctx)
		e.activeByPlatform[platform] = item.TaskID
		e.activeCancels[item.TaskID] = taskCancel

		// Immediately update task in store and notify frontend so it transitions to running instantly
		e.emitUpdate(item.TaskID, func(t *tasks.Task) {
			t.Status = tasks.StatusRunning
			if t.Progress < 5 {
				t.Progress = 5
			}
			t.Message = "正在初始化下载流水线..."
		})

		e.workerWG.Add(1)
		go e.runWorker(taskCtx, item.TaskID, item.FormatID, platform)
	}

	e.pendingQueue = remaining
}

func (e *Executor) runWorker(taskCtx context.Context, taskID, formatID, platform string) {
	defer e.workerWG.Done()
	defer func() {
		e.mu.Lock()
		delete(e.activeCancels, taskID)
		if e.activeByPlatform[platform] == taskID {
			delete(e.activeByPlatform, platform)
		}
		e.mu.Unlock()
		e.notifySchedule()
	}()

	e.processTask(taskCtx, taskID, formatID)
}

func (e *Executor) Enqueue(taskID string) error {
	return e.EnqueueWithFormat(taskID, "best")
}

func (e *Executor) EnqueueWithFormat(taskID, formatID string) error {
	e.mu.Lock()
	if e.clearing {
		e.mu.Unlock()
		return fmt.Errorf("task executor is clearing")
	}
	if len(e.pendingQueue) >= 64 {
		e.mu.Unlock()
		return fmt.Errorf("task queue is full")
	}
	e.taskFormat[taskID] = formatID
	e.pendingQueue = append(e.pendingQueue, QueuedTaskItem{TaskID: taskID, FormatID: formatID})
	e.mu.Unlock()

	e.notifySchedule()
	return nil
}

func (e *Executor) QueueDepth() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.pendingQueue)
}

func (e *Executor) PauseTask(taskID string) error {
	e.mu.Lock()
	if cancel, running := e.activeCancels[taskID]; running && cancel != nil {
		cancel()
		delete(e.activeCancels, taskID)
		for p, id := range e.activeByPlatform {
			if id == taskID {
				delete(e.activeByPlatform, p)
				break
			}
		}
	}

	for i, item := range e.pendingQueue {
		if item.TaskID == taskID {
			e.pendingQueue = append(e.pendingQueue[:i], e.pendingQueue[i+1:]...)
			break
		}
	}
	e.mu.Unlock()

	task, ok := e.store.Get(taskID)
	if !ok {
		return fmt.Errorf("task not found")
	}
	if task.Status == tasks.StatusQueued || task.Status == tasks.StatusRunning {
		e.emitUpdate(taskID, func(item *tasks.Task) {
			item.Status = tasks.StatusPaused
			item.Message = "下载已暂停 (支持随时继续)"
		})
	}
	e.notifySchedule()
	return nil
}

func (e *Executor) ResumeTask(taskID string) error {
	task, ok := e.store.Get(taskID)
	if !ok {
		return fmt.Errorf("task not found")
	}
	if task.Status == tasks.StatusRunning || task.Status == tasks.StatusCompleted {
		return nil
	}

	formatID := "best"
	e.mu.Lock()
	if f, ok := e.taskFormat[taskID]; ok && f != "" {
		formatID = f
	}
	e.mu.Unlock()

	e.emitUpdate(taskID, func(item *tasks.Task) {
		item.Status = tasks.StatusQueued
		item.Message = "已加入恢复队列，准备继续下载..."
	})

	return e.EnqueueWithFormat(taskID, formatID)
}

func (e *Executor) CancelAllAndWait() {
	e.mu.Lock()
	e.clearing = true
	for _, cancel := range e.activeCancels {
		if cancel != nil {
			cancel()
		}
	}
	e.pendingQueue = nil
	e.taskFormat = make(map[string]string)
	e.mu.Unlock()

	e.workerWG.Wait()

	e.mu.Lock()
	e.activeCancels = make(map[string]context.CancelFunc)
	e.activeByPlatform = make(map[string]string)
	e.clearing = false
	e.mu.Unlock()
}
func (e *Executor) CancelIfActive(taskID string) {
	e.mu.Lock()
	if cancel, ok := e.activeCancels[taskID]; ok && cancel != nil {
		cancel()
		delete(e.activeCancels, taskID)
		for p, id := range e.activeByPlatform {
			if id == taskID {
				delete(e.activeByPlatform, p)
				break
			}
		}
	}
	for i, item := range e.pendingQueue {
		if item.TaskID == taskID {
			e.pendingQueue = append(e.pendingQueue[:i], e.pendingQueue[i+1:]...)
			break
		}
	}
	e.mu.Unlock()
	e.notifySchedule()
}

func (e *Executor) processTask(taskCtx context.Context, taskID, formatID string) {
	task, ok := e.store.Get(taskID)
	if !ok {
		return
	}

	if task.Status == tasks.StatusPaused || task.Status == tasks.StatusCompleted {
		return
	}

	e.logger.Printf("Processing task %s (%s, format: %s)", task.ID, task.Platform, formatID)

	router := extractors.NewRouter(e.outputDir, e.ytdlpPath, e.ffmpegPath)

	outPath, err := router.DownloadWithFormat(taskCtx, task.URL, task.OutputDir, formatID, func(percent int, msg string) {
		if taskCtx.Err() != nil {
			return
		}
		e.emitUpdate(taskID, func(item *tasks.Task) {
			if taskCtx.Err() != nil {
				return
			}
			item.Status = tasks.StatusRunning
			if percent > item.Progress {
				item.Progress = percent
			}
			item.Message = msg
		})
	})

	if err != nil {
		if taskCtx.Err() != nil || errors.Is(err, context.Canceled) {
			e.logger.Printf("Task %s paused by user", taskID)
			e.emitUpdate(taskID, func(item *tasks.Task) {
				item.Status = tasks.StatusPaused
				item.Message = "下载已暂停 (支持随时继续)"
			})
			return
		}

		e.logger.Printf("Task %s failed: %v", taskID, err)
		e.emitUpdate(taskID, func(item *tasks.Task) {
			item.Status = tasks.StatusFailed
			item.ErrorCode = "download_failed"
			item.Message = err.Error()
		})
		return
	}

	e.emitUpdate(taskID, func(item *tasks.Task) {
		item.Status = tasks.StatusCompleted
		item.Progress = 100
		item.Message = "下载完成，文件已保存在本地。"
		item.OutputPath = outPath
	})
}

func (e *Executor) emitUpdate(taskID string, updateFn func(item *tasks.Task)) {
	updatedTask, ok := e.store.Update(taskID, updateFn)
	if !ok {
		return
	}

	e.ctxMu.Lock()
	wCtx := e.wailsCtx
	e.ctxMu.Unlock()

	if wCtx != nil {
		runtime.EventsEmit(wCtx, "task:progress", updatedTask)
	}
}
