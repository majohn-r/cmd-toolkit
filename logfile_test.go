package cmd_toolkit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func Test_initWriter(t *testing.T) {
	originalLogPath := logPath
	originalFileSystem := fileSystem
	originalStateHome := xdg.StateHome
	defer func() {
		logPath = originalLogPath
		fileSystem = originalFileSystem
		xdg.StateHome = originalStateHome
	}()
	fileSystem = afero.NewMemMapFs()
	tests := map[string]struct {
		preTest         func()
		postTest        func()
		applicationName string
		wantNil         bool
		wantLogPath     string
		output.WantedRecording
	}{
		"app name not defined": {
			preTest:         func() {},
			postTest:        func() {},
			applicationName: "",
			wantNil:         true,
			wantLogPath:     "",
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Log initialization is not possible due to a coding error; " +
					"the application name \"\" is not valid.\n",
			},
		},
		"success": {
			preTest: func() {
				xdg.StateHome = "goodLogs"
				_ = fileSystem.Mkdir("goodLogs", StdDirPermissions)
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
			applicationName: "myApp",
			wantNil:         false,
			wantLogPath:     "goodLogs\\myApp\\logs",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			logPath = ""
			o := output.NewRecorder()
			w, p := initWriter(o, tt.applicationName)
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
				Error: "" +
					"The directory \"no such directory\" cannot be read: " +
					"'*fs.PathError: open no such directory: file does not exist'.\n",
				Log: "level='error' directory='no such directory' error='open no such directory: " +
					"file does not exist' msg='cannot read directory'\n",
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
				prefix := logFilePrefix("")
				for k := 0; k < maxLogFiles; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					_ = afero.WriteFile(
						fileSystem,
						filepath.Join("maxLogFiles", fileName),
						[]byte{0, 1, 2},
						StdFilePermissions,
					)
				}
			},
			postTest:  func(_ *testing.T) {},
			path:      "maxLogFiles",
			wantFound: maxLogFiles,
		},
		"lots of files present": {
			preTest: func() {
				_ = fileSystem.Mkdir("manyLogFiles", StdDirPermissions)
				prefix := logFilePrefix("")
				for k := 0; k < maxLogFiles+1; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					_ = afero.WriteFile(
						fileSystem,
						filepath.Join("manyLogFiles", fileName),
						[]byte{0, 1, 2},
						StdFilePermissions,
					)
					time.Sleep(100 * time.Millisecond)
				}
			},
			postTest: func(t *testing.T) {
				fileName := fmt.Sprintf("%s0%s", logFilePrefix(""), logFileExtension)
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
			gotFound, gotDeleted := cleanup(o, tt.path, "")
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
				Error: "" +
					"The log file \"no such file\" cannot be deleted: " +
					"'*fs.PathError: remove no such file: file does not exist'.\n",
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
				name: fmt.Sprintf("%s-dir-%s", logFilePrefix(""), logFileExtension),
				mode: fs.ModeDir,
			},
		},
		"symbolic link": {
			file: fi{
				name: fmt.Sprintf("%s-dir-%s", logFilePrefix(""), logFileExtension),
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
				name: fmt.Sprintf("%sxx%s", logFilePrefix(""), logFileExtension),
				mode: 0,
			},
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotOk := isLogFile(tt.file, ""); gotOk != tt.wantOk {
				t.Errorf("isLogFile() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_logFilePrefix(t *testing.T) {
	tests := map[string]struct {
		applicationName string
		want            string
	}{
		"bad app name": {
			applicationName: "",
			want:            "_log_.",
		},
		"good app name": {
			applicationName: "myApp",
			want:            "myApp.",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := logFilePrefix(tt.applicationName); got != tt.want {
				t.Errorf("logFilePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findLogFilePath(t *testing.T) {
	originalStateHome := xdg.StateHome
	defer func() {
		xdg.StateHome = originalStateHome
	}()
	testDir := "logFile_findLogFilePath"
	if err := Mkdir(testDir); err != nil {
		t.Errorf("Mkdir() error creating %q = %v", err, testDir)
	}
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	badAppName := "badAppName"
	if f, err := os.Create(filepath.Join(testDir, badAppName)); err != nil {
		t.Errorf("os.Create() error creating %q = %v", err, badAppName)
	} else {
		_ = f.Close()
	}
	tests := map[string]struct {
		stateHome       string
		applicationName string
		want            string
		output.WantedRecording
	}{
		"bad state home": {
			stateHome:       "",
			applicationName: "myApp",
			want:            "",
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The state home value \"\" is not a directory, nor can it be created as a directory.\n" +
					"What to do:\n" +
					"The value of XDG_STATE_HOME should be a directory path, typically %HOMEPATH%\\AppData\\Local.\n" +
					"Either it should contain a subdirectory named \"myApp\", " +
					"which in turn contains a subdirectory named \"logs\".\n" +
					"Or, if they do not exist, it must be possible to create those subdirectories.\n",
			},
		},
		"cannot create full dir": {
			stateHome:       testDir,
			applicationName: badAppName,
			want:            "",
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The state home value \"logFile_findLogFilePath\" " +
					"cannot be used to create a directory for log files.\n" +
					"What to do:\n" +
					"The value of XDG_STATE_HOME should be a directory path, typically %HOMEPATH%\\AppData\\Local.\n" +
					"Either it should contain a subdirectory named \"badAppName\", " +
					"which in turn contains a subdirectory named \"logs\".\n" +
					"Or, if they do not exist, it must be possible to create those subdirectories.\n",
			},
		},
		"success": {
			stateHome:       testDir,
			applicationName: "myApp",
			want:            filepath.Join(testDir, "myApp", "logs"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			xdg.StateHome = tt.stateHome
			if got := findLogFilePath(o, tt.applicationName); got != tt.want {
				t.Errorf("findLogFilePath() = %v, want %v", got, tt.want)
			}
			o.Report(t, "findLogFilePath()."+name, tt.WantedRecording)
		})
	}
}
