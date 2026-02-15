package main

import (
	"fmt"
	"os"

	"github.com/t3yamoto/gt/internal/command"
	"github.com/urfave/cli/v2"
)

var version = "dev"

func main() {
	app := &cli.App{
		Name:    "gt",
		Usage:   "Google Tasks CLI",
		Version: version,
		Commands: []*cli.Command{
			command.ListCommand(),
			command.AddCommand(),
			command.DoneCommand(),
			command.EditCommand(),
			command.DeleteCommand(),
		},
		Action: func(c *cli.Context) error {
			// Default action: run list command
			return command.ListCommand().Action(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
