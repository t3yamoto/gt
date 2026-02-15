# Contributing to gt

## Development Setup

### Prerequisites

- Go 1.21 or later
- Google Cloud project with Tasks API enabled
- OAuth 2.0 credentials

### Build

```bash
git clone https://github.com/t3yamoto/gt.git
cd gt
go build
```

### Run

```bash
go run . list
```

### Install locally

```bash
go install .
```

## Project Structure

```
gt/
├── main.go                    # Entry point
├── internal/
│   ├── auth/
│   │   └── oauth.go           # OAuth 2.0 authentication
│   ├── cache/
│   │   └── cache.go           # File-based caching
│   ├── client/
│   │   ├── constants.go       # Constants and helpers
│   │   └── tasks.go           # Google Tasks API client
│   ├── command/
│   │   ├── add.go             # add command
│   │   ├── delete.go          # delete command
│   │   ├── done.go            # done command
│   │   ├── edit.go            # edit command
│   │   ├── list.go            # list command
│   │   └── resolver.go        # Task resolution helper
│   ├── editor/
│   │   ├── editor.go          # $EDITOR integration
│   │   └── markdown.go        # Markdown/frontmatter parsing
│   ├── output/
│   │   ├── json.go            # JSON output
│   │   └── table.go           # Table output
│   └── selector/
│       └── selector.go        # Fuzzy finder wrapper
└── go.mod
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for exported functions

## Key Packages

### internal/client

Wraps the Google Tasks API. Key types:
- `Client`: API client with caching
- `Task`: Task representation
- `TaskList`: Task list representation

Constants:
- `DefaultTaskList`: "@default"
- `StatusNeedsAction`: "needsAction"
- `StatusCompleted`: "completed"

### internal/editor

Handles editor integration:
- `Open()`: Opens content in $EDITOR
- `GenerateMarkdown()`: Creates markdown from task
- `ParseMarkdown()`: Parses markdown to task

### internal/cache

File-based caching at `~/.cache/gt/cache.json`:
- TTL: 5 minutes
- Updated on write operations (add, edit, done, delete)

## Testing

Currently no tests. Contributions welcome!

```bash
go test ./...
```

## Making Changes

1. Create a feature branch
2. Make your changes
3. Run `go build` to verify
4. Submit a pull request

## Commit Messages

Use conventional commit format:
- `feat:` New feature
- `fix:` Bug fix
- `refactor:` Code refactoring
- `docs:` Documentation
- `chore:` Maintenance
