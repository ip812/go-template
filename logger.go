package main

import (
	"fmt"
	"log"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

func NewLogger() Logger {
	if Profile == "local" {
		return &LoggerLocal{}
	}
	return &LoggerProd{}
}

// Logger implementation for local development
type LoggerLocal struct{}

func (l *LoggerLocal) Debug(msg string, args ...interface{}) {
	log.Println("DEBUG: " + fmt.Sprintf(msg, args...))
}

func (l *LoggerLocal) Info(msg string, args ...interface{}) {
	log.Println("INFO: " + fmt.Sprintf(msg, args...))
}

func (l *LoggerLocal) Warn(msg string, args ...interface{}) {
	log.Println("WARN: " + fmt.Sprintf(msg, args...))
}

func (l *LoggerLocal) Error(msg string, args ...interface{}) {
	log.Println("ERROR: " + fmt.Sprintf(msg, args...))
}

// Logger implementation for prod (send logs to AWS Cloudwatch)
type LoggerProd struct{}

func (l *LoggerProd) Debug(msg string, args ...interface{}) {
	log.Println("DEBUG: " + fmt.Sprintf(msg, args...))
}

func (l *LoggerProd) Info(msg string, args ...interface{}) {
	log.Println("INFO: " + fmt.Sprintf(msg, args...))
}

func (l *LoggerProd) Warn(msg string, args ...interface{}) {
	log.Println("WARN " + fmt.Sprintf(msg, args...))
}

func (l *LoggerProd) Error(msg string, args ...interface{}) {
	log.Println("ERROR: " + fmt.Sprintf(msg, args...))
}
