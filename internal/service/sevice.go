package service

import (
	"log/slog"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/logsource"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/persistence"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/util"
)

type LogRetrievalService struct {
	logStore *util.ConcurrentMap[string, logsource.LogSource]
}

// NewLogRetrievalService initializes a new LogRetrievalService, loading any saved log sources from disk.
func NewLogRetrievalService() *LogRetrievalService {
	// Initialize the log store
	logStore := util.NewConcurrentMap[string, logsource.LogSource]()

	// Load persistence
	savedLogSources, err := persistence.LoadSavedLogSources()
	if err == nil {
		slog.Info("Loaded saved log sources from disk", "count", len(savedLogSources))
		logStore.PutAll(savedLogSources)
	} else {
		slog.Info("No saved log sources found on disk. Starting with an empty store.", "error", err)
	}

	return &LogRetrievalService{
		logStore: logStore,
	}
}

// GetAllLogSources retrieves all log sources from the in-memory store.
func (s *LogRetrievalService) GetAllLogSources() []logsource.LogSource {
	slog.Info("Retrieving all log sources")
	logSources := s.logStore.GetAllValues()
	slog.Info("Returning all log sources", "count", len(logSources))
	return logSources
}

// GetLogSourceByID retrieves a log source by its ID from the in-memory store.
func (s *LogRetrievalService) GetLogSourceByID(id string) (logsource.LogSource, error) {
	slog.Info("Retrieving log source with ID", "logSourceID", id)

	exists := s.logStore.ContainsKey(id)
	if !exists {
		slog.Error("Log source with the specified ID does not exist.", "logSourceID", id)
		return logsource.LogSource{}, NotFoundError("Log source with the id " + id + " does not exist")
	}

	logSource := s.logStore.Get(id)

	slog.Info("Returning log source", "logSourceID", id)
	return logSource, nil
}

// AddNewLogSource adds a new log source to the in-memory store and persists it to disk.
// TODO - Maybe return a pointer
func (s *LogRetrievalService) AddNewLogSource(logSource logsource.LogSource) (logsource.LogSource, error) {

	// TODO: Validate the new log source fields (e.g., ID, Name, Type, etc.) before proceeding.
	// TODO: Generate a unique ID for the new log source if not provided, or validate the provided ID.

	slog.Info("Adding new log source", "logSourceID", logSource.ID)

	idAlreadyExists := s.logStore.ContainsKey(logSource.ID)
	if idAlreadyExists {
		slog.Error("Log source with the same ID already exists.", "logSourceID", logSource.ID)
		return logSource, AlreadyExistsError("log source with the id " + logSource.ID + " already exists")
	}

	// Add the new log source to the in-memory store and persist it to disk
	newLogSource := logSource
	err := persistence.SaveLogSource(newLogSource)
	if err != nil {
		slog.Error("Failed to save log source. Returning 500 Internal Server Error.", "error", err)
		return newLogSource, GenericServiceError("Failed to save log source. Root Cause: " + err.Error())
	}

	s.logStore.Put(newLogSource.ID, newLogSource)
	slog.Info("Log source added successfully.", "logSourceID", newLogSource.ID)
	return newLogSource, nil
}

// UpdateLogSource updates an existing log source in the in-memory store and persists it to disk.
// TODO - Maybe return a pointer
func (s *LogRetrievalService) UpdateLogSource(logSource logsource.LogSource) (logsource.LogSource, error) {

	// TODO: Validate the new log source fields (e.g., ID, Name, Type, etc.) before proceeding.
	// TODO: Generate a unique ID for the new log source if not provided, or validate the provided ID.

	slog.Info("Updating log source", "logSourceID", logSource.ID)

	idAlreadyExists := s.logStore.ContainsKey(logSource.ID)
	if !idAlreadyExists {
		slog.Error("Log source with the same ID does not exist.", "logSourceID", logSource.ID)
		return logSource, NotFoundError("Log source with the id " + logSource.ID + " does not exist")
	}

	// Add the new log source to the in-memory store and persist it to disk
	newLogSource := logSource
	err := persistence.SaveLogSource(newLogSource)
	if err != nil {
		slog.Error("Failed to save log source. Returning 500 Internal Server Error.", "error", err)
		return newLogSource, GenericServiceError("Failed to save log source. Root Cause: " + err.Error())
	}

	s.logStore.Put(newLogSource.ID, newLogSource)
	slog.Info("Log source updated successfully.", "logSourceID", newLogSource.ID)
	return newLogSource, nil
}

// DeleteLogSource deletes a log source from the in-memory store and removes it from disk.
func (s *LogRetrievalService) DeleteLogSourceByID(id string) error {
	slog.Info("Deleting log source", "logSourceID", id)

	idExists := s.logStore.ContainsKey(id)
	if !idExists {
		slog.Error("Log source with the same ID does not exist.", "logSourceID", id)
		return NotFoundError("Log source with the id " + id + " does not exist")
	}

	err := persistence.DeleteSavedLogSource(id)
	if err != nil {
		slog.Error("Failed to delete log source from disk. Returning 500 Internal Server Error.", "error", err)
		return GenericServiceError("Failed to delete log source from disk. Root Cause: " + err.Error())
	}
	s.logStore.Delete(id)

	return nil
}
