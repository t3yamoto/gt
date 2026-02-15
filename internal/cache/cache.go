package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDir  = ".cache/gt"
	cacheFile = "cache.json"
	cacheTTL  = 5 * time.Minute
)

// TaskListCache represents a cached task list
type TaskListCache struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// TaskCache represents a cached task
type TaskCache struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Notes        string `json:"notes"`
	Due          string `json:"due"`
	Status       string `json:"status"`
	TaskListID   string `json:"task_list_id"`
	TaskListName string `json:"task_list_name"`
}

// CacheData represents the entire cache structure
type CacheData struct {
	TaskLists []TaskListCache `json:"task_lists"`
	Tasks     []TaskCache     `json:"tasks"`
	CachedAt  time.Time       `json:"cached_at"`
}

// Cache provides file-based caching
type Cache struct {
	path string
}

// New creates a new Cache instance
func New() (*Cache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, cacheDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	return &Cache{
		path: filepath.Join(dir, cacheFile),
	}, nil
}

// Load reads cache from file, returns nil if expired or not found
func (c *Cache) Load() *CacheData {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil
	}

	var cache CacheData
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}

	// Check TTL
	if time.Since(cache.CachedAt) > cacheTTL {
		return nil
	}

	return &cache
}

// Save writes cache to file
func (c *Cache) Save(data *CacheData) error {
	data.CachedAt = time.Now()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, jsonData, 0600)
}

// AddTask adds a task to the cache
func (c *Cache) AddTask(task TaskCache) {
	data := c.Load()
	if data == nil {
		return // No cache to update
	}

	data.Tasks = append(data.Tasks, task)
	c.Save(data)
}

// UpdateTask updates a task in the cache
func (c *Cache) UpdateTask(task TaskCache) {
	data := c.Load()
	if data == nil {
		return
	}

	for i, t := range data.Tasks {
		if t.ID == task.ID {
			data.Tasks[i] = task
			c.Save(data)
			return
		}
	}
}

// RemoveTask removes a task from the cache
func (c *Cache) RemoveTask(taskID string) {
	data := c.Load()
	if data == nil {
		return
	}

	for i, t := range data.Tasks {
		if t.ID == taskID {
			data.Tasks = append(data.Tasks[:i], data.Tasks[i+1:]...)
			c.Save(data)
			return
		}
	}
}

// Invalidate removes the cache file
func (c *Cache) Invalidate() error {
	if err := os.Remove(c.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
