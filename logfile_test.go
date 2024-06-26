package cmd_toolkit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func Test_initWriter(t *testing.T) {
	originalTmp := NewEnvVarMemento("TMP")
	originalTemp := NewEnvVarMemento("TEMP")
	originalAppName := appName
	originalLogPath := logPath
	originalFileSystem := fileSystem
	defer func() {
		originalTmp.Restore()
		originalTemp.Restore()
		appName = originalAppName
		logPath = originalLogPath
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest     func()
		postTest    func()
		wantNil     bool
		wantLogPath string
		output.WantedRecording
	}{
		"no temp folder defined": {
			preTest: func() {
				_ = os.Unsetenv("TMP")
				_ = os.Unsetenv("TEMP")
			},
			postTest:        func() {},
			wantNil:         true,
			WantedRecording: output.WantedRecording{Error: "Neither the TMP nor TEMP environment variables are defined.\n"},
		},
		"uninitialized appName": {
			preTest: func() {
				_ = os.Setenv("TMP", "logs1")
				_ = os.Unsetenv("TEMP")
				appName = ""
			},
			postTest:        func() {},
			wantNil:         true,
			WantedRecording: output.WantedRecording{Error: "A programming error has occurred: app name has not been initialized.\n"},
		},
		"bad TMP setting": {
			preTest: func() {
				_ = os.Setenv("TMP", "logs2")
				_ = os.Unsetenv("TEMP")
				appName = "myApp"
				_ = afero.WriteFile(fileSystem, "logs2", []byte{}, StdFilePermissions)
			},
			postTest: func() {
			},
			wantNil:     true,
			wantLogPath: "",
			WantedRecording: output.WantedRecording{
				Error: "The temporary folder \"logs2\" exists as a plain file.\n",
			},
		},
		"success": {
			preTest: func() {
				_ = os.Setenv("TMP", "goodLogs")
				_ = os.Unsetenv("TEMP")
				appName = "myApp"
			},
			postTest: func() {
				// critical to close logWriter, otherwise, "goodLogs" cannot be
				// removed, as logWriter will continue hold the current log file
				// open
				if closeErr := logWriter.Close(); closeErr != nil {
					t.Errorf("error closing logWriter: %v", closeErr)
				} else {
					// this is necessary because the logging library creates the
					// directory in the os file system, not in the one our tests
					// use
					if fileErr := afero.NewOsFs().RemoveAll("goodLogs"); fileErr != nil {
						t.Errorf("Error removing goodLogs: %v", fileErr)
					}
				}
			},
			wantNil:     false,
			wantLogPath: "goodLogs\\myApp\\logs",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			logPath = ""
			o := output.NewRecorder()
			w, p := initWriter(o)
			if gotNil := w == nil; !gotNil == tt.wantNil {
				t.Errorf("initWriter() gotNil= %t, wantNil %t", gotNil, tt.wantNil)
			}
			if p != tt.wantLogPath {
				t.Errorf("initWriter() got logPath=%q, want %q", p, tt.wantLogPath)
			}
			o.Report(t, "initWriter()", tt.WantedRecording)
		})
	}
}

func Test_cleanup(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest     func()
		postTest    func(t *testing.T)
		path        string
		wantFound   int
		wantDeleted int
		output.WantedRecording
	}{
		"non-existent directory": {
			preTest:  func() {},
			postTest: func(t *testing.T) {},
			path:     "no such directory",
			WantedRecording: output.WantedRecording{
				Error: "The directory \"no such directory\" cannot be read: open no such directory: file does not exist.\n",
				Log:   "level='error' directory='no such directory' error='open no such directory: file does not exist' msg='cannot read directory'\n",
			},
		},
		"empty directory": {
			preTest: func() {
				_ = fileSystem.Mkdir("empty", StdDirPermissions)
			},
			postTest: func(_ *testing.T) {},
			path:     "empty",
		},
		"maxLogFiles present": {
			preTest: func() {
				_ = fileSystem.Mkdir("maxLogFiles", StdDirPermissions)
				prefix := logFilePrefix()
				for k := 0; k < maxLogFiles; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					_ = afero.WriteFile(fileSystem, filepath.Join("maxLogFiles", fileName), []byte{0, 1, 2}, StdFilePermissions)
				}
			},
			postTest:  func(_ *testing.T) {},
			path:      "maxLogFiles",
			wantFound: maxLogFiles,
		},
		"lots of files present": {
			preTest: func() {
				_ = fileSystem.Mkdir("manyLogFiles", StdDirPermissions)
				prefix := logFilePrefix()
				for k := 0; k < maxLogFiles+1; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					_ = afero.WriteFile(fileSystem, filepath.Join("manyLogFiles", fileName), []byte{0, 1, 2}, StdFilePermissions)
					time.Sleep(100 * time.Millisecond)
				}
			},
			postTest: func(t *testing.T) {
				fileName := fmt.Sprintf("%s0%s", logFilePrefix(), logFileExtension)
				if PlainFileExists(filepath.Join("manyLogFiles", fileName)) {
					t.Logf("cleanup() %s should have been deleted", fileName)
					remaining, _ := ReadDirectory(output.NewNilBus(), "manyLogFiles")
					for _, entry := range remaining {
						t.Logf("- %s remains", entry.Name())
					}
					t.Fail()
				}
			},
			path:        "manyLogFiles",
			wantFound:   maxLogFiles + 1,
			wantDeleted: 1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest(t)
			o := output.NewRecorder()
			gotFound, gotDeleted := cleanup(o, tt.path)
			if gotFound != tt.wantFound {
				t.Errorf("cleanup() found %d want %d", gotFound, tt.wantFound)
			}
			if gotDeleted != tt.wantDeleted {
				t.Errorf("cleanup() deleted %d want %d", gotDeleted, tt.wantDeleted)
			}
			o.Report(t, "cleanup()", tt.WantedRecording)
		})
	}
}

