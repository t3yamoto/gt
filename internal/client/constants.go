package client

// Task list constants
const (
	DefaultTaskList = "@default"
)

// Task status constants
const (
	StatusNeedsAction = "needsAction"
	StatusCompleted   = "completed"
)

// FormatDueDate converts a date string (YYYY-MM-DD) to API format
func FormatDueDate(date string) string {
	if date == "" {
		return ""
	}
	return date + "T00:00:00.000Z"
}

// ParseDueDate extracts the date part (YYYY-MM-DD) from API format
func ParseDueDate(apiDate string) string {
	if apiDate == "" || len(apiDate) < 10 {
		return ""
	}
	return apiDate[:10]
}
