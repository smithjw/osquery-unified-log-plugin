//go:build !darwin

package oslog

// Type mirrors os_log_type_t for cross-platform compatibility.
type Type uint8

const (
	TypeDefault Type = iota
	TypeInfo
	TypeDebug
	TypeError
	TypeFault
)

// Logger is a no-op logger for non-Darwin platforms.
type Logger struct{}

// New creates a no-op logger on non-Darwin platforms.
func New(subsystem, category string) *Logger {
	return &Logger{}
}

// Log is a no-op on non-Darwin platforms.
func (l *Logger) Log(t Type, msg string) {}

// LogStructured is a no-op on non-Darwin platforms.
func (l *Logger) LogStructured(t Type, msg string) {}
