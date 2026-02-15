package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/t3yamoto/gt/internal/auth"
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
}

// NewClient creates a new Tasks API client
func NewClient(ctx context.Context) (*Client, error) {
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	service, err := tasks.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("Tasks サービスの初期化に失敗しました: %w", err)
	}

	return &Client{service: service}, nil
}

// GetTaskLists returns all task lists
func (c *Client) GetTaskLists(ctx context.Context) ([]*TaskList, error) {
	resp, err := c.service.Tasklists.List().Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("タスクリストの取得に失敗しました: %w", err)
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

	return "", fmt.Errorf("タスクリスト '%s' が見つかりません", name)
}

// GetTaskListName returns the task list name for display
func (c *Client) GetTaskListName(ctx context.Context, id string) (string, error) {
	if id == "@default" {
		tl, err := c.service.Tasklists.Get("@default").Context(ctx).Do()
		if err != nil {
			return "@default", nil
		}
		return tl.Title, nil
	}

	tl, err := c.service.Tasklists.Get(id).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("タスクリストの取得に失敗しました: %w", err)
	}
	return tl.Title, nil
}

// ListAllTasks returns all incomplete tasks from all task lists
func (c *Client) ListAllTasks(ctx context.Context) ([]*Task, error) {
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
		return nil, fmt.Errorf("タスクの取得に失敗しました: %w", err)
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
		return nil, fmt.Errorf("タスクの取得に失敗しました: %w", err)
	}
	listName, _ := c.GetTaskListName(ctx, taskListID)
	return convertTask(t, taskListID, listName), nil
}

// FindTask searches for a task by ID across all task lists
func (c *Client) FindTask(ctx context.Context, taskID string) (*Task, error) {
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

	return nil, fmt.Errorf("タスク '%s' が見つかりません", taskID)
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
		return nil, fmt.Errorf("タスクの作成に失敗しました: %w", err)
	}
	listName, _ := c.GetTaskListName(ctx, taskListID)
	return convertTask(t, taskListID, listName), nil
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
		return nil, fmt.Errorf("タスクの取得に失敗しました: %w", err)
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
		return nil, fmt.Errorf("タスクの更新に失敗しました: %w", err)
	}
	listName, _ := c.GetTaskListName(ctx, taskListID)
	return convertTask(t, taskListID, listName), nil
}

// CompleteTask marks a task as completed
func (c *Client) CompleteTask(ctx context.Context, taskListID, taskID string) (*Task, error) {
	fullID, err := c.ResolveTaskID(ctx, taskListID, taskID)
	if err != nil {
		return nil, err
	}

	existing, err := c.service.Tasks.Get(taskListID, fullID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("タスクの取得に失敗しました: %w", err)
	}

	existing.Status = "completed"
	t, err := c.service.Tasks.Update(taskListID, fullID, existing).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("タスクの完了に失敗しました: %w", err)
	}
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
		return fmt.Errorf("タスクの削除に失敗しました: %w", err)
	}
	return nil
}

// ResolveTaskID resolves a short task ID to a full ID
func (c *Client) ResolveTaskID(ctx context.Context, taskListID, shortID string) (string, error) {
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
		return "", fmt.Errorf("タスクの検索に失敗しました: %w", err)
	}

	var matches []string
	for _, t := range resp.Items {
		if strings.HasPrefix(t.Id, shortID) {
			matches = append(matches, t.Id)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("タスク '%s' が見つかりません", shortID)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("タスクID '%s' が複数のタスクにマッチします。より長いIDを指定してください", shortID)
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
