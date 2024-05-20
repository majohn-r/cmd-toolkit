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
	originalAppname := appname
	originalLogPath := logPath
	originalFileSystem := fileSystem
	defer func() {
		originalTmp.Restore()
		originalTemp.Restore()
		appname = originalAppname
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
				os.Unsetenv("TMP")
				os.Unsetenv("TEMP")
			},
			postTest:        func() {},
			wantNil:         true,
			WantedRecording: output.WantedRecording{Error: "Neither the TMP nor TEMP environment variables are defined.\n"},
		},
		"uninitialized appname": {
			preTest: func() {
				os.Setenv("TMP", "logs1")
				os.Unsetenv("TEMP")
				appname = ""
			},
			postTest:        func() {},
			wantNil:         true,
			WantedRecording: output.WantedRecording{Error: "A programming error has occurred: app name has not been initialized.\n"},
		},
		"bad TMP setting": {
			preTest: func() {
				os.Setenv("TMP", "logs2")
				os.Unsetenv("TEMP")
				appname = "myApp"
				afero.WriteFile(fileSystem, "logs2", []byte{}, StdFilePermissions)
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
				os.Setenv("TMP", "goodLogs")
				os.Unsetenv("TEMP")
				appname = "myApp"
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
			if gotNil := initWriter(o) == nil; !gotNil == tt.wantNil {
				t.Errorf("initWriter() gotNil= %t, wantNil %t", gotNil, tt.wantNil)
			}
			if got := LogPath(); got != tt.wantLogPath {
				t.Errorf("initWriter() got logPath=%q, want %q", got, tt.wantLogPath)
			}
			if issues, verified := o.Verify(tt.WantedRecording); !verified {
				for _, issue := range issues {
					t.Errorf("initWriter() %s", issue)
				}
			}
		})
	}
}

func Test_cleanup(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	type args struct {
		path string
	}
	tests := map[string]struct {
		preTest  func()
		postTest func(t *testing.T)
		args
		wantFound   int
		wantDeleted int
		output.WantedRecording
	}{
		"non-existent directory": {
			preTest:  func() {},
			postTest: func(t *testing.T) {},
			args:     args{path: "no such directory"},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"no such directory\" cannot be read: open no such directory: file does not exist.\n",
				Log:   "level='error' directory='no such directory' error='open no such directory: file does not exist' msg='cannot read directory'\n",
			},
		},
		"empty directory": {
			preTest: func() {
				fileSystem.Mkdir("empty", StdDirPermissions)
			},
			postTest: func(_ *testing.T) {},
			args:     args{path: "empty"},
		},
		"maxLogFiles present": {
			preTest: func() {
				fileSystem.Mkdir("maxLogFiles", StdDirPermissions)
				prefix := logFilePrefix()
				for k := 0; k < maxLogFiles; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					afero.WriteFile(fileSystem, filepath.Join("maxLogFiles", fileName), []byte{0, 1, 2}, StdFilePermissions)
				}
			},
			postTest:  func(_ *testing.T) {},
			args:      args{path: "maxLogFiles"},
			wantFound: maxLogFiles,
		},
		"lots of files present": {
			preTest: func() {
				fileSystem.Mkdir("manyLogFiles", StdDirPermissions)
				prefix := logFilePrefix()
				for k := 0; k < maxLogFiles+1; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					afero.WriteFile(fileSystem, filepath.Join("manyLogFiles", fileName), []byte{0, 1, 2}, StdFilePermissions)
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
			args:        args{path: "manyLogFiles"},
			wantFound:   maxLogFiles + 1,
			wantDeleted: 1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest(t)
			o := output.NewRecorder()
			gotFound, gotDeleted := cleanup(o, tt.args.path)
			if gotFound != tt.wantFound {
				t.Errorf("cleanup() found %d want %d", gotFound, tt.wantFound)
			}
			if gotDeleted != tt.wantDeleted {
				t.Errorf("cleanup() deleted %d want %d", gotDeleted, tt.wantDeleted)
			}
			if issues, verified := o.Verify(tt.WantedRecording); !verified {
				for _, issue := range issues {
					t.Errorf("cleanup() %s", issue)
				}
			}
		})
	}
}

func Test_deleteLogFile(t *testing.T) {
	originalFileSystem := fileSystem
	defer func() {
		fileSystem = originalFileSystem
	}()
	fileSystem = afero.NewMemMapFs()
	type args struct {
		logFile string
	}
	tests := map[string]struct {
		preTest func()
		args
		output.WantedRecording
	}{
		"failure": {
			preTest: func() {},
			args:    args{logFile: "no such file"},
			WantedRecording: output.WantedRecording{
				Error: "The log file \"no such file\" cannot be deleted: remove no such file: file does not exist.\n",
			},
		},
		"success": {
			preTest: func() {
				fileSystem.Mkdir("logs", StdDirPermissions)
				afero.WriteFile(fileSystem, filepath.Join("logs", "file.log"), []byte{}, StdFilePermissions)
			},
			args: args{logFile: filepath.Join("logs", "file.log")},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			o := output.NewRecorder()
			deleteLogFile(o, tt.args.logFile)
			if issues, verified := o.Verify(tt.WantedRecording); !verified {
				for _, issue := range issues {
					t.Errorf("deleteLogFile() %s", issue)
				}
			}
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
				os.Unsetenv("TMP")
				os.Unsetenv("TEMP")
			},
			WantedRecording: output.WantedRecording{Error: "Neither the TMP nor TEMP environment variables are defined.\n"},
		},
		"TMP, no TEMP": {
			preTest: func() {
				os.Setenv("TMP", "tmp")
				os.Unsetenv("TEMP")
			},
			want:  "tmp",
			want1: true,
		},
		"TEMP, no TMP": {
			preTest: func() {
				os.Setenv("TEMP", "temp")
				os.Unsetenv("TMP")
			},
			want:  "temp",
			want1: true,
		},
		"TMP and TEMP": {
			preTest: func() {
				os.Setenv("TMP", "tmp")
				os.Setenv("TEMP", "temp")
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
			if issues, verified := o.Verify(tt.WantedRecording); !verified {
				for _, issue := range issues {
					t.Errorf("findTemp() %s", issue)
				}
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
	type args struct {
		file fs.FileInfo
	}
	tests := map[string]struct {
		args
		wantOk bool
	}{
		"directory": {
			args: args{file: fi{name: fmt.Sprintf("%s-dir-%s", logFilePrefix(), logFileExtension), mode: fs.ModeDir}},
		},
		"symbolic link": {
			args: args{file: fi{name: fmt.Sprintf("%s-dir-%s", logFilePrefix(), logFileExtension), mode: fs.ModeSymlink}},
		},
		"badly named file": {
			args: args{file: fi{name: "foo", mode: 0}},
		},
		"well named file": {
			args:   args{file: fi{name: fmt.Sprintf("%sxx%s", logFilePrefix(), logFileExtension), mode: 0}},
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotOk := isLogFile(tt.args.file); gotOk != tt.wantOk {
				t.Errorf("isLogFile() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_logFilePrefix(t *testing.T) {
	savedAppname := appname
	defer func() {
		appname = savedAppname
	}()
	tests := map[string]struct {
		preTest func()
		want    string
	}{
		"bad app name": {
			preTest: func() {
				appname = ""
			},
			want: "_log_.",
		},
		"good app name": {
			preTest: func() {
				appname = "myApp"
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
