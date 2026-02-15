package command

import (
	"context"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/selector"
)

// ResolveTask resolves a task either by ID or interactive selection
// If taskID is provided, it searches for the task (optionally within taskListName)
// If taskID is empty, it presents an interactive selector
func ResolveTask(ctx context.Context, c *client.Client, taskID, taskListName string) (*client.Task, string, error) {
	if taskID != "" {
		return resolveTaskByID(ctx, c, taskID, taskListName)
	}
	return resolveTaskInteractive(ctx, c, taskListName)
}

func resolveTaskByID(ctx context.Context, c *client.Client, taskID, taskListName string) (*client.Task, string, error) {
	if taskListName != "" {
		taskListID, err := c.ResolveTaskListID(ctx, taskListName)
		if err != nil {
			return nil, "", err
		}
		task, err := c.GetTask(ctx, taskListID, taskID)
		if err != nil {
			return nil, "", err
		}
		return task, taskListID, nil
	}

	// Search across all task lists
	task, err := c.FindTask(ctx, taskID)
	if err != nil {
		return nil, "", err
	}
	return task, task.TaskListID, nil
}

func resolveTaskInteractive(ctx context.Context, c *client.Client, taskListName string) (*client.Task, string, error) {
	var tasks []*client.Task
	var taskListID string
	var err error

	if taskListName != "" {
		taskListID, err = c.ResolveTaskListID(ctx, taskListName)
		if err != nil {
			return nil, "", err
		}
		tasks, err = c.ListTasks(ctx, taskListID)
	} else {
		tasks, err = c.ListAllTasks(ctx)
	}
	if err != nil {
		return nil, "", err
	}

	task, err := selector.SelectTask(tasks)
	if err != nil {
		return nil, "", err
	}
	return task, task.TaskListID, nil
}
