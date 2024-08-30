package cmd_toolkit

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/majohn-r/output"
)

type exitFunc func(int)

type simpleLogger struct {
	writer          io.Writer
	exitFunction    exitFunc
	currentLogLevel output.Level
	lock            *sync.RWMutex
}

const defaultLoggingLevel = output.Info

// ProductionLogger is the production implementation of the output.Logger interface
var (
	ProductionLogger = &simpleLogger{
		exitFunction:    os.Exit,
		currentLogLevel: defaultLoggingLevel,
		lock:            &sync.RWMutex{},
	}
	logPath string
)

// LogWriterInitFn is a function to get an io.Writer with which to initialize the logger;
// this variable setting makes it easy to substitute another function in unit tests
var LogWriterInitFn = initWriter

// LogPath returns the path for log files; see https://github.com/majohn-r/cmd-toolkit/issues/16
func LogPath() string {
	return logPath
}

// InitLogging sets up logging at the default log level
func InitLogging(o output.Bus, applicationName string) (ok bool) {
	return InitLoggingWithLevel(o, defaultLoggingLevel, applicationName)
}

// InitLoggingWithLevel initializes logging with a specific log level
func InitLoggingWithLevel(o output.Bus, l output.Level, applicationName string) (ok bool) {
	if w, p := LogWriterInitFn(o, applicationName); w != nil {
		logPath = p
		ProductionLogger.writer = w
		ProductionLogger.currentLogLevel = l
		ok = true
	}
	return
}

// WillLog returns true if the implementation will log messages at a specified level
func (sl *simpleLogger) WillLog(l output.Level) bool {
	return l <= sl.currentLogLevel
}

var typicalChars = regexp.MustCompile(`^[a-zA-Z0-9._/@^+-]+$`)

func requiresQuotes(s string) bool {
	return !typicalChars.MatchString(s)
}

func toString(v any) string {
	value, ok := v.(string)
	if !ok {
		value = fmt.Sprint(v)
	}
	if requiresQuotes(value) {
		return fmt.Sprintf("%q", value)
	}
	return value
}

var levelsToString = map[output.Level]string{
	output.Debug:   "debug",
	output.Error:   "error",
	output.Fatal:   "fatal",
	output.Info:    "info",
	output.Panic:   "panic",
	output.Warning: "warning",
	output.Trace:   "trace",
}

func (sl *simpleLogger) log(l output.Level, msg string, fields map[string]any) {
	if sl.WillLog(l) {
		sl.doLog(l, time.Now(), msg, fields)
	}
}

func (sl *simpleLogger) doLog(l output.Level, timestamp time.Time, msg string, fields map[string]any) {
	var fieldMap = map[string]string{}
	fieldKeys := make([]string, len(fields))
	if len(fields) > 0 {
		index := 0
		for k, v := range fields {
			fieldKeys[index] = k
			index++
			fieldMap[k] = toString(v)
		}
		sort.Strings(fieldKeys)
	}
	levelValue := levelsToString[l]
	msgValue := toString(msg)
	sl.lock.Lock()
	defer sl.lock.Unlock()
	tValue := timestamp.Format(time.RFC3339)
	loggedFields := make([]string, 3+len(fieldKeys))
	loggedFields[0] = fmt.Sprintf("time=%q", tValue)
	loggedFields[1] = fmt.Sprintf("level=%s", levelValue)
	loggedFields[2] = fmt.Sprintf("msg=%s", msgValue)
	for index, k := range fieldKeys {
		loggedFields[index+3] = fmt.Sprintf("%s=%s", k, fieldMap[k])
	}
	fmt.Fprintln(sl.writer, strings.Join(loggedFields, " "))
}

// Debug outputs a debug log message
func (sl *simpleLogger) Debug(msg string, fields map[string]any) {
	sl.log(output.Debug, msg, fields)
}

// Error outputs an error log message
func (sl *simpleLogger) Error(msg string, fields map[string]any) {
	sl.log(output.Error, msg, fields)
}

// Fatal outputs a fatal log message and terminates the program
func (sl *simpleLogger) Fatal(msg string, fields map[string]any) {
	sl.log(output.Fatal, msg, fields)
	sl.exitFunction(0)
}

// Info outputs an info log message
func (sl *simpleLogger) Info(msg string, fields map[string]any) {
	sl.log(output.Info, msg, fields)
}

// Panic outputs a panic log message and calls panic()
func (sl *simpleLogger) Panic(msg string, fields map[string]any) {
	sl.log(output.Panic, msg, fields)
	panic(msg)
}

// Trace outputs a trace log message
func (sl *simpleLogger) Trace(msg string, fields map[string]any) {
	sl.log(output.Trace, msg, fields)
}

// Warning outputs a warning log message
func (sl *simpleLogger) Warning(msg string, fields map[string]any) {
	sl.log(output.Warning, msg, fields)
}

// ErrorToString converts an error into an easy-to-read string
func ErrorToString(e error) string {
	if e == nil {
		return "'nil error'"
	}
	eType := fmt.Sprintf("%T", e)
	if eType == "*errors.errorString" {
		return fmt.Sprintf("'%v'", e)
	}
	return fmt.Sprintf("'%s: %v'", eType, e)
}
