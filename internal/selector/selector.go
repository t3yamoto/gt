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
		return nil, fmt.Errorf("no tasks found")
	}

	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err != nil {
		return nil, fmt.Errorf("fzf is not installed")
	}

	// Build input for fzf
	// Format: index<TAB>display<TAB>preview_base64
	// Display format: [LIST] TITLE
	var input strings.Builder
	for i, t := range tasks {
		preview := editor.GenerateMarkdown(t, t.TaskListName)
		previewB64 := base64.StdEncoding.EncodeToString([]byte(preview))
		display := fmt.Sprintf("[%s] %s", t.TaskListName, t.Title)
		input.WriteString(fmt.Sprintf("%d\t%s\t%s\n", i, display, previewB64))
	}

	// Run fzf with shell
	// Use bat for syntax highlighting if available, otherwise fall back to base64 -d
	cmd := exec.Command("sh", "-c",
		`fzf --delimiter='\t' --with-nth=2 --preview='echo {3} | base64 -d | bat -l md --style=plain --color=always 2>/dev/null || echo {3} | base64 -d' --preview-window=right:50%:wrap`,
	)

	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				return nil, fmt.Errorf("cancelled")
			}
		}
		return nil, fmt.Errorf("failed to run fzf: %w", err)
	}

	// Parse selected line
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return nil, fmt.Errorf("cancelled")
	}

	parts := strings.Split(selected, "\t")
	if len(parts) < 1 {
		return nil, fmt.Errorf("failed to parse selection")
	}

	index, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse selection: %w", err)
	}

	if index < 0 || index >= len(tasks) {
		return nil, fmt.Errorf("invalid selection")
	}

	return tasks[index], nil
}
