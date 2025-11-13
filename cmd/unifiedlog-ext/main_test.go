package main

import (
	"testing"

	"github.com/smithjw/osquery-unified-log-plugin/internal/oslog"
)

func TestLoggerInitialization(t *testing.T) {
	l := oslog.New("com.github.smithjw.osquery-unified-log-plugin", "results")
	if l == nil {
		t.Fatal("failed to create logger")
	}
}

func TestLoggerTypes(t *testing.T) {
	l := oslog.New("com.test.osquery", "test")

	tests := []struct {
		name string
		typ  oslog.Type
		msg  string
	}{
		{"info", oslog.TypeInfo, `{"name":"test","time":123}`},
		{"default", oslog.TypeDefault, "status message"},
		{"empty", oslog.TypeInfo, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			l.Log(tt.typ, tt.msg)
		})
	}
}