func Test_deleteLogFile(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest func()
		logFile string
		output.WantedRecording
	}{
		"failure": {
			preTest: func() {},
			logFile: "no such file",
			WantedRecording: output.WantedRecording{
				Error: "The log file \"no such file\" cannot be deleted: remove no such file: file does not exist.\n",
			},
		},
		"success": {
			preTest: func() {
				_ = fileSystem.Mkdir("logs", StdDirPermissions)
				_ = afero.WriteFile(fileSystem, filepath.Join("logs", "file.log"), []byte{}, StdFilePermissions)
			},
			logFile: filepath.Join("logs", "file.log"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			deleteLogFile(o, tt.logFile)
			o.Report(t, "deleteLogFile()", tt.WantedRecording)
		})
	}
}

func Test_findTemp(t *testing.T) {
	savedTmp := NewEnvVarMemento("TMP")
	savedTemp := NewEnvVarMemento("TEMP")
	defer func() {
		savedTmp.Restore()
		savedTemp.Restore()
	}()
	tests := map[string]struct {
		preTest func()
		want    string
		want1   bool
		output.WantedRecording
	}{
		"no temp vars": {
			preTest: func() {
				_ = os.Unsetenv("TMP")
				_ = os.Unsetenv("TEMP")
			},
			WantedRecording: output.WantedRecording{Error: "Neither the TMP nor TEMP environment variables are defined.\n"},
		},
		"TMP, no TEMP": {
			preTest: func() {
				_ = os.Setenv("TMP", "tmp")
				_ = os.Unsetenv("TEMP")
			},
			want:  "tmp",
			want1: true,
		},
		"TEMP, no TMP": {
			preTest: func() {
				_ = os.Setenv("TEMP", "temp")
				_ = os.Unsetenv("TMP")
			},
			want:  "temp",
			want1: true,
		},
		"TMP and TEMP": {
			preTest: func() {
				_ = os.Setenv("TMP", "tmp")
				_ = os.Setenv("TEMP", "temp")
			},
			want:  "tmp",
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			got, got1 := findTemp(o)
			if got != tt.want {
				t.Errorf("findTemp() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("findTemp() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "findTemp()", tt.WantedRecording)
		})
	}
}

type fi struct {
	name string
	mode fs.FileMode
}

func (f fi) Name() string {
	return f.name
}

func (f fi) Size() int64 {
	return 0
}

func (f fi) Mode() fs.FileMode {
	return f.mode
}

func (f fi) ModTime() time.Time {
	return time.Now()
}

func (f fi) IsDir() bool {
	return false
}

func (f fi) Sys() any {
	return nil
}

func Test_isLogFile(t *testing.T) {
	tests := map[string]struct {
		file   fs.FileInfo
		wantOk bool
	}{
		"directory": {
			file: fi{
				name: fmt.Sprintf("%s-dir-%s", logFilePrefix(), logFileExtension),
				mode: fs.ModeDir,
			},
		},
		"symbolic link": {
			file: fi{
				name: fmt.Sprintf("%s-dir-%s", logFilePrefix(), logFileExtension),
				mode: fs.ModeSymlink,
			},
		},
		"badly named file": {
			file: fi{
				name: "foo",
				mode: 0,
			},
		},
		"well named file": {
			file: fi{
				name: fmt.Sprintf("%sxx%s", logFilePrefix(), logFileExtension),
				mode: 0,
			},
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotOk := isLogFile(tt.file); gotOk != tt.wantOk {
				t.Errorf("isLogFile() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_logFilePrefix(t *testing.T) {
	savedAppName := appName
	defer func() {
		appName = savedAppName
	}()
	tests := map[string]struct {
		preTest func()
		want    string
	}{
		"bad app name": {
			preTest: func() {
				appName = ""
			},
			want: "_log_.",
		},
		"good app name": {
			preTest: func() {
				appName = "myApp"
			},
			want: "myApp.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			if got := logFilePrefix(); got != tt.want {
				t.Errorf("logFilePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
