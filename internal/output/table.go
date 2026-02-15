package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/t3yamoto/gt/internal/client"
)

// PrintTasksTable prints tasks in a table format
func PrintTasksTable(w io.Writer, tasks []*client.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks found.")
		return
	}

	// Calculate column widths
	idWidth := 8
	listWidth := 16
	titleWidth := 32
	dueWidth := 10

	// Print header
	fmt.Fprintf(w, "%s  %s  %s  %s\n",
		padRight("ID", idWidth),
		padRight("LIST", listWidth),
		padRight("TITLE", titleWidth),
		"DUE")
	fmt.Fprintln(w, strings.Repeat("-", idWidth+listWidth+titleWidth+dueWidth+6))

	// Print tasks
	for _, t := range tasks {
		list := truncate(t.TaskListName, listWidth)
		title := truncate(t.Title, titleWidth)
		due := t.Due
		if due == "" {
			due = "-"
		}

		fmt.Fprintf(w, "%s  %s  %s  %s\n",
			padRight(client.ShortID(t.ID), idWidth),
			padRight(list, listWidth),
			padRight(title, titleWidth),
			due)
	}
}

// padRight pads a string to the specified display width
func padRight(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// truncate truncates a string to the specified display width, adding "..." if necessary
func truncate(s string, width int) string {
	w := runewidth.StringWidth(s)
	if w <= width {
		return s
	}
	if width <= 3 {
		return runewidth.Truncate(s, width, "")
	}
	return runewidth.Truncate(s, width, "...")
}
