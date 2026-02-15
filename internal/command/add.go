package command

import (
	"fmt"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/editor"
	"github.com/urfave/cli/v2"
)

func AddCommand() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a task (opens editor if no argument)",
		ArgsUsage: "[title]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "tasklist",
				Aliases: []string{"l"},
				Value:   "@default",
				Usage:   "Target task list name",
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

			// Get actual task list name for display
			displayName, _ := taskClient.GetTaskListName(ctx, taskListID)

			var newTask *client.Task

			if c.Args().Len() > 0 {
				// Simple mode: create task with title from argument
				title := c.Args().First()
				newTask = &client.Task{
					Title: title,
				}
			} else {
				// Editor mode
				initialContent := editor.GenerateEmptyMarkdown(displayName)
				editedContent, err := editor.Open(initialContent)
				if err != nil {
					return err
				}

				if editor.IsEmpty(editedContent) {
					fmt.Println("Cancelled.")
					return nil
				}

				parsed, err := editor.ParseMarkdown(editedContent)
				if err != nil {
					return err
				}

				// If task list was changed in editor, resolve new task list
				editorTaskList := parsed.GetTaskListName()
				if editorTaskList != displayName && editorTaskList != "@default" {
					taskListID, err = taskClient.ResolveTaskListID(ctx, editorTaskList)
					if err != nil {
						return err
					}
				}

				newTask = parsed.ToTask()
			}

			// Create task
			created, err := taskClient.CreateTask(ctx, taskListID, newTask)
			if err != nil {
				return err
			}

			fmt.Printf("Task added: %s (ID: %s)\n", created.Title, client.ShortID(created.ID))
			return nil
		},
	}
}
