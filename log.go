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

// ProductionLogger is the production implementation of the output.Logger
// interface
type simpleLogger struct {
	writer          io.Writer
	exitFunction    exitFunc
	currentLogLevel output.Level
	lock            *sync.RWMutex
}

const defaultLoggingLevel = output.Info

var ProductionLogger *simpleLogger = &simpleLogger{
	exitFunction:    os.Exit,
	currentLogLevel: defaultLoggingLevel,
	lock:            &sync.RWMutex{},
}

// function to get an io.Writer with which to initialize the logger; this makes
// it easy to substitute another function in unit tests
var writerGetter func(o output.Bus) io.Writer = initWriter

// InitLogging sets up logging at the default log level
func InitLogging(o output.Bus) (ok bool) {
	return InitLoggingWithLevel(o, defaultLoggingLevel)
}

// InitLoggingWithLevel initializes logging with a specific log level
func InitLoggingWithLevel(o output.Bus, l output.Level) (ok bool) {
	if w := writerGetter(o); w != nil {
		ProductionLogger.writer = w
		ProductionLogger.currentLogLevel = l
		ok = true
	}
	return
}

func (sl *simpleLogger) willLog(l output.Level) bool {
	return l <= sl.currentLogLevel
}

var typicalChars = regexp.MustCompile(`^[a-zA-Z0-9._/@\^+-]+$`)

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
	if !sl.willLog(l) {
		return
	}
	var fieldMap map[string]string = map[string]string{}
	var fieldKeys []string
	if len(fields) > 0 {
		for k, v := range fields {
			fieldKeys = append(fieldKeys, k)
			fieldMap[k] = toString(v)
		}
		sort.Strings(fieldKeys)
	}
	levelValue := levelsToString[l]
	msgValue := toString(msg)
	sl.lock.Lock()
	defer sl.lock.Unlock()
	tValue := time.Now().Format(time.RFC3339)
	loggedFields := []string{fmt.Sprintf("time=%q", tValue), fmt.Sprintf("level=%s", levelValue), fmt.Sprintf("msg=%s", msgValue)}
	for _, k := range fieldKeys {
		loggedFields = append(loggedFields, fmt.Sprintf("%s=%s", k, fieldMap[k]))
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

func (sl *simpleLogger) ExitFunc() exitFunc {
	return sl.exitFunction
}
func (sl *simpleLogger) SetExitFunc(f exitFunc) {
	sl.exitFunction = f
}
