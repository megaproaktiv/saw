# Saw Quick Reference Guide

## New Features Overview

This guide provides a quick reference for the new features added to Saw.

---

## 📦 Build & Install

### Build to dist directory

```bash
# Build saw binary to dist/saw
task build

# Clean build artifacts
task clean
```

### Install to user local bin

```bash
# Install to ~/.local/bin/saw
task install

# Manually ensure ~/.local/bin is in PATH
export PATH="$HOME/.local/bin:$PATH"
```

**Location**: Binary installed at `~/.local/bin/saw`

---

## 🎨 Automatic Log Level Colorization

Log levels are automatically colorized for better visibility:

| Level   | Background | Foreground | Example Display |
|---------|------------|------------|-----------------|
| `INFO`  | Green      | Black      | [Green bg]INFO[/] |
| `ERROR` | Red        | White      | [Red bg]ERROR[/] |

### Features

- ✅ Case-sensitive (only `INFO` and `ERROR` in all caps)
- ✅ Works with all output modes (raw, pretty, JSON)
- ✅ Automatic color detection (TTY vs pipe)
- ✅ No configuration needed

### Disable Colors

```bash
# Using environment variable
NO_COLOR=1 saw get production

# For piping/scripting (auto-detected)
saw get production > logs.txt  # Colors automatically disabled
```

---

## ✂️ Line Shortening (`--shorten` / `-s`)

Truncate lines exceeding 512 characters to keep output clean.

### Usage

```bash
# Long form
saw get production --shorten
saw watch production --shorten

# Short form
saw get production -s
saw watch production -s
```

### How It Works

- Lines > 512 chars are truncated at position 512
- Truncated lines end with `...`
- Lines ≤ 512 chars pass through unchanged

### Examples

```bash
# Get logs with shortened lines
saw get production --start -1h -s

# Watch with shortening and filtering
saw watch production --filter ERROR -s

# Combine with pretty mode
saw get production -s --pretty

# Multiple flags
saw watch api-gateway --prefix lambda -s --filter ERROR
```

### When to Use

- ✅ Large JSON payloads in logs
- ✅ Long stack traces
- ✅ Base64 encoded data
- ✅ SQL query logs
- ✅ Quick log scanning
- ❌ Debugging (need full content)
- ❌ Piping to other tools
- ❌ Archiving/saving logs

---

## 🚀 AWS SDK v2

Saw now uses AWS SDK for Go v2 (migrated from v1).

### Benefits

- Modern AWS SDK
- Better performance
- Improved error handling
- Context-based API calls
- Active maintenance and updates

### No Changes Required

The migration is transparent - all existing commands work the same:

```bash
# Same commands as before
saw groups
saw streams production
saw get production --start -1h
saw watch production --prefix api
```

### Configuration

Still uses standard AWS configuration:

```bash
# Profile support
saw groups --profile myprofile

# Region override
saw get production --region us-west-2

# Endpoint override
saw watch production --endpoint-url http://localhost:4566
```

---

## 📚 Command Quick Reference

### Common Workflows

#### 1. Quick Error Check
```bash
saw get production --start -30m --filter ERROR -s --pretty
```
Shows errors from last 30 minutes with shortened lines and timestamps.

#### 2. Real-time Monitoring
```bash
saw watch production --prefix api -s
```
Stream logs from API streams with line shortening.

#### 3. Debug Specific Service
```bash
saw watch production --prefix lambda/myfunction --filter "timeout"
```
Watch specific Lambda function for timeout errors.

#### 4. Export Logs
```bash
saw get production --start -1h > export.log
```
Save logs to file (colors auto-disabled).

#### 5. Scan Large Logs
```bash
saw get production --start -6h -s | less
```
Review 6 hours of logs with shortening, paginated.

---

## 🔧 All Flags Reference

### Global Flags

| Flag | Description |
|------|-------------|
| `--profile` | AWS profile to use |
| `--region` | AWS region override |
| `--endpoint-url` | Custom endpoint (LocalStack, etc.) |

