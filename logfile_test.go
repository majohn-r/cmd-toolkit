package cmd_toolkit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func Test_initWriter(t *testing.T) {
	originalTmp := NewEnvVarMemento("TMP")
	originalTemp := NewEnvVarMemento("TEMP")
	originalLogPath := logPath
	originalFileSystem := fileSystem
	defer func() {
		originalTmp.Restore()
		originalTemp.Restore()
		logPath = originalLogPath
		fileSystem = originalFileSystem
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
				Error: "Log initialization is not possible due to a coding error; the application name \"\" is not valid.\n",
			},
		},
		"no temp folder defined": {
			preTest: func() {
				_ = os.Unsetenv("TMP")
				_ = os.Unsetenv("TEMP")
			},
			postTest:        func() {},
			applicationName: "myApp",
			wantNil:         true,
			wantLogPath:     "",
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Log initialization is not possible because neither the TMP nor TEMP environment variables are defined.\n" +
					"What to do:\n" +
					"Define at least one of TMP and TEMP, setting the value to a directory path, e.g., '/tmp'.\n" +
					"Either it should contain a subdirectory named \"myApp\", which in turn contains a subdirectory named \"logs\".\n" +
					"Or, if they do not exist, it must be possible to create those subdirectories.\n",
			},
		},
		"bad TMP setting, no TEMP setting": {
			preTest: func() {
				_ = os.Setenv("TMP", "logs2")
				_ = os.Unsetenv("TEMP")
				_ = afero.WriteFile(fileSystem, "logs2", []byte{}, StdFilePermissions)
			},
			postTest:        func() {},
			applicationName: "myApp",
			wantNil:         true,
			wantLogPath:     "",
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The TMP environment variable value \"logs2\" is not a directory, nor can it be created as a directory.\n" +
					"What to do:\n" +
					"The values of TMP and TEMP should be a directory path, e.g., '/tmp'.\n" +
					"Either it should contain a subdirectory named \"myApp\", which in turn contains a subdirectory named \"logs\".\n" +
					"Or, if they do not exist, it must be possible to create those subdirectories.\n",
			},
		},
		"TMP does not exist yet, TEMP not ok": {
			preTest: func() {
				_ = os.Setenv("TMP", "tmp")
				_ = os.Setenv("TEMP", "temp")
				_ = afero.WriteFile(fileSystem, "temp", []byte("temp"), StdFilePermissions)
			},
			postTest: func() {
				if closeErr := logWriter.Close(); closeErr != nil {
					t.Errorf("error closing logWriter: %v", closeErr)
				} else {
					// this is necessary because the logging library creates the
					// directory in the os file system, not in the one our tests
					// use
					if fileErr := afero.NewOsFs().RemoveAll("tmp"); fileErr != nil {
						t.Errorf("Error removing tmp: %v", fileErr)
					}
				}
			},
			applicationName: "myApp",
			wantNil:         false,
			wantLogPath:     "tmp\\myApp\\logs",
		},
		"TMP ok, TEMP not ok": {
			preTest: func() {
				_ = os.Setenv("TMP", "tmp")
				_ = os.Setenv("TEMP", "temp")
				_ = fileSystem.Mkdir("tmp", StdDirPermissions)
				_ = afero.WriteFile(fileSystem, "temp", []byte("temp"), StdFilePermissions)
			},
			postTest: func() {
				if closeErr := logWriter.Close(); closeErr != nil {
					t.Errorf("error closing logWriter: %v", closeErr)
				} else {
					// this is necessary because the logging library creates the
					// directory in the os file system, not in the one our tests
					// use
					if fileErr := afero.NewOsFs().RemoveAll("tmp"); fileErr != nil {
						t.Errorf("Error removing tmp: %v", fileErr)
					}
				}
			},
			applicationName: "myApp",
			wantNil:         false,
			wantLogPath:     "tmp\\myApp\\logs",
		},
		"TEMP ok, TMP not ok": {
			preTest: func() {
				_ = os.Setenv("TMP", "temp")
				_ = os.Setenv("TEMP", "tmp")
				_ = fileSystem.Mkdir("tmp", StdDirPermissions)
				_ = afero.WriteFile(fileSystem, "temp", []byte("temp"), StdFilePermissions)
			},
			postTest: func() {
				if closeErr := logWriter.Close(); closeErr != nil {
					t.Errorf("error closing logWriter: %v", closeErr)
				} else {
					// this is necessary because the logging library creates the
					// directory in the os file system, not in the one our tests
					// use
					if fileErr := afero.NewOsFs().RemoveAll("tmp"); fileErr != nil {
						t.Errorf("Error removing tmp: %v", fileErr)
					}
				}
			},
			applicationName: "myApp",
			wantNil:         false,
			wantLogPath:     "tmp\\myApp\\logs",
			WantedRecording: output.WantedRecording{
				Error: "The TMP environment variable value \"temp\" is not a directory, nor can it be created as a directory.\n",
			},
		},
		"neither TEMP nor TMP ok": {
			preTest: func() {
				_ = os.Setenv("TMP", "temp")
				_ = os.Setenv("TEMP", "temp")
				_ = afero.WriteFile(fileSystem, "temp", []byte("temp"), StdFilePermissions)
			},
			postTest:        func() {},
			applicationName: "myApp",
			wantNil:         true,
			wantLogPath:     "",
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The TMP environment variable value \"temp\" is not a directory, nor can it be created as a directory.\n" +
					"The TEMP environment variable value \"temp\" is not a directory, nor can it be created as a directory.\n" +
					"What to do:\n" +
					"The values of TMP and TEMP should be a directory path, e.g., '/tmp'.\n" +
					"Either it should contain a subdirectory named \"myApp\", which in turn contains a subdirectory named \"logs\".\n" +
					"Or, if they do not exist, it must be possible to create those subdirectories.\n",
			},
		},
		"cannot create TMP/myapp, but can create TEMP/myapp": {
			preTest: func() {
				fileSystem = afero.NewOsFs()
				_ = os.Setenv("TMP", ".\\tmp")
				_ = os.Setenv("TEMP", "temp")
				_ = fileSystem.MkdirAll(filepath.Join("tmp", "myApp"), StdDirPermissions)
				_ = afero.WriteFile(fileSystem, filepath.Join("tmp", "myApp", "logs"), []byte("tmp"), StdFilePermissions)
				_ = fileSystem.Mkdir("temp", StdDirPermissions)
			},
			postTest: func() {
				_ = fileSystem.RemoveAll("tmp")
				_ = fileSystem.RemoveAll("temp")
				if closeErr := logWriter.Close(); closeErr != nil {
					t.Errorf("error closing logWriter: %v", closeErr)
				} else {
					// this is necessary because the logging library creates the
					// directory in the os file system, not in the one our tests
					// use
					if fileErr := fileSystem.RemoveAll("temp"); fileErr != nil {
						t.Errorf("Error removing temp: %v", fileErr)
					}
				}
				fileSystem = afero.NewMemMapFs()
			},
			applicationName: "myApp",
			wantNil:         false,
			wantLogPath:     "temp\\myApp\\logs",
			WantedRecording: output.WantedRecording{
				Error: "The TMP environment variable value \".\\\\tmp\" cannot be used to create a directory for log files.\n",
			},
		},
		"success": {
			preTest: func() {
				_ = os.Setenv("TMP", "goodLogs")
				_ = os.Unsetenv("TEMP")
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
				prefix := logFilePrefix("")
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
				prefix := logFilePrefix("")
				for k := 0; k < maxLogFiles+1; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					_ = afero.WriteFile(fileSystem, filepath.Join("manyLogFiles", fileName), []byte{0, 1, 2}, StdFilePermissions)
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
		want    map[string]string
	}{
		"no temp vars": {
			preTest: func() {
				_ = os.Unsetenv("TMP")
				_ = os.Unsetenv("TEMP")
			},
			want: map[string]string{},
		},
		"TMP, no TEMP": {
			preTest: func() {
				_ = os.Setenv("TMP", "tmp")
				_ = os.Unsetenv("TEMP")
			},
			want: map[string]string{"TMP": "tmp"},
		},
		"TEMP, no TMP": {
			preTest: func() {
				_ = os.Setenv("TEMP", "temp")
				_ = os.Unsetenv("TMP")
			},
			want: map[string]string{"TEMP": "temp"},
		},
		"TMP and TEMP": {
			preTest: func() {
				_ = os.Setenv("TMP", "tmp")
				_ = os.Setenv("TEMP", "temp")
			},
			want: map[string]string{"TMP": "tmp", "TEMP": "temp"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			got := findTemp()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findTemp() got = %v, want %v", got, tt.want)
			}
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
