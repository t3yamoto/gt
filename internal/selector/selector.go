package selector

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/editor"
)

// SelectTask presents an interactive task selector and returns the selected task
func SelectTask(tasks []*client.Task) (*client.Task, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("タスクがありません")
	}

	idx, err := fuzzyfinder.Find(
		tasks,
		func(i int) string {
			return tasks[i].Title
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return editor.GenerateMarkdown(tasks[i], tasks[i].TaskListName)
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
