package service

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/logsource"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/persistence"
)

func testFileLogSource(id string, name string) logsource.LogSource {
	return logsource.LogSource{
		ID:   id,
		Name: name,
		Type: logsource.File,
		FileLogSource: logsource.FileLogSource{
			FilePath:    "/var/log",
			FilePattern: name + ".log",
		},
	}
}

func assertServiceError(t *testing.T, err error, expectedCode int, expectedMessage string) {
	t.Helper()

	if err == nil {
		t.Fatal("Expected a service error, got nil")
	}

	var serviceErr ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("Expected a ServiceError, got %T: %v", err, err)
	}
	if serviceErr.Code != expectedCode {
		t.Errorf("Expected error code %d, got %d", expectedCode, serviceErr.Code)
	}
	if serviceErr.Message != expectedMessage {
		t.Errorf("Expected error message %q, got %q", expectedMessage, serviceErr.Message)
	}
}

func TestNewLogRetrievalService_NoSavedLogSources_StartsEmpty(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())

	service := NewLogRetrievalService()

	if logSources := service.GetAllLogSources(); len(logSources) != 0 {
		t.Errorf("Expected an empty service, got %v", logSources)
	}
}

func TestNewLogRetrievalService_LoadsSavedLogSources(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)

	expectedLogSource := testFileLogSource("test-id", "test-source")
	if err := persistence.SaveLogSource(expectedLogSource); err != nil {
		t.Fatalf("Failed to save test log source: %v", err)
	}

	service := NewLogRetrievalService()
	logSources := service.GetAllLogSources()

	if !reflect.DeepEqual(logSources, []logsource.LogSource{expectedLogSource}) {
		t.Errorf("Expected log sources %v, got %v", []logsource.LogSource{expectedLogSource}, logSources)
	}
}

func TestLogRetrievalService_GetAllLogSources_ReturnsAllSources(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()
	first := testFileLogSource("first-id", "first-source")
	second := testFileLogSource("second-id", "second-source")
	service.logStore.Put(first.ID, first)
	service.logStore.Put(second.ID, second)

	logSources := service.GetAllLogSources()
	if len(logSources) != 2 {
		t.Fatalf("Expected 2 log sources, got %d", len(logSources))
	}
	for _, expected := range []logsource.LogSource{first, second} {
		if !containsLogSource(logSources, expected) {
			t.Errorf("Expected log source %v in %v", expected, logSources)
		}
	}
}

func TestLogRetrievalService_GetLogSourceByID_Success(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()
	expectedLogSource := testFileLogSource("test-id", "test-source")
	service.logStore.Put(expectedLogSource.ID, expectedLogSource)

	logSource, err := service.GetLogSourceByID(expectedLogSource.ID)
	if err != nil {
		t.Fatalf("Failed to get log source: %v", err)
	}
	if !reflect.DeepEqual(logSource, expectedLogSource) {
		t.Errorf("Expected log source %v, got %v", expectedLogSource, logSource)
	}
}

func TestLogRetrievalService_GetLogSourceByID_NotFoundError(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()

	logSource, err := service.GetLogSourceByID("missing-id")

	if !reflect.DeepEqual(logSource, logsource.LogSource{}) {
		t.Errorf("Expected an empty log source, got %v", logSource)
	}
	assertServiceError(t, err, 404, "Log source with the id missing-id does not exist")
}

func TestLogRetrievalService_AddNewLogSource_Success(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()
	expectedLogSource := testFileLogSource("test-id", "test-source")

	logSource, err := service.AddNewLogSource(expectedLogSource)
	if err != nil {
		t.Fatalf("Failed to add log source: %v", err)
	}
	if !reflect.DeepEqual(logSource, expectedLogSource) {
		t.Errorf("Expected added log source %v, got %v", expectedLogSource, logSource)
	}
	if !service.logStore.ContainsKey(expectedLogSource.ID) {
		t.Error("Expected added log source to be in the service store")
	}
}

func TestLogRetrievalService_AddNewLogSource_AlreadyExists_Error(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()
	logSource := testFileLogSource("test-id", "test-source")
	service.logStore.Put(logSource.ID, logSource)

	_, err := service.AddNewLogSource(logSource)

	assertServiceError(t, err, 409, "log source with the id test-id already exists")
}

func TestLogRetrievalService_AddNewLogSource_PersistenceError(t *testing.T) {
	tempDir := t.TempDir() + "/missing"
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)
	service := NewLogRetrievalService()
	logSource := testFileLogSource("test-id", "test-source")

	_, err := service.AddNewLogSource(logSource)

	if err == nil {
		t.Fatal("Expected a persistence error")
	}
	var serviceErr ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("Expected a ServiceError, got %T: %v", err, err)
	}
	if serviceErr.Code != 500 {
		t.Errorf("Expected error code 500, got %d", serviceErr.Code)
	}
}

func TestLogRetrievalService_UpdateLogSource_Success(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()
	originalLogSource := testFileLogSource("test-id", "original-source")
	updatedLogSource := testFileLogSource("test-id", "updated-source")
	service.logStore.Put(originalLogSource.ID, originalLogSource)

	logSource, err := service.UpdateLogSource(updatedLogSource)
	if err != nil {
		t.Fatalf("Failed to update log source: %v", err)
	}
	if !reflect.DeepEqual(logSource, updatedLogSource) {
		t.Errorf("Expected updated log source %v, got %v", updatedLogSource, logSource)
	}
	if storedLogSource := service.logStore.Get(updatedLogSource.ID); !reflect.DeepEqual(storedLogSource, updatedLogSource) {
		t.Errorf("Expected stored log source %v, got %v", updatedLogSource, storedLogSource)
	}
}

