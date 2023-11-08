package cmd_toolkit

import (
	"io"

	"github.com/majohn-r/output"
	"github.com/sirupsen/logrus"
)

// OldProductionLogger is the production implementation of the output.Logger
// interface
type OldProductionLogger struct{}

// function to get an io.Writer with which to initialize the logger; this makes
// it easy to substitute another function in unit tests
var OldWriterGetter func(o output.Bus) io.Writer = initWriter

const OldDefaultLoggingLevel = output.Info

// OldInitLogging sets up logging at the default log level
func OldInitLogging(o output.Bus) (ok bool) {
	return OldInitLoggingWithLevel(o, OldDefaultLoggingLevel)
}

var OldLogrusLogLevelMap map[output.Level]logrus.Level = map[output.Level]logrus.Level{
	output.Panic:   logrus.PanicLevel,
	output.Fatal:   logrus.FatalLevel,
	output.Error:   logrus.ErrorLevel,
	output.Warning: logrus.WarnLevel,
	output.Info:    logrus.InfoLevel,
	output.Debug:   logrus.DebugLevel,
	output.Trace:   logrus.TraceLevel,
}

// OldInitLoggingWithLevel initializes logging with a specific log level
func OldInitLoggingWithLevel(o output.Bus, l output.Level) (ok bool) {
	if w := OldWriterGetter(o); w != nil {
		logrus.SetOutput(w)
		logrus.SetLevel(OldLogrusLogLevelMap[l])
		ok = true
	}
	return
}

// Debug outputs a debug log message
func (pl OldProductionLogger) Debug(msg string, fields map[string]any) {
	logrus.WithFields(fields).Debug(msg)
}

// Error outputs an error log message
func (pl OldProductionLogger) Error(msg string, fields map[string]any) {
	logrus.WithFields(fields).Error(msg)
}

// Fatal outputs a fatal log message and terminates the program
func (pl OldProductionLogger) Fatal(msg string, fields map[string]any) {
	logrus.WithFields(fields).Fatal(msg)
}

// Info outputs an info log message
func (pl OldProductionLogger) Info(msg string, fields map[string]any) {
	logrus.WithFields(fields).Info(msg)
}

// Panic outputs a panic log message and calls panic()
func (pl OldProductionLogger) Panic(msg string, fields map[string]any) {
	logrus.WithFields(fields).Panic(msg)
}

// Trace outputs a trace log message
func (pl OldProductionLogger) Trace(msg string, fields map[string]any) {
	logrus.WithFields(fields).Trace(msg)
}

// Warning outputs a warning log message
func (pl OldProductionLogger) Warning(msg string, fields map[string]any) {
	logrus.WithFields(fields).Warning(msg)
}

func (pl OldProductionLogger) ExitFunc() (f func(int)) {
	return logrus.StandardLogger().ExitFunc
}

func (pl OldProductionLogger) SetExitFunc(f func(int)) {
	logrus.StandardLogger().ExitFunc = f
}
