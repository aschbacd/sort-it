package logger

import (
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	// Colors
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Loggers
	InfoLogger = log.New(os.Stdout, cyan("INFO: "), 0)
	WarnLogger = log.New(os.Stderr, yellow("WARN: "), 0)
	ErrorLogger = log.New(os.Stderr, red("ERROR: "), 0)
}

func Info(str string) {
	InfoLogger.Println(str)
}

func Warn(str string) {
	WarnLogger.Println(str)
}

func Error(str string) {
	ErrorLogger.Println(str)
}

func Fatal(str string) {
	Error(str)
	os.Exit(1)
}
