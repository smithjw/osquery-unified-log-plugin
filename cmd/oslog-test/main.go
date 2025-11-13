//go:build darwin

package main

import (
	"fmt"
	"time"

	"github.com/smithjw/osquery-unified-log-plugin/internal/oslog"
)

func main() {
	fmt.Println("Testing Unified Log output...")
	fmt.Println("Run this command in another terminal to view logs:")
	fmt.Println("  log stream --predicate 'subsystem == \"com.github.smithjw.osquery-unified-log-plugin\"' --level info")
	fmt.Println()

	l := oslog.New("com.github.smithjw.osquery-unified-log-plugin", "results")
	if l == nil {
		fmt.Println("ERROR: Failed to create logger")
		return
	}

	fmt.Println("Sending test messages to Unified Log...")

	// Test 1: Simple string
	fmt.Println("1. Sending simple string...")
	l.Log(oslog.TypeInfo, "Test message from osquery-unified-log-plugin")
	time.Sleep(500 * time.Millisecond)

	// Test 2: Raw JSON
	fmt.Println("2. Sending raw JSON...")
	jsonMsg := `{"name":"test_query","hostIdentifier":"test-mac","calendarTime":"2025-11-12 17:30:00","unixTime":"1731437400","action":"added","columns":{"hostname":"test-mac","cpu_type":"arm64"}}`
	l.Log(oslog.TypeInfo, jsonMsg)
	time.Sleep(500 * time.Millisecond)

	// Test 3: Structured logging
	fmt.Println("3. Sending structured JSON...")
	l.LogStructured(oslog.TypeInfo, jsonMsg)
	time.Sleep(500 * time.Millisecond)

	// Test 4: Different log levels
	fmt.Println("4. Sending different log levels...")
	l.Log(oslog.TypeDefault, "Default level message")
	l.Log(oslog.TypeDebug, "Debug level message")
	l.Log(oslog.TypeError, "Error level message")
	time.Sleep(500 * time.Millisecond)

	fmt.Println()
	fmt.Println("Test complete! Check the log stream output.")
	fmt.Println("To verify, run:")
	fmt.Println("  log show --last 1m --predicate 'subsystem == \"com.github.smithjw.osquery-unified-log-plugin\"' --info")
}
