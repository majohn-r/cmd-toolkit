package cmd_toolkit

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/majohn-r/output"
	"github.com/utahta/go-cronowriter"
)

const (
	logDirName       = "logs"
	logFileExtension = ".log"
	symlinkName      = "latest" + logFileExtension
	maxLogFiles      = 10
)

var logWriter io.WriteCloser
var tmpEnvironmentVariableNames = []string{"TMP", "TEMP"}

func initWriter(o output.Bus) (w io.Writer, path string) {
	// pre-check!
	if _, err := AppName(); err != nil {
		o.WriteCanonicalError("Log initialization is not possible due to a coding error: %v", err)
		return
	}
	// get the temporary folder values
	tmpFolderMap := findTemp()
	if len(tmpFolderMap) == 0 {
		o.WriteCanonicalError("Log initialization is not possible because neither the TMP nor TEMP environment variables are defined")
		return
	}
	path = findLogFilePath(o, tmpFolderMap)
	if path != "" {
		cleanup(o, path)
		logWriter = cronowriter.MustNew(
			filepath.Join(path, logFilePrefix()+"%Y%m%d"+logFileExtension),
			cronowriter.WithSymlink(filepath.Join(path, symlinkName)),
			cronowriter.WithInit())
		w = logWriter
	}
	return
}

func findLogFilePath(o output.Bus, tmpFolderMap map[string]string) string {
	for _, variableName := range tmpEnvironmentVariableNames {
		if tmpFolder, found := tmpFolderMap[variableName]; found {
			if !DirExists(tmpFolder) {
				o.WriteCanonicalError("The %s environment variable value %q is not a directory", variableName, tmpFolder)
			} else {
				// this is safe because we know the app name has been set
				tmp, _ := CreateAppSpecificPath(tmpFolder)
				path := filepath.Join(tmp, logDirName)
				_ = fileSystem.MkdirAll(path, StdDirPermissions)
				if DirExists(path) {
					return path
				}
				o.WriteCanonicalError("The %s environment variable value %q cannot be used to create a directory for log files", variableName, tmpFolder)
			}
		}
	}
	return ""
}

func cleanup(o output.Bus, logPath string) (found, deleted int) {
	if files, dirRead := ReadDirectory(o, logPath); dirRead {
		var fileMap = make(map[time.Time]fs.FileInfo)
		times := make([]time.Time, 0, len(files))
		for _, file := range files {
			if isLogFile(file) {
				modificationTime := file.ModTime()
				fileMap[modificationTime] = file
				times = append(times, modificationTime)
			}
		}
		found = len(times)
		if found > maxLogFiles && len(times) > 0 {
			sort.Slice(times, func(i, j int) bool {
				return times[i].Before(times[j])
			})
			limit := len(times) - maxLogFiles
			for k := 0; k < limit; k++ {
				entry := fileMap[times[k]]
				if entry != nil {
					logFile := filepath.Join(logPath, entry.Name())
					if deleteLogFile(o, logFile) {
						deleted++
					}
				}
			}
		}
	}
	return
}

func deleteLogFile(o output.Bus, logFile string) bool {
	if fileErr := fileSystem.Remove(logFile); fileErr != nil {
		o.WriteCanonicalError("The log file %q cannot be deleted: %v", logFile, fileErr)
		return false
	}
	return true
}

func findTemp() map[string]string {
	result := map[string]string{}
	for _, variableName := range tmpEnvironmentVariableNames {
		if tmpFolder, found := os.LookupEnv(variableName); found {
			result[variableName] = tmpFolder
		}
	}
	return result
}

func isLogFile(file fs.FileInfo) (ok bool) {
	if file.Mode().IsRegular() {
		fileName := file.Name()
		ok = strings.HasPrefix(fileName, logFilePrefix()) && strings.HasSuffix(fileName, logFileExtension)
	}
	return
}

func logFilePrefix() string {
	prefix, appNameInitErr := AppName()
	if appNameInitErr != nil {
		return "_log_."
	}
	return prefix + "."
}
