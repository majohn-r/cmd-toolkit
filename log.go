package cmd_toolkit

import (
	"io"

	"github.com/majohn-r/output"
	"github.com/sirupsen/logrus"
)

// ProductionLogger is the production implementation of the output.Logger
// interface
type ProductionLogger struct{}

// function to get an io.Writer with which to initialize the logger; this makes
// it easy to substitute another function in unit tests
var writerGetter func(o output.Bus) io.Writer = initWriter

const defaultLoggingLevel = output.Info

// InitLogging sets up logging at the default log level
func InitLogging(o output.Bus) (ok bool) {
	return InitLoggingWithLevel(o, defaultLoggingLevel)
}

var logrusLogLevelMap map[output.Level]logrus.Level = map[output.Level]logrus.Level{
	output.Panic:   logrus.PanicLevel,
	output.Fatal:   logrus.FatalLevel,
	output.Error:   logrus.ErrorLevel,
	output.Warning: logrus.WarnLevel,
	output.Info:    logrus.InfoLevel,
	output.Debug:   logrus.DebugLevel,
	output.Trace:   logrus.TraceLevel,
}

// InitLoggingWithLevel initializes logging with a specific log level
func InitLoggingWithLevel(o output.Bus, l output.Level) (ok bool) {
	if w := writerGetter(o); w != nil {
		logrus.SetOutput(w)
		logrus.SetLevel(logrusLogLevelMap[l])
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

func (pl ProductionLogger) ExitFunc() (f func(int)) {
	return logrus.StandardLogger().ExitFunc
}

func (pl ProductionLogger) SetExitFunc(f func(int)) {
	logrus.StandardLogger().ExitFunc = f
}
