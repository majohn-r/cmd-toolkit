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

func initWriter(o output.Bus) (w io.Writer, path string) {
	var tmpFolder string
	var found bool
	if tmpFolder, found = findTemp(o); !found {
		return nil, ""
	}
	if PlainFileExists(tmpFolder) {
		o.WriteCanonicalError("The temporary folder %q exists as a plain file", tmpFolder)
		return nil, ""
	}
	tmp, creationError := CreateAppSpecificPath(tmpFolder)
	if creationError != nil {
		o.WriteCanonicalError("A programming error has occurred: %v", creationError)
		return nil, ""
	}
	path = filepath.Join(tmp, logDirName)
	_ = fileSystem.MkdirAll(path, StdDirPermissions)
	cleanup(o, path)
	logWriter = cronowriter.MustNew(
		filepath.Join(path, logFilePrefix()+"%Y%m%d"+logFileExtension),
		cronowriter.WithSymlink(filepath.Join(path, symlinkName)),
		cronowriter.WithInit())
	return logWriter, path
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

func findTemp(o output.Bus) (string, bool) {
	for _, v := range []string{"TMP", "TEMP"} {
		if tmpFolder, found := os.LookupEnv(v); found {
			return tmpFolder, found
		}
	}
	o.WriteCanonicalError("Neither the TMP nor TEMP environment variables are defined")
	return "", false
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
