package logsource

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func createTestLogFile(directory string, filename string, contentPrefix string) string {
	logFilePath := filepath.Join(directory, filename)
	testLog := fmt.Sprintf("%s 1\n%s 2\n%s 3", contentPrefix, contentPrefix, contentPrefix)
	if err := os.WriteFile(logFilePath, []byte(testLog), 0644); err != nil {
		panic(err)
	}

	return logFilePath
}

func assertLogs(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected logs %v, got %v", want, got)
	}
}

func TestNewFileLogSource(t *testing.T) {
	source := NewFileLogSource("1", "Test Log Source", "/var/log", "app-{yyyy}-{MM}-{dd}.log")

	expected := &LogSource{
		ID:   "1",
		Name: "Test Log Source",
		Type: File,
		FileLogSource: FileLogSource{
			FilePath:    "/var/log",
			FilePattern: "app-{yyyy}-{MM}-{dd}.log",
		},
	}
	if !reflect.DeepEqual(source, expected) {
		t.Errorf("Expected source %#v, got %#v", expected, source)
	}
}

func TestNewPodLogSource(t *testing.T) {
	source := NewPodLogSource("1", "Test Pod", "default", "api-123")

	expected := &LogSource{
		ID:   "1",
		Name: "Test Pod",
		Type: Pod,
		PodLogSource: PodLogSource{
			Namespace: "default",
			PodName:   "api-123",
		},
	}
	if !reflect.DeepEqual(source, expected) {
		t.Errorf("Expected source %#v, got %#v", expected, source)
	}
}

func TestRetrieveAllLogsFromFileSource_NoMatchingFiles_Success(t *testing.T) {
	testSource := NewFileLogSource("1", "Test Log Source", t.TempDir(), "missing.log")

	logs, err := testSource.RetrieveAllLogs()
	if err != nil {
		t.Fatalf("Failed to retrieve logs from file: %v", err)
	}

	assertLogs(t, logs, []string{})
}

func TestRetrieveAllLogsFromFileSource_DatePattern_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := fmt.Sprintf("app-%s-%s-%s.log",
		time.Now().Format("2006"),
		time.Now().Format("01"),
		time.Now().Format("02"),
	)
	createTestLogFile(tmpDir, filename, "INFO")
	testSource := NewFileLogSource("1", "Test Log Source", tmpDir, "app-{yyyy}-{MM}-{dd}.log")

	logs, err := testSource.RetrieveAllLogs()
	if err != nil {
		t.Fatalf("Failed to retrieve logs from file: %v", err)
	}

	assertLogs(t, logs, []string{"INFO 1", "INFO 2", "INFO 3"})
}

func TestRetrieveAllLogsFromFileSource_ReadError(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, "test.log"), 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	testSource := NewFileLogSource("1", "Test Log Source", tmpDir, "test.log")

	logs, err := testSource.RetrieveAllLogs()
	if err == nil {
		t.Fatal("Expected an error when a matching path cannot be read")
	}
	assertLogs(t, logs, []string{})
	if !strings.Contains(err.Error(), "Failed to read log file") {
		t.Errorf("Expected read error, got %v", err)
	}
}

func TestRetrieveNLogsFromFileSource_Success(t *testing.T) {
	tmpDir := t.TempDir()
	createTestLogFile(tmpDir, "test.log", "INFO")
	testSource := NewFileLogSource("1", "Test Log Source", tmpDir, "test.log")

	tests := []struct {
		name string
		n    int
		want []string
	}{
		{name: "last two", n: 2, want: []string{"INFO 2", "INFO 3"}},
		{name: "all lines", n: 3, want: []string{"INFO 1", "INFO 2", "INFO 3"}},
		{name: "more than available", n: 5, want: []string{"INFO 1", "INFO 2", "INFO 3"}},
		{name: "zero lines", n: 0, want: []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logs, err := testSource.RetrieveNLogs(test.n)
			if err != nil {
				t.Fatalf("Failed to retrieve logs from file: %v", err)
			}
			assertLogs(t, logs, test.want)
		})
	}
}

