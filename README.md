# Em – Emoji Picker

A simple terminal-based emoji picker written in Go. Search and copy emojis directly from your terminal.

## Installation

```bash
go install github.com/bart-jaskulski/em@latest
```

## Key Bindings

- `↑/↓/←/→`: Navigate through emojis
- `Enter`: Copy selected emoji
- `Tab`: Toggle focus between search and emoji grid
- `Esc`: Quit

## Data Storage

Emoji data is cached in `~/.local/share/emoji-picker/` (or `$XDG_DATA_HOME/emoji-picker/` if set).
