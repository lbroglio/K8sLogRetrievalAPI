package main

// TODO - Consider JSON logging

import (
	"log/slog"
	"net/http"

	"github.com/lbroglio/K8sLogRetrievalAPI.git/internal/api"
)

func main() {

	// Configure and launch server
	mux := http.NewServeMux()
	mux.HandleFunc("/log-sources", api.HandleLogSource)
	mux.HandleFunc("/log-sources/{id}", api.HandleLogSourceById)

	slog.Info("Starting Server - Listening on port 8080")
	http.ListenAndServe(":8080", mux)
}
