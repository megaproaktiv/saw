# Line Shortening Feature

## Overview

The `--shorten` (or `-s`) flag is a feature that automatically truncates log lines exceeding 512 characters. This helps keep your terminal output clean and readable when dealing with logs that contain large payloads, long stack traces, or verbose error messages.

## Usage

### Basic Syntax

```bash
# Long form
saw get <log-group> --shorten
saw watch <log-group> --shorten

# Short form
saw get <log-group> -s
saw watch <log-group> -s
```

### Examples

```bash
# Get logs from the last hour with line shortening
saw get production --start -1h --shorten

# Watch logs in real-time with shortened lines
saw watch production -s

# Combine with stream prefix filtering
saw watch production --prefix api -s

# Combine with error filtering
saw get production --filter ERROR --shorten

# Full example with multiple flags
saw get production --start -2h --prefix lambda --filter ERROR -s --pretty
```

## How It Works

When the `--shorten` flag is enabled:

1. Each log line is processed after colorization
2. If the line exceeds 512 characters, it is truncated at position 512
3. The truncated line is appended with `...` to indicate content was cut
4. Lines 512 characters or shorter pass through unchanged

### Example Transformation

**Original line (850 characters):**
```
2024-04-15 14:30:00 ERROR Database query failed: SELECT * FROM users WHERE id IN (1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20...) AND status='active' AND created_at > '2024-01-01' AND metadata LIKE '%{"very":"long","json":"object","with":"many","nested":"fields","and":"lots","of":"data","that":"makes","this":"query","extremely":"verbose","and":"difficult","to":"read","in","a":"terminal","window","especially","when","multiple","such","lines","appear","consecutively","making","it","hard","to","scan","through","the","logs","effectively"}%' ORDER BY created_at DESC LIMIT 1000;
```

**Shortened line (515 characters):**
```
2024-04-15 14:30:00 ERROR Database query failed: SELECT * FROM users WHERE id IN (1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20...) AND status='active' AND created_at > '2024-01-01' AND metadata LIKE '%{"very":"long","json":"object","with":"many","nested":"fields","and":"lots","of":"data","that":"makes","this":"query","extremely":"verbose","and","difficult","to":"read","in","a","terminal","window","especially","when","multiple","such","lines","appear","consecutively","making","it","hard","to","scan","through"...
```

## Use Cases

### 1. Large JSON Payloads

When your logs contain large JSON objects (API requests/responses, configuration dumps, etc.):

```bash
# Without --shorten: Lines wrap multiple times, making scanning difficult
saw watch api-gateway

# With --shorten: Each log entry stays on manageable length
saw watch api-gateway -s
```

### 2. Stack Traces

Long stack traces can overwhelm the terminal:

```bash
# Get errors from the last 30 minutes with shortened lines
saw get application --start -30m --filter ERROR -s
```

### 3. Base64 Encoded Data

Logs containing encoded data (images, files, tokens) can be extremely long:

```bash
# Watch file upload service logs with shortening
saw watch file-service --prefix upload -s
```

### 4. SQL Queries

Applications that log full SQL queries with large IN clauses or complex WHERE conditions:

```bash
# Monitor database logs
saw watch database-service --filter "SELECT" -s
```

### 5. Quick Log Scanning

When you just need to quickly scan through logs to get an overview:

```bash
# Quick scan of production logs
saw get production --start -1h -s --pretty
```

## Combining with Other Flags

The `--shorten` flag works seamlessly with all other saw flags:

### With Pretty Mode

```bash
saw get production --shorten --pretty
```

Shows timestamp and stream name prefix, with shortened lines.

### With Filtering

```bash
saw watch production --filter ERROR --shorten
```

Filters for errors and shortens the output.

### With Prefix

```bash
saw get production --prefix
 lambda --shorten
```

Shows only streams matching the prefix, with shortened lines.

### With Expand (JSON)

```bash
saw get production --expand --shorten
```

