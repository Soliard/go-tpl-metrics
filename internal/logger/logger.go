package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Component string

const (
	ComponentServer Component = "server"
	ComponentAgent  Component = "agent"
)

type Logger struct {
	file   *os.File
	logger *log.Logger
}

func New(component Component) (*Logger, error) {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", component, timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:   file,
		logger: log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

func (l *Logger) Close() error {
	return l.file.Close()
}

func (l *Logger) log(level string, args ...interface{}) {
	text := fmt.Sprintf("%s: %v\n", level, fmt.Sprint(args...))
	fmt.Print(text)
	l.logger.Print(text)
}

func (l *Logger) Debug(args ...interface{}) {
	l.log("DEBUG", args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.log("INFO", args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.log("ERROR", args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.log("FATAL", args...)
	os.Exit(1)
}
