# osquery-unified-log-plugin

A lightweight osquery logger extension that forwards query results and status logs to the macOS Unified Logging system.

## Features

- Forwards osquery query results to macOS Unified Log
- Defaults to sending the results as JSON stings to the Unified Log
- Supports structured logging with automatic JSON parsing
- Minimal overhead with direct CGO bridge to `os_log`
- Universal binary supporting both Apple Silicon (arm64) and Intel (amd64)
- Cross-platform buildable (no-op on non-Darwin platforms)

## Requirements

- macOS 10.12 (Sierra) or later
- osquery 5.x or later
- Go 1.25.x (for building from source)

## Quick Start

> [!WARNING]
> Please do not blindly follow the instructions below that tell you to curl things with `sudo` and move them around your system.
> You should look at the source code first, then determine if this extension is the right tool for the job!

```bash
# 1. Install extension
curl -L https://github.com/smithjw/osquery-unified-log-plugin/releases/latest/download/unifiedlog.ext -o ~/Downloads/unifiedlog.ext
sudo mkdir -p /var/osquery/extensions
sudo mv ~/Downloads/unifiedlog.ext /var/osquery/extensions/unifiedlog.ext
sudo chmod +x /var/osquery/extensions/unifiedlog.ext

# 2. Configure osquery to load the extension
echo "/var/osquery/extensions/unifiedlog.ext" | sudo tee /var/osquery/extensions.load

# 3. Add to osquery flags
echo "--extensions_autoload=/var/osquery/extensions.load" | sudo tee -a /var/osquery/osquery.flags
echo "--logger_plugin=filesystem,unifiedlog" | sudo tee -a /var/osquery/osquery.flags

# 4. Restart osquery
# If osquery is already running
sudo launchctl kickstart -k system/io.osquery.agent
# If it's not currently running
sudo launchctl bootstrap system /Library/LaunchDaemons/io.osquery.agent.plist

# 5. View logs
log stream --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin"'
```

## Build from Source

```bash
git clone https://github.com/smithjw/osquery-unified-log-plugin.git
cd osquery-unified-log-plugin
go build -trimpath -o unifiedlog.ext ./cmd/unifiedlog-ext
```

For universal binary (both architectures):

```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC="clang -arch arm64" \
  go build -trimpath -ldflags="-s -w" -o dist/unifiedlog.ext.arm64 ./cmd/unifiedlog-ext

CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="clang -arch x86_64" \
  go build -trimpath -ldflags="-s -w" -o dist/unifiedlog.ext.amd64 ./cmd/unifiedlog-ext

lipo -create -output unifiedlog.ext dist/unifiedlog.ext.arm64 dist/unifiedlog.ext.amd64
```

## Configuration

### 1. Register the Extension

Add the extension path to osquery's extensions autoload file:

```bash
# Default mode (raw JSON logging)
echo "/var/osquery/extensions/unifiedlog.ext" | sudo tee /var/osquery/extensions.load

# Or with structured logging enabled (parses JSON fields)
echo "/var/osquery/extensions/unifiedlog.ext -structured" | sudo tee /var/osquery/extensions.load
```

### 2. Configure osquery Flags

Edit `/var/osquery/osquery.flags` to enable the extension and logger plugin:

```bash
--extensions_autoload=/var/osquery/extensions.load
--logger_plugin=filesystem,unifiedlog
--logger_path=/var/log/osquery
```

**Note**: Using `filesystem,unifiedlog` enables both loggers simultaneously. Remove `filesystem` if you only want Unified Log output.

### 3. Configure osquery Queries (Optional)

Edit `/var/osquery/osquery.conf` or create query packs:

```json
{
  "schedule": {
    "system_info": {
      "query": "SELECT hostname, cpu_type, physical_memory FROM system_info;",
      "interval": 3600
    }
  }
}
```

### 4. Restart osquery

```bash
sudo launchctl kickstart -k system/io.osquery.agent
```

Verify the extension loaded:

```bash
osqueryi --extensions_socket=/var/osquery/osquery.em
osquery> SELECT * FROM osquery_extensions;
```

## Usage

### Viewing Logs

After configuration, query results will appear in the Unified Log:

```bash
# Show all logs from the last minute
log show --last 1m --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin"'

# Show only snapshot results
log show --last 1m --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin" AND category == "snapshots"'

# Show only regular query results
log show --last 1m --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin" AND category == "results"'

# Stream live logs (all categories)
log stream --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin"'

# Stream only snapshots
log stream --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin" AND category == "snapshots"'
```

### Log Categories

The extension uses a single subsystem with three categories:

