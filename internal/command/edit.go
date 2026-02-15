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
		Usage:     "タスクを編集（引数なしでインタラクティブ選択）",
		ArgsUsage: "[task-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "tasklist",
				Aliases: []string{"l"},
				Value:   "@default",
				Usage:   "対象タスクリスト名",
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context
			taskListName := c.String("tasklist")

			taskClient, err := client.NewClient(ctx)
			if err != nil {
				return err
			}

			// Resolve task list name to ID
			taskListID, err := taskClient.ResolveTaskListID(ctx, taskListName)
			if err != nil {
				return err
			}

			var task *client.Task

			if c.Args().Len() > 0 {
				// Direct mode: get task by ID
				taskID := c.Args().First()
				task, err = taskClient.GetTask(ctx, taskListID, taskID)
				if err != nil {
					return err
				}
			} else {
				// Interactive mode
				tasks, err := taskClient.ListTasks(ctx, taskListID)
				if err != nil {
					return err
				}

				task, err = selector.SelectTask(tasks)
				if err != nil {
					return err
				}
			}

			// Get task list display name
			displayName, _ := taskClient.GetTaskListName(ctx, taskListID)

			// Open in editor
			initialContent := editor.GenerateMarkdown(task, displayName)
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
			if editorTaskList != taskListName && editorTaskList != "@default" && editorTaskList != displayName {
				newTaskListID, err := taskClient.ResolveTaskListID(ctx, editorTaskList)
				if err != nil {
					return err
				}
				// Need to delete from old list and create in new list
				if err := taskClient.DeleteTask(ctx, taskListID, task.ID); err != nil {
					return fmt.Errorf("タスクの移動に失敗しました: %w", err)
				}
				created, err := taskClient.CreateTask(ctx, newTaskListID, updatedTask)
				if err != nil {
					return fmt.Errorf("タスクの移動に失敗しました: %w", err)
				}
				fmt.Printf("タスクを更新しました: %s (新しいID: %s)\n", created.Title, client.ShortID(created.ID))
				return nil
			}

			// Update in same list
			updated, err := taskClient.UpdateTask(ctx, taskListID, updatedTask)
			if err != nil {
				return err
			}

			fmt.Printf("タスクを更新しました: %s\n", updated.Title)
			return nil
		},
	}
}
