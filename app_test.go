package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"videodownloader/internal/config"
	"videodownloader/internal/executor"
	"videodownloader/internal/extractors"
	"videodownloader/internal/tasks"
)

func TestSubmitBatchDownloadsMarksQueueOverflowFailed(t *testing.T) {
	store := tasks.NewStore()
	taskExecutor := executor.New(log.New(io.Discard, "", 0), store, t.TempDir(), "", "")
	for i := 0; i < 64; i++ {
		task, err := store.Create(tasks.CreateTaskRequest{URL: fmt.Sprintf("https://example.com/occupied/%d", i)})
		if err != nil {
			t.Fatal(err)
		}
		if err := taskExecutor.Enqueue(task.ID); err != nil {
			t.Fatalf("fill queue: %v", err)
		}
	}
	app := &App{taskStore: store, taskExecutor: taskExecutor}

	created, err := app.SubmitBatchDownloads("https://example.com/video", t.TempDir())
	if err != nil {
		t.Fatalf("SubmitBatchDownloads failed: %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("created %d tasks; want 1", len(created))
	}
	if created[0].Status != tasks.StatusFailed || created[0].ErrorCode != "queue_full" {
		t.Fatalf("overflow task status = %q, code = %q; want failed/queue_full", created[0].Status, created[0].ErrorCode)
	}
}

func TestPauseResumeDeleteTask(t *testing.T) {
	store := tasks.NewStore()
	tempDir := t.TempDir()
	taskExecutor := executor.New(log.New(io.Discard, "", 0), store, tempDir, "", "")
	app := &App{
		taskStore:    store,
		taskExecutor: taskExecutor,
		paths:        config.Paths{DownloadsDir: tempDir},
	}

	task, err := store.Create(tasks.CreateTaskRequest{
		URL:       "https://missav.live/cesd-935",
		OutputDir: tempDir,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	// 1. Test Pause
	paused, err := app.PauseTask(task.ID)
	if err != nil {
		t.Fatalf("pause task: %v", err)
	}
	if paused.Status != tasks.StatusPaused {
		t.Fatalf("paused status = %q; want %q", paused.Status, tasks.StatusPaused)
	}

	// 2. Test Resume
	resumed, err := app.ResumeTask(task.ID)
	if err != nil {
		t.Fatalf("resume task: %v", err)
	}
	if resumed.Status != tasks.StatusQueued {
		t.Fatalf("resumed status = %q; want %q", resumed.Status, tasks.StatusQueued)
	}

	// 3. Create a fake unfinished resume directory
	cacheDir := extractors.GetTaskResumeDir(tempDir, task.Platform, task.URL)
	_ = os.MkdirAll(cacheDir, 0755)
	_ = os.WriteFile(filepath.Join(cacheDir, "seg_00001.ts"), []byte("test"), 0644)

	// 4. Test Delete Task and Cache cleanup
	deleted, err := app.DeleteTask(task.ID)
	if err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if !deleted {
		t.Fatalf("delete task returned false")
	}

	// Verify task removed from store
	if _, ok := store.Get(task.ID); ok {
		t.Fatalf("task still exists in store after delete")
	}

	// Verify resume cache removed
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Fatalf("resume cache dir %s was not cleaned up after delete", cacheDir)
	}
}

func TestClearAllTasksAndCaches(t *testing.T) {
	store := tasks.NewStore()
	tempDir := t.TempDir()
	taskExecutor := executor.New(log.New(io.Discard, "", 0), store, tempDir, "", "")
	app := &App{
		taskStore:    store,
		taskExecutor: taskExecutor,
		paths:        config.Paths{DownloadsDir: tempDir},
	}

	// Create unfinished tasks with caches
	t1, _ := store.Create(tasks.CreateTaskRequest{URL: "https://missav.live/cesd-935", OutputDir: tempDir})
	t2, _ := store.Create(tasks.CreateTaskRequest{URL: "https://ggjav.com/main/video?id=289780", OutputDir: tempDir})

	c1 := extractors.GetTaskResumeDir(tempDir, t1.Platform, t1.URL)
	c2 := extractors.GetTaskResumeDir(tempDir, t2.Platform, t2.URL)
	_ = os.MkdirAll(c1, 0755)
	_ = os.MkdirAll(c2, 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "video.partial.mp4"), []byte("part"), 0644)

	// Test ClearAllTasksAndCaches
	if err := app.ClearAllTasksAndCaches(); err != nil {
		t.Fatalf("ClearAllTasksAndCaches error: %v", err)
	}

	// Verify store empty
	if store.Count() != 0 {
		t.Fatalf("store count = %d; want 0", store.Count())
	}

	// Verify caches removed
	if _, err := os.Stat(c1); !os.IsNotExist(err) {
		t.Fatalf("cache c1 was not removed")
	}
	if _, err := os.Stat(c2); !os.IsNotExist(err) {
		t.Fatalf("cache c2 was not removed")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "video.partial.mp4")); err != nil {
		t.Fatalf("unowned partial video was removed")
	}
}

func TestGetSystemLocale(t *testing.T) {
	app := &App{}
	locale := app.GetSystemLocale()
	if locale == "" {
		t.Fatalf("expected non-empty locale")
	}
	t.Logf("Detected system locale: %s", locale)
}

func TestDeleteTaskPropagatesPersistenceError(t *testing.T) {
	store := tasks.NewStore()
	task, _ := store.Create(tasks.CreateTaskRequest{URL: "https://example.com/video"})
	blocker := filepath.Join(t.TempDir(), "blocker")
	os.WriteFile(blocker, []byte("x"), 0600)
	store.SetPersistPath(filepath.Join(blocker, "tasks.json"))
	app := &App{taskStore: store}
	deleted, err := app.DeleteTask(task.ID)
	if err == nil || deleted {
		t.Fatalf("deleted=%v err=%v", deleted, err)
	}
	if _, ok := store.Get(task.ID); !ok {
		t.Fatal("failed delete lost task")
	}
}