func TestLogRetrievalService_UpdateLogSource_NotFoundError(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()
	logSource := testFileLogSource("missing-id", "missing-source")

	_, err := service.UpdateLogSource(logSource)

	assertServiceError(t, err, 404, "Log source with the id missing-id does not exist")
}

func TestLogRetrievalService_UpdateLogSource_PersistenceError(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)
	service := NewLogRetrievalService()
	logSource := testFileLogSource("test-id", "test-source")
	service.logStore.Put(logSource.ID, logSource)

	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("Failed to remove persistence root: %v", err)
	}

	_, err := service.UpdateLogSource(logSource)
	if err == nil {
		t.Fatal("Expected a persistence error")
	}
	var serviceErr ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("Expected a ServiceError, got %T: %v", err, err)
	}
	if serviceErr.Code != 500 {
		t.Errorf("Expected error code 500, got %d", serviceErr.Code)
	}
}

func TestLogRetrievalService_DeleteLogSourceByID_Success(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()
	logSource := testFileLogSource("test-id", "test-source")
	if _, err := service.AddNewLogSource(logSource); err != nil {
		t.Fatalf("Failed to add log source: %v", err)
	}

	err := service.DeleteLogSourceByID(logSource.ID)
	if err != nil {
		t.Fatalf("Failed to delete log source: %v", err)
	}
	if service.logStore.ContainsKey(logSource.ID) {
		t.Error("Expected deleted log source to be removed from the service store")
	}
}

func TestLogRetrievalService_DeleteLogSourceByID_NotFoundError(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()

	err := service.DeleteLogSourceByID("missing-id")

	assertServiceError(t, err, 404, "Log source with the id missing-id does not exist")
}

func TestLogRetrievalService_DeleteLogSourceByID_PersistenceError(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", tempDir)
	service := NewLogRetrievalService()
	logSource := testFileLogSource("test-id", "test-source")
	service.logStore.Put(logSource.ID, logSource)

	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("Failed to remove persistence root: %v", err)
	}

	err := service.DeleteLogSourceByID(logSource.ID)
	if err == nil {
		t.Fatal("Expected a persistence error")
	}
	var serviceErr ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("Expected a ServiceError, got %T: %v", err, err)
	}
	if serviceErr.Code != 500 {
		t.Errorf("Expected error code 500, got %d", serviceErr.Code)
	}
	if !service.logStore.ContainsKey(logSource.ID) {
		t.Error("Expected log source to remain in the service store")
	}
}

func TestLogRetrievalService_RetrieveAllLogsFromSourceById_Success(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	logDirectory := t.TempDir()
	logSource := testFileLogSource("test-id", "test-source")
	logSource.FileLogSource.FilePath = logDirectory
	if err := os.WriteFile(filepath.Join(logDirectory, "test-source.log"), []byte("INFO 1\nINFO 2"), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	service := NewLogRetrievalService()
	service.logStore.Put(logSource.ID, logSource)

	logs, err := service.RetrieveAllLogsFromSourceById(logSource.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve logs: %v", err)
	}
	if !reflect.DeepEqual(logs, []string{"INFO 1", "INFO 2"}) {
		t.Errorf("Expected logs %v, got %v", []string{"INFO 1", "INFO 2"}, logs)
	}
}

func TestLogRetrievalService_RetrieveAllLogsFromSourceById_NotFoundError(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()

	logs, err := service.RetrieveAllLogsFromSourceById("missing-id")

	if logs != nil {
		t.Errorf("Expected nil logs, got %v", logs)
	}
	assertServiceError(t, err, 404, "Log source with the id missing-id does not exist")
}

func TestLogRetrievalService_RetrieveNLogsFromSourceById_Success(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	logDirectory := t.TempDir()
	logSource := testFileLogSource("test-id", "test-source")
	logSource.FileLogSource.FilePath = logDirectory
	if err := os.WriteFile(filepath.Join(logDirectory, "test-source.log"), []byte("INFO 1\nINFO 2\nINFO 3"), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}
	service := NewLogRetrievalService()
	service.logStore.Put(logSource.ID, logSource)

	logs, err := service.RetrieveNLogsFromSourceById(logSource.ID, 2)
	if err != nil {
		t.Fatalf("Failed to retrieve logs: %v", err)
	}
	if !reflect.DeepEqual(logs, []string{"INFO 2", "INFO 3"}) {
		t.Errorf("Expected logs %v, got %v", []string{"INFO 2", "INFO 3"}, logs)
	}
}

func TestLogRetrievalService_RetrieveNLogsFromSourceById_NotFoundError(t *testing.T) {
	t.Setenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT", t.TempDir())
	service := NewLogRetrievalService()

	logs, err := service.RetrieveNLogsFromSourceById("missing-id", 2)

	if logs != nil {
		t.Errorf("Expected nil logs, got %v", logs)
	}
	assertServiceError(t, err, 404, "Log source with the id missing-id does not exist")
}

func containsLogSource(logSources []logsource.LogSource, expected logsource.LogSource) bool {
	for _, logSource := range logSources {
		if reflect.DeepEqual(logSource, expected) {
			return true
		}
	}
	return false
}
