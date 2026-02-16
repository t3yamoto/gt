package editor

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/t3yamoto/gt/internal/client"
	"gopkg.in/yaml.v3"
)

// TaskFrontMatter represents the YAML front matter for a task
type TaskFrontMatter struct {
	Title     string `yaml:"title"`
	Due       string `yaml:"due,omitempty"`
	TaskList  string `yaml:"tasklist,omitempty"`
	Completed bool   `yaml:"completed,omitempty"`
}

// TaskMarkdown represents a parsed task markdown document
type TaskMarkdown struct {
	FrontMatter TaskFrontMatter
	Body        string // Notes content
}

// templateData holds data for the markdown template
type templateData struct {
	Title     string
	Due       string
	TaskList  string
	Completed bool
	Notes     string
}

var taskTemplate = template.Must(template.New("task").Parse(`---
title: {{.Title}}
due: {{.Due}}
tasklist: {{.TaskList}}
completed: {{.Completed}}
---

{{.Notes}}`))

// GenerateMarkdown creates a markdown document from a task
func GenerateMarkdown(task *client.Task, taskListName string) string {
	data := templateData{
		Title:     task.Title,
		Due:       task.Due,
		TaskList:  taskListName,
		Completed: task.Status == client.StatusCompleted,
		Notes:     task.Notes,
	}

	var buf bytes.Buffer
	taskTemplate.Execute(&buf, data)
	return buf.String()
}

// GenerateEmptyMarkdown creates an empty markdown template
func GenerateEmptyMarkdown(taskListName string) string {
	data := templateData{
		Title:     "",
		Due:       "",
		TaskList:  taskListName,
		Completed: false,
		Notes:     "",
	}

	var buf bytes.Buffer
	taskTemplate.Execute(&buf, data)
	return buf.String()
}

// ParseMarkdown parses a markdown document into TaskMarkdown
func ParseMarkdown(content string) (*TaskMarkdown, error) {
	content = strings.TrimSpace(content)

	// Check for front matter
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("front matter not found, please start with ---")
	}

	// Find the end of front matter
	rest := content[3:] // Skip opening ---
	rest = strings.TrimLeft(rest, "\n\r")

	endIdx := strings.Index(rest, "---")
	if endIdx == -1 {
		return nil, fmt.Errorf("front matter end (---) not found")
	}

	frontMatterStr := strings.TrimSpace(rest[:endIdx])
	body := strings.TrimSpace(rest[endIdx+3:])

	// Parse front matter YAML
	var fm TaskFrontMatter
	if err := yaml.Unmarshal([]byte(frontMatterStr), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}

	if fm.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	return &TaskMarkdown{
		FrontMatter: fm,
		Body:        body,
	}, nil
}

// ToTask converts TaskMarkdown to a Task
func (tm *TaskMarkdown) ToTask() *client.Task {
	status := client.StatusNeedsAction
	if tm.FrontMatter.Completed {
		status = client.StatusCompleted
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
	return client.DefaultTaskList
}
