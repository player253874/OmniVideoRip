package tasks

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRestartPreservesFormatAndMakesQueuedTasksResumable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	store := NewStore(path)
	task, err := store.Create(CreateTaskRequest{URL: "https://example.com/video", FormatID: "720p"})
	if err != nil {
		t.Fatal(err)
	}
	restored := NewStore(path)
	got, ok := restored.Get(task.ID)
	if !ok || got.Status != StatusPaused || got.FormatID != "720p" {
		t.Fatalf("restored %+v", got)
	}
}

func TestProgressCheckpointAndDurableClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	store := NewStore(path)
	task, err := store.Create(CreateTaskRequest{URL: "https://example.com/video"})
	if err != nil {
		t.Fatal(err)
	}
	initial, _ := os.ReadFile(path)
	for i := 0; i < 100; i++ {
		store.UpdateProgress(task.ID, func(t *Task) { t.Progress = i })
	}
	checkpoint, _ := os.ReadFile(path)
	if !bytes.Equal(initial, checkpoint) {
		t.Fatal("every progress event wrote the whole file")
	}
	if err := store.Flush(); err != nil {
		t.Fatal(err)
	}
	got, _ := NewStore(path).Get(task.ID)
	if got.Progress != 99 {
		t.Fatalf("progress=%d", got.Progress)
	}
	store.UpdateProgress(task.ID, func(t *Task) { t.Progress = 100 })
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := store.Flush(); err != nil {
		t.Fatal(err)
	}
	if NewStore(path).Count() != 0 {
		t.Fatal("pending checkpoint resurrected cleared tasks")
	}
}

func TestDeleteReturnsPersistenceFailure(t *testing.T) {
	store := NewStore()
	task, _ := store.Create(CreateTaskRequest{URL: "https://example.com/video"})
	blocker := filepath.Join(t.TempDir(), "blocker")
	os.WriteFile(blocker, []byte("x"), 0600)
	store.SetPersistPath(filepath.Join(blocker, "tasks.json"))
	deleted, err := store.DeleteWithError(task.ID)
	if err == nil || deleted {
		t.Fatalf("deleted=%v err=%v", deleted, err)
	}
	if _, ok := store.Get(task.ID); !ok {
		t.Fatal("failed delete lost task")
	}
}
