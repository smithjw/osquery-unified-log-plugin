//go:build !darwin

package oslog

import "testing"

func TestNewStub(t *testing.T) {
	l := New("com.test.subsystem", "test-category")
	if l == nil {
		t.Fatal("New() returned nil")
	}
}

func TestLogStub(t *testing.T) {
	l := New("com.test.oslog", "test")
	// Should not panic
	l.Log(TypeInfo, "test message")
}

func TestLogNilLoggerStub(t *testing.T) {
	var l *Logger
	// Should not panic
	l.Log(TypeInfo, "test")
}

func TestLogStructuredStub(t *testing.T) {
	l := New("com.test.oslog", "test")
	// Should not panic
	l.LogStructured(TypeInfo, `{"name":"test"}`)
}
