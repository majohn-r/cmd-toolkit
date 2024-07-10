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

var (
	logWriter                   io.WriteCloser
	tmpEnvironmentVariableNames = []string{"TMP", "TEMP"}
)

func initWriter(o output.Bus, applicationName string) (w io.Writer, path string) {
	// pre-check!
	if !isLegalApplicationName(applicationName) {
		o.WriteCanonicalError("Log initialization is not possible due to a coding error; the application name %q is not valid", applicationName)
		return
	}
	// get the temporary folder values
	tmpFolderMap := findTemp()
	if len(tmpFolderMap) == 0 {
		o.WriteCanonicalError("Log initialization is not possible because neither the TMP nor TEMP environment variables are defined")
		o.WriteCanonicalError("What to do:\nDefine at least one of TMP and TEMP, setting the value to a directory path, e.g., '/tmp'")
		o.WriteCanonicalError("Either it should contain a subdirectory named %q, which in turn contains a subdirectory named %q", applicationName, logDirName)
		o.WriteCanonicalError("Or, if they do not exist, it must be possible to create those subdirectories")
		return
	}
	path = findLogFilePath(o, tmpFolderMap, applicationName)
	if path != "" {
		cleanup(o, path, applicationName)
		logWriter = cronowriter.MustNew(
			filepath.Join(path, logFilePrefix(applicationName)+"%Y%m%d"+logFileExtension),
			cronowriter.WithSymlink(filepath.Join(path, symlinkName)),
			cronowriter.WithInit())
		w = logWriter
	}
	return
}

func findLogFilePath(o output.Bus, tmpFolderMap map[string]string, applicationName string) string {
	for _, variableName := range tmpEnvironmentVariableNames {
		if tmpFolder, found := tmpFolderMap[variableName]; found {
			if err := Mkdir(tmpFolder); err != nil {
				o.WriteCanonicalError("The %s environment variable value %q is not a directory, nor can it be created as a directory", variableName, tmpFolder)
			} else {
				// this is safe because we know the application name has been validated
				tmp, _ := createAppSpecificPath(tmpFolder, applicationName)
				path := filepath.Join(tmp, logDirName)
				_ = fileSystem.MkdirAll(path, StdDirPermissions)
				if DirExists(path) {
					return path
				}
				o.WriteCanonicalError("The %s environment variable value %q cannot be used to create a directory for log files", variableName, tmpFolder)
			}
		}
	}
	o.WriteCanonicalError("What to do:\nThe values of TMP and TEMP should be a directory path, e.g., '/tmp'")
	o.WriteCanonicalError("Either it should contain a subdirectory named %q, which in turn contains a subdirectory named %q", applicationName, logDirName)
	o.WriteCanonicalError("Or, if they do not exist, it must be possible to create those subdirectories")
	return ""
}

func cleanup(o output.Bus, logPath, applicationName string) (found, deleted int) {
	if files, dirRead := ReadDirectory(o, logPath); dirRead {
		var fileMap = make(map[time.Time]fs.FileInfo)
		times := make([]time.Time, 0, len(files))
		for _, file := range files {
			if isLogFile(file, applicationName) {
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

func isLogFile(file fs.FileInfo, applicationName string) (ok bool) {
	if file.Mode().IsRegular() {
		fileName := file.Name()
		ok = strings.HasPrefix(fileName, logFilePrefix(applicationName)) && strings.HasSuffix(fileName, logFileExtension)
	}
	return
}

func logFilePrefix(applicationName string) string {
	if isLegalApplicationName(applicationName) {
		return applicationName + "."
	}
	return "_log_."
}
