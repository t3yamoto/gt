# gt - Google Tasks CLI

A simple command-line interface for Google Tasks.

## Features

- List tasks from all task lists or a specific list
- Add tasks (simple mode or editor mode with markdown)
- Edit tasks with your favorite editor
- Mark tasks as done
- Delete tasks
- Interactive task selection with fuzzy finder
- File-based caching for faster responses

## Requirements

- **fzf** - Required for interactive task selection
- **bat** - Optional, for syntax highlighting in preview (falls back to plain text)

```bash
# macOS
brew install fzf bat

# Ubuntu/Debian
sudo apt install fzf bat
```

## Installation

```bash
go install github.com/t3yamoto/gt@latest
```

## Setup

### 1. Create OAuth credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Tasks API
4. Create OAuth 2.0 credentials (Desktop application)
5. Download the credentials JSON file

### 2. Save credentials

Save the downloaded file to:
```
~/.config/gt/credentials.json
```

### 3. Authenticate

Run any command and authenticate via browser:
```bash
gt list
```

## Usage

### List tasks

```bash
# List all tasks (sorted by due date, then list name)
gt list

# List tasks from a specific list
gt list -l "My List"

# Output as JSON
gt list --json
```

### Add a task

```bash
# Simple mode: add with title
gt add "Buy groceries"

# Editor mode: opens $EDITOR with markdown template
gt add

# Add to a specific list
gt add -l "Shopping" "Buy milk"
```

#### Editor format

```markdown
---
title: Task title
due: 2024-02-20
tasklist: My List
done: false
---

Notes go here.
Multiple lines supported.
```

### Mark task as done

```bash
# Interactive selection
gt done

# By task ID
gt done abc123

# From a specific list
gt done -l "My List" abc123
```

### Edit a task

```bash
# Interactive selection
gt edit

# By task ID
gt edit abc123
```

### Delete a task

```bash
# Interactive selection
gt delete

# By task ID
gt delete abc123
```

## Configuration

### Cache

Task data is cached locally for 5 minutes:
```
~/.cache/gt/cache.json
```

### Authentication tokens

OAuth tokens are stored at:
```
~/.config/gt/token.json
```

## Task ID

Task IDs are displayed as 8-character short IDs for convenience. You can use these short IDs in commands.

## License

MIT
