# Saw Examples

This directory contains example programs and demonstrations for the Saw CloudWatch Logs tool.

## Color Demo

The `color_demo.go` program demonstrates the log level colorization feature in Saw.

### Running the Demo

```bash
go run examples/color_demo.go
```

### Features Demonstrated

- **INFO colorization**: All text matching "INFO" (all caps) is displayed with a green background and black text
- **ERROR colorization**: All text matching "ERROR" (all caps) is displayed with a red background and white text

### Color Examples

The demo shows colorization across different log formats:

1. **Standard timestamps**: `2024-04-15 14:30:00 INFO Application started`
2. **RFC3339 timestamps**: `[2024-04-15T14:30:15Z] (stream-1) INFO Starting worker`
3. **JSON logs**: `{"level": "INFO", "message": "API request processed"}`

### Implementation

The colorization is implemented using the [fatih/color](https://github.com/fatih/color) package:

- `color.New(color.FgBlack, color.BgGreen)` - Green background, black foreground for INFO
- `color.New(color.FgWhite, color.BgRed)` - Red background, white foreground for ERROR

### Integration with Saw

This colorization is automatically applied to all log output in the Saw tool when using:

- `saw get <log-group>` - Get log events
- `saw watch <log-group>` - Stream log events in real-time

### Disabling Colors

If you need to disable colorization (e.g., for piping to files), you can set the `NO_COLOR` environment variable:

```bash
NO_COLOR=1 saw get my-log-group
```

Or use the fatih/color's built-in support by setting:

```bash
TERM=dumb saw get my-log-group
```

## Shorten Demo

The `shorten_demo.go` program demonstrates the line shortening feature in Saw.

### Running the Demo

```bash
go run examples/shorten_demo.go
```

### Features Demonstrated

- **Line truncation**: Lines exceeding 512 characters are automatically truncated
- **Visual marker**: Truncated lines are marked with "..." at the end
- **Multiple formats**: Shows shortening across different log formats (standard logs, JSON, etc.)

### Use Cases

The `--shorten` (or `-s`) flag is particularly useful when:

1. **Dealing with large payloads**: Logs containing large JSON objects or base64-encoded data
2. **Terminal readability**: Preventing line wrapping that clutters the terminal
3. **Quick scanning**: Making it easier to scan through logs without excessive scrolling
4. **Stack traces**: Truncating very long stack traces while keeping the beginning visible

### Integration with Saw

The shortening feature works with both commands:

```bash
# Get logs with line shortening
saw get production --shorten
saw get production -s

# Watch logs with line shortening
saw watch production --shorten
saw watch production -s

# Combine with other flags
saw get production -s --pretty --filter ERROR
saw watch production -s --prefix api
```

### Implementation

Lines are checked after colorization but before output. If a line exceeds 512 characters, it's truncated:

```go
func shortenLine(line string) string {
    const maxLength = 512
    if len(line) > maxLength {
        return line[:maxLength] + "..."
    }
    return line
}
```

This ensures you can still see the beginning of each log message while keeping the output clean and manageable.
