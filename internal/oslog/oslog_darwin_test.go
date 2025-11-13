//go:build darwin

package oslog

import "testing"

func TestNew(t *testing.T) {
	l := New("com.test.subsystem", "test-category")
	if l == nil {
		t.Fatal("New() returned nil")
	}
	if l.h == nil {
		t.Fatal("New() returned Logger with nil handle")
	}
}

func TestLog(t *testing.T) {
	tests := []struct {
		name string
		typ  Type
		msg  string
	}{
		{"default", TypeDefault, "default message"},
		{"info", TypeInfo, "info message"},
		{"debug", TypeDebug, "debug message"},
		{"error", TypeError, "error message"},
		{"fault", TypeFault, "fault message"},
		{"empty", TypeInfo, ""},
		{"long", TypeInfo, "very long message " + string(make([]byte, 1000))},
	}

	l := New("com.test.oslog", "test")
	if l == nil {
		t.Fatal("failed to create logger")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			l.Log(tt.typ, tt.msg)
		})
	}
}

func TestLogNilLogger(t *testing.T) {
	var l *Logger
	// Should not panic
	l.Log(TypeInfo, "test")
}

func TestLogNilHandle(t *testing.T) {
	l := &Logger{h: nil}
	// Should not panic
	l.Log(TypeInfo, "test")
}

func TestLogStructured(t *testing.T) {
	l := New("com.test.oslog", "structured")
	if l == nil {
		t.Fatal("failed to create logger")
	}

	tests := []struct {
		name string
		msg  string
	}{
		{
			"valid json",
			`{"name":"test_query","hostIdentifier":"mac-001","calendarTime":"2024-01-01","action":"added","columns":{"hostname":"mac-001"}}`,
		},
		{
			"minimal json",
			`{"name":"test"}`,
		},
		{
			"invalid json",
			`not valid json`,
		},
		{
			"empty json",
			`{}`,
		},
		{
			"empty string",
			``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			l.LogStructured(TypeInfo, tt.msg)
		})
	}
}

func TestLogStructuredNilLogger(t *testing.T) {
	var l *Logger
	// Should not panic
	l.LogStructured(TypeInfo, `{"test":"value"}`)
}

func TestLogStructuredNilHandle(t *testing.T) {
	l := &Logger{h: nil}
	// Should not panic
	l.LogStructured(TypeInfo, `{"test":"value"}`)
}
