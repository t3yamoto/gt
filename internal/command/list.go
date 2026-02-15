package command

import (
	"os"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/output"
	"github.com/urfave/cli/v2"
)

func ListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "タスク一覧を表示",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "tasklist",
				Aliases: []string{"l"},
				Usage:   "対象タスクリスト名（省略時は全リスト）",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "JSON形式で出力",
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context

			taskClient, err := client.NewClient(ctx)
			if err != nil {
				return err
			}

			var tasks []*client.Task

			if c.String("tasklist") != "" {
				// Specific task list
				taskListID, err := taskClient.ResolveTaskListID(ctx, c.String("tasklist"))
				if err != nil {
					return err
				}
				tasks, err = taskClient.ListTasks(ctx, taskListID)
				if err != nil {
					return err
				}
			} else {
				// All task lists
				tasks, err = taskClient.ListAllTasks(ctx)
				if err != nil {
					return err
				}
			}

			// Output
			if c.Bool("json") {
				return output.PrintTasksJSON(os.Stdout, tasks)
			}
			output.PrintTasksTable(os.Stdout, tasks)
			return nil
		},
	}
}
