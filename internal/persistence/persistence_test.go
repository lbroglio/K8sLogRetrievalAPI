package persistence_test

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/logsource"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/persistence"
)

func assertSavedFileContent(t *testing.T, expectedFile string, expectedLogSources map[string]logsource.LogSource) {
	t.Helper()

	// Load the saved log sources and verify the saved log source is present
	contentJSON, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read saved log sources: %v", err)
	}

	savedLogSources := make(map[string]logsource.LogSource)
	err = json.Unmarshal(contentJSON, &savedLogSources)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved log sources: %v", err)
	}

	result := compareSourcesMaps(t, expectedLogSources, savedLogSources)
	if !result {
		t.Fatalf("Saved log sources do not match expected. Expected: %v, Got: %v", expectedLogSources, savedLogSources)
	}
}

func compareSourcesMaps(t *testing.T, a map[string]logsource.LogSource, b map[string]logsource.LogSource) bool {
	t.Helper()

	return maps.EqualFunc(a, b, func(a logsource.LogSource, b logsource.LogSource) bool {
		return a.ID == b.ID
	})
}

func TestSaveLogSource_SingleLogSource_Success(t *testing.T) {
	// Create a temporary directory for writing the test file
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	// Create a sample log source
	logSource := logsource.LogSource{
		ID:   "test-id",
		Name: "test-source",
		Type: logsource.File,
	}

	expectedLogSources := make(map[string]logsource.LogSource)
	expectedLogSources[logSource.ID] = logSource

	err := persistence.SaveLogSource(logSource)
	if err != nil {
		t.Fatalf("Failed to save log source: %v", err)
	}

	// Load the saved log sources and verify the saved log source is present
	expectedFile := filepath.Join(tempDir, "log_sources.json")
	assertSavedFileContent(t, expectedFile, expectedLogSources)
}

func TestSaveLogSource_SingleExistingLogSource_Success(t *testing.T) {
	// Create a temporary directory for writing the test file
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	// Create a sample log source
	logSource := logsource.LogSource{
		ID:   "test-id",
		Name: "test-source",
		Type: logsource.File,
	}
	otherLogSource := logsource.LogSource{
		ID:   "test-id-2",
		Name: "test-source-2",
		Type: logsource.File,
	}

	expectedLogSources := make(map[string]logsource.LogSource)
	expectedLogSources[logSource.ID] = logSource
	expectedLogSources[otherLogSource.ID] = otherLogSource

	err := persistence.SaveLogSource(logSource)
	if err != nil {
		t.Fatalf("Failed to save log source: %v", err)
	}
	err = persistence.SaveLogSource(otherLogSource)
	if err != nil {
		t.Fatalf("Failed to save other log source: %v", err)
	}

	// Load the saved log sources and verify the saved log source is present
	expectedFile := filepath.Join(tempDir, "log_sources.json")
	assertSavedFileContent(t, expectedFile, expectedLogSources)
}

func TestSaveLogSource_MultipleExistingLogSource_Success(t *testing.T) {
	// Create a temporary directory for writing the test file
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	// Create a sample log source
	logSource := logsource.LogSource{
		ID:   "test-id",
		Name: "test-source",
		Type: logsource.File,
	}
	otherLogSource := logsource.LogSource{
		ID:   "test-id-2",
		Name: "test-source-2",
		Type: logsource.File,
	}
	thirdLogSource := logsource.LogSource{
		ID:   "test-id-3",
		Name: "test-source-3",
		Type: logsource.File,
	}

	expectedLogSources := make(map[string]logsource.LogSource)
	expectedLogSources[logSource.ID] = logSource
	expectedLogSources[otherLogSource.ID] = otherLogSource
	expectedLogSources[thirdLogSource.ID] = thirdLogSource

	err := persistence.SaveLogSource(logSource)
	if err != nil {
		t.Fatalf("Failed to save log source: %v", err)
	}
	err = persistence.SaveLogSource(otherLogSource)
	if err != nil {
		t.Fatalf("Failed to save other log source: %v", err)
	}
	err = persistence.SaveLogSource(thirdLogSource)
	if err != nil {
		t.Fatalf("Failed to save third log source: %v", err)
	}
	// Load the saved log sources and verify the saved log source is present
	expectedFile := filepath.Join(tempDir, "log_sources.json")
	assertSavedFileContent(t, expectedFile, expectedLogSources)
}

func TestSaveLogSource_PodLogSource_Success(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	logSource := logsource.LogSource{
		ID:   "test-pod-id",
		Name: "test-pod",
		Type: logsource.Pod,
		PodLogSource: logsource.PodLogSource{
			Namespace: "default",
			PodName:   "test-pod-123",
		},
	}

	err := persistence.SaveLogSource(logSource)
	if err != nil {
		t.Fatalf("Failed to save pod log source: %v", err)
	}

	expectedFile := filepath.Join(tempDir, "log_sources.json")
	assertSavedFileContent(t, expectedFile, map[string]logsource.LogSource{logSource.ID: logSource})
}

