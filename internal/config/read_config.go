// Package repsonsible for fetching configuration values from the environment variables.
//
// This packkage contains isolated functions which other packages can use to get
// configurations without needing to be coupled to the environment variable names directly.
//
// The reason to use environment variables for config is to make it easier to run this
// application in a containerized environment.
package config

import (
	"os"
)

// Get the root folder where the application will store its persistent data.
func GetPersistenceRootPath() string {
	return os.Getenv("K8SLOGRETRIEVALAPI_PERSISTENCE_ROOT")
}
