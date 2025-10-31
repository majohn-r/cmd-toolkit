package cmd_toolkit

import (
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/xdg"
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
	logWriter io.WriteCloser
)

func initWriter(o output.Bus, applicationName string) (w io.Writer, path string) {
	// pre-check!
	if !isLegalApplicationName(applicationName) {
		o.ErrorPrintf(
			"Log initialization is not possible due to a coding error; the application name %q is not valid.\n",
			applicationName,
		)
		return
	}
	path = findLogFilePath(o, applicationName)
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

func findLogFilePath(o output.Bus, applicationName string) string {
	base := xdg.StateHome
	if err := Mkdir(base); err != nil {
		o.ErrorPrintf("The state home value %q is not a directory, nor can it be created as a directory.\n", base)
	} else {
		// this is safe because we know the application name has been validated
		appSpecificPath, _ := createAppSpecificPath(base, applicationName)
		path := filepath.Join(appSpecificPath, logDirName)
		_ = fileSystem.MkdirAll(path, StdDirPermissions)
		if DirExists(path) {
			return path
		}
		o.ErrorPrintf("The state home value %q cannot be used to create a directory for log files.\n", base)
	}
	o.ErrorPrintln("What to do:")
	o.ErrorPrintln("The value of XDG_STATE_HOME should be a directory path, typically %HOMEPATH%\\AppData\\Local.")
	o.ErrorPrintf(
		"Either it should contain a subdirectory named %q, which in turn contains a subdirectory named %q.\n",
		applicationName,
		logDirName,
	)
	o.ErrorPrintln("Or, if they do not exist, it must be possible to create those subdirectories.")
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
		o.ErrorPrintf("The log file %q cannot be deleted: %s.\n", logFile, ErrorToString(fileErr))
		return false
	}
	return true
}

func isLogFile(file fs.FileInfo, applicationName string) (ok bool) {
	if file.Mode().IsRegular() {
		fileName := file.Name()
		ok = strings.HasPrefix(
			fileName,
			logFilePrefix(applicationName),
		) && strings.HasSuffix(
			fileName,
			logFileExtension,
		)
	}
	return
}

func logFilePrefix(applicationName string) string {
	if isLegalApplicationName(applicationName) {
		return applicationName + "."
	}
	return "_log_."
}
