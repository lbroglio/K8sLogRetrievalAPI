package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/service"
)

// ErrorResponse A standardized API error structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// TODO - Docu
func returnErrorResponse(writer http.ResponseWriter, statusCode int, errorMessage string) {
	writer.WriteHeader(statusCode)
	writer.Header().Set("Content-Type", "application/json")
	errorResponse := ErrorResponse{Error: errorMessage}
	json.NewEncoder(writer).Encode(errorResponse)
}

// TODO - Docu
func returnErrorResponseFromCustomError(writer http.ResponseWriter, customErr error) {
	var serviceErr service.ServiceError
	if errors.As(customErr, &serviceErr) {
		returnErrorResponse(writer, serviceErr.HttpStatusCode(), serviceErr.Error())
	} else {
		// If the error is not a ServiceError, return a generic 500 Internal Server Error
		returnErrorResponse(writer, http.StatusInternalServerError, "Unexpected error occurred")
	}
}
