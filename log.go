package cmd_toolkit

import (
	"github.com/majohn-r/output"
	"github.com/sirupsen/logrus"
)

// ProductionLogger is the production implementation of the output.Logger
// interface
type ProductionLogger struct{}

// InitLogging sets up logging
func InitLogging(o output.Bus) (ok bool) {
	if w := initWriter(o); w != nil {
		logrus.SetOutput(w)
		ok = true
	}
	return
}

// Debug outputs a debug log message
func (pl ProductionLogger) Debug(msg string, fields map[string]any) {
	logrus.WithFields(fields).Debug(msg)
}

// Error outputs an error log message
func (pl ProductionLogger) Error(msg string, fields map[string]any) {
	logrus.WithFields(fields).Error(msg)
}

// Fatal outputs a fatal log message and terminates the program
func (pl ProductionLogger) Fatal(msg string, fields map[string]any) {
	logrus.WithFields(fields).Fatal(msg)
}

// Info outputs an info log message
func (pl ProductionLogger) Info(msg string, fields map[string]any) {
	logrus.WithFields(fields).Info(msg)
}

// Panic outputs a panic log message and calls panic()
func (pl ProductionLogger) Panic(msg string, fields map[string]any) {
	logrus.WithFields(fields).Panic(msg)
}

// Trace outputs a trace log message
func (pl ProductionLogger) Trace(msg string, fields map[string]any) {
	logrus.WithFields(fields).Trace(msg)
}

// Warning outputs a warning log message
func (pl ProductionLogger) Warning(msg string, fields map[string]any) {
	logrus.WithFields(fields).Warning(msg)
}
