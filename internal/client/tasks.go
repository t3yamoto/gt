package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/t3yamoto/gt/internal/auth"
	"github.com/t3yamoto/gt/internal/cache"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

// Task represents a task with additional metadata
type Task struct {
	ID           string
	Title        string
	Notes        string
	Due          string
	Status       string
	Completed    string
	TaskListID   string
	TaskListName string
}

// TaskList represents a task list
type TaskList struct {
	ID    string
	Title string
}

// Client wraps the Google Tasks API service
type Client struct {
	service *tasks.Service
	cache   *cache.Cache
}

// NewClient creates a new Tasks API client
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	service, err := tasks.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Tasks service: %w", err)
	}

	c, _ := cache.New() // Ignore cache initialization errors

	return &Client{service: service, cache: c}, nil
}

// GetTaskLists returns all task lists
func (c *Client) GetTaskLists(ctx context.Context) ([]*TaskList, error) {
	// Try cache first
	if c.cache != nil {
		if cached := c.cache.Load(); cached != nil && len(cached.TaskLists) > 0 {
			var lists []*TaskList
			for _, tl := range cached.TaskLists {
				lists = append(lists, &TaskList{
					ID:    tl.ID,
					Title: tl.Title,
				})
			}
			return lists, nil
		}
	}

	resp, err := c.service.Tasklists.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get task lists: %w", err)
	}

	var lists []*TaskList
	for _, tl := range resp.Items {
		lists = append(lists, &TaskList{
			ID:    tl.Id,
			Title: tl.Title,
		})
	}
	return lists, nil
}

// ResolveTaskListID resolves a task list name to its ID
// @default returns the default task list
func (c *Client) ResolveTaskListID(ctx context.Context, name string) (string, error) {
	if name == "@default" || name == "" {
		return "@default", nil
	}

	lists, err := c.GetTaskLists(ctx)
	if err != nil {
		return "", err
	}

	for _, list := range lists {
		if list.Title == name {
			return list.ID, nil
		}
	}

	return "", fmt.Errorf("task list '%s' not found", name)
}

// GetTaskListName returns the task list name for display
func (c *Client) GetTaskListName(ctx context.Context, id string) (string, error) {
	// Try to get from cache first
	if c.cache != nil {
		if cached := c.cache.Load(); cached != nil {
			for _, tl := range cached.TaskLists {
				if tl.ID == id {
					return tl.Title, nil
				}
			}
		}
	}

	if id == "@default" {
		tl, err := c.service.Tasklists.Get("@default").Context(ctx).Do()
		if err != nil {
			return "@default", nil
		}
		return tl.Title, nil
	}

	tl, err := c.service.Tasklists.Get(id).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get task list: %w", err)
	}
	return tl.Title, nil
}

// ListAllTasks returns all incomplete tasks from all task lists
func (c *Client) ListAllTasks(ctx context.Context) ([]*Task, error) {
	// Try cache first
	if c.cache != nil {
		if cached := c.cache.Load(); cached != nil && len(cached.Tasks) > 0 {
			var tasks []*Task
			for _, t := range cached.Tasks {
				tasks = append(tasks, &Task{
					ID:           t.ID,
					Title:        t.Title,
					Notes:        t.Notes,
					Due:          t.Due,
					Status:       t.Status,
					TaskListID:   t.TaskListID,
					TaskListName: t.TaskListName,
				})
			}
			return tasks, nil
		}
	}

	lists, err := c.GetTaskLists(ctx)
	if err != nil {
		return nil, err
	}

	var allTasks []*Task
	for _, list := range lists {
		tasks, err := c.listTasksFromList(ctx, list.ID, list.Title)
		if err != nil {
			return nil, err
		}
		allTasks = append(allTasks, tasks...)
	}

	// Save to cache
	c.saveToCache(lists, allTasks)

	return allTasks, nil
}

// ListTasks returns all incomplete tasks in the specified task list
func (c *Client) ListTasks(ctx context.Context, taskListID string) ([]*Task, error) {
	listName, _ := c.GetTaskListName(ctx, taskListID)
	return c.listTasksFromList(ctx, taskListID, listName)
}

