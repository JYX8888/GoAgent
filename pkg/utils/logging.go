package utils

import (
	"fmt"
	"log"
	"os"
)

var defaultLogger *Logger

type Logger struct {
	name  string
	level string
}

func init() {
	defaultLogger = NewLogger("hello_agents", "INFO")
}

func NewLogger(name, level string) *Logger {
	return &Logger{name: name, level: level}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level == "DEBUG" {
		fmt.Printf("[DEBUG] "+format+"\n", v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level == "INFO" || l.level == "DEBUG" {
		fmt.Printf("[INFO] "+format+"\n", v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	fmt.Printf("[FATAL] "+format+"\n", v...)
	os.Exit(1)
}

func SetupLogger(name, level, format string) *Logger {
	logger := NewLogger(name, level)
	return logger
}

func GetLogger(name string) *Logger {
	if name == "" {
		name = "hello_agents"
	}
	return defaultLogger
}

func DefaultLogger() *Logger {
	return defaultLogger
}

type SimpleLogger struct {
	logger *log.Logger
	level  string
}

func NewSimpleLogger(prefix string, level string) *SimpleLogger {
	return &SimpleLogger{
		logger: log.New(os.Stdout, prefix, log.LstdFlags),
		level:  level,
	}
}

func (l *SimpleLogger) Debug(v ...interface{}) {
	if l.level == "DEBUG" {
		l.logger.Println(v...)
	}
}

func (l *SimpleLogger) Info(v ...interface{}) {
	l.logger.Println(v...)
}

func (l *SimpleLogger) Warn(v ...interface{}) {
	l.logger.Println(append([]interface{}{"WARN:"}, v...)...)
}

func (l *SimpleLogger) Error(v ...interface{}) {
	l.logger.Println(append([]interface{}{"ERROR:"}, v...)...)
}
