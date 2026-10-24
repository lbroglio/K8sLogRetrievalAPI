package api

import (
	"encoding/json"
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
		allLogSources := logStore.GetAllValues()
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(allLogSources)
	case http.MethodPost:
		// Parse the request body to get the new log source
		newLogSource := logsource.LogSource{}
		err := json.NewDecoder(request.Body).Decode(&newLogSource)
		if err != nil {
			returnErrorResponse(writer, http.StatusBadRequest, "Invalid request body")
			return
		}
		// TODO: Validate the new log source fields (e.g., ID, Name, Type, etc.) before proceeding.
		// TODO: Generate a unique ID for the new log source if not provided, or validate the provided ID.

		idAlreadyExists := logStore.ContainsKey(newLogSource.ID)
		if idAlreadyExists {
			returnErrorResponse(writer, http.StatusConflict, "Log source with the same ID already exists")
			return
		}

		// Add the new log source to the in-memory store and persist it to disk
		err = persistence.SaveLogSource(newLogSource)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Header().Set("Content-Type", "application/json")
			errorResponse := ErrorResponse{Error: "Failed to save log source"}
			json.NewEncoder(writer).Encode(errorResponse)
			return
		}
		logStore.Put(newLogSource.ID, newLogSource)

		writer.WriteHeader(http.StatusCreated)
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(newLogSource)
	case http.MethodPut:
		// Parse the request body to get the updated log source
		updatedLogSource := logsource.LogSource{}
		err := json.NewDecoder(request.Body).Decode(&updatedLogSource)
		if err != nil {
			returnErrorResponse(writer, http.StatusBadRequest, "Invalid request body")
			return
		}

		idExists := logStore.ContainsKey(updatedLogSource.ID)
		if !idExists {
			returnErrorResponse(writer, http.StatusNotFound, "Log source with the given ID does not exist")
			return
		}

		// TODO: Validate the updated log source fields (e.g., Name, Type, etc.) before proceeding.

		// Update the log source in the in-memory store and persist it to disk
		err = persistence.SaveLogSource(updatedLogSource)
		if err != nil {
			returnErrorResponse(writer, http.StatusInternalServerError, "Failed to update log source")
			return
		}
		logStore.Put(updatedLogSource.ID, updatedLogSource)

	// Throw 405 Method Not Allowed for unsupported methods
	default:
		returnErrorResponse(writer, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func main() {
	logStore = util.NewConcurrentMap[string, logsource.LogSource]()
}
