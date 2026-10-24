package main

// TODO - Consider JSON logging

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/logsource"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/persistence"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/util"
)

var logStore *util.ConcurrentMap[string, logsource.LogSource]

// ErrorResponse A standardized API error structure
type ErrorResponse struct {
	Error string `json:"error"`
}

func returnErrorResponse(writer http.ResponseWriter, statusCode int, errorMessage string) {
	writer.WriteHeader(statusCode)
	writer.Header().Set("Content-Type", "application/json")
	errorResponse := ErrorResponse{Error: errorMessage}
	json.NewEncoder(writer).Encode(errorResponse)
}

// TODO: Maybe break this out into separate functions
// handleLogSource handles HTTP to the /log-sources endpoint
func handleLogSource(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		slog.Info("Received GET request to retrieve all log sources")
		allLogSources := logStore.GetAllValues()
		slog.Info("Returning all log sources", "count", len(allLogSources))

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(allLogSources)
	// TODO - Make wrapper request format to abstract embedded type for POST and PUT here
	case http.MethodPost:

		slog.Info("Received POST request to add a new log source")
		// Parse the request body to get the new log source
		newLogSource := logsource.LogSource{}
		err := json.NewDecoder(request.Body).Decode(&newLogSource)
		if err != nil {
			slog.Error("Failed to decode new log source request body. Returning 400 Bad Request.", "error", err)
			returnErrorResponse(writer, http.StatusBadRequest, "Invalid request body")
			return
		}
		// TODO: Validate the new log source fields (e.g., ID, Name, Type, etc.) before proceeding.
		// TODO: Generate a unique ID for the new log source if not provided, or validate the provided ID.

		idAlreadyExists := logStore.ContainsKey(newLogSource.ID)
		if idAlreadyExists {
			slog.Error("Log source with the same ID already exists. Returning 409 Conflict.", "logSourceID", newLogSource.ID)
			returnErrorResponse(writer, http.StatusConflict, "Log source with the same ID already exists")
			return
		}

		// Add the new log source to the in-memory store and persist it to disk
		err = persistence.SaveLogSource(newLogSource)
		if err != nil {
			slog.Error("Failed to save log source. Returning 500 Internal Server Error.", "error", err)
			returnErrorResponse(writer, http.StatusInternalServerError, "Failed to save log source")
			return
		}
		logStore.Put(newLogSource.ID, newLogSource)
		slog.Info("Log source added successfully.", "logSourceID", newLogSource.ID)

		writer.WriteHeader(http.StatusCreated)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(newLogSource)
	case http.MethodPut:
		slog.Info("Received PUT request to update log source.")

		// Parse the request body to get the updated log source
		updatedLogSource := logsource.LogSource{}
		err := json.NewDecoder(request.Body).Decode(&updatedLogSource)
		if err != nil {
			slog.Error("Failed to decode updated log source request body. Returning 400 Bad Request.", "error", err)
			returnErrorResponse(writer, http.StatusBadRequest, "Invalid request body")
			return
		}

		idExists := logStore.ContainsKey(updatedLogSource.ID)
		if !idExists {
			slog.Error("Log source with the given ID does not exist. Returning 404 Not Found.", "logSourceID", updatedLogSource.ID)
			returnErrorResponse(writer, http.StatusNotFound, "Log source with the given ID does not exist")
			return
		}

		// TODO: Validate the updated log source fields (e.g., Name, Type, etc.) before proceeding.

		// Update the log source in the in-memory store and persist it to disk
		err = persistence.SaveLogSource(updatedLogSource)
		if err != nil {
			slog.Error("Failed to update log source. Returning 500 Internal Server Error.", "error", err)
			returnErrorResponse(writer, http.StatusInternalServerError, "Failed to update log source")
			return
		}
		logStore.Put(updatedLogSource.ID, updatedLogSource)

		slog.Info("Log source updated successfully.", "logSourceID", updatedLogSource.ID)
	// Throw 405 Method Not Allowed for unsupported methods
	default:
		returnErrorResponse(writer, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func main() {
	logStore = util.NewConcurrentMap[string, logsource.LogSource]()

	// Load persistence
	savedLogSources, err := persistence.LoadSavedLogSources()
	if err == nil {
		slog.Info("Loaded saved log sources from disk", "count", len(savedLogSources))
		logStore.PutAll(savedLogSources)
	} else {
		slog.Info("No saved log sources found on disk. Starting with an empty store.", "error", err)
	}

	// Configure and launch server
	mux := http.NewServeMux()
	mux.HandleFunc("/log-sources", handleLogSource)

	slog.Info("Starting Server - Listening on port 8080")
	http.ListenAndServe(":8080", mux)
}
