package selector

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/t3yamoto/gt/internal/client"
	"github.com/t3yamoto/gt/internal/editor"
)

// SelectTask presents an interactive task selector using fzf and returns the selected task
func SelectTask(tasks []*client.Task) (*client.Task, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("タスクがありません")
	}

	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err != nil {
		return nil, fmt.Errorf("fzf がインストールされていません")
	}

	// Build input for fzf
	// Format: index<TAB>title<TAB>preview_base64
	var input strings.Builder
	for i, t := range tasks {
		preview := editor.GenerateMarkdown(t, t.TaskListName)
		previewB64 := base64.StdEncoding.EncodeToString([]byte(preview))
		input.WriteString(fmt.Sprintf("%d\t%s\t%s\n", i, t.Title, previewB64))
	}

	// Run fzf with shell
	cmd := exec.Command("sh", "-c",
		`fzf --delimiter='\t' --with-nth=2 --preview='echo {3} | base64 -d' --preview-window=right:50%:wrap`,
	)

	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				return nil, fmt.Errorf("キャンセルされました")
			}
		}
		return nil, fmt.Errorf("fzf の実行に失敗しました: %w", err)
	}

	// Parse selected line
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return nil, fmt.Errorf("キャンセルされました")
	}

	parts := strings.Split(selected, "\t")
	if len(parts) < 1 {
		return nil, fmt.Errorf("選択の解析に失敗しました")
	}

	index, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("選択の解析に失敗しました: %w", err)
	}

	if index < 0 || index >= len(tasks) {
		return nil, fmt.Errorf("無効な選択です")
	}

	return tasks[index], nil
}
