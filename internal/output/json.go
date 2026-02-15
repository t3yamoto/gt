package output

import (
	"encoding/json"
	"io"

	"github.com/t3yamoto/gt/internal/client"
)

// TaskJSON represents a task in JSON format
type TaskJSON struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Notes        string `json:"notes,omitempty"`
	Due          string `json:"due,omitempty"`
	Status       string `json:"status"`
	TaskListID   string `json:"tasklistId"`
	TaskListName string `json:"tasklistName"`
}

// PrintTasksJSON prints tasks in JSON format
func PrintTasksJSON(w io.Writer, tasks []*client.Task) error {
	jsonTasks := make([]TaskJSON, len(tasks))
	for i, t := range tasks {
		jsonTasks[i] = TaskJSON{
			ID:           t.ID,
			Title:        t.Title,
			Notes:        t.Notes,
			Due:          t.Due,
			Status:       t.Status,
			TaskListID:   t.TaskListID,
			TaskListName: t.TaskListName,
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonTasks)
}
