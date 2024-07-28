package main

import (
	"os"
	"testing"
	"time"
)

// Test for intcmp function
func TestIntcmp(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"1", "2", -1},
		{"2", "1", 1},
		{"123", "123", 0},
		{"-1", "1", -1},
		{"1", "-1", 1},
		{"0", "0", 0},
		{"+1", "1", 0},
		{"-1", "-1", 0},
		{"1", "001", 0},
		{"001", "1", 0},
		{"123", "0123", 0},
		{"0123", "123", 0},
	}

	for _, test := range tests {
		result := intcmp(test.a, test.b)
		if result != test.expected {
			t.Errorf("intcmp(%q, %q) = %d; expected %d", test.a, test.b, result, test.expected)
		}
	}
}

// Test for isDigit function
func TestIsDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"abc", false},
		{"123abc", false},
		{"", false},
		{"0", true},
		{"-1", false},
	}

	for _, test := range tests {
		result := isDigit(test.input)
		if result != test.expected {
			t.Errorf("isDigit(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

// Test for mtimecmp function
func TestMtimecmp(t *testing.T) {
	// Create two temporary files with different modification times
	file1, _ := os.CreateTemp("", "file1")
	file2, _ := os.CreateTemp("", "file2")
	defer os.Remove(file1.Name())
	defer os.Remove(file2.Name())

	file1Time := time.Now().Add(-1 * time.Hour)
	file2Time := time.Now()

	os.Chtimes(file1.Name(), file1Time, file1Time)
	os.Chtimes(file2.Name(), file2Time, file2Time)

	info1, _ := os.Stat(file1.Name())
	info2, _ := os.Stat(file2.Name())

	if mtimecmp(info1, info2) != -1 {
		t.Error("Expected file1 to be older than file2")
	}
	if mtimecmp(info2, info1) != 1 {
		t.Error("Expected file2 to be newer than file1")
	}
	if mtimecmp(info1, info1) != 0 {
		t.Error("Expected file1 to be the same age as itself")
	}
}

// Test for file type check functions
func TestFileChecks(t *testing.T) {
	// Creating a temporary file and directory for testing
	tempFile, _ := os.CreateTemp("", "tempfile")
	tempDir, _ := os.MkdirTemp("", "tempdir")
	defer os.Remove(tempFile.Name())
	defer os.RemoveAll(tempDir)

	// Write some content to the temporary file
	tempFile.WriteString("test content")
	tempFile.Close()

	if !isRegularFile(tempFile.Name()) {
		t.Errorf("Expected %s to be a regular file", tempFile.Name())
	}

	if !isDirectory(tempDir) {
		t.Errorf("Expected %s to be a directory", tempDir)
	}

	if !fileExists(tempFile.Name()) {
		t.Errorf("Expected %s to exist", tempFile.Name())
	}

	if !isReadable(tempFile.Name()) {
		t.Errorf("Expected %s to be readable", tempFile.Name())
	}

	if !isWritable(tempFile.Name()) {
		t.Errorf("Expected %s to be writable", tempFile.Name())
	}

	if !isNonEmptyString("test") {
		t.Errorf("Expected 'test' to be a non-empty string")
	}

	if !isEmptyString("") {
		t.Errorf("Expected '' to be an empty string")
	}

	if !hasSize(tempFile.Name()) {
		t.Errorf("Expected %s to have a size", tempFile.Name())
	}

	// Skip isTerminal test if not in a terminal environment
	if os.Getenv("TERM") != "" {
		if !isTerminal("1") {
			t.Errorf("Expected file descriptor 0 to be a terminal")
		}
	}
}

// Test for binary test functions
func TestBinaryTests(t *testing.T) {
	// String comparisons
	if !stringsEqual("test", "test") {
		t.Error(`Expected stringsEqual("test", "test") to be true`)
	}
	if stringsNotEqual("test", "test") {
		t.Error(`Expected stringsNotEqual("test", "test") to be false`)
	}

	// Integer comparisons
	if !integersEqual("123", "123") {
		t.Error(`Expected integersEqual("123", "123") to be true`)
	}
	if integersNotEqual("123", "123") {
		t.Error(`Expected integersNotEqual("123", "123") to be false`)
	}
	if !integerGreaterThan("124", "123") {
		t.Error(`Expected integerGreaterThan("124", "123") to be true`)
	}
	if !integerGreaterOrEqual("124", "124") {
		t.Error(`Expected integerGreaterOrEqual("124", "124") to be true`)
	}
	if !integerLessThan("123", "124") {
		t.Error(`Expected integerLessThan("123", "124") to be true`)
	}
	if !integerLessOrEqual("124", "124") {
		t.Error(`Expected integerLessOrEqual("124", "124") to be true`)
	}

	// File comparisons
	file1, err := os.CreateTemp("", "file1")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(file1.Name())

	file2, err := os.CreateTemp("", "file2")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(file2.Name())

	if !filesEqual(file1.Name(), file1.Name()) {
		t.Errorf("Expected %s to be equal to itself", file1.Name())
	}
	if filesEqual(file1.Name(), file2.Name()) {
		t.Errorf("Expected %s and %s to not be equal", file1.Name(), file2.Name())
	}

	file1Time := time.Now().Add(-1 * time.Hour)
	file2Time := time.Now()

	if err := os.Chtimes(file1.Name(), file1Time, file1Time); err != nil {
		t.Fatalf("Failed to change file time: %v", err)
	}
	if err := os.Chtimes(file2.Name(), file2Time, file2Time); err != nil {
		t.Fatalf("Failed to change file time: %v", err)
	}

	if !fileOlderThan(file1.Name(), file2.Name()) {
		t.Errorf("Expected %s to be older than %s", file1.Name(), file2.Name())
	}
	if !fileNewerThan(file2.Name(), file1.Name()) {
		t.Errorf("Expected %s to be newer than %s", file2.Name(), file1.Name())
	}
}

// Test for helper functions
func TestHelperFunctions(t *testing.T) {
	// Test isType and isMode
	tempFile, _ := os.CreateTemp("", "tempfile")
	defer os.Remove(tempFile.Name())

	if !isMode(tempFile.Name(), os.ModePerm) {
		t.Errorf("Expected %s to have permissions", tempFile.Name())
	}

	// Skip isattyFd test if not in a terminal environment
	if os.Getenv("TERM") != "" {
		if !isattyFd(1) {
			t.Errorf("Expected file descriptor 0 to be a TTY")
		}
	}
}

// Test for performTest function
func TestPerformTest(t *testing.T) {
	tests := []struct {
		args     []string
		expected bool
	}{
		{[]string{"-f", "test_test.go"}, true},
		{[]string{"-d", "test_test.go"}, false},
		{[]string{"-r", "test_test.go"}, true},
		{[]string{"-w", "test_test.go"}, true},
		{[]string{"-e", "test_test.go"}, true},
		{[]string{"-n", "test"}, true},
		{[]string{"-z", ""}, true},
		{[]string{"test", "=", "test"}, true},
		{[]string{"test", "!=", "test"}, false},
		{[]string{"123", "-eq", "123"}, true},
		{[]string{"123", "-ne", "123"}, false},
		{[]string{"124", "-gt", "123"}, true},
		{[]string{"124", "-ge", "124"}, true},
		{[]string{"123", "-lt", "124"}, true},
		{[]string{"124", "-le", "124"}, true},
		{[]string{"!", "-f", "test_test.go"}, false},
		{[]string{"!", "-d", "test_test.go"}, true},
	}

	for _, test := range tests {
		result := performTest(test.args)
		if result != test.expected {
			t.Errorf("performTest(%v) = %v; expected %v", test.args, result, test.expected)
		}
	}
}
