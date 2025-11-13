# AI Agent Guide

Essential, actionable knowledge for contributing to this macOS-only osquery logging extension. Keep the code lean and latency minimal.

## Purpose & Data Flow
Forward osquery query results/status into macOS Unified Logging via a CGO bridge to `os_log`.
`osquery` → extension socket `/var/osquery/osquery.em` → logger plugin callback in `main.go` → `Logger.Log(...)` → Unified Log (subsystem `com.github.smithjw.osquery-unified-log-plugin`, categories: `results`, `snapshots`, `status`).

## Core Files
`cmd/unifiedlog-ext/main.go` – flags (`-socket`, `-name`, `-structured`), creates extension server, registers single `logger.NewPlugin` callback.
`internal/oslog/oslog_darwin.go` – Darwin-only CGO wrapper; `Logger`, `Log`, `LogStructured`, severity constants (`TypeDefault`, `TypeInfo`, `TypeDebug`, `TypeError`, `TypeFault`).
`internal/oslog/oslog_stub.go` – Non-Darwin no-op implementation for cross-platform builds.
`.github/workflows/release.yml` – dual-arch builds + universal binary via `lipo` on tag push.

## Build & Run (macOS)
```
go build -trimpath -o dist/unifiedlog.ext ./cmd/unifiedlog-ext
sudo mkdir -p /var/osquery/extensions
sudo cp dist/unifiedlog.ext /var/osquery/extensions/
echo "/var/osquery/extensions/unifiedlog.ext" | sudo tee /var/osquery/extensions.load
echo "--extensions_autoload=/var/osquery/extensions.load" | sudo tee -a /var/osquery/osquery.flags
echo "--logger_plugin=filesystem,unifiedlog" | sudo tee -a /var/osquery/osquery.flags
sudo launchctl unload /Library/LaunchDaemons/io.osquery.agent.plist
sudo launchctl load /Library/LaunchDaemons/io.osquery.agent.plist

# With structured logging (parses JSON fields)
echo "/var/osquery/extensions/unifiedlog.ext -structured" | sudo tee /var/osquery/extensions.load
```
Standard paths: extension in `/var/osquery/extensions/`, autoload file at `/var/osquery/extensions.load`, flags in `/var/osquery/osquery.flags`, daemon plist at `/Library/LaunchDaemons/io.osquery.agent.plist`. Logger plugins configured in flags file, not config file (osquery best practice).

## Release (manual mimic of CI)
```
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC="clang -arch arm64" go build -trimpath -ldflags="-s -w" -o dist/unifiedlog.ext.arm64 ./cmd/unifiedlog-ext
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 CC="clang -arch x86_64" go build -trimpath -ldflags="-s -w" -o dist/unifiedlog.ext.amd64 ./cmd/unifiedlog-ext
lipo -create -output dist/unifiedlog.ext dist/unifiedlog.ext.arm64 dist/unifiedlog.ext.amd64
shasum -a 256 dist/unifiedlog.ext > dist/unifiedlog.ext.sha256
```

## Conventions & Patterns
Darwin-specific code guarded by `//go:build darwin`; add a `!darwin` stub file for portability (same API, no-op).
Always free every `C.CString` (see `New`, `Log`, and `LogStructured`). Mirror existing pattern for new CGO calls.
Severity and category mapping: all log types use `TypeDefault` for visibility without flags; `LogTypeString` → `results` category, `LogTypeSnapshot` → `snapshots` category, `LogTypeStatus` → `status` category.
Runtime path should not panic; only startup uses `log.Fatalf`. `Logger.Log` returns silently if handle nil.
Structured logging is opt-in via `-structured` flag; defaults to raw JSON logging to minimize overhead.

## Extending
Add new plugin(s): create another `logger.NewPlugin` and `srv.RegisterPlugin(...)` before `srv.Run()`.
Adjust subsystem/category: create new loggers with `oslog.New("subsystem", "category")` (use reverse-DNS for subsystem). Currently uses three categories: `results`, `snapshots`, and `status`.
Add richer severity logic: parse status strings before mapping; extend constants sequentially if mapping required.
Cross-platform: implement `internal/oslog/oslog_stub.go` under `//go:build !darwin` returning a struct with no-op `Log`.

## Verifying Output
After running, inspect Unified Log:
```
log show --last 1m --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin"'
log show --last 1m --predicate 'category == "snapshots"'
log stream --predicate 'subsystem == "com.github.smithjw.osquery-unified-log-plugin" AND category == "results"'
```
Adjust `--last` duration as needed. All logs use `TypeDefault` level (no `--info` or `--debug` flags needed).

## When Unsure
Look at upstream `github.com/osquery/osquery-go` examples and replicate minimal patterns; avoid abstractions not present upstream.

---
Update this file when adding structured parsing, metrics, or multi-platform support.
