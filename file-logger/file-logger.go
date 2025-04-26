package logger

import (
	"log"
	"os"
)

var (
	// Logger is the default logger
	lgr      *log.Logger
	colorMap = map[string]string{
		"INFO":  "\033[32m", // Green
		"ERROR": "\033[31m", // Red
		"DEBUG": "\033[34m", // Blue
		"WARN":  "\033[33m", // Yellow
		"FATAL": "\033[35m", // Magenta
	}
	// Reset color
	resetColor = "\033[0m" // Reset color
)

func Setup(path string) {
	// Set up the logger to write to a file
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Create a new logger
	lgr = log.New(file, "", log.LstdFlags)

	// Set the logger as the default logger
	log.SetOutput(lgr.Writer())
}

// slf4j style logging
func logg(level string, msg ...any) {
	color := colorMap[level]
	if color == "" {
		color = resetColor
	}

	// Format the log message
	lgr.Printf("%s [%s] %s%v\n", color, level, resetColor, msg)
}

func Info(msg ...any) {
	logg("INFO", msg...)
}

func Error(msg ...any) {
	logg("ERROR", msg...)
}

func Debug(msg ...any) {
	logg("DEBUG", msg...)
}

func Warn(msg ...any) {
	logg("WARN", msg...)
}

func Fatal(msg ...any) {
	logg("FATAL", msg...)
	os.Exit(1)
}
