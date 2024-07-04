package cmd_toolkit_test

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"os"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestApplicationPath(t *testing.T) {
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	defer cmdtoolkit.UnsafeSetApplicationPath(originalApplicationPath)
	tests := map[string]struct {
		applicationPath string
		want            string
	}{"dummy": {applicationPath: "foo/bar", want: "foo/bar"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.UnsafeSetApplicationPath(tt.applicationPath)
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
		cmdtoolkit.UnsafeSetApplicationPath(originalApplicationPath)
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
					"Files used by beautifulApp cannot be read or written because the environment variable APPDATA has not been set.\n" +
					"What to do:\n" +
					"Define APPDATA, giving it a value that is a directory path, typically %HOMEPATH%\\AppData\\Roaming.\n",
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
					"The APPDATA environment variable value \"foo.bar\" is not a directory, nor can it be created as a directory.\n" +
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
				Error: "The directory \"myApp1\" cannot be created: file exists and is not a directory.\n",
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
			cmdtoolkit.UnsafeSetApplicationPath("")
			if tt.appDataSet {
				_ = os.Setenv("APPDATA", tt.appDataValue)
			} else {
				_ = os.Unsetenv("APPDATA")
			}
			tt.preTest()
			o := output.NewRecorder()
			if gotInitialized := cmdtoolkit.InitApplicationPath(o, tt.applicationName); gotInitialized != tt.wantInitialized {
				t.Errorf("InitApplicationPath() = %v, want %v", gotInitialized, tt.wantInitialized)
			}
			o.Report(t, "InitApplicationPath()", tt.WantedRecording)
		})
	}
}

func TestSetApplicationPath(t *testing.T) {
	originalApplicationPath := cmdtoolkit.ApplicationPath()
	defer cmdtoolkit.UnsafeSetApplicationPath(originalApplicationPath)
	tests := map[string]struct {
		applicationPath string
		s               string
		wantPrevious    string
	}{"simple": {applicationPath: "foo", s: "bar", wantPrevious: "foo"}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmdtoolkit.UnsafeSetApplicationPath(tt.applicationPath)
			if gotPrevious := cmdtoolkit.SetApplicationPath(tt.s); gotPrevious != tt.wantPrevious {
				t.Errorf("SetApplicationPath() = %v, want %v", gotPrevious, tt.wantPrevious)
			}
			if gotNew := cmdtoolkit.ApplicationPath(); gotNew != tt.s {
				t.Errorf("SetApplicationPath() gotNew = %v, want %v", gotNew, tt.s)
			}
		})
	}
}
