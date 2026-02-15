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
		fmt.Fprintln(w, "タスクがありません。")
		return
	}

	// Calculate column widths
	idWidth := 8
	titleWidth := 24
	listWidth := 12
	dueWidth := 10
	statusWidth := 6

	// Print header
	fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
		padRight("ID", idWidth),
		padRight("TITLE", titleWidth),
		padRight("LIST", listWidth),
		padRight("DUE", dueWidth),
		"STATUS")
	fmt.Fprintln(w, strings.Repeat("-", idWidth+titleWidth+listWidth+dueWidth+statusWidth+10))

	// Print tasks
	for _, t := range tasks {
		title := truncate(t.Title, titleWidth)
		list := truncate(t.TaskListName, listWidth)
		due := t.Due
		if due == "" {
			due = "-"
		}
		status := "[ ]"
		if t.Status == "completed" {
			status = "[x]"
		}

		fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
			padRight(client.ShortID(t.ID), idWidth),
			padRight(title, titleWidth),
			padRight(list, listWidth),
			padRight(due, dueWidth),
			status)
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