### Get Command

| Flag | Short | Description |
|------|-------|-------------|
| `--start` | | Start time (absolute or relative) |
| `--stop` | | Stop time (default: now) |
| `--filter` | | CloudWatch filter pattern |
| `--prefix` | | Log stream prefix filter |
| `--pretty` | | Show timestamp and stream prefix |
| `--expand` | | Indent JSON output |
| `--invert` | | Invert colors (light theme) |
| `--rawString` | | Print JSON strings unescaped |
| `--shorten` | `-s` | Truncate lines > 512 chars |

### Watch Command

| Flag | Short | Description |
|------|-------|-------------|
| `--filter` | | CloudWatch filter pattern |
| `--prefix` | | Log stream prefix filter |
| `--raw` | | No timestamp/stream prefix |
| `--expand` | | Indent JSON output |
| `--invert` | | Invert colors (light theme) |
| `--rawString` | | Print JSON strings unescaped |
| `--shorten` | `-s` | Truncate lines > 512 chars |

---

## 💡 Pro Tips

### Tip 1: Combine Shortening with Pretty Mode
```bash
saw watch production -s --pretty
```
Best of both worlds: timestamps + clean output.

### Tip 2: Use Relative Time Ranges
```bash
saw get production --start -2h --stop -1h
```
Get logs from 2 hours ago to 1 hour ago.

### Tip 3: Filter + Shorten for Clean Error Reports
```bash
saw get production --filter ERROR -s --start -1d | wc -l
```
Count errors from last day without overwhelming output.

### Tip 4: Stream Prefix for Multi-Instance Apps
```bash
saw watch production --prefix "api-server-" -s
```
Monitor all API server instances.

### Tip 5: Export with NO_COLOR for Clean Files
```bash
NO_COLOR=1 saw get production --start -1h > clean-logs.txt
```
Ensures no ANSI color codes in saved file.

---

## 🐛 Troubleshooting

### Colors Not Showing
- Check if terminal supports colors
- Verify `NO_COLOR` env var is not set
- Ensure not piping to file/another command

### Lines Still Too Long
- Verify `-s` or `--shorten` flag is used
- Remember: limit is 512 chars including color codes
- Consider terminal width settings

### Build/Install Issues
```bash
# Clean and rebuild
task clean
task build

# Check binary
ls -lh dist/saw

# Manual install
cp dist/saw ~/.local/bin/
chmod +x ~/.local/bin/saw
```

### AWS Credentials
```bash
# Test AWS access
aws sts get-caller-identity --profile myprofile

# Use same profile with saw
saw groups --profile myprofile
```

---

## 📖 Additional Resources

- **Main README**: [../README.md](../README.md)
- **Shorten Feature**: [SHORTEN.md](SHORTEN.md)
- **Color Demo**: [../examples/color_demo.go](../examples/color_demo.go)
- **Shorten Demo**: [../examples/shorten_demo.go](../examples/shorten_demo.go)
- **Color Guide**: [../examples/COLORS.md](../examples/COLORS.md)

---

## 📝 Examples by Use Case

### DevOps: Monitor Deployments
```bash
saw watch /aws/ecs/cluster-name --prefix app-v2 -s --pretty
```

### SRE: Investigate Outage
```bash
saw get production --start "2024-04-15 14:00:00" --stop "2024-04-15 14:30:00" --filter ERROR -s
```

### Developer: Debug Lambda
```bash
saw watch /aws/lambda/my-function --expand -s
```

### Security: Audit Access Logs
```bash
saw get api-gateway --filter "401\|403" --start -24h
```

### Data Engineer: ETL Monitoring
```bash
saw watch /aws/glue/jobs --prefix etl-pipeline -s --filter "ERROR\|WARN"
```

---

**Version**: Saw with AWS SDK v2, Colorization, and Line Shortening  
**Last Updated**: 2024-04-15