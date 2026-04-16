# Changelog


## v0.3.0 (Unreleased)

### Major Changes

- **AWS SDK v2 Migration** - Upgraded from AWS SDK for Go v1 to v2
  - Modern AWS SDK with better performance and error handling
  - Context-based API calls
  - Improved pagination using paginators
  - Active maintenance and security updates
  - All existing commands work unchanged

### New Features

- **Automatic Log Level Colorization** - INFO and ERROR are automatically highlighted
  - `INFO` appears with green background and black text
  - `ERROR` appears with red background and white text
  - Works across all output modes (raw, pretty, JSON)
  - Case-sensitive matching (only all-caps INFO/ERROR)
  - Automatic color detection (disabled when piping)
  - Use `NO_COLOR=1` environment variable to disable

- **Line Shortening** - New `--shorten` / `-s` flag to truncate long lines
  - Truncates lines exceeding 512 characters
  - Appends "..." to truncated lines
  - Prevents terminal clutter from large payloads
  - Available for both `get` and `watch` commands
  - Useful for logs with JSON payloads, stack traces, or base64 data

- **Dual Split-Pane View** - New `dual` command for monitoring two log groups
  - Watch two log groups simultaneously with exact half-screen horizontal split
  - Each pane takes exactly 50% of terminal height for perfect symmetry
  - Fullscreen terminal UI powered by Bubble Tea
  - Interactive keyboard navigation (vim-style controls)
  - Independent filtering and scrolling for each pane
  - Real-time updates every second
  - **Tail mode**: Defaults to last 5 minutes of logs, auto-scrolls to bottom on new messages
  - **Single-line display**: Newlines in log messages replaced with spaces for clean rendering
  - Keyboard controls: `q` quit, `Tab` switch pane, `↑↓/jk` scroll, `g/G` top/bottom
  - Per-pane configuration: `--filter1/2`, `--prefix1/2`, `--start1/2`, `--shorten1/2`
  - Help text displayed in bottom pane title bar to maintain exact 50/50 split
  - Exact height maintained - no resizing or extra line feeds
  - Perfect for deployment monitoring, error correlation, and microservices debugging

### Build & Installation

- **Task-based Build System** - New Taskfile.yml for streamlined building
  - `task build` - Build binary to `dist/saw`
  - `task install` - Install binary to `~/.local/bin/saw`
  - `task clean` - Remove build artifacts

### Documentation

- Comprehensive documentation for all new features
- Example programs for color and shortening demonstrations
- Visual guides for dual view interface
- Quick reference guide for all commands and features

## v0.2.2

 - Added support for parsing additonal time formats (@andrewpage)

## v0.2.1

- Feature - de-duplicate newlines event messages (@klichukb)
- Feature - Added Dockerfile (@shnhrrsn)

## v0.2.0

- Added --raw flag to watch subcommand, disables decorations (@will-ockmore)
- Added --pretty flag to get subcommand, enables decorations (@will-ockmore)

The defaults are for the watch output to be pretty and the get output to be raw.

## v0.1.8

- Support filter option for get (@cynipe)

## v0.1.7

- Fix usage output for get command
- Rename get command `end` flag to `stop`
- Unexport some exported vars in `cmd` package

## v0.1.6

- Add MFA (assumerole) support (@perriea)
- Add usage for get command (@will-ockmore)

## v0.1.5

- Add region and profile support
