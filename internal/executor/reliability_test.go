package executor

import (
	"context"
	"io"
	"log"
	"testing"
	"videodownloader/internal/tasks"
)

func TestResumeWaitsForOldWorkerBeforeReusingTaskID(t *testing.T) {
	store := tasks.NewStore()
	task, _ := store.Create(tasks.CreateTaskRequest{URL: "https://example.com/video", FormatID: "720p"})
	e := New(log.New(io.Discard, "", 0), store, t.TempDir(), "missing", "missing")
	ctx, cancel := context.WithCancel(context.Background())
	e.activeCancels[task.ID] = cancel
	e.activeByPlatform[task.Platform] = task.ID
	e.activeDone[task.ID] = make(chan struct{})
	store.SetStatus(task.ID, tasks.StatusRunning, "running")
	if err := e.PauseTask(task.ID); err != nil {
		t.Fatal(err)
	}
	if err := e.ResumeTask(task.ID); err != nil {
		t.Fatal(err)
	}
	e.schedule(context.Background())
	if len(e.pendingQueue) != 1 || e.pendingQueue[0].FormatID != "720p" {
		t.Fatal("resumed task ran before old worker finished")
	}
	e.workerWG.Add(1)
	e.runWorker(ctx, task.ID, "720p", task.Platform)
	got, _ := store.Get(task.ID)
	if got.Status != tasks.StatusQueued {
		t.Fatalf("old worker changed resumed status to %s", got.Status)
	}
	if len(e.activeDone) != 0 {
		t.Fatal("stopping worker not retired")
	}
}

func TestClearBarrierCoversCleanupCallback(t *testing.T) {
	store := tasks.NewStore()
	task, _ := store.Create(tasks.CreateTaskRequest{URL: "https://example.com/video"})
	e := New(log.New(io.Discard, "", 0), store, t.TempDir(), "", "")
	err := e.WithStoppedWorkers(func() error {
		if err := e.Enqueue(task.ID); err == nil {
			t.Fatal("accepted work during clear")
		}
		return store.Clear()
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.QueueDepth() != 0 || store.Count() != 0 {
		t.Fatal("clear did not empty queue and store")
	}
}
