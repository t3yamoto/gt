package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/t3yamoto/gt/internal/client"
)

// PrintTasksTable prints tasks in a table format
func PrintTasksTable(w io.Writer, tasks []*client.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "タスクがありません。")
		return
	}

	// Calculate column widths
	idWidth := 8
	titleWidth := 30
	dueWidth := 10
	statusWidth := 6

	// Print header
	fmt.Fprintf(w, "%-*s  %-*s  %-*s  %s\n",
		idWidth, "ID",
		titleWidth, "TITLE",
		dueWidth, "DUE",
		"STATUS")
	fmt.Fprintln(w, strings.Repeat("-", idWidth+titleWidth+dueWidth+statusWidth+8))

	// Print tasks
	for _, t := range tasks {
		title := truncate(t.Title, titleWidth)
		due := t.Due
		if due == "" {
			due = "-"
		}
		status := "[ ]"
		if t.Status == "completed" {
			status = "[x]"
		}

		fmt.Fprintf(w, "%-*s  %-*s  %-*s  %s\n",
			idWidth, client.ShortID(t.ID),
			titleWidth, title,
			dueWidth, due,
			status)
	}
}

// truncate truncates a string to the specified width, adding "..." if necessary
func truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}
