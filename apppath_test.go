package cmd_toolkit_test

import (
	"os"
	"path/filepath"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestApplicationPath(t *testing.T) {
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	defer cmdtoolkit.SetApplicationPath(originalApplicationPath)
	tests := map[string]struct {
		applicationPath string
		want            string
	}{"dummy": {applicationPath: "foo/bar", want: "foo/bar"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.SetApplicationPath(tt.applicationPath)
			if got := cmdtoolkit.ApplicationPath(); got != tt.want {
				t.Errorf("ApplicationPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitApplicationPath(t *testing.T) {
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	originalFileSystem := cmdtoolkit.FileSystem()
	appDataMemento := cmdtoolkit.NewEnvVarMemento("APPDATA")
	defer func() {
		cmdtoolkit.SetApplicationPath(originalApplicationPath)
		appDataMemento.Restore()
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	tests := map[string]struct {
		applicationName string
		appDataValue    string
		appDataSet      bool
		wantInitialized bool
		preTest         func() // things to do before calling
		output.WantedRecording
	}{
		"no appdata": {
			applicationName: "beautifulApp",
			appDataSet:      false,
			wantInitialized: false,
			preTest:         func() {},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Files used by beautifulApp cannot be read or written because the environment variable APPDATA " +
					"has not been set.\n" +
					"What to do:\n" +
					"Define APPDATA, giving it a value that is a directory path, " +
					"typically %HOMEPATH%\\AppData\\Roaming.\n",
				Log: "level='error' environmentVariable='APPDATA' msg='not set'\n",
			},
		},
		"no app name": {
			applicationName: "",
			appDataSet:      true,
			appDataValue:    "foo", // doesn't matter ...
			wantInitialized: false,
			preTest:         func() {},
			WantedRecording: output.WantedRecording{
				Log: "level='error' error='application name \"\" is not valid' msg='program error'\n",
			},
		},
		"appData not a directory": {
			applicationName: "myApp",
			appDataSet:      true,
			appDataValue:    "foo.bar",
			wantInitialized: false,
			preTest: func() {
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), "foo.bar", []byte("foo"), cmdtoolkit.StdFilePermissions)
			},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The APPDATA environment variable value \"foo.bar\" is not a directory, " +
					"nor can it be created as a directory.\n" +
					"What to do:\n" +
					"The value of APPDATA should be a directory path, typically %HOMEPATH%\\AppData\\Roaming.\n" +
					"Either it should contain a subdirectory named \"myApp\".\n" +
					"Or, if it does not exist, it must be possible to create that subdirectory.\n",
				Log: "" +
					"level='error'" +
					" error='file exists and is not a directory'" +
					" fileName='foo.bar'" +
					" msg='directory check failed'\n",
			},
		},
		"cannot create subdirectory": {
			applicationName: "myApp1",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: false,
			preTest: func() {
				_ = afero.WriteFile(cmdtoolkit.FileSystem(), "myApp1", []byte{1, 2, 3}, cmdtoolkit.StdFilePermissions)
			},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"myApp1\" cannot be created: 'file exists and is not a directory'.\n",
				Log: "" +
					"level='error'" +
					" directory='myApp1'" +
					" error='file exists and is not a directory'" +
					" msg='cannot create directory'\n",
			},
		},
		"subdirectory already exists": {
			applicationName: "myApp2",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest: func() {
				_ = cmdtoolkit.FileSystem().Mkdir("myApp2", cmdtoolkit.StdDirPermissions)
			},
		},
		"subdirectory does not yet exist": {
			applicationName: "myApp3",
			appDataSet:      true,
			appDataValue:    ".",
			wantInitialized: true,
			preTest:         func() {},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.SetApplicationPath("")
			if tt.appDataSet {
				_ = os.Setenv("APPDATA", tt.appDataValue)
			} else {
				_ = os.Unsetenv("APPDATA")
			}
			tt.preTest()
			o := output.NewRecorder()
			if got := cmdtoolkit.InitApplicationPath(o, tt.applicationName); got != tt.wantInitialized {
				t.Errorf("InitApplicationPath() = %v, want %v", got, tt.wantInitialized)
			}
			o.Report(t, "InitApplicationPath()", tt.WantedRecording)
		})
	}
}

func TestSetApplicationPath(t *testing.T) {
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	defer cmdtoolkit.SetApplicationPath(originalApplicationPath)
	tests := map[string]struct {
		applicationPath string
		s               string
		wantPrevious    string
	}{"simple": {applicationPath: "foo", s: "bar", wantPrevious: "foo"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.SetApplicationPath(tt.applicationPath)
			if gotPrevious := cmdtoolkit.SetApplicationPath(tt.s); gotPrevious != tt.wantPrevious {
				t.Errorf("SetApplicationPath() = %v, want %v", gotPrevious, tt.wantPrevious)
			}
			if gotNew := cmdtoolkit.ApplicationPath(); gotNew != tt.s {
				t.Errorf("SetApplicationPath() gotNew = %v, want %v", gotNew, tt.s)
			}
		})
	}
}

func TestAppName(t *testing.T) {
	savedArgs := os.Args
	defer func() { os.Args = savedArgs }()
	ps := string(os.PathSeparator)
	tests := map[string]struct {
		args []string
		want string
	}{
		"nil": {
			args: nil,
			want: "",
		},
		"empty": {
			args: []string{},
			want: "",
		},
		"simple": {
			args: []string{"foo"},
			want: "foo",
		},
		"with extension": {
			args: []string{"foo.bar"},
			want: "foo",
		},
		"silly extension": {
			args: []string{"foo."},
			want: "foo",
		},
		"with path": {
			args: []string{filepath.Join("bar", "foo")},
			want: "foo",
		},
		"with path and extension": {
			args: []string{filepath.Join("bar", "foo.baz")},
			want: "foo",
		},
		"app name starts with '.' and has no extension": {
			args: []string{".bash"},
			want: ".bash",
		},
		"app name starts with lots of '.' and has no extension": {
			args: []string{"....bash"},
			want: "....bash",
		},
		"app name starts with lots of '.' and has an extension": {
			args: []string{"....bash.foo"},
			want: "....bash",
		},
		"path ends in '.'": {
			args: []string{"foo" + ps + "."},
			want: "",
		},
		"path ends in '..'": {
			args: []string{"foo" + ps + ".."},
			want: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = tt.args
			if got := cmdtoolkit.AppName(); got != tt.want {
				t.Errorf("AppName() = %v, want %v", got, tt.want)
			}
		})
	}
}