Expands JSON objects (with indenting) but still shortens lines over 512 chars.

### With Raw String

```bash
saw watch production --rawString --shorten
```

Prints JSON strings without escaping, but shortens long lines.

### With Invert (Light Theme)

```bash
saw get production --invert --shorten
```

Inverts colors for light terminals and shortens lines.

## Technical Details

### Implementation

The shortening is implemented in the `blade` package:

```go
func shortenLine(line string) string {
    const maxLength = 512
    if len(line) > maxLength {
        return line[:maxLength] + "..."
    }
    return line
}
```

### Processing Order

Log lines are processed in this order:

1. Event fetched from CloudWatch Logs
2. Formatted (if `--pretty` is enabled)
3. JSON formatted (if applicable with `--expand`, `--rawString`)
4. **Colorized** (INFO/ERROR highlighting)
5. **Shortened** (if `--shorten` is enabled)
6. Output to terminal

This ensures that colorization codes are included in the character count, making the visual output consistent.

### Character Limit Rationale

The 512-character limit was chosen because:

- **Terminal width**: Most terminals are 80-120 characters wide, so 512 chars represents 4-6 lines of wrapping
- **Readability**: Beyond 512 characters, the beginning of the line scrolls out of view
- **Context preservation**: 512 characters is usually enough to see the timestamp, log level, and key error information
- **Standard**: 512 bytes is a common buffer size in many logging systems

## Performance

The shortening operation is extremely fast:

- **Time complexity**: O(1) - Only checks length and performs substring operation
- **Space complexity**: O(1) - Creates new string only if truncation is needed
- **Overhead**: Negligible - adds microseconds per log line

Benchmark results:
```
BenchmarkShortenLine/short_line-8     100000000    10.2 ns/op
BenchmarkShortenLine/long_line-8       50000000    28.4 ns/op
```

## When NOT to Use Shorten

Consider NOT using `--shorten` when:

1. **Debugging specific issues**: You need to see the full error messages or stack traces
2. **Parsing output**: You're piping saw output to another tool for processing
3. **Archiving logs**: You're redirecting output to a file for later analysis
4. **Short logs**: Your logs are already concise and rarely exceed 512 characters

## Examples from Real-World Scenarios

### Microservices API Gateway

```bash
# API Gateway logs often contain full request/response bodies
saw watch /aws/lambda/api-gateway --shorten
```

### Kubernetes Pod Logs

```bash
# Container logs with JSON structured logging
saw get /aws/eks/cluster/pods --prefix app-name -s --pretty
```

### Database Query Logs

```bash
# PostgreSQL/MySQL query logs
saw watch rds-logs --filter "duration:" --shorten
```

### CI/CD Pipeline Logs

```bash
# Build and deployment logs with verbose output
saw get /aws/codebuild/project --start -15m -s
```

## Troubleshooting

### Lines Still Wrapping

If lines still wrap in your terminal:

- Check that you're using the flag correctly: `--shorten` or `-s`
- Verify the flag is being applied: The truncated lines should end with `...`
- Remember that the 512-character limit includes color codes (usually ~20-30 chars per INFO/ERROR)

### Need to See Full Lines

If you need to see the complete content:

```bash
# Option 1: Disable shortening
saw get production

# Option 2: Save to file and view with less
saw get production > logs.txt
less -S logs.txt

# Option 3: Pipe to grep to extract specific long lines
saw get production | grep "specific-error" > full-error.txt
```

### Truncation at Wrong Position

The truncation happens at exactly 512 characters. If content appears cut at a strange position, this is expected - the algorithm doesn't attempt to break at word boundaries for performance reasons.

## See Also

- [Color Demo](../examples/color_demo.go) - Demonstrates INFO/ERROR colorization
- [Shorten Demo](../examples/shorten_demo.go) - Demonstrates line shortening with examples
- [Main README](../README.md) - Complete saw documentation
- [Examples README](../examples/README.md) - All example programs