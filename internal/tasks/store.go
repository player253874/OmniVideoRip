package tasks

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu          sync.RWMutex
	tasks       map[string]Task
	persistPath string
}

func NewStore(persistPath ...string) *Store {
	var p string
	if len(persistPath) > 0 {
		p = persistPath[0]
	}
	s := &Store{
		tasks:       make(map[string]Task),
		persistPath: p,
	}
	if p != "" {
		_ = s.Load()
	}
	return s
}

func (s *Store) SetPersistPath(path string) {
	s.mu.Lock()
	s.persistPath = path
	s.mu.Unlock()
	_ = s.Load()
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.persistPath == "" {
		return nil
	}

	data, err := os.ReadFile(s.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var loadedTasks map[string]Task
	if err := json.Unmarshal(data, &loadedTasks); err != nil {
		return err
	}

	if s.tasks == nil {
		s.tasks = make(map[string]Task)
	}

	for id, task := range loadedTasks {
		// If a task was running/downloading when app was closed, restore as paused
		if task.Status == StatusRunning || task.Status == "downloading" || task.Status == "merging" {
			task.Status = StatusPaused
			task.Message = "未完成任务 (已暂停，点击继续)"
		}
		s.tasks[id] = task
	}

	return nil
}

func (s *Store) persistLocked() error {
	if s.persistPath == "" {
		return nil
	}

	dir := filepath.Dir(s.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create task data directory: %w", err)
	}

	data, err := json.MarshalIndent(s.tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("encode tasks: %w", err)
	}

	tmpFile := fmt.Sprintf("%s.tmp.%d", s.persistPath, time.Now().UnixNano())
	defer os.Remove(tmpFile)
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("write task data: %w", err)
	}
	if err := os.Rename(tmpFile, s.persistPath); err != nil {
		return fmt.Errorf("replace task data: %w", err)
	}
	return nil
}
func (s *Store) Create(req CreateTaskRequest) (Task, error) {
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		return Task{}, fmt.Errorf("url is required")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return Task{}, fmt.Errorf("url must be a valid absolute address")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return Task{}, fmt.Errorf("url scheme must be http or https")
	}

	platform, supported := detectPlatform(parsed.Hostname())
	if !supported {
		return Task{}, fmt.Errorf("unsupported platform for MVP")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	task := Task{
		ID:            newTaskID(),
		URL:           rawURL,
		Platform:      platform,
		Status:        StatusQueued,
		Message:       "Task accepted and queued.",
		Progress:      0,
		Source:        strings.TrimSpace(req.Source),
		OutputDir:     strings.TrimSpace(req.OutputDir),
		ClientVersion: strings.TrimSpace(req.ClientVersion),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	persistErr := s.persistLocked()
	if persistErr != nil {
		delete(s.tasks, task.ID)
	}
	s.mu.Unlock()
	if persistErr != nil {
		return Task{}, persistErr
	}
	return task, nil
}

func (s *Store) Get(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	return task, ok
}

func (s *Store) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasksList := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasksList = append(tasksList, task)
	}
	return tasksList
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, ok := s.tasks[id]; ok {
		delete(s.tasks, id)
		if err := s.persistLocked(); err != nil {
			s.tasks[id] = task
			return false
		}
		return true
	}
	return false
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	previous := s.tasks
	s.tasks = make(map[string]Task)
	if err := s.persistLocked(); err != nil {
		s.tasks = previous
		return err
	}
	return nil
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.tasks)
}

func (s *Store) Summary() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := map[string]int{
		"total": len(s.tasks),
	}

	for _, task := range s.tasks {
		summary[string(task.Status)]++
	}

	return summary
}

func (s *Store) Update(id string, updateFn func(task *Task)) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, false
	}

	original := task
	updateFn(&task)
	task.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	s.tasks[id] = task
	if err := s.persistLocked(); err != nil {
		s.tasks[id] = original
		return Task{}, false
	}

	return task, true
}

func (s *Store) SetStatus(id string, status Status, message string) (Task, bool) {
	return s.Update(id, func(task *Task) {
		task.Status = status
		task.Message = strings.TrimSpace(message)
	})
}

func (s *Store) SetProgress(id string, progress int, message string) (Task, bool) {
	return s.Update(id, func(task *Task) {
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}

		task.Progress = progress
		if strings.TrimSpace(message) != "" {
			task.Message = strings.TrimSpace(message)
		}
	})
}

func (s *Store) Complete(id string, outputPath string, message string) (Task, bool) {
	return s.Update(id, func(task *Task) {
		task.Status = StatusCompleted
		task.Progress = 100
		task.Message = strings.TrimSpace(message)
		task.OutputPath = strings.TrimSpace(outputPath)
		task.ErrorCode = ""
	})
}

func (s *Store) Fail(id string, errorCode string, message string) (Task, bool) {
	return s.Update(id, func(task *Task) {
		task.Status = StatusFailed
		task.ErrorCode = strings.TrimSpace(errorCode)
		task.Message = strings.TrimSpace(message)
	})
}

func detectPlatform(host string) (string, bool) {
	normalizedHost := strings.ToLower(strings.TrimSpace(host))
	switch {
	case matchesHost(normalizedHost, "youtube.com"), matchesHost(normalizedHost, "youtu.be"):
		return "youtube", true
	case matchesHost(normalizedHost, "bilibili.com"), matchesHost(normalizedHost, "b23.tv"):
		return "bilibili", true
	case matchesHost(normalizedHost, "ggjav.com"), matchesHost(normalizedHost, "ggjav.tv"):
		return "ggjav", true
	case matchesHost(normalizedHost, "missav.ws"), matchesHost(normalizedHost, "missav.live"), matchesHost(normalizedHost, "missav.ai"), matchesHost(normalizedHost, "missav.com"):
		return "missav", true
	case matchesHost(normalizedHost, "rou.video"):
		return "rouvideo", true
	case matchesHost(normalizedHost, "tiktok.com"), matchesHost(normalizedHost, "vt.tiktok.com"):
		return "tiktok", true
	case matchesHost(normalizedHost, "x.com"), matchesHost(normalizedHost, "twitter.com"):
		return "x", true
	case matchesHost(normalizedHost, "instagram.com"), matchesHost(normalizedHost, "instagr.am"):
		return "instagram", true
	case matchesHost(normalizedHost, "facebook.com"), matchesHost(normalizedHost, "fb.watch"):
		return "facebook", true
	case matchesHost(normalizedHost, "twitch.tv"):
		return "twitch", true
	case matchesHost(normalizedHost, "vimeo.com"):
		return "vimeo", true
	case matchesHost(normalizedHost, "reddit.com"):
		return "reddit", true
	default:
		if normalizedHost != "" {
			cleanHost := strings.TrimPrefix(normalizedHost, "www.")
			return cleanHost, true
		}
		return "", false
	}
}

func matchesHost(host string, expected string) bool {
	normalizedHost := strings.ToLower(strings.TrimSpace(host))
	normalizedExpected := strings.ToLower(strings.TrimSpace(expected))

	return normalizedHost == normalizedExpected || strings.HasSuffix(normalizedHost, "."+normalizedExpected)
}

func newTaskID() string {
	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	return fmt.Sprintf("task_%d_%s", time.Now().Unix(), hex.EncodeToString(random))
}