- **Subsystem**: `com.github.smithjw.osquery-unified-log-plugin`
- **Categories**:
  - `results` - Regular scheduled query results
  - `snapshots` - Snapshot query results
  - `status` - Status messages and other logs

### Log Types

Different osquery log types are mapped to Unified Log categories and levels:

| osquery Type      | Category    | Unified Log Level | Description                     |
| ----------------- | ----------- | ----------------- | ------------------------------- |
| `LogTypeString`   | `results`   | Default           | Regular scheduled query results |
| `LogTypeSnapshot` | `snapshots` | Default           | Snapshot query results          |
| `LogTypeStatus`   | `status`    | Default           | Status messages                 |
| Other             | `status`    | Default           | Fallback                        |

### Structured Logging

The extension supports two logging modes:

#### Raw Mode (Default)
Logs complete JSON results as-is without parsing:
```bash
./unifiedlog.ext -socket /var/osquery/osquery.em -name unifiedlog
```

#### Structured Mode
Parses JSON and extracts key fields for searchable metadata (requires `-structured` flag):
```bash
./unifiedlog.ext -socket /var/osquery/osquery.em -name unifiedlog -structured
```

When structured mode is enabled:
- Extracts up to 8 key fields (name, hostIdentifier, calendarTime, etc.)
- Logs structured data as: `osquery name=system_info hostIdentifier=mac-001 ...`
- **Also logs the complete JSON** for full record retention
- Falls back to raw logging if JSON parsing fails

This allows powerful filtering in structured mode:
```bash
# Filter snapshots by specific host
log show --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin" AND category == "snapshots" AND hostIdentifier == "mac-001"'

# Filter results by query name
log show --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin" AND category == "results" AND name == "system_info"'

# Show all snapshots
log stream --predicate 'category == "snapshots"'
```

**Note**: Structured mode logs each result twice (once with extracted fields, once with full JSON), which may increase log volume.

## Troubleshooting

### Extension Not Loading

1. Check osquery daemon logs:
   ```bash
   sudo tail -f /var/log/osquery/osqueryd.INFO
   ```

2. Verify the extension file exists and is executable:
   ```bash
   ls -la /var/osquery/extensions/unifiedlog.ext
   ```

3. Check the extensions autoload file:
   ```bash
   cat /var/osquery/extensions.load
   ```

4. Verify the extension loaded successfully:
   ```bash
   osqueryi --extensions_socket=/var/osquery/osquery.em
   osquery> SELECT * FROM osquery_extensions;
   ```

### No Logs Appearing in Unified Log

1. Verify the logger plugin is configured in `/var/osquery/osquery.flags`:
   ```bash
   grep logger_plugin /var/osquery/osquery.flags
   # Should show: --logger_plugin=filesystem,unifiedlog
   ```

2. Check if logs are being generated at all:
   ```bash
   sudo log stream --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin"'
   ```

3. Verify osquery is running queries:
   ```bash
   tail -f /var/log/osquery/osqueryd.results.log
   ```

4. Check osquery flags are loaded:
   ```bash
   osqueryi --extensions_socket=/var/osquery/osquery.em
   osquery> SELECT * FROM osquery_flags WHERE name = 'logger_plugin';
   osquery> SELECT * FROM osquery_flags WHERE name = 'extensions_autoload';
   ```

### Testing the Extension Manually

Stop the osquery daemon and run the extension standalone for testing:

```bash
# Stop osquery daemon
sudo launchctl unload /Library/LaunchDaemons/io.osquery.agent.plist

# Run extension manually
/var/osquery/extensions/unifiedlog.ext -socket /var/osquery/osquery.em -name unifiedlog &

# Start osqueryd manually with verbose logging
sudo osqueryd --flagfile=/var/osquery/osquery.flags --verbose

# Clean up when done
sudo launchctl load /Library/LaunchDaemons/io.osquery.agent.plist
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run on Darwin only
go test -tags darwin ./internal/oslog
```

### Project Structure

```
.
├── cmd/unifiedlog-ext/     # Main extension binary
│   ├── main.go            # Entry point, plugin registration
│   └── main_test.go       # Basic integration tests
├── internal/oslog/        # CGO wrapper for os_log
│   ├── oslog_darwin.go    # Darwin implementation
│   ├── oslog_darwin_test.go
│   ├── oslog_stub.go      # Non-Darwin no-op
│   └── oslog_stub_test.go
└── .github/workflows/     # CI/CD
    └── release.yml        # Automated universal binary builds
```

## Contributing

Contributions are welcome! Please:

1. Ensure tests pass: `go test ./...`
2. Follow existing code style
3. Update documentation for new features
4. Keep the codebase minimal and focused

## License

See LICENSE file for details.
