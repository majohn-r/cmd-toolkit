package cmd_toolkit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/majohn-r/output"
)

func Test_initWriter(t *testing.T) {
	savedTmp := NewEnvVarMemento("TMP")
	savedTemp := NewEnvVarMemento("TEMP")
	savedAppname := appname
	defer func() {
		savedTmp.Restore()
		savedTemp.Restore()
		appname = savedAppname
	}()
	tests := map[string]struct {
		preTest  func()
		postTest func()
		wantNil  bool
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
				os.Setenv("TMP", "logs")
				os.Unsetenv("TEMP")
				appname = ""
			},
			postTest:        func() {},
			wantNil:         true,
			WantedRecording: output.WantedRecording{Error: "A programming error has occurred: app name has not been initialized.\n"},
		},
		"bad TMP setting": {
			preTest: func() {
				os.Setenv("TMP", "logs")
				os.Unsetenv("TEMP")
				appname = "myApp"
				_ = os.WriteFile("logs", []byte{}, StdFilePermissions)
			},
			postTest: func() {
				os.Remove("logs")
			},
			wantNil: true,
			WantedRecording: output.WantedRecording{
				Error: "The directory \"logs\\\\myApp\\\\logs\" cannot be created: mkdir logs: The system cannot find the path specified.\n",
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
				if err := logWriter.Close(); err != nil {
					t.Errorf("error closing logWriter: %v", err)
				} else {
					if err := os.RemoveAll("goodLogs"); err != nil {
						t.Errorf("Error removing goodLogs: %v", err)
					}
				}
			},
			wantNil: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			o := output.NewRecorder()
			if gotNil := initWriter(o) == nil; !gotNil == tt.wantNil {
				t.Errorf("initWriter() gotNil= %t, wantNil %t", gotNil, tt.wantNil)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("initWriter() %s", issue)
				}
			}
		})
	}
}

func Test_cleanup(t *testing.T) {
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
				Error: "The directory \"no such directory\" cannot be read: open no such directory: The system cannot find the file specified.\n",
				Log:   "level='error' directory='no such directory' error='open no such directory: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		"empty directory": {
			preTest: func() {
				_ = os.Mkdir("empty", StdDirPermissions)
			},
			postTest: func(_ *testing.T) {
				_ = os.RemoveAll("empty")
			},
			args: args{path: "empty"},
		},
		"maxLogFiles present": {
			preTest: func() {
				_ = os.Mkdir("maxLogFiles", StdDirPermissions)
				prefix := logFilePrefix()
				for k := 0; k < maxLogFiles; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					_ = os.WriteFile(filepath.Join("maxLogFiles", fileName), []byte{0, 1, 2}, StdFilePermissions)
				}
			},
			postTest: func(_ *testing.T) {
				_ = os.RemoveAll("maxLogFiles")
			},
			args:      args{path: "maxLogFiles"},
			wantFound: maxLogFiles,
		},
		"lots of files present": {
			preTest: func() {
				_ = os.Mkdir("manyLogFiles", StdDirPermissions)
				prefix := logFilePrefix()
				for k := 0; k < maxLogFiles+1; k++ {
					fileName := fmt.Sprintf("%s%d%s", prefix, k, logFileExtension)
					_ = os.WriteFile(filepath.Join("manyLogFiles", fileName), []byte{0, 1, 2}, StdFilePermissions)
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
				_ = os.RemoveAll("manyLogFiles")
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("cleanup() %s", issue)
				}
			}
		})
	}
}

func Test_deleteLogFile(t *testing.T) {
	type args struct {
		logFile string
	}
	tests := map[string]struct {
		preTest  func()
		postTest func()
		args
		output.WantedRecording
	}{
		"failure": {
			preTest:  func() {},
			postTest: func() {},
			args:     args{logFile: "no such file"},
			WantedRecording: output.WantedRecording{
				Error: "The log file \"no such file\" cannot be deleted: remove no such file: The system cannot find the file specified.\n",
			},
		},
		"success": {
			preTest: func() {
				_ = os.Mkdir("logs", StdDirPermissions)
				_ = os.WriteFile(filepath.Join("logs", "file.log"), []byte{}, StdFilePermissions)
			},
			postTest: func() {
				_ = os.RemoveAll("logs")
			},
			args: args{logFile: filepath.Join("logs", "file.log")},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest()
			defer tt.postTest()
			o := output.NewRecorder()
			deleteLogFile(o, tt.args.logFile)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("findTemp() %s", issue)
				}
			}
		})
	}
}

type entry struct {
	name string
	mode fs.FileMode
}

func (e entry) Name() string {
	return e.name
}

func (e entry) IsDir() bool {
	return e.mode.IsDir()
}

func (e entry) Type() fs.FileMode {
	return e.mode
}

type fi struct {
}

func (f fi) Name() string {
	return ""
}

func (f fi) Size() int64 {
	return 0
}

func (f fi) Mode() fs.FileMode {
	return 0
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

func (e entry) Info() (fs.FileInfo, error) {
	return fi{}, nil
}

func Test_isLogFile(t *testing.T) {
	type args struct {
		file fs.DirEntry
	}
	tests := map[string]struct {
		args
		wantOk bool
	}{
		"directory": {
			args: args{file: entry{name: fmt.Sprintf("%s-dir-%s", logFilePrefix(), logFileExtension), mode: fs.ModeDir}},
		},
		"symbolic link": {
			args: args{file: entry{name: fmt.Sprintf("%s-dir-%s", logFilePrefix(), logFileExtension), mode: fs.ModeSymlink}},
		},
		"badly named file": {
			args: args{file: entry{name: "foo", mode: 0}},
		},
		"well named file": {
			args:   args{file: entry{name: fmt.Sprintf("%sxx%s", logFilePrefix(), logFileExtension), mode: 0}},
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
