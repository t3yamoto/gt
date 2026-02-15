package command

import (
	"fmt"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/selector"
	"github.com/urfave/cli/v2"
)

func DeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "タスクを削除（引数なしでインタラクティブ選択）",
		ArgsUsage: "[task-id]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "tasklist",
				Aliases: []string{"l"},
				Usage:   "対象タスクリスト名（省略時は全リスト）",
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context

			taskClient, err := client.NewClient(ctx)
			if err != nil {
				return err
			}

			var taskID string
			var taskTitle string
			var taskListID string

			if c.Args().Len() > 0 {
				// Direct mode: use provided task ID
				taskID = c.Args().First()
				if c.String("tasklist") != "" {
					taskListID, err = taskClient.ResolveTaskListID(ctx, c.String("tasklist"))
					if err != nil {
						return err
					}
				} else {
					taskListID = "@default"
				}
				// Get task to display title
				task, err := taskClient.GetTask(ctx, taskListID, taskID)
				if err != nil {
					return err
				}
				taskTitle = task.Title
				taskID = task.ID // Use full ID
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

				selected, err := selector.SelectTask(tasks)
				if err != nil {
					return err
				}
				taskID = selected.ID
				taskTitle = selected.Title
				taskListID = selected.TaskListID
			}

			// Delete the task
			if err := taskClient.DeleteTask(ctx, taskListID, taskID); err != nil {
				return err
			}

			fmt.Printf("タスクを削除しました: %s\n", taskTitle)
			return nil
		},
	}
}
