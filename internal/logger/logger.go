package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func InitLogger(componentName string) error {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", componentName, timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}

func LogConfig(component string, config interface{}) {
	text := fmt.Sprintf("[%s] Configuration: %+v\n", component, config)
	fmt.Println(text)
	InfoLogger.Print(text)
}

func LogError(component string, err error) {
	text := fmt.Sprintf("[%s] Error: %v\n", component, err)
	fmt.Println(text)
	ErrorLogger.Print(text)
}

func LogInfo(component string, message string) {
	text := fmt.Sprintf("[%s] %s\n", component, message)
	fmt.Println(text)
	InfoLogger.Print(text)
}
