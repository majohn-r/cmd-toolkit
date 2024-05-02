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
	logPath string
	// exposed so that unit tests can close the writer!
	logWriter io.WriteCloser
)

// https://github.com/majohn-r/cmd-toolkit/issues/16
func LogPath() string {
	return logPath
}

func initWriter(o output.Bus) io.Writer {
	var tmpFolder string
	var found bool
	if tmpFolder, found = findTemp(o); !found {
		return nil
	}
	var tmp string
	var err error
	if PlainFileExists(tmpFolder) {
		o.WriteCanonicalError("The temporary folder %q exists as a plain file", tmpFolder)
		return nil
	}
	if tmp, err = CreateAppSpecificPath(tmpFolder); err != nil {
		o.WriteCanonicalError("A programming error has occurred: %v", err)
		return nil
	}
	logPath = filepath.Join(tmp, logDirName)
	fileSystem.MkdirAll(logPath, StdDirPermissions)
	cleanup(o, logPath)
	logWriter = cronowriter.MustNew(
		filepath.Join(logPath, logFilePrefix()+"%Y%m%d"+logFileExtension),
		cronowriter.WithSymlink(filepath.Join(logPath, symlinkName)),
		cronowriter.WithInit())
	return logWriter
}

func cleanup(o output.Bus, logPath string) (found, deleted int) {
	if files, ok := ReadDirectory(o, logPath); ok {
		var fileMap map[time.Time]fs.FileInfo = make(map[time.Time]fs.FileInfo)
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
	if err := fileSystem.Remove(logFile); err != nil {
		o.WriteCanonicalError("The log file %q cannot be deleted: %v", logFile, err)
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
	s, err := AppName()
	if err != nil {
		return "_log_."
	}
	return s + "."
}