func (c *Client) listTasksFromList(ctx context.Context, taskListID, taskListName string) ([]*Task, error) {
	resp, err := c.service.Tasks.List(taskListID).
		ShowCompleted(false).
		ShowHidden(false).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var tasksList []*Task
	for _, t := range resp.Items {
		tasksList = append(tasksList, convertTask(t, taskListID, taskListName))
	}
	return tasksList, nil
}

// GetTask returns a task by ID from a specific task list
func (c *Client) GetTask(ctx context.Context, taskListID, taskID string) (*Task, error) {
	// Try to resolve short ID
	fullID, err := c.ResolveTaskID(ctx, taskListID, taskID)
	if err != nil {
		return nil, err
	}

	t, err := c.service.Tasks.Get(taskListID, fullID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	listName, _ := c.GetTaskListName(ctx, taskListID)
	return convertTask(t, taskListID, listName), nil
}

// FindTask searches for a task by ID across all task lists
func (c *Client) FindTask(ctx context.Context, taskID string) (*Task, error) {
	// Try cache first for task lookup
	if c.cache != nil {
		if cached := c.cache.Load(); cached != nil {
			for _, t := range cached.Tasks {
				if t.ID == taskID || strings.HasPrefix(t.ID, taskID) {
					return &Task{
						ID:           t.ID,
						Title:        t.Title,
						Notes:        t.Notes,
						Due:          t.Due,
						Status:       t.Status,
						TaskListID:   t.TaskListID,
						TaskListName: t.TaskListName,
					}, nil
				}
			}
		}
	}

	lists, err := c.GetTaskLists(ctx)
	if err != nil {
		return nil, err
	}

	for _, list := range lists {
		task, err := c.GetTask(ctx, list.ID, taskID)
		if err == nil {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task '%s' not found", taskID)
}

// CreateTask creates a new task
func (c *Client) CreateTask(ctx context.Context, taskListID string, task *Task) (*Task, error) {
	newTask := &tasks.Task{
		Title: task.Title,
		Notes: task.Notes,
	}
	if task.Due != "" {
		newTask.Due = task.Due + "T00:00:00.000Z"
	}
	if task.Status == "completed" {
		newTask.Status = "completed"
	}

	t, err := c.service.Tasks.Insert(taskListID, newTask).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	listName, _ := c.GetTaskListName(ctx, taskListID)
	created := convertTask(t, taskListID, listName)

	// Add to cache
	c.addTaskToCache(created)

	return created, nil
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(ctx context.Context, taskListID string, task *Task) (*Task, error) {
	fullID, err := c.ResolveTaskID(ctx, taskListID, task.ID)
	if err != nil {
		return nil, err
	}

	// Get existing task first
	existing, err := c.service.Tasks.Get(taskListID, fullID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update fields
	existing.Title = task.Title
	existing.Notes = task.Notes
	if task.Due != "" {
		existing.Due = task.Due + "T00:00:00.000Z"
	} else {
		existing.Due = ""
	}
	if task.Status == "completed" {
		existing.Status = "completed"
	} else {
		existing.Status = "needsAction"
	}

	t, err := c.service.Tasks.Update(taskListID, fullID, existing).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	listName, _ := c.GetTaskListName(ctx, taskListID)
	updated := convertTask(t, taskListID, listName)

	// Update cache (remove if completed, update otherwise)
	if updated.Status == "completed" {
		c.removeTaskFromCache(updated.ID)
	} else {
		c.updateTaskInCache(updated)
	}

	return updated, nil
}

// CompleteTask marks a task as completed
func (c *Client) CompleteTask(ctx context.Context, taskListID, taskID string) (*Task, error) {
	fullID, err := c.ResolveTaskID(ctx, taskListID, taskID)
	if err != nil {
		return nil, err
	}

	existing, err := c.service.Tasks.Get(taskListID, fullID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	existing.Status = "completed"
	t, err := c.service.Tasks.Update(taskListID, fullID, existing).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to complete task: %w", err)
	}

	// Remove from cache (completed tasks are not cached)
	c.removeTaskFromCache(fullID)

	listName, _ := c.GetTaskListName(ctx, taskListID)
	return convertTask(t, taskListID, listName), nil
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(ctx context.Context, taskListID, taskID string) error {
	fullID, err := c.ResolveTaskID(ctx, taskListID, taskID)
	if err != nil {
		return err
	}

	if err := c.service.Tasks.Delete(taskListID, fullID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Remove from cache
	c.removeTaskFromCache(fullID)

	return nil
}

// ResolveTaskID resolves a short task ID to a full ID
func (c *Client) ResolveTaskID(ctx context.Context, taskListID, shortID string) (string, error) {
	// Try cache first
	if c.cache != nil {
		if cached := c.cache.Load(); cached != nil {
			for _, t := range cached.Tasks {
				if t.TaskListID == taskListID && strings.HasPrefix(t.ID, shortID) {
					return t.ID, nil
				}
			}
		}
	}

	// First try the ID as-is
	_, err := c.service.Tasks.Get(taskListID, shortID).Context(ctx).Do()
	if err == nil {
		return shortID, nil
	}

	// Search for matching task
	resp, err := c.service.Tasks.List(taskListID).
		ShowCompleted(true).
		ShowHidden(true).
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("failed to search tasks: %w", err)
	}

	var matches []string
	for _, t := range resp.Items {
		if strings.HasPrefix(t.Id, shortID) {
			matches = append(matches, t.Id)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("task '%s' not found", shortID)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("task ID '%s' matches multiple tasks, please use a longer ID", shortID)
	}
	return matches[0], nil
}

// ShortID returns the first 8 characters of a task ID
func ShortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// saveToCache saves task lists and tasks to cache
func (c *Client) saveToCache(lists []*TaskList, tasks []*Task) {
	if c.cache == nil {
		return
	}

	data := &cache.CacheData{}

	for _, l := range lists {
		data.TaskLists = append(data.TaskLists, cache.TaskListCache{
			ID:    l.ID,
			Title: l.Title,
		})
	}

	for _, t := range tasks {
		data.Tasks = append(data.Tasks, cache.TaskCache{
			ID:           t.ID,
			Title:        t.Title,
			Notes:        t.Notes,
			Due:          t.Due,
			Status:       t.Status,
			TaskListID:   t.TaskListID,
			TaskListName: t.TaskListName,
		})
	}

	c.cache.Save(data)
}

// addTaskToCache adds a task to the cache
func (c *Client) addTaskToCache(task *Task) {
	if c.cache == nil {
		return
	}
	c.cache.AddTask(cache.TaskCache{
		ID:           task.ID,
		Title:        task.Title,
		Notes:        task.Notes,
		Due:          task.Due,
		Status:       task.Status,
		TaskListID:   task.TaskListID,
		TaskListName: task.TaskListName,
	})
}

// updateTaskInCache updates a task in the cache
func (c *Client) updateTaskInCache(task *Task) {
	if c.cache == nil {
		return
	}
	c.cache.UpdateTask(cache.TaskCache{
		ID:           task.ID,
		Title:        task.Title,
		Notes:        task.Notes,
		Due:          task.Due,
		Status:       task.Status,
		TaskListID:   task.TaskListID,
		TaskListName: task.TaskListName,
	})
}

// removeTaskFromCache removes a task from the cache
func (c *Client) removeTaskFromCache(taskID string) {
	if c.cache == nil {
		return
	}
	c.cache.RemoveTask(taskID)
}

func convertTask(t *tasks.Task, taskListID, taskListName string) *Task {
	due := ""
	if t.Due != "" {
		// Extract date part only
		if len(t.Due) >= 10 {
			due = t.Due[:10]
		}
	}

	completed := ""
	if t.Completed != nil {
		completed = *t.Completed
	}

	return &Task{
		ID:           t.Id,
		Title:        t.Title,
		Notes:        t.Notes,
		Due:          due,
		Status:       t.Status,
		Completed:    completed,
		TaskListID:   taskListID,
		TaskListName: taskListName,
	}
}
