//go:build darwin

package oslog

/*
#cgo CFLAGS: -mmacosx-version-min=10.12
#cgo LDFLAGS: -mmacosx-version-min=10.12
#include <os/log.h>
#include <stdlib.h>

static os_log_t go_os_log_create(const char* subsystem, const char* category) {
  return os_log_create(subsystem, category);
}

static void go_os_log_with_type(os_log_t h, uint8_t t, const char* msg) {
  // %{public}s prevents privacy redaction in Unified Log.
  // By default, string arguments are redacted as "<private>" in logs.
  // The {public} qualifier makes the string visible in log output.
  os_log_with_type(h, (os_log_type_t)t, "%{public}s", msg);
}

// Structured logging with up to 8 key-value pairs
static void go_os_log_structured(os_log_t h, uint8_t t, const char* msg,
    const char* k1, const char* v1,
    const char* k2, const char* v2,
    const char* k3, const char* v3,
    const char* k4, const char* v4,
    const char* k5, const char* v5,
    const char* k6, const char* v6,
    const char* k7, const char* v7,
    const char* k8, const char* v8) {
  // Build a format string based on how many non-NULL keys we have
  if (k1 && v1) {
    if (k2 && v2) {
      if (k3 && v3) {
        if (k4 && v4) {
          if (k5 && v5) {
            if (k6 && v6) {
              if (k7 && v7) {
                if (k8 && v8) {
                  os_log_with_type(h, t, "%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s",
                    msg, k1, v1, k2, v2, k3, v3, k4, v4, k5, v5, k6, v6, k7, v7, k8, v8);
                } else {
                  os_log_with_type(h, t, "%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s",
                    msg, k1, v1, k2, v2, k3, v3, k4, v4, k5, v5, k6, v6, k7, v7);
                }
              } else {
                os_log_with_type(h, t, "%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s",
                  msg, k1, v1, k2, v2, k3, v3, k4, v4, k5, v5, k6, v6);
              }
            } else {
              os_log_with_type(h, t, "%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s",
                msg, k1, v1, k2, v2, k3, v3, k4, v4, k5, v5);
            }
          } else {
            os_log_with_type(h, t, "%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s",
              msg, k1, v1, k2, v2, k3, v3, k4, v4);
          }
        } else {
          os_log_with_type(h, t, "%{public}s %{public}s=%{public}s %{public}s=%{public}s %{public}s=%{public}s",
            msg, k1, v1, k2, v2, k3, v3);
        }
      } else {
        os_log_with_type(h, t, "%{public}s %{public}s=%{public}s %{public}s=%{public}s",
          msg, k1, v1, k2, v2);
      }
    } else {
      os_log_with_type(h, t, "%{public}s %{public}s=%{public}s",
        msg, k1, v1);
    }
  } else {
    os_log_with_type(h, t, "%{public}s", msg);
  }
}
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// Type mirrors os_log_type_t.
type Type uint8

const (
	TypeDefault Type = iota
	TypeInfo
	TypeDebug
	TypeError
	TypeFault
)

type Logger struct{ h C.os_log_t }

func New(subsystem, category string) *Logger {
	cs := C.CString(subsystem)
	cc := C.CString(category)
	defer C.free(unsafe.Pointer(cs))
	defer C.free(unsafe.Pointer(cc))
	return &Logger{h: C.go_os_log_create(cs, cc)}
}

// Log writes a simple string message.
func (l *Logger) Log(t Type, msg string) {
	if l == nil || l.h == nil {
		return
	}
	cm := C.CString(msg)
	defer C.free(unsafe.Pointer(cm))
	C.go_os_log_with_type(l.h, C.uchar(uint8(t)), cm)
}

// LogStructured attempts to parse msg as JSON and log with structured fields.
// Falls back to Log if parsing fails. Supports up to 8 key-value pairs.
func (l *Logger) LogStructured(t Type, msg string) {
	if l == nil || l.h == nil {
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		// Not valid JSON, fall back to simple logging
		l.Log(t, msg)
		return
	}

	// Extract up to 8 interesting fields for structured logging
	keys := []string{"name", "hostIdentifier", "calendarTime", "unixTime", "action", "columns", "decorations", "epoch"}
	pairs := make([][2]string, 0, 8)

	for _, key := range keys {
		if val, ok := data[key]; ok {
			// Convert value to string
			var valStr string
			switch v := val.(type) {
			case string:
				valStr = v
			case float64:
				valStr = fmt.Sprintf("%.0f", v)
			case bool:
				valStr = fmt.Sprintf("%t", v)
			default:
				// For complex types, marshal back to JSON
				if b, err := json.Marshal(v); err == nil {
					valStr = string(b)
				}
			}
			if valStr != "" {
				pairs = append(pairs, [2]string{key, valStr})
				if len(pairs) >= 8 {
					break
				}
			}
		}
	}

	// If no structured fields found, log the raw JSON
	if len(pairs) == 0 {
		l.Log(t, msg)
		return
	}

	// Prepare C strings for the structured log call
	baseMsg := "osquery"
	cMsg := C.CString(baseMsg)
	defer C.free(unsafe.Pointer(cMsg))

	// Allocate space for up to 8 key-value pairs
	cKeys := make([]*C.char, 8)
	cVals := make([]*C.char, 8)
	defer func() {
		for i := 0; i < 8; i++ {
			if cKeys[i] != nil {
				C.free(unsafe.Pointer(cKeys[i]))
			}
			if cVals[i] != nil {
				C.free(unsafe.Pointer(cVals[i]))
			}
		}
	}()

	// Populate the pairs
	for i := 0; i < len(pairs) && i < 8; i++ {
		cKeys[i] = C.CString(pairs[i][0])
		cVals[i] = C.CString(pairs[i][1])
	}

	// Call the structured logging function
	C.go_os_log_structured(
		l.h, C.uchar(uint8(t)), cMsg,
		cKeys[0], cVals[0],
		cKeys[1], cVals[1],
		cKeys[2], cVals[2],
		cKeys[3], cVals[3],
		cKeys[4], cVals[4],
		cKeys[5], cVals[5],
		cKeys[6], cVals[6],
		cKeys[7], cVals[7],
	)

	// Also log the full JSON for complete record
	l.Log(t, msg)
}