func TestSaveLogSource_InvalidExistingFile_Error(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	sourcesFilePath := filepath.Join(tempDir, "log_sources.json")
	err := os.WriteFile(sourcesFilePath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid log sources: %v", err)
	}

	err = persistence.SaveLogSource(logsource.LogSource{
		ID:   "test-id",
		Name: "test-source",
		Type: logsource.File,
	})
	if err == nil {
		t.Fatal("Expected an error when existing log sources are invalid")
	}
	if !strings.Contains(err.Error(), "Failed to load saved log sources") {
		t.Errorf("Expected saved-log-sources error, got %v", err)
	}
}

func TestSaveLogSource_PersistenceRootDoesNotExist_Error(t *testing.T) {
	tempDir := filepath.Join(t.TempDir(), "missing")
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	err := persistence.SaveLogSource(logsource.LogSource{
		ID:   "test-id",
		Name: "test-source",
		Type: logsource.File,
	})
	if err == nil {
		t.Fatal("Expected an error when persistence root does not exist")
	}
	if !strings.Contains(err.Error(), "Failed to write log sources to disk") {
		t.Errorf("Expected write error, got %v", err)
	}
}

func TestLoadSavedLogSources_SingleLogSource_Success(t *testing.T) {
	// Create a temporary directory for writing the test file
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	// Create a sample log source
	logSource := logsource.LogSource{
		ID:   "test-id",
		Name: "test-source",
		Type: logsource.File,
	}

	expectedLogSources := make(map[string]logsource.LogSource)
	expectedLogSources[logSource.ID] = logSource

	err := persistence.SaveLogSource(logSource)
	if err != nil {
		t.Fatalf("Failed to save log source: %v", err)
	}

	// Load the saved log sources and verify the saved log source is present
	savedLogSources, err := persistence.LoadSavedLogSources()
	if err != nil {
		t.Fatalf("Failed to load saved log sources: %v", err)
	}

	result := compareSourcesMaps(t, expectedLogSources, savedLogSources)
	if !result {
		t.Fatalf("Saved log sources do not match expected. Expected: %v, Got: %v", expectedLogSources, savedLogSources)
	}
}

func TestLoadSavedLogSources_NoSavedLogSources_Error(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	_, err := persistence.LoadSavedLogSources()
	if err == nil {
		t.Fatal("Expected an error when saved log sources do not exist")
	}
	if !strings.Contains(err.Error(), "Failed to read log sources from disk") {
		t.Errorf("Expected read error, got %v", err)
	}
}

func TestLoadSavedLogSources_InvalidJSON_Error(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	sourcesFilePath := filepath.Join(tempDir, "log_sources.json")
	err := os.WriteFile(sourcesFilePath, []byte("{invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid log sources: %v", err)
	}

	_, err = persistence.LoadSavedLogSources()
	if err == nil {
		t.Fatal("Expected an error when saved log sources contain invalid JSON")
	}
	if !strings.Contains(err.Error(), "Failed to unmarshal log sources") {
		t.Errorf("Expected unmarshal error, got %v", err)
	}
}

func TestLoadSavedLogSources_EmptyMap_Success(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	sourcesFilePath := filepath.Join(tempDir, "log_sources.json")
	err := os.WriteFile(sourcesFilePath, []byte("{}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write empty log sources: %v", err)
	}

	savedLogSources, err := persistence.LoadSavedLogSources()
	if err != nil {
		t.Fatalf("Failed to load empty log sources: %v", err)
	}
	if len(savedLogSources) != 0 {
		t.Errorf("Expected no saved log sources, got %v", savedLogSources)
	}
}

func TestLoadSavedLogSources_MultipleLogSources_Success(t *testing.T) {
	// Create a temporary directory for writing the test file
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	// Create a sample log source
	logSource := logsource.LogSource{
		ID:   "test-id",
		Name: "test-source",
		Type: logsource.File,
	}
	otherLogSource := logsource.LogSource{
		ID:   "test-id-2",
		Name: "test-source-2",
		Type: logsource.File,
	}
	thirdLogSource := logsource.LogSource{
		ID:   "test-id-3",
		Name: "test-source-3",
		Type: logsource.File,
	}

	expectedLogSources := make(map[string]logsource.LogSource)
	expectedLogSources[logSource.ID] = logSource
	expectedLogSources[otherLogSource.ID] = otherLogSource
	expectedLogSources[thirdLogSource.ID] = thirdLogSource

	sourcesJSON, err := json.Marshal(expectedLogSources)
	if err != nil {
		t.Fatalf("Failed to marshal log sources: %v", err)
	}
	sourcesFilePath := filepath.Join(tempDir, "log_sources.json")
	err = os.WriteFile(sourcesFilePath, sourcesJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write log sources to disk: %v", err)
	}

	// Load the saved log sources and verify the saved log source is present
	savedLogSources, err := persistence.LoadSavedLogSources()
	if err != nil {
		t.Fatalf("Failed to load saved log sources: %v", err)
	}

	result := compareSourcesMaps(t, expectedLogSources, savedLogSources)
	if !result {
		t.Fatalf("Saved log sources do not match expected. Expected: %v, Got: %v", expectedLogSources, savedLogSources)
	}
}
