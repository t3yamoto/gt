package command

import (
	"fmt"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/selector"
	"github.com/urfave/cli/v2"
)

func DoneCommand() *cli.Command {
	return &cli.Command{
		Name:      "done",
		Usage:     "タスクを完了にする（引数なしでインタラクティブ選択）",
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
					// Search across all task lists
					task, err := taskClient.FindTask(ctx, taskID)
					if err != nil {
						return err
					}
					taskID = task.ID
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

				selected, err := selector.SelectTask(tasks)
				if err != nil {
					return err
				}
				taskID = selected.ID
				taskListID = selected.TaskListID
			}

			// Complete the task
			completed, err := taskClient.CompleteTask(ctx, taskListID, taskID)
			if err != nil {
				return err
			}

			fmt.Printf("タスクを完了しました: %s\n", completed.Title)
			return nil
		},
	}
}
