package command

import (
	"fmt"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/editor"
	"github.com/t3yamoto/gt/internal/selector"
	"github.com/urfave/cli/v2"
)

func EditCommand() *cli.Command {
	return &cli.Command{
		Name:      "edit",
		Usage:     "Edit a task (interactive selection if no argument)",
		ArgsUsage: "[task-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "tasklist",
				Aliases: []string{"l"},
				Usage:   "Target task list name (default: all lists)",
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context

			taskClient, err := client.NewClient(ctx)
			if err != nil {
				return err
			}

			var task *client.Task
			var taskListID string

			if c.Args().Len() > 0 {
				// Direct mode: get task by ID
				taskID := c.Args().First()
				if c.String("tasklist") != "" {
					taskListID, err = taskClient.ResolveTaskListID(ctx, c.String("tasklist"))
					if err != nil {
						return err
					}
					task, err = taskClient.GetTask(ctx, taskListID, taskID)
					if err != nil {
						return err
					}
				} else {
					// Search across all task lists
					task, err = taskClient.FindTask(ctx, taskID)
					if err != nil {
						return err
					}
					taskListID = task.TaskListID
				}
			} else {
				// Interactive mode
				var tasks []*client.Task
				if c.String("tasklist") != "" {
					taskListID, err = taskClient.ResolveTaskListID(ctx, c.String("tasklist"))
					if err != nil {
						return err
					}
					tasks, err = taskClient.ListTasks(ctx, taskListID)
				} else {
					tasks, err = taskClient.ListAllTasks(ctx)
				}
				if err != nil {
					return err
				}

				task, err = selector.SelectTask(tasks)
				if err != nil {
					return err
				}
				taskListID = task.TaskListID
			}

			// Open in editor
			initialContent := editor.GenerateMarkdown(task, task.TaskListName)
			editedContent, err := editor.Open(initialContent)
			if err != nil {
				return err
			}

			// Parse edited content
			parsed, err := editor.ParseMarkdown(editedContent)
			if err != nil {
				return err
			}

			// Update task
			updatedTask := parsed.ToTask()
			updatedTask.ID = task.ID

			// Check if task list was changed
			editorTaskList := parsed.GetTaskListName()
			if editorTaskList != "@default" && editorTaskList != task.TaskListName {
				newTaskListID, err := taskClient.ResolveTaskListID(ctx, editorTaskList)
				if err != nil {
					return err
				}
				// Need to delete from old list and create in new list
				if err := taskClient.DeleteTask(ctx, taskListID, task.ID); err != nil {
					return fmt.Errorf("failed to move task: %w", err)
				}
				created, err := taskClient.CreateTask(ctx, newTaskListID, updatedTask)
				if err != nil {
					return fmt.Errorf("failed to move task: %w", err)
				}
				fmt.Printf("Task updated: %s (new ID: %s)\n", created.Title, client.ShortID(created.ID))
				return nil
			}

			// Update in same list
			updated, err := taskClient.UpdateTask(ctx, taskListID, updatedTask)
			if err != nil {
				return err
			}

			fmt.Printf("Task updated: %s\n", updated.Title)
			return nil
		},
	}
}
