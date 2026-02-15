package editor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Open opens the specified content in the user's editor and returns the edited content
func Open(content string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default fallback
	}

	// Create temp file
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "gt-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write initial content
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Open editor
	cmd := exec.Command(editor, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to start editor: %w", err)
	}

	// Read edited content
	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	return string(edited), nil
}

// GetEditorName returns the name of the configured editor
func GetEditorName() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return "vi"
	}
	return filepath.Base(editor)
}

// IsEmpty checks if the content is effectively empty (only whitespace/frontmatter delimiters)
func IsEmpty(content string) bool {
	// Remove front matter delimiters and whitespace
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "---")
	content = strings.TrimSuffix(content, "---")
	content = strings.TrimSpace(content)
	return content == ""
}
