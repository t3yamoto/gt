# CLAUDE.md

Instructions for Claude Code and other AI assistants working on this project.

## Project Overview

`gt` is a Google Tasks CLI tool written in Go. It provides a simple interface to manage tasks from the terminal.

## Tech Stack

- **Language**: Go
- **CLI Framework**: urfave/cli v2
- **Authentication**: OAuth 2.0 (browser-based flow)
- **Interactive Selection**: fzf (external command, required)
- **Syntax Highlighting**: bat (optional, for preview)

## Architecture

```
main.go              → CLI app setup
internal/
  auth/              → OAuth 2.0 (credentials.json, token.json)
  cache/             → File cache (~/.cache/gt/cache.json, 5min TTL)
  client/            → Google Tasks API wrapper
  command/           → CLI commands (list, add, done, edit, delete)
  editor/            → $EDITOR integration with markdown frontmatter
  output/            → Table and JSON formatters
  selector/          → Fuzzy finder for interactive selection
```

## Key Constants (internal/client/constants.go)

```go
DefaultTaskList   = "@default"
StatusNeedsAction = "needsAction"
StatusCompleted   = "completed"
```

## Common Patterns

### Task Resolution

Use `command.ResolveTask()` for commands that need to find a task:
```go
task, taskListID, err := ResolveTask(ctx, client, taskID, taskListName)
```

### Cache Conversion

Use helpers in `client/tasks.go`:
```go
taskFromCache(c cache.TaskCache) *Task
taskToCache(t *Task) cache.TaskCache
```

### Date Handling

Use helpers in `client/constants.go`:
```go
FormatDueDate(date string) string  // "2024-01-01" → "2024-01-01T00:00:00.000Z"
ParseDueDate(apiDate string) string // "2024-01-01T00:00:00.000Z" → "2024-01-01"
```

## Build & Run

```bash
go build          # Build
go run . list     # Run without installing
go install .      # Install to $GOBIN
```

## Important Notes

- Tasks API does not support timeboxing (time slots are stored in Google Calendar)
- Task IDs are displayed as 8-character short IDs (`client.ShortID()`)
- Cache is updated (not invalidated) on write operations
- Error handling for `GetTaskListName` is intentionally ignored (fallback to ID)

## File Locations

- Config: `~/.config/gt/`
  - `credentials.json` - OAuth client credentials
  - `token.json` - OAuth tokens
- Cache: `~/.cache/gt/cache.json`

## Language

- Code comments: English
- Commit messages: English
- CLI messages: English
