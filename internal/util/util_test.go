package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists_FileExists_ReturnsTrue(t *testing.T) {
	tmpDir := t.TempDir()

	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	// Create a temporary file
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer file.Close()

	result := FileExists(tmpFile)

	if !result {
		t.Errorf("Expected FileExists to return true for existing file, got false")
	}
}

func TestFileExists_FileDoesNotExist_ReturnsFalse(t *testing.T) {
	result := FileExists("/path/to/nonexistent/file.txt")

	if result {
		t.Errorf("Expected FileExists to return false for non-existent file, got true")
	}
}
