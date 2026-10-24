package persistence

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/config"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/logsource"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/util"
)

// SaveLogSource saves the given log source to disk in a JSON file.
func SaveLogSource(source logsource.LogSource) error {
	persistenceRoot := config.GetPersistenceRootPath()

	saveFilePath := filepath.Join(persistenceRoot, "log_sources.json")
	logSources := make(map[string]logsource.LogSource)

	// If there are already saved log sources, load them
	if util.FileExists(saveFilePath) {
		savedLogSources, err := LoadSavedLogSources()
		if err != nil {
			return fmt.Errorf("Failed to load saved log sources. Root Cause: %w", err)
		}
		maps.Copy(logSources, savedLogSources)
	}
	logSources[source.ID] = source

	logSourcesJSON, err := json.Marshal(logSources)
	if err != nil {
		return fmt.Errorf("Failed to marshal log sources. Root Cause: %w", err)
	}

	err = os.WriteFile(saveFilePath, logSourcesJSON, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write log sources to disk. Root Cause: %w", err)
	}

	return nil
}

// LoadSavedLogSources loads the saved log sources from disk and returns them as a slice of LogSource objects.
func LoadSavedLogSources() (map[string]logsource.LogSource, error) {
	persistenceRoot := config.GetPersistenceRootPath()

	saveFilePath := filepath.Join(persistenceRoot, "log_sources.json")

	contents, err := os.ReadFile(saveFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log sources from disk. Root Cause: %w", err)
	}

	logSources := make(map[string]logsource.LogSource)
	err = json.Unmarshal(contents, &logSources)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal log sources. Root Cause: %w", err)
	}

	return logSources, nil
}
