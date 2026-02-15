package command

import (
	"fmt"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/urfave/cli/v2"
)

func DeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Usage:     "Delete a task (interactive selection if no argument)",
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

			task, taskListID, err := ResolveTask(ctx, taskClient, c.Args().First(), c.String("tasklist"))
			if err != nil {
				return err
			}

			if err := taskClient.DeleteTask(ctx, taskListID, task.ID); err != nil {
				return err
			}

			fmt.Printf("Task deleted: %s\n", task.Title)
			return nil
		},
	}
}
