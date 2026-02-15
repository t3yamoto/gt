package selector

import (
	"fmt"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/t3yamoto/gt/internal/client"
)

// SelectTask presents an interactive task selector and returns the selected task
func SelectTask(tasks []*client.Task) (*client.Task, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("タスクがありません")
	}

	idx, err := fuzzyfinder.Find(
		tasks,
		func(i int) string {
			t := tasks[i]
			due := t.Due
			if due == "" {
				due = "----------"
			}
			status := "[ ]"
			if t.Status == "completed" {
				status = "[x]"
			}
			return fmt.Sprintf("%s  %s  %s", t.Title, due, status)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return formatPreview(tasks[i], w, h)
		}),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return nil, fmt.Errorf("キャンセルされました")
		}
		return nil, fmt.Errorf("選択に失敗しました: %w", err)
	}

	return tasks[idx], nil
}

// formatPreview formats the task preview for the fuzzyfinder preview window
func formatPreview(task *client.Task, width, height int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Title: %s\n", task.Title))
	sb.WriteString(fmt.Sprintf("ID: %s\n", client.ShortID(task.ID)))

	if task.Due != "" {
		sb.WriteString(fmt.Sprintf("Due: %s\n", task.Due))
	}

	status := "未完了"
	if task.Status == "completed" {
		status = "完了"
	}
	sb.WriteString(fmt.Sprintf("Status: %s\n", status))

	if task.Notes != "" {
		sb.WriteString("\n--- Notes ---\n")
		// Truncate notes if too long
		notes := task.Notes
		maxLines := height - 6
		if maxLines < 1 {
			maxLines = 1
		}
		lines := strings.Split(notes, "\n")
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			lines = append(lines, "...")
		}
		sb.WriteString(strings.Join(lines, "\n"))
	}

	return sb.String()
}
