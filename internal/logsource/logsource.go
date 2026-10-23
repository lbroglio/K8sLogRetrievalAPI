package logsource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TODO - Ability to retrieve logs from since a certain given a pattern for the timestamp.

type LogSourceType int

const (
	File LogSourceType = iota
	Pod
)

// FileLogSource represents a log source that reads logs from files in a specified directory.
type FileLogSource struct {
	// Path to the directory holding the log files
	FilePath string `json:"file_path"`
	// Pattern to match log files including replacement patterns (e.g., "{dd}-{MM}-{yyyy}.log")
	// All replacement patterns should be enclosed in curly braces {}.
	// Currently supported replacement patterns are:
	// 	- dd: Replaced with the current day of the month (01-31)
	// 	- MM: Replaced with the current month (01-12)
	// 	- yyyy: Replaced with the current year (e.g., 2024)
	// 	- *: Replaced with a wildcard that matches any sequence of characters. All matching log files will be processed.
	FilePattern string `json:"file_pattern"`
}

// PodLogSource represents a log source that reads logs from a specific pod in a Kubernetes cluster using the Kubernetes API.
type PodLogSource struct {
	// Namespace of the pod
	Namespace string `json:"namespace"`
	// Name of the pod
	PodName string `json:"pod_name"`
}

// LogSource represents a source of logs, which can be either from files or from a Kubernetes pod.
type LogSource struct {
	// Unique identifier for the log source
	ID string `json:"id"`
	// Human-readable name for the log source
	Name string `json:"name"`
	// Type of the log source (e.g., File, Pod)
	Type LogSourceType `json:"type"`
	// Details for the file log source (if Type is File)
	FileLogSource FileLogSource `json:"file_log_source,omitzero"`
	// Details for the pod log source (if Type is Pod)
	PodLogSource PodLogSource `json:"pod_log_source,omitzero"`
}

// TODO - Maybe generate id
// NewFileLogSource creates a new LogSource of type File with the provided parameters.
func NewFileLogSource(id string, name string, filePath string, filePattern string) *LogSource {
	FileLogSource := FileLogSource{
		FilePath:    filePath,
		FilePattern: filePattern,
	}

	return &LogSource{
		ID:            id,
		Name:          name,
		Type:          File,
		FileLogSource: FileLogSource,
	}
}

// TODO - Maybe generate id
// NewPodLogSource creates a new LogSource of type Pod with the provided parameters.
func NewPodLogSource(id string, name string, namespace string, podName string) *LogSource {
	PodLogSource := PodLogSource{
		Namespace: namespace,
		PodName:   podName,
	}

	return &LogSource{
		ID:           id,
		Name:         name,
		Type:         Pod,
		PodLogSource: PodLogSource,
	}
}

// TODO - Docu
func (source LogSource) RetrieveAllLogs() ([]string, error) {
	switch source.Type {
	case File:
		logs, err := retrieveLogsFromFileSource(source)
		if err != nil {
			return []string{}, err
		}
		return logs, nil
	case Pod:
		// TODO - Implement pod log retrieval
		return []string{}, fmt.Errorf("Pod log retrieval not implemented yet")
	default:
		return []string{}, fmt.Errorf("Unsupported log source type: %v", source.Type)
	}

}

// TODO - Documentation
func (source LogSource) RetrieveNLogs(n int) ([]string, error) {
	switch source.Type {
	case File:
		logs, err := retrieveLastNLinesFromFileSource(source, n)
		if err != nil {
			return []string{}, err
		}
		return logs, nil
	case Pod:
		// TODO - Implement pod log retrieval
		return []string{}, fmt.Errorf("Pod log retrieval not implemented yet")
	default:
		return []string{}, fmt.Errorf("Unsupported log source type: %v", source.Type)
	}
}

// retrieveLogsFromFileSource retrieves logs from a file-based log source.
// It reads all log files matching the specified pattern in the given directory and
// returns their contents as a slice of strings.
func retrieveLogsFromFileSource(source LogSource) ([]string, error) {
	effectiveFilePattern := replaceDateInPattern(source.FileLogSource.FilePattern)

	matchingFiles, err := getAllFilesMatchingPattern(source.FileLogSource.FilePath, effectiveFilePattern)
	if err != nil {
		return []string{}, err
	}

	logLines := []string{}
	for _, file := range matchingFiles {
		content, err := os.ReadFile(file)

		if err != nil {
			return []string{}, fmt.Errorf("Failed to read log file: %s. Root cause: %v", file, err)
		}
		logLines = append(logLines, strings.Split(string(content), "\n")...)
	}
	return logLines, nil
}

// retrieveLastNLinesFromFileSource retrieves the last N lines from a file-based log source.
// This function will throw an error if multiple files match the specified pattern, as it is ambiguous
// which file to read from.
func retrieveLastNLinesFromFileSource(source LogSource, n int) ([]string, error) {
	effectiveFilePattern := replaceDateInPattern(source.FileLogSource.FilePattern)

	matchingFiles, err := getAllFilesMatchingPattern(source.FileLogSource.FilePath, effectiveFilePattern)
	if err != nil {
		return []string{}, err
	}

	if len(matchingFiles) > 1 {
		return []string{},
			fmt.Errorf("Multiple log files match the pattern: %s. Refine your file pattern to match a single log file.", effectiveFilePattern)
	}
	if len(matchingFiles) == 0 {
		return []string{}, nil
	}
	targetFile := matchingFiles[0]

	fileContent, err := os.ReadFile(targetFile)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to read log file: %s. Root cause: %v", matchingFiles[0], err)
	}
	fileLines := strings.Split(string(fileContent), "\n")
	if n > len(fileLines) {
		n = len(fileLines)
	}
	logLines := []string{}
	logLines = append(logLines, fileLines[len(fileLines)-n:]...)

	return logLines, nil
}

// replaceInPattern replaces a target pattern in the original string with a replacement string. All
// All replacement patterns should be enclosed in curly braces {}.
// For example, if the original string is "log-{dd}-{MM}-{yyyy}.log", the target is "dd", and the replacement is "15", the function will return "log-15-{MM}-{yyyy}.log".
func replaceInPattern(original string, target string, replacement string) string {
	return strings.ReplaceAll(original, fmt.Sprintf("{%s}", target), replacement)
}

// replaceDateInPattern replaces date-related patterns in the original string with the current date values.
// It replaces "MM" with the current month, "dd" with the current day, and "yyyy" with the current year.
func replaceDateInPattern(original string) string {
	updatedFilePattern := original
	updatedFilePattern = replaceInPattern(updatedFilePattern, "MM", time.Now().Format("01"))
	updatedFilePattern = replaceInPattern(updatedFilePattern, "dd", time.Now().Format("02"))
	updatedFilePattern = replaceInPattern(updatedFilePattern, "yyyy", time.Now().Format("2006"))

	return updatedFilePattern
}

func getAllFilesMatchingPattern(filePath string, filePattern string) ([]string, error) {
	effectiveFilePath := filepath.Join(filePath, filePattern)

	matchingFiles, err := filepath.Glob(effectiveFilePath)
	if err != nil {
		return []string{}, fmt.Errorf("Failed to fetch log files from path: %s. Root cause: %v", effectiveFilePath, err)
	}

	return matchingFiles, nil
}