func TestRetrieveNLogsFromFileSource_MultipleFiles_Error(t *testing.T) {
	tmpDir := t.TempDir()
	createTestLogFile(tmpDir, "test1.log", "INFO")
	createTestLogFile(tmpDir, "test2.log", "WARN")
	testSource := NewFileLogSource("1", "Test Log Source", tmpDir, "test*.log")

	logs, err := testSource.RetrieveNLogs(1)
	if err == nil {
		t.Fatal("Expected an error when multiple files match")
	}
	assertLogs(t, logs, []string{})
	if !strings.Contains(err.Error(), "Multiple log files match") {
		t.Errorf("Expected multiple-file error, got %v", err)
	}
}

func TestRetrieveNLogsFromFileSource_NoMatchingFiles_Success(t *testing.T) {
	testSource := NewFileLogSource("1", "Test Log Source", t.TempDir(), "missing.log")

	logs, err := testSource.RetrieveNLogs(1)
	if err != nil {
		t.Fatalf("Failed to retrieve logs from file: %v", err)
	}
	assertLogs(t, logs, []string{})
}

func TestRetrieveLogs_UnsupportedSourceType(t *testing.T) {
	source := LogSource{Type: LogSourceType(99)}

	tests := []struct {
		name string
		call func() ([]string, error)
	}{
		{name: "all logs", call: source.RetrieveAllLogs},
		{name: "n logs", call: func() ([]string, error) { return source.RetrieveNLogs(1) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logs, err := test.call()
			if err == nil {
				t.Fatal("Expected an unsupported-source error")
			}
			assertLogs(t, logs, []string{})
			if !strings.Contains(err.Error(), "Unsupported log source type") {
				t.Errorf("Expected unsupported-source error, got %v", err)
			}
		})
	}
}

func TestReplaceInPattern(t *testing.T) {
	got := replaceInPattern("{yyyy}/{MM}/{dd}-{MM}", "MM", "10")
	want := "{yyyy}/10/{dd}-10"
	if got != want {
		t.Errorf("Expected %q, got %q", want, got)
	}
}

func TestReplaceDateInPattern(t *testing.T) {
	got := replaceDateInPattern("{yyyy}-{MM}-{dd}.log")
	want := fmt.Sprintf("%s-%s-%s.log",
		time.Now().Format("2006"),
		time.Now().Format("01"),
		time.Now().Format("02"),
	)
	if got != want {
		t.Errorf("Expected %q, got %q", want, got)
	}
}

func TestGetAllFilesMatchingPattern_InvalidPattern(t *testing.T) {
	files, err := getAllFilesMatchingPattern(t.TempDir(), "[")
	if err == nil {
		t.Fatal("Expected an invalid-pattern error")
	}
	assertLogs(t, files, []string{})
	if !strings.Contains(err.Error(), "Failed to fetch log files") {
		t.Errorf("Expected pattern error, got %v", err)
	}
}

func TestGetAllFilesMatchingPattern_MatchesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	first := createTestLogFile(tmpDir, "one.log", "INFO")
	second := createTestLogFile(tmpDir, "two.log", "WARN")

	files, err := getAllFilesMatchingPattern(tmpDir, "*.log")
	if err != nil {
		t.Fatalf("Failed to find matching files: %v", err)
	}
	assertLogs(t, files, []string{first, second})
}

func TestRetrieveAllLogsFromFileSource_OneFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	createTestLogFile(tmpDir, "test.log", "INFO")
	testSource := NewFileLogSource("1", "Test Log Source", tmpDir, "test.log")

	logs, err := testSource.RetrieveAllLogs()
	if err != nil {
		t.Fatalf("Failed to retrieve logs from file: %v", err)
	}

	// Check the number of logs
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}

	// Check the content of the logs
	assertLogs(t, logs, []string{"INFO 1", "INFO 2", "INFO 3"})
}

func TestRetrieveAllLogsFromFileSource_MultipleFiles_Success(t *testing.T) {
	tmpDir := t.TempDir()
	createTestLogFile(tmpDir, "test1.log", "INFO")
	createTestLogFile(tmpDir, "test2.log", "WARN")
	testSource := NewFileLogSource("1", "Test Log Source", tmpDir, "test*.log")

	logs, err := testSource.RetrieveAllLogs()
	if err != nil {
		t.Fatalf("Failed to retrieve logs from file: %v", err)
	}

	// Check the number of logs
	if len(logs) != 6 {
		t.Errorf("Expected 6 logs, got %d", len(logs))
	}

	// Check the content of the logs
	assertLogs(t, logs, []string{"INFO 1", "INFO 2", "INFO 3", "WARN 1", "WARN 2", "WARN 3"})
}
