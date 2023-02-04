package cmd_toolkit

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/majohn-r/output"
	"github.com/sirupsen/logrus"
	"github.com/utahta/go-cronowriter"
)

const (
	logDirName       = "logs"
	logFileExtension = ".log"
	symlinkName      = "latest" + logFileExtension
	maxLogFiles      = 10
)

// exposed so that unit tests can close the writer!
var logger *cronowriter.CronoWriter

// ProductionLogger is the production implementation of the output.Logger
// interface
type ProductionLogger struct{}

// InitLogging sets up logging
func InitLogging(o output.Bus) (ok bool) {
	if tmpFolder, found := findTemp(o); found {
		if tmp, err := CreateAppSpecificPath(tmpFolder); err != nil {
			o.WriteCanonicalError("A programming error has occurred: %v", err)
		} else {
			logPath := filepath.Join(tmp, logDirName)
			if err := os.MkdirAll(logPath, StdDirPermissions); err != nil {
				WriteDirectoryCreationError(o, logPath, err)
			} else {
				logger = cronowriter.MustNew(
					filepath.Join(logPath, logFilePrefix()+"%Y%m%d"+logFileExtension),
					cronowriter.WithSymlink(filepath.Join(logPath, symlinkName)),
					cronowriter.WithInit())
				logrus.SetOutput(logger)
				cleanup(o, logPath)
				ok = true
			}
		}
	}
	return
}

func cleanup(o output.Bus, logPath string) (found, deleted int) {
	if files, ok := ReadDirectory(o, logPath); ok {
		var fileMap map[time.Time]fs.DirEntry = make(map[time.Time]fs.DirEntry)
		var times []time.Time
		for _, file := range files {
			if isLogFile(file) {
				if f, fErr := file.Info(); fErr == nil {
					modificationTime := f.ModTime()
					fileMap[modificationTime] = file
					times = append(times, modificationTime)
				}
			}
		}
		found = len(times)
		if found > maxLogFiles {
			sort.Slice(times, func(i, j int) bool {
				return times[i].Before(times[j])
			})
			limit := len(times) - maxLogFiles
			for k := 0; k < limit; k++ {
				logFile := filepath.Join(logPath, fileMap[times[k]].Name())
				if deleteLogFile(o, logFile) {
					deleted++
				}
			}
		}
	}
	return
}

func deleteLogFile(o output.Bus, logFile string) (ok bool) {
	if err := os.Remove(logFile); err != nil {
		LogFileDeletionFailure(o, logFile, err)
		o.WriteCanonicalError("The log file %q cannot be deleted: %v", logFile, err)
	} else {
		ok = true
		o.Log(output.Info, "successfully deleted log file", map[string]any{
			"fileName": logFile,
		})
	}
	return
}

func findTemp(o output.Bus) (string, bool) {
	for _, v := range []string{"TMP", "TEMP"} {
		if tmpFolder, found := os.LookupEnv(v); found {
			return tmpFolder, found
		}
	}
	o.WriteCanonicalError("Neither the TMP nor TEMP environment variables are defined")
	return "", false
}

func isLogFile(file fs.DirEntry) (ok bool) {
	if file.Type().IsRegular() {
		fileName := file.Name()
		ok = strings.HasPrefix(fileName, logFilePrefix()) && strings.HasSuffix(fileName, logFileExtension)
	}
	return
}

func logFilePrefix() string {
	if s, err := AppName(); err != nil {
		return "_log_."
	} else {
		return s + "."
	}
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
