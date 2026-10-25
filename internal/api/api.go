package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/logsource"
	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/service"
)

var logService = service.NewLogRetrievalService()

// TODO: Maybe break this out into separate functions
// handleLogSource handles HTTP to the /log-sources endpoint
func HandleLogSource(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		slog.Info("Received GET request to retrieve all log sources")
		allLogSources := logService.GetAllLogSources()

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(allLogSources)
	// TODO - Make wrapper request format to abstract embedded type for POST and PUT here
	case http.MethodPost:
		slog.Info("Received POST request to add a new log source")
		newLogSource := logsource.LogSource{}
		err := json.NewDecoder(request.Body).Decode(&newLogSource)
		if err != nil {
			slog.Error("Failed to decode new log source request body. Returning 400 Bad Request.", "error", err)
			returnErrorResponse(writer, http.StatusBadRequest, "Invalid request body")
			return
		}
		// Add the new log source to the in-memory store and persist it to disk
		_, err = logService.AddNewLogSource(newLogSource)
		if err != nil {
			returnErrorResponseFromCustomError(writer, err)
			return
		}
		writer.WriteHeader(http.StatusCreated)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(newLogSource)
	case http.MethodPut:
		slog.Info("Received PUT request to update log source.")

		// Parse the request body to get the updated log source
		updatedLogSource := logsource.LogSource{}
		err := json.NewDecoder(request.Body).Decode(&updatedLogSource)
		if err != nil {
			slog.Error("Failed to decode update log source request body. Returning 400 Bad Request.", "error", err)
			returnErrorResponse(writer, http.StatusBadRequest, "Invalid request body")
			return
		}
		_, err = logService.UpdateLogSource(updatedLogSource)
		if err != nil {
			returnErrorResponseFromCustomError(writer, err)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(updatedLogSource)

	// Throw 405 Method Not Allowed for unsupported methods
	default:
		slog.Error("Received unsupported HTTP method request for /log-sources. Returning 405 Method Not Allowed.", "method", request.Method)
		returnErrorResponse(writer, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleLogSourceById handles HTTP requests to the /log-sources/{id} endpoint
func HandleLogSourceById(writer http.ResponseWriter, request *http.Request) {
	// Extract the log source ID from the URL path
	id := request.PathValue("id")

	switch request.Method {
	case http.MethodGet:
		slog.Info("Received GET request to retrieve log source by ID", "id", id)
		logSource, err := logService.GetLogSourceByID(id)
		if err != nil {
			returnErrorResponseFromCustomError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(logSource)
	// Throw 405 Method Not Allowed for unsupported methods
	default:
		slog.Error("Received unsupported HTTP method request for /log-sources/{id}. Returning 405 Method Not Allowed.", "method", request.Method)
		returnErrorResponse(writer, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
