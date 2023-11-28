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

// exposed so that unit tests can close the writer!
var logWriter io.WriteCloser

func initWriter(o output.Bus) io.Writer {
	var tmpFolder string
	var found bool
	if tmpFolder, found = findTemp(o); !found {
		return nil
	}
	var tmp string
	var err error
	if tmp, err = CreateAppSpecificPath(tmpFolder); err != nil {
		o.WriteCanonicalError("A programming error has occurred: %v", err)
		return nil
	}
	logPath := filepath.Join(tmp, logDirName)
	if err = os.MkdirAll(logPath, StdDirPermissions); err != nil {
		WriteDirectoryCreationError(o, logPath, err)
		return nil
	}
	cleanup(o, logPath)
	logWriter = cronowriter.MustNew(
		filepath.Join(logPath, logFilePrefix()+"%Y%m%d"+logFileExtension),
		cronowriter.WithSymlink(filepath.Join(logPath, symlinkName)),
		cronowriter.WithInit())
	return logWriter
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

func deleteLogFile(o output.Bus, logFile string) (ok bool) {
	if err := os.Remove(logFile); err != nil {
		o.WriteCanonicalError("The log file %q cannot be deleted: %v", logFile, err)
	} else {
		ok = true
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
