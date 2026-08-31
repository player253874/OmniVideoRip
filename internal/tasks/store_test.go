package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestStoreCreateValidTask(t *testing.T) {
	store := NewStore()

	req := CreateTaskRequest{
		URL:           "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		OutputDir:     "C:\\Downloads",
		Source:        "test",
		ClientVersion: "2.0.0",
	}

	task, err := store.Create(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if task.ID == "" {
		t.Errorf("Expected non-empty task ID")
	}

	if task.Platform != "youtube" {
		t.Errorf("Expected platform 'youtube', got '%s'", task.Platform)
	}

	if task.Status != StatusQueued {
		t.Errorf("Expected status 'queued', got '%s'", task.Status)
	}
}

func TestStoreCreateInvalidURL(t *testing.T) {
	store := NewStore()

	req := CreateTaskRequest{
		URL: "invalid-url-string",
	}

	_, err := store.Create(req)
	if err == nil {
		t.Errorf("Expected error for invalid URL, got nil")
	}
}

func TestStoreRejectsNonHTTPURL(t *testing.T) {
	store := NewStore()
	for _, rawURL := range []string{"file:///C:/Windows/win.ini", "ftp://example.com/video"} {
		if _, err := store.Create(CreateTaskRequest{URL: rawURL}); err == nil {
			t.Errorf("Create(%q) unexpectedly succeeded", rawURL)
		}
	}
}

func TestStoreConcurrency(t *testing.T) {
	store := NewStore()
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := CreateTaskRequest{
				URL:       fmt.Sprintf("https://www.bilibili.com/video/BV1xx411c7m%d", idx),
				OutputDir: "C:\\Downloads",
			}
			task, err := store.Create(req)
			if err == nil {
				_, _ = store.Get(task.ID)
				store.SetProgress(task.ID, idx*2, "Downloading...")
				store.List()
			}
		}(i)
	}

	wg.Wait()

	if store.Count() != numGoroutines {
		t.Errorf("Expected store count %d, got %d", numGoroutines, store.Count())
	}
}

func TestStorePersistence(t *testing.T) {
	tempFile := fmt.Sprintf("%s/tasks_%d.json", t.TempDir(), 123)

	// 1. Create store with file and add tasks
	store1 := NewStore(tempFile)
	t1, err := store1.Create(CreateTaskRequest{URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", OutputDir: "C:\\Downloads"})
	if err != nil {
		t.Fatalf("Create task: %v", err)
	}
	store1.SetStatus(t1.ID, StatusPaused, "User paused")
	store1.SetProgress(t1.ID, 45, "45% done")

	// 2. Create store2 pointing to same file, verify loaded
	store2 := NewStore(tempFile)
	if store2.Count() != 1 {
		t.Fatalf("Expected store2 count 1, got %d", store2.Count())
	}
	loaded, ok := store2.Get(t1.ID)
	if !ok {
		t.Fatalf("Expected to find task %s in store2", t1.ID)
	}
	if loaded.Status != StatusPaused {
		t.Errorf("Expected status paused, got %s", loaded.Status)
	}
	if loaded.Progress != 45 {
		t.Errorf("Expected progress 45, got %d", loaded.Progress)
	}

	// 3. Test delete in store2 and verify persisted
	store2.Delete(t1.ID)
	store3 := NewStore(tempFile)
	if store3.Count() != 0 {
		t.Errorf("Expected store3 count 0 after delete, got %d", store3.Count())
	}
}

func TestStorePlatformDetectionAndQueueSegregation(t *testing.T) {
	store := NewStore()

	testCases := []struct {
		url              string
		expectedPlatform string
	}{
		{"https://rou.video/v/cmt6sx2gg0000s6zm7hpzvx3x", "rouvideo"},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "youtube"},
		{"https://missav.ai/ja/abc-123", "missav"},
		{"https://ggjav.com/main/video?id=123", "ggjav"},
		{"https://www.tiktok.com/@user/video/123", "tiktok"},
		{"https://spankbang.com/abc/video/test", "spankbang.com"},
		{"https://vimeo.com/123456", "vimeo"},
		{"https://example.org/myvideo.mp4", "example.org"},
	}

	for _, tc := range testCases {
		task, err := store.Create(CreateTaskRequest{URL: tc.url, OutputDir: "C:\\Downloads"})
		if err != nil {
			t.Fatalf("Failed to create task for %s: %v", tc.url, err)
		}
		if task.Platform != tc.expectedPlatform {
			t.Errorf("URL %s got platform %q, want %q", tc.url, task.Platform, tc.expectedPlatform)
		}
	}
}
func TestStoreClearIsDurableAndRejectsLateUpdates(t *testing.T) {
	persistPath := filepath.Join(t.TempDir(), "tasks.json")
	store := NewStore(persistPath)
	task, err := store.Create(CreateTaskRequest{URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if store.Count() != 0 {
		t.Fatalf("store count after clear = %d", store.Count())
	}
	if _, ok := store.SetProgress(task.ID, 90, "late worker update"); ok {
		t.Fatal("late worker update recreated a cleared task")
	}
	reloaded := NewStore(persistPath)
	if reloaded.Count() != 0 {
		t.Fatalf("reloaded store count after clear = %d", reloaded.Count())
	}
}

func TestStoreClearRollsBackWhenPersistenceFails(t *testing.T) {
	store := NewStore()
	if _, err := store.Create(CreateTaskRequest{URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocker, []byte("block"), 0600); err != nil {
		t.Fatal(err)
	}
	store.SetPersistPath(filepath.Join(blocker, "tasks.json"))
	if err := store.Clear(); err == nil {
		t.Fatal("Clear unexpectedly succeeded with an invalid persistence path")
	}
	if store.Count() != 1 {
		t.Fatalf("failed clear lost in-memory tasks; count = %d", store.Count())
	}
}
