package logs

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Define level for severity of log
type Level int8

const (
	// Levels for severity of log
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)


// Return string representation of level
func (l Level) String() string {
	switch l {
		case LevelInfo:
			return "INFO"
		case LevelError:
			return "ERROR"
		case LevelFatal:
			return "FATAL"
		default:
			return ""
	}
}


// Custom logger for JSON format
type Logger struct {
	out io.Writer
	minlevel Level
	mu sync.Mutex
}


// Factory function for creating new logger
func New(out io.Writer, minlevel Level) *Logger {
	return &Logger{
		out: out, 
		minlevel: minlevel,
	}
}


// Print the log message for INFO level
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}


// Print the log message for ERROR level
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}


// Print the log message for FATAL level
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}


// Internal function for printing log message
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error){
	// Check if level is valid
	if level < l.minlevel {
		return 0, nil
	}


	// Creating an auxilary struct for logging in JSON format
	aux := struct {
		Level string `json:"level"`
		Time string `json:"time"`
		Message string `json:"message"`
		Properties map[string]string `json:"properties"`
		Trace string `json:"trace"`
	}{
		Level: level.String(),
		Time: time.Now().UTC().Format(time.RFC3339),
		Message: message,
		Properties: properties,
	}


	// Include stack trace if level is ERROR or FATAL
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}


	// Variable for the log entry text
	var line []byte

	
	// Marshal the auxilary struct to JSON format
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": failed to marshal log message:" + err.Error())
	}


	// Locking the mutex to prevent to prevent redundant logs in case of concurrent calls
	l.mu.Lock()
	defer l.mu.Unlock()


	// Writing the log entry to the output
	return l.out.Write(append(line, '\n'))
}


// Writer Method for writing the log entry to the output
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}