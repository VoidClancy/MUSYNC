package logger

import (
	"fmt"
	"os"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
)

// Info logs a structured message at INFO level
func Info(msg string, keysAndValues ...any) {
	logStructured("INFO", colorGreen, msg, keysAndValues...)
}

// Infof logs a formatted message at INFO level
func Infof(format string, a ...any) {
	logFormatted("INFO", colorGreen, format, a...)
}

// Error logs a structured message at ERROR level
func Error(msg string, keysAndValues ...any) {
	logStructured("ERROR", colorRed, msg, keysAndValues...)
}

// Errorf logs a formatted message at ERROR level
func Errorf(format string, a ...any) {
	logFormatted("ERROR", colorRed, format, a...)
}

// Warn logs a structured message at WARN level
func Warn(msg string, keysAndValues ...any) {
	logStructured("WARN", colorYellow, msg, keysAndValues...)
}

// Warnf logs a formatted message at WARN level
func Warnf(format string, a ...any) {
	logFormatted("WARN", colorYellow, format, a...)
}

// Debug logs a structured message at DEBUG level
func Debug(msg string, keysAndValues ...any) {
	logStructured("DEBUG", colorGray, msg, keysAndValues...)
}

// Debugf logs a formatted message at DEBUG level
func Debugf(format string, a ...any) {
	logFormatted("DEBUG", colorGray, format, a...)
}

func formatLevel(level string) string {
	switch level {
	case "INFO":
		return "[ INFO ]"
	case "WARN":
		return "[ WARN ]"
	case "ERROR":
		return "[ ERROR ]"
	case "DEBUG":
		return "[ DEBUG ]"
	default:
		return "[" + level + "]"
	}
}

func logStructured(level, color, msg string, keysAndValues ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	var kvStr string
	if len(keysAndValues) > 0 {
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				kvStr += fmt.Sprintf(" %s%s%s=%v", colorCyan, keysAndValues[i], colorReset, keysAndValues[i+1])
			} else {
				kvStr += fmt.Sprintf(" %s%v", colorCyan, keysAndValues[i])
			}
		}
	}

	fmt.Fprintf(os.Stdout, "%s%s%s %s%s%s %s%s\n", colorGray, timestamp, colorReset, color, formatLevel(level), colorReset, msg, kvStr)
}

func logFormatted(level, color, format string, a ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stdout, "%s%s%s %s%s%s %s\n", colorGray, timestamp, colorReset, color, formatLevel(level), colorReset, msg)
}
