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
				Value:   "@default",
				Usage:   "対象タスクリスト名",
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context

			taskClient, err := client.NewClient(ctx)
			if err != nil {
				return err
			}

			// Resolve task list name to ID
			taskListID, err := taskClient.ResolveTaskListID(ctx, c.String("tasklist"))
			if err != nil {
				return err
			}

			var taskID string

			if c.Args().Len() > 0 {
				// Direct mode: use provided task ID
				taskID = c.Args().First()
			} else {
				// Interactive mode
				tasks, err := taskClient.ListTasks(ctx, taskListID)
				if err != nil {
					return err
				}

				selected, err := selector.SelectTask(tasks)
				if err != nil {
					return err
				}
				taskID = selected.ID
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
