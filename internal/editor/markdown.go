package editor

import (
	"fmt"
	"strings"

	"github.com/t3yamoto/gt/internal/client"
	"gopkg.in/yaml.v3"
)

// TaskFrontMatter represents the YAML front matter for a task
type TaskFrontMatter struct {
	Title    string `yaml:"title"`
	Due      string `yaml:"due,omitempty"`
	TaskList string `yaml:"tasklist,omitempty"`
	Done     bool   `yaml:"done,omitempty"`
}

// TaskMarkdown represents a parsed task markdown document
type TaskMarkdown struct {
	FrontMatter TaskFrontMatter
	Body        string // Notes content
}

// GenerateMarkdown creates a markdown document from a task
func GenerateMarkdown(task *client.Task, taskListName string) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("title: %s\n", task.Title))
	sb.WriteString(fmt.Sprintf("due: %s\n", task.Due))
	sb.WriteString(fmt.Sprintf("tasklist: %s\n", taskListName))
	sb.WriteString(fmt.Sprintf("done: %t\n", task.Status == "completed"))
	sb.WriteString("---\n\n")
	sb.WriteString(task.Notes)

	return sb.String()
}

// GenerateEmptyMarkdown creates an empty markdown template
func GenerateEmptyMarkdown(taskListName string) string {
	var sb strings.Builder

	sb.WriteString("---\n")
	sb.WriteString("title: \n")
	sb.WriteString("due: \n")
	sb.WriteString(fmt.Sprintf("tasklist: %s\n", taskListName))
	sb.WriteString("done: false\n")
	sb.WriteString("---\n\n")

	return sb.String()
}

// ParseMarkdown parses a markdown document into TaskMarkdown
func ParseMarkdown(content string) (*TaskMarkdown, error) {
	content = strings.TrimSpace(content)

	// Check for front matter
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("front matter が見つかりません。--- で始めてください")
	}

	// Find the end of front matter
	rest := content[3:] // Skip opening ---
	rest = strings.TrimLeft(rest, "\n\r")

	endIdx := strings.Index(rest, "---")
	if endIdx == -1 {
		return nil, fmt.Errorf("front matter の終了 (---) が見つかりません")
	}

	frontMatterStr := strings.TrimSpace(rest[:endIdx])
	body := strings.TrimSpace(rest[endIdx+3:])

	// Parse front matter YAML
	var fm TaskFrontMatter
	if err := yaml.Unmarshal([]byte(frontMatterStr), &fm); err != nil {
		return nil, fmt.Errorf("front matter のパースに失敗しました: %w", err)
	}

	if fm.Title == "" {
		return nil, fmt.Errorf("title は必須です")
	}

	return &TaskMarkdown{
		FrontMatter: fm,
		Body:        body,
	}, nil
}

// ToTask converts TaskMarkdown to a Task
func (tm *TaskMarkdown) ToTask() *client.Task {
	status := "needsAction"
	if tm.FrontMatter.Done {
		status = "completed"
	}

	return &client.Task{
		Title:        tm.FrontMatter.Title,
		Notes:        tm.Body,
		Due:          tm.FrontMatter.Due,
		Status:       status,
		TaskListName: tm.FrontMatter.TaskList,
	}
}

// GetTaskListName returns the task list name from front matter, or default
func (tm *TaskMarkdown) GetTaskListName() string {
	if tm.FrontMatter.TaskList != "" {
		return tm.FrontMatter.TaskList
	}
	return "@default"
}
