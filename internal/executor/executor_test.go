package executor

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
	"videodownloader/internal/tasks"
)

func TestExecutorEnqueueAndDepth(t *testing.T) {
	logger := log.New(os.Stdout, "[Test] ", log.LstdFlags)
	store := tasks.NewStore()

	exec := New(logger, store, t.TempDir(), "dummy_yt", "dummy_ff")

	req := tasks.CreateTaskRequest{
		URL:       "https://www.bilibili.com/video/BV1xx411c7mD",
		OutputDir: t.TempDir(),
	}
	task, err := store.Create(req)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	err = exec.Enqueue(task.ID)
	if err != nil {
		t.Fatalf("Failed to enqueue task: %v", err)
	}

	if exec.QueueDepth() != 1 {
		t.Errorf("Expected queue depth 1, got %d", exec.QueueDepth())
	}
}

func TestExecutorPerPlatformScheduling(t *testing.T) {
	logger := log.New(os.Stdout, "[Test] ", log.LstdFlags)
	store := tasks.NewStore()
	tempDir := t.TempDir()

	exec := New(logger, store, tempDir, "dummy_yt", "dummy_ff")

	// Create 2 YouTube tasks and 1 Bilibili task
	t1, _ := store.Create(tasks.CreateTaskRequest{URL: "https://www.youtube.com/watch?v=111", OutputDir: tempDir})
	t2, _ := store.Create(tasks.CreateTaskRequest{URL: "https://www.youtube.com/watch?v=222", OutputDir: tempDir})
	t3, _ := store.Create(tasks.CreateTaskRequest{URL: "https://www.bilibili.com/video/BV1xx", OutputDir: tempDir})

	// Add to pendingQueue
	exec.mu.Lock()
	exec.pendingQueue = append(exec.pendingQueue,
		QueuedTaskItem{TaskID: t1.ID, FormatID: "best"},
		QueuedTaskItem{TaskID: t2.ID, FormatID: "best"},
		QueuedTaskItem{TaskID: t3.ID, FormatID: "best"},
	)
	exec.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exec.schedule(ctx)

	exec.mu.Lock()
	activeYT := exec.activeByPlatform["youtube"]
	activeBili := exec.activeByPlatform["bilibili"]
	remainingLen := len(exec.pendingQueue)
	exec.mu.Unlock()

	// YouTube t1 should be allocated
	if activeYT != t1.ID {
		t.Errorf("Expected active YouTube task to be %s, got %s", t1.ID, activeYT)
	}
	// Bilibili t3 should be allocated concurrently
	if activeBili != t3.ID {
		t.Errorf("Expected active Bilibili task to be %s, got %s", t3.ID, activeBili)
	}
	// YouTube t2 should remain in queue because YouTube is currently busy
	if remainingLen != 1 {
		t.Errorf("Expected 1 task (t2) remaining in queue, got %d", remainingLen)
	}

	// Now Pause t1:
	_ = exec.PauseTask(t1.ID)

	// After pause, t1's platform slot should be released, and next schedule should promote t2
	exec.schedule(ctx)

	exec.mu.Lock()
	newActiveYT := exec.activeByPlatform["youtube"]
	exec.mu.Unlock()

	if newActiveYT != t2.ID {
		t.Errorf("Expected YouTube t2 to be promoted after t1 paused, got %s", newActiveYT)
	}
}
func TestCancelAllAndWaitClearsQueueAndWaitsForWorkers(t *testing.T) {
	logger := log.New(os.Stdout, "[Test] ", log.LstdFlags)
	store := tasks.NewStore()
	exec := New(logger, store, t.TempDir(), "dummy_yt", "dummy_ff")
	task, err := store.Create(tasks.CreateTaskRequest{URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})
	if err != nil {
		t.Fatal(err)
	}
	if err := exec.Enqueue(task.ID); err != nil {
		t.Fatal(err)
	}

	exec.workerWG.Add(1)
	go func() {
		time.Sleep(60 * time.Millisecond)
		exec.workerWG.Done()
	}()
	started := time.Now()
	exec.CancelAllAndWait()
	if elapsed := time.Since(started); elapsed < 50*time.Millisecond {
		t.Fatalf("CancelAllAndWait returned before workers stopped: %v", elapsed)
	}
	if depth := exec.QueueDepth(); depth != 0 {
		t.Fatalf("queue depth after clear = %d", depth)
	}
	if err := exec.Enqueue(task.ID); err != nil {
		t.Fatalf("executor did not resume accepting work after clear: %v", err)
	}
}
